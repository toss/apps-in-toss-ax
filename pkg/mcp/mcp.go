package mcp

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/toss/apps-in-toss-ax/pkg/instrumentation"
	"github.com/toss/apps-in-toss-ax/pkg/search"
)

const (
	name  = "ax"
	title = "ax"

	// Fallback when the caller doesn't wire in the CLI version via WithVersion.
	defaultVersion = "0.0.0-dev"
)

type lazySearcher struct {
	mu     sync.Mutex
	s      *search.Searcher
	initFn func() (*search.Searcher, error)
}

func newLazySearcher(initFn func() (*search.Searcher, error)) *lazySearcher {
	return &lazySearcher{initFn: initFn}
}

func (ls *lazySearcher) get(ctx context.Context) (*search.Searcher, error) {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	if ls.s != nil {
		return ls.s, nil
	}

	s, err := ls.initFn()
	if err != nil {
		return nil, err
	}
	if err := s.EnsureIndex(ctx); err != nil {
		return nil, err
	}
	ls.s = s
	return ls.s, nil
}

type Protocol struct {
	OnInit    func(context.Context)
	Transport mcp.Transport
	Server    *mcp.Server

	completions *CompletionRegistry
	docSearcher *lazySearcher
	tdsRn       *lazySearcher
	tdsWeb      *lazySearcher
	analytics   *instrumentation.Analytics
	sessionID   string
	version     string
}

type Option func(*Protocol)

func WithTransport(transport mcp.Transport) Option {
	return func(s *Protocol) {
		s.Transport = transport
	}
}

func WithAnalytics(analytics *instrumentation.Analytics) Option {
	return func(s *Protocol) {
		s.analytics = analytics
	}
}

func WithVersion(version string) Option {
	return func(s *Protocol) {
		s.version = version
	}
}

func New(options ...Option) *Protocol {
	p := &Protocol{
		Transport:   &mcp.StdioTransport{},
		OnInit:      func(_ context.Context) {},
		completions: NewCompletionRegistry(),
		docSearcher: newLazySearcher(search.New),
		tdsRn:       newLazySearcher(search.NewTDSSearcher),
		tdsWeb:      newLazySearcher(search.NewTDSMobileSearcher),
		sessionID:   newTelemetrySessionID(),
		version:     defaultVersion,
	}

	for _, o := range options {
		o(p)
	}

	i := mcp.NewServer(
		&mcp.Implementation{
			Name:    name,
			Title:   title,
			Version: p.version,
		},
		&mcp.ServerOptions{
			Instructions:      instructions(),
			HasPrompts:        true,
			HasResources:      true,
			HasTools:          true,
			CompletionHandler: p.completions.Handler,
		})
	i.AddReceivingMiddleware(p.analyticsMiddleware())

	i.AddPrompt(miniappActionPlan, miniappActionPlanHandler)
	p.completions.RegisterAll(miniappActionPlanCompletions)

	mcp.AddTool(i, searchDocs, p.searchDocsHandler)
	mcp.AddTool(i, searchTdsRnDocs, p.searchTdsRnDocsHandler)
	mcp.AddTool(i, searchTdsWebDocs, p.searchTdsWebDocsHandler)
	mcp.AddTool(i, getDoc, p.getDocHandler)
	mcp.AddTool(i, getTdsRnDoc, p.getTdsRnDocHandler)
	mcp.AddTool(i, getTdsWebDoc, p.getTdsWebDocHandler)

	p.Server = i
	return p
}

func (p *Protocol) analyticsMiddleware() mcp.Middleware {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			if p.analytics == nil || method != "tools/call" {
				return next(ctx, method, req)
			}

			skillName, query := toolCallInfo(req.GetParams())
			invokedProperties := p.skillEventProperties(req, skillName)
			if query != "" {
				invokedProperties["query"] = query
			}
			p.analytics.Track(ctx, "skill_invoked", invokedProperties)

			result, err := next(ctx, method, req)
			if err == nil {
				renderedProperties := p.skillEventProperties(req, skillName)
				if count, ok := searchResultCount(result); ok {
					renderedProperties["search_result_count"] = count
				}
				p.analytics.Track(ctx, "skill_response_rendered", renderedProperties)
			}
			return result, err
		}
	}
}

func (p *Protocol) skillEventProperties(req mcp.Request, skillName string) map[string]any {
	properties := map[string]any{
		"session_id": p.analyticsSessionID(req),
	}
	if skillName != "" {
		properties["skill_name"] = skillName
	}
	if clientInfo := mcpClientInfo(req); clientInfo != nil {
		if clientInfo.Name != "" {
			properties["mcp_client_name"] = clientInfo.Name
		}
		if clientInfo.Version != "" {
			properties["mcp_client_version"] = clientInfo.Version
		}
	}
	return properties
}

func (p *Protocol) analyticsSessionID(req mcp.Request) string {
	if req != nil {
		if sessionID := sdkSessionID(req.GetSession()); sessionID != "" {
			return sessionID
		}
	}
	return p.sessionID
}

func sdkSessionID(session mcp.Session) (sessionID string) {
	if session == nil {
		return ""
	}
	defer func() {
		if recover() != nil {
			sessionID = ""
		}
	}()
	value := reflect.ValueOf(session)
	if value.Kind() == reflect.Ptr && value.IsNil() {
		return ""
	}
	return session.ID()
}

func mcpClientInfo(req mcp.Request) *mcp.Implementation {
	if req == nil {
		return nil
	}
	session := req.GetSession()
	if session == nil {
		return nil
	}
	value := reflect.ValueOf(session)
	if value.Kind() == reflect.Ptr && value.IsNil() {
		return nil
	}
	serverSession, ok := session.(*mcp.ServerSession)
	if !ok || serverSession == nil {
		return nil
	}
	params := serverSession.InitializeParams()
	if params == nil {
		return nil
	}
	return params.ClientInfo
}

func toolCallInfo(params mcp.Params) (skillName string, query string) {
	switch params := params.(type) {
	case *mcp.CallToolParamsRaw:
		return params.Name, queryFromArguments(params.Arguments)
	case *mcp.CallToolParams:
		return params.Name, queryFromArguments(params.Arguments)
	default:
		return "", ""
	}
}

func queryFromArguments(arguments any) string {
	if arguments == nil {
		return ""
	}

	var payload map[string]any
	switch arguments := arguments.(type) {
	case json.RawMessage:
		if len(arguments) == 0 {
			return ""
		}
		if err := json.Unmarshal(arguments, &payload); err != nil {
			return ""
		}
	case map[string]any:
		payload = arguments
	default:
		data, err := json.Marshal(arguments)
		if err != nil {
			return ""
		}
		if err := json.Unmarshal(data, &payload); err != nil {
			return ""
		}
	}

	query, _ := payload["query"].(string)
	return query
}

func searchResultCount(result mcp.Result) (int, bool) {
	toolResult, ok := result.(*mcp.CallToolResult)
	if !ok || toolResult == nil {
		return 0, false
	}

	return searchResultCountFromStructuredContent(toolResult.StructuredContent)
}

type searchResultCountPayload struct {
	Total   *int              `json:"total"`
	Results []json.RawMessage `json:"results"`
}

func searchResultCountFromStructuredContent(content any) (int, bool) {
	if content == nil {
		return 0, false
	}

	data, err := json.Marshal(content)
	if err != nil {
		return 0, false
	}

	var payload searchResultCountPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return 0, false
	}
	if payload.Total != nil {
		return *payload.Total, true
	}
	if payload.Results != nil {
		return len(payload.Results), true
	}
	return 0, false
}

func newTelemetrySessionID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "ax-mcp-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	return "ax-mcp-" + hex.EncodeToString(b[:])
}
