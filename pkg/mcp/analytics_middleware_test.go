package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/toss/apps-in-toss-ax/pkg/instrumentation"
)

func TestAnalyticsMiddlewareTracksSkillEvents(t *testing.T) {
	requests := make(chan amplitudeRequest, 4)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body amplitudeRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode amplitude request: %v", err)
		}
		requests <- body
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	analytics, err := instrumentation.NewAnalyticsWithHTTPClient(instrumentation.Config{
		Enabled:     true,
		AnonymousID: "ax-test-user",
		TimeoutMS:   int(time.Second / time.Millisecond),
		Adapters: []instrumentation.AnalyticsAdapterConf{
			{
				Type:     instrumentation.AdapterAmplitude,
				APIKey:   "amp-key",
				Endpoint: server.URL,
			},
		},
	}, server.Client())
	if err != nil {
		t.Fatal(err)
	}
	defer analytics.Close(context.Background())

	p := New(WithAnalytics(analytics))
	handler := p.analyticsMiddleware()(func(context.Context, string, mcpsdk.Request) (mcpsdk.Result, error) {
		return &mcpsdk.CallToolResult{
			StructuredContent: SearchOutput{
				Total: 3,
			},
		}, nil
	})
	req := &mcpsdk.ServerRequest[*mcpsdk.CallToolParamsRaw]{
		Params: &mcpsdk.CallToolParamsRaw{
			Name:      "search_docs",
			Arguments: json.RawMessage(`{"query":"결제 연동","limit":3}`),
		},
	}

	if _, err := handler(context.Background(), "tools/call", req); err != nil {
		t.Fatal(err)
	}
	if err := analytics.Flush(context.Background()); err != nil {
		t.Fatal(err)
	}

	events := collectAmplitudeEvents(t, requests, 2)
	invoked := findAmplitudeEvent(t, events, "skill_invoked")
	if invoked.EventProperties["session_id"] == "" {
		t.Fatal("skill_invoked session_id is empty")
	}
	if invoked.EventProperties["skill_name"] != "search_docs" {
		t.Errorf("skill_invoked skill_name = %v, want search_docs", invoked.EventProperties["skill_name"])
	}
	if invoked.EventProperties["query"] != "결제 연동" {
		t.Errorf("skill_invoked query = %v, want 결제 연동", invoked.EventProperties["query"])
	}

	rendered := findAmplitudeEvent(t, events, "skill_response_rendered")
	if rendered.EventProperties["session_id"] != invoked.EventProperties["session_id"] {
		t.Errorf("skill_response_rendered session_id = %v, want %v", rendered.EventProperties["session_id"], invoked.EventProperties["session_id"])
	}
	if rendered.EventProperties["skill_name"] != "search_docs" {
		t.Errorf("skill_response_rendered skill_name = %v, want search_docs", rendered.EventProperties["skill_name"])
	}
	if _, ok := rendered.EventProperties["query"]; ok {
		t.Errorf("skill_response_rendered query = %v, want omitted", rendered.EventProperties["query"])
	}
	if rendered.EventProperties["search_result_count"] != float64(3) {
		t.Errorf("skill_response_rendered search_result_count = %v, want 3", rendered.EventProperties["search_result_count"])
	}
}

func TestSearchResultCount(t *testing.T) {
	tests := []struct {
		name   string
		result mcpsdk.Result
		want   int
		wantOK bool
	}{
		{
			name: "search output total",
			result: &mcpsdk.CallToolResult{
				StructuredContent: SearchOutput{Total: 3},
			},
			want:   3,
			wantOK: true,
		},
		{
			name: "map total",
			result: &mcpsdk.CallToolResult{
				StructuredContent: map[string]any{"total": 4},
			},
			want:   4,
			wantOK: true,
		},
		{
			name: "results fallback",
			result: &mcpsdk.CallToolResult{
				StructuredContent: map[string]any{
					"results": []any{
						map[string]any{"id": "first"},
						map[string]any{"id": "second"},
					},
				},
			},
			want:   2,
			wantOK: true,
		},
		{
			name:   "nil result",
			result: nil,
			wantOK: false,
		},
		{
			name:   "missing structured content",
			result: &mcpsdk.CallToolResult{},
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := searchResultCount(tt.result)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if got != tt.want {
				t.Fatalf("count = %d, want %d", got, tt.want)
			}
		})
	}
}

type amplitudeRequest struct {
	Events []amplitudeEvent `json:"events"`
}

type amplitudeEvent struct {
	EventType       string         `json:"event_type"`
	EventProperties map[string]any `json:"event_properties"`
}

func collectAmplitudeEvents(t *testing.T, requests <-chan amplitudeRequest, count int) []amplitudeEvent {
	t.Helper()

	var events []amplitudeEvent
	deadline := time.After(2 * time.Second)
	for len(events) < count {
		select {
		case req := <-requests:
			events = append(events, req.Events...)
		case <-deadline:
			t.Fatalf("timed out waiting for %d amplitude events, got %d", count, len(events))
		}
	}
	return events
}

func findAmplitudeEvent(t *testing.T, events []amplitudeEvent, eventType string) amplitudeEvent {
	t.Helper()

	for _, event := range events {
		if event.EventType == eventType {
			return event
		}
	}
	t.Fatalf("event %s not found in %+v", eventType, events)
	return amplitudeEvent{}
}
