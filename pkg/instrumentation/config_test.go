package instrumentation

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDefaultConfigEnablesAnalytics(t *testing.T) {
	cfg := DefaultConfig()
	if !cfg.Enabled {
		t.Fatal("DefaultConfig().Enabled = false, want true")
	}
}

func TestConfigUpsertAdapterMergesExistingAdapter(t *testing.T) {
	cfg := Config{}
	cfg.UpsertAdapter(AnalyticsAdapterConf{
		Type:    "amp",
		Enabled: boolPtr(true),
	})
	cfg.UpsertAdapter(AnalyticsAdapterConf{
		Type:     "amplitude",
		Endpoint: "https://example.com/amplitude",
	})

	if len(cfg.Adapters) != 1 {
		t.Fatalf("adapters length = %d, want 1", len(cfg.Adapters))
	}
	adapter := cfg.Adapters[0]
	if adapter.Type != AdapterAmplitude {
		t.Errorf("type = %s, want %s", adapter.Type, AdapterAmplitude)
	}
	if adapter.Enabled == nil || !*adapter.Enabled {
		t.Errorf("enabled = %v, want true", adapter.Enabled)
	}
	if adapter.Endpoint != "https://example.com/amplitude" {
		t.Errorf("endpoint = %s, want https://example.com/amplitude", adapter.Endpoint)
	}
}

func TestConfigRedactedHidesSecrets(t *testing.T) {
	cfg := Config{
		Adapters: []AnalyticsAdapterConf{
			{
				Type:   AdapterAmplitude,
				APIKey: "api-key",
			},
		},
	}

	redacted := cfg.Redacted()
	adapter := redacted.Adapters[0]
	if adapter.APIKey != "********" {
		t.Errorf("api_key = %s, want redacted", adapter.APIKey)
	}
}

func TestConfigWithRemoteFetchesStaticConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{
			"enabled": true,
			"default_properties": {
				"release_channel": "stable"
			},
			"adapters": [
				{"type": "amplitude", "enabled": true},
				{"type": "future_adapter", "enabled": false}
			]
		}`)
	}))
	defer server.Close()

	restoreInstrumentationConfigURL(server.URL)
	defer restoreInstrumentationConfigURL(defaultInstrumentationConfigURL)

	cfg := configWithRemote(Config{
		Enabled: false,
		Adapters: []AnalyticsAdapterConf{
			{
				Type:   AdapterAmplitude,
				APIKey: "existing-amp-key",
			},
			{
				Type:     "future_adapter",
				Endpoint: "https://collector.example.com",
			},
		},
	})

	if !cfg.Enabled {
		t.Fatal("enabled = false, want true")
	}
	if cfg.DefaultProperties["release_channel"] != "stable" {
		t.Errorf("release_channel = %v, want stable", cfg.DefaultProperties["release_channel"])
	}
	if len(cfg.Adapters) != 2 {
		t.Fatalf("adapters length = %d, want 2", len(cfg.Adapters))
	}

	amplitude := findTestAdapter(t, cfg, AdapterAmplitude)
	if amplitude.APIKey != "existing-amp-key" {
		t.Errorf("amplitude api_key = %s, want existing-amp-key", amplitude.APIKey)
	}
	if amplitude.Enabled == nil || !*amplitude.Enabled {
		t.Errorf("amplitude enabled = %v, want true", amplitude.Enabled)
	}

	future := findTestAdapter(t, cfg, "future_adapter")
	if future.Endpoint != "https://collector.example.com" {
		t.Errorf("future endpoint = %s, want https://collector.example.com", future.Endpoint)
	}
	if future.Enabled == nil || *future.Enabled {
		t.Errorf("future enabled = %v, want false", future.Enabled)
	}
}

func TestConfigWithRemoteIgnoresUnavailableConfig(t *testing.T) {
	restoreInstrumentationConfigURL("http://127.0.0.1:1/instrumentation.json")
	defer restoreInstrumentationConfigURL(defaultInstrumentationConfigURL)

	cfg := Config{
		Enabled: false,
		Adapters: []AnalyticsAdapterConf{
			{Type: AdapterAmplitude},
		},
	}

	got := configWithRemote(cfg)
	if got.Enabled != cfg.Enabled {
		t.Errorf("enabled = %v, want %v", got.Enabled, cfg.Enabled)
	}
	if len(got.Adapters) != 1 || got.Adapters[0].Type != AdapterAmplitude {
		t.Fatalf("adapters = %+v, want original adapters", got.Adapters)
	}
}

func TestConfigWithBuildTimeFillsSelectedAdapterSecrets(t *testing.T) {
	restoreBuildTimeSecrets("build-amp-key")
	defer restoreBuildTimeSecrets("")

	cfg := configWithBuildTime(Config{
		Adapters: []AnalyticsAdapterConf{
			{Type: AdapterAmplitude},
			{Type: "future_adapter"},
		},
	})

	amplitude := findTestAdapter(t, cfg, AdapterAmplitude)
	if amplitude.APIKey != "build-amp-key" {
		t.Errorf("amplitude api_key = %s, want build-amp-key", amplitude.APIKey)
	}
	if amplitude.Enabled != nil {
		t.Errorf("amplitude enabled = %v, want nil to preserve user setting", amplitude.Enabled)
	}

	future := findTestAdapter(t, cfg, "future_adapter")
	if future.APIKey != "" {
		t.Errorf("future api_key = %s, want empty", future.APIKey)
	}
}

func TestConfigWithBuildTimeDoesNotSelectAdapters(t *testing.T) {
	restoreBuildTimeSecrets("build-amp-key")
	defer restoreBuildTimeSecrets("")

	cfg := configWithBuildTime(Config{})
	if len(cfg.Adapters) != 0 {
		t.Fatalf("adapters length = %d, want 0", len(cfg.Adapters))
	}
}

func TestConfigWithBuildTimePreservesDisabledAdapter(t *testing.T) {
	restoreBuildTimeSecrets("build-amp-key")
	defer restoreBuildTimeSecrets("")

	cfg := configWithBuildTime(Config{
		Adapters: []AnalyticsAdapterConf{
			{
				Type:    AdapterAmplitude,
				Enabled: boolPtr(false),
			},
		},
	})

	amplitude := findTestAdapter(t, cfg, AdapterAmplitude)
	if amplitude.Enabled == nil || *amplitude.Enabled {
		t.Errorf("amplitude enabled = %v, want false", amplitude.Enabled)
	}
	if amplitude.APIKey != "build-amp-key" {
		t.Errorf("amplitude api_key = %s, want build-amp-key", amplitude.APIKey)
	}
}

func TestConfigWithoutRuntimeSecrets(t *testing.T) {
	restoreBuildTimeSecrets("build-amp-key")
	defer restoreBuildTimeSecrets("")

	cfg := configWithoutRuntimeSecrets(Config{
		Adapters: []AnalyticsAdapterConf{
			{
				Type:   AdapterAmplitude,
				APIKey: "build-amp-key",
			},
			{
				Type:   AdapterAmplitude,
				APIKey: "manual-amp-key",
			},
			{
				Type:   "future_adapter",
				APIKey: "future-key",
			},
		},
	})

	if got := cfg.Adapters[0].APIKey; got != "" {
		t.Errorf("build-time amplitude api_key = %s, want empty", got)
	}
	if got := cfg.Adapters[1].APIKey; got != "manual-amp-key" {
		t.Errorf("manual amplitude api_key = %s, want manual-amp-key", got)
	}
	if got := cfg.Adapters[2].APIKey; got != "future-key" {
		t.Errorf("future api_key = %s, want future-key", got)
	}
}

func restoreBuildTimeSecrets(amplitudeAPIKey string) {
	AMPLITUDE_API_KEY = amplitudeAPIKey
}

func restoreInstrumentationConfigURL(url string) {
	INSTRUMENTATION_CONFIG_URL = url
}

func findTestAdapter(t *testing.T, cfg Config, adapterType string) AnalyticsAdapterConf {
	t.Helper()
	adapterType = NormalizeAdapterType(adapterType)
	for _, adapter := range cfg.Adapters {
		if NormalizeAdapterType(adapter.Type) == adapterType {
			return adapter
		}
	}
	t.Fatalf("adapter %s not found", adapterType)
	return AnalyticsAdapterConf{}
}
