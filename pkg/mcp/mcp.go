package mcp

import (
	"context"
	"sync"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/toss/apps-in-toss-ax/pkg/search"
)

const (
	name    = "ax"
	title   = "ax"
	version = "0.1.0"
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

	docSearcher *lazySearcher
	tdsRn       *lazySearcher
	tdsWeb      *lazySearcher
}

type Option func(*Protocol)

func WithTransport(transport mcp.Transport) Option {
	return func(s *Protocol) {
		s.Transport = transport
	}
}

func New(options ...Option) *Protocol {
	p := &Protocol{
		Transport:   &mcp.StdioTransport{},
		OnInit:      func(_ context.Context) {},
		docSearcher: newLazySearcher(search.New),
		tdsRn:       newLazySearcher(search.NewTDSSearcher),
		tdsWeb:      newLazySearcher(search.NewTDSMobileSearcher),
	}

	for _, o := range options {
		o(p)
	}

	i := mcp.NewServer(
		&mcp.Implementation{
			Name:    name,
			Title:   title,
			Version: version,
		},
		&mcp.ServerOptions{
			Instructions: instructions(),
			HasPrompts:   true,
			HasResources: true,
			HasTools:     true,
		})

	mcp.AddTool(i, listExamples, p.listExamplesHandler)
	mcp.AddTool(i, getExample, p.getExampleHandler)
	mcp.AddTool(i, searchDocs, p.searchDocsHandler)
	mcp.AddTool(i, searchTdsRnDocs, p.searchTdsRnDocsHandler)
	mcp.AddTool(i, searchTdsWebDocs, p.searchTdsWebDocsHandler)
	mcp.AddTool(i, getDoc, p.getDocHandler)
	mcp.AddTool(i, getTdsRnDoc, p.getTdsRnDocHandler)
	mcp.AddTool(i, getTdsWebDoc, p.getTdsWebDocHandler)

	p.Server = i
	return p
}
