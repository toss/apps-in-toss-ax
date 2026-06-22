package instrumentation

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAmplitudeAdapterTrack(t *testing.T) {
	requests := make(chan amplitudeRequest, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want %s", r.Method, http.MethodPost)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("content-type = %s, want application/json", got)
		}

		var body amplitudeRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		requests <- body
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	adapter, err := NewAmplitudeAdapter(AnalyticsAdapterConf{
		APIKey:   "amp-key",
		Endpoint: server.URL,
	}, server.Client(), time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer adapter.Close(context.Background())

	event := testEvent()
	if err := adapter.Track(context.Background(), event); err != nil {
		t.Fatal(err)
	}
	if err := adapter.Flush(context.Background()); err != nil {
		t.Fatal(err)
	}

	body := <-requests
	if body.APIKey != "amp-key" {
		t.Fatalf("api_key = %s, want amp-key", body.APIKey)
	}
	if len(body.Events) != 1 {
		t.Fatalf("events length = %d, want 1", len(body.Events))
	}
	got := body.Events[0]
	if got.DeviceID != event.AnonymousID {
		t.Errorf("device_id = %s, want %s", got.DeviceID, event.AnonymousID)
	}
	if got.EventType != event.Name {
		t.Errorf("event_type = %s, want %s", got.EventType, event.Name)
	}
	if got.Time != event.Timestamp.UnixMilli() {
		t.Errorf("time = %d, want %d", got.Time, event.Timestamp.UnixMilli())
	}
	if got.EventProperties["command"] != "ax search docs" {
		t.Errorf("event_properties.command = %v", got.EventProperties["command"])
	}
}

func TestNewAdapterValidatesRequiredConfiguration(t *testing.T) {
	tests := []struct {
		name string
		cfg  AnalyticsAdapterConf
	}{
		{
			name: "amplitude api key",
			cfg:  AnalyticsAdapterConf{Type: AdapterAmplitude},
		},
		{
			name: "unsupported adapter",
			cfg:  AnalyticsAdapterConf{Type: "future_adapter"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := NewAdapter(tt.cfg, nil); err == nil {
				t.Fatal("NewAdapter returned nil error")
			}
		})
	}
}

type amplitudeRequest struct {
	APIKey string `json:"api_key"`
	Events []struct {
		DeviceID        string         `json:"device_id"`
		EventType       string         `json:"event_type"`
		Time            int64          `json:"time"`
		EventProperties map[string]any `json:"event_properties"`
	} `json:"events"`
}

func testEvent() Event {
	return Event{
		Name:        "cli.command-started",
		Timestamp:   time.Unix(1710000000, 123456789).UTC(),
		AnonymousID: "ax-test-user",
		Properties: map[string]any{
			"command": "ax search docs",
		},
	}
}
