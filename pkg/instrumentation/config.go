package instrumentation

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	AdapterAmplitude = "amplitude"

	defaultTimeout = 2 * time.Second

	defaultInstrumentationConfigURL = "https://static.toss.im/appsintoss/ax/instrumentation.json"
	maxRemoteConfigBytes            = 64 * 1024
)

var (
	AMPLITUDE_API_KEY          string
	INSTRUMENTATION_CONFIG_URL = defaultInstrumentationConfigURL
)

type Config struct {
	Enabled           bool                   `json:"enabled"`
	AnonymousID       string                 `json:"anonymous_id,omitempty"`
	TimeoutMS         int                    `json:"timeout_ms,omitempty"`
	DefaultProperties map[string]any         `json:"default_properties,omitempty"`
	Adapters          []AnalyticsAdapterConf `json:"adapters,omitempty"`
}

type AnalyticsAdapterConf struct {
	Type     string `json:"type"`
	Enabled  *bool  `json:"enabled,omitempty"`
	APIKey   string `json:"api_key,omitempty"`
	Endpoint string `json:"endpoint,omitempty"`
}

type remoteInstrumentationConfig struct {
	Enabled           *bool                        `json:"enabled,omitempty"`
	TimeoutMS         *int                         `json:"timeout_ms,omitempty"`
	DefaultProperties map[string]any               `json:"default_properties,omitempty"`
	Adapters          []remoteAnalyticsAdapterConf `json:"adapters,omitempty"`
}

type remoteAnalyticsAdapterConf struct {
	Type    string `json:"type"`
	Enabled *bool  `json:"enabled,omitempty"`
}

func ConfigDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get config dir: %w", err)
	}
	return filepath.Join(configDir, "ax"), nil
}

func ConfigPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "analytics.json"), nil
}

func LoadConfig() (Config, error) {
	cfg := DefaultConfig()

	path, err := ConfigPath()
	if err != nil {
		return cfg, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return configWithRuntime(configWithRemote(cfg)), nil
		}
		return cfg, fmt.Errorf("failed to read analytics config: %w", err)
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("failed to parse analytics config: %w", err)
	}

	if cfg.AnonymousID == "" {
		cfg.AnonymousID = newAnonymousID()
	}

	return configWithRuntime(configWithRemote(cfg)), nil
}

func SaveConfig(cfg Config) error {
	if cfg.AnonymousID == "" {
		cfg.AnonymousID = newAnonymousID()
	}
	cfg = configWithoutRuntimeSecrets(cfg)

	dir, err := ConfigDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal analytics config: %w", err)
	}

	path, err := ConfigPath()
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write analytics config: %w", err)
	}

	return nil
}

func DefaultConfig() Config {
	return Config{
		Enabled:     true,
		AnonymousID: newAnonymousID(),
		TimeoutMS:   int(defaultTimeout / time.Millisecond),
		DefaultProperties: map[string]any{
			"app": "ax",
		},
	}
}

func (c Config) Timeout() time.Duration {
	if c.TimeoutMS <= 0 {
		return defaultTimeout
	}
	return time.Duration(c.TimeoutMS) * time.Millisecond
}

func (c Config) Redacted() Config {
	redacted := c
	redacted.Adapters = make([]AnalyticsAdapterConf, 0, len(c.Adapters))
	for _, adapter := range c.Adapters {
		if adapter.APIKey != "" {
			adapter.APIKey = "********"
		}
		redacted.Adapters = append(redacted.Adapters, adapter)
	}
	return redacted
}

func (c *Config) UpsertAdapter(adapter AnalyticsAdapterConf) {
	adapter.Type = NormalizeAdapterType(adapter.Type)
	for i := range c.Adapters {
		if NormalizeAdapterType(c.Adapters[i].Type) == adapter.Type {
			c.Adapters[i] = mergeAdapterConfig(c.Adapters[i], adapter)
			return
		}
	}
	c.Adapters = append(c.Adapters, adapter)
}

func (c *Config) SetAdapterEnabled(adapterType string, enabled bool) bool {
	adapterType = NormalizeAdapterType(adapterType)
	for i := range c.Adapters {
		if NormalizeAdapterType(c.Adapters[i].Type) == adapterType {
			c.Adapters[i].Enabled = boolPtr(enabled)
			return true
		}
	}
	return false
}

func NormalizeAdapterType(adapterType string) string {
	switch strings.ToLower(strings.TrimSpace(adapterType)) {
	case "amp", "amplitude":
		return AdapterAmplitude
	default:
		return strings.ToLower(strings.TrimSpace(adapterType))
	}
}

func AdapterEnabled(adapter AnalyticsAdapterConf) bool {
	return adapter.Enabled == nil || *adapter.Enabled
}

func boolPtr(v bool) *bool {
	return &v
}

func mergeAdapterConfig(base AnalyticsAdapterConf, next AnalyticsAdapterConf) AnalyticsAdapterConf {
	base.Type = NormalizeAdapterType(next.Type)
	if next.Enabled != nil {
		base.Enabled = next.Enabled
	}
	if next.APIKey != "" {
		base.APIKey = next.APIKey
	}
	if next.Endpoint != "" {
		base.Endpoint = next.Endpoint
	}
	return base
}

func configWithRuntime(cfg Config) Config {
	return configWithBuildTime(cfg)
}

func configWithRemote(cfg Config) Config {
	url := instrumentationConfigURL()
	if url == "" {
		return cfg
	}

	remote, err := fetchRemoteInstrumentationConfig(http.DefaultClient, url)
	if err != nil {
		return cfg
	}
	return remote.apply(cfg)
}

func fetchRemoteInstrumentationConfig(client *http.Client, url string) (remoteInstrumentationConfig, error) {
	var remote remoteInstrumentationConfig
	if client == nil {
		client = http.DefaultClient
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return remote, err
	}

	clientWithTimeout := *client
	if clientWithTimeout.Timeout == 0 {
		clientWithTimeout.Timeout = defaultTimeout
	}

	resp, err := clientWithTimeout.Do(req)
	if err != nil {
		return remote, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return remote, fmt.Errorf("remote instrumentation config returned status %d", resp.StatusCode)
	}

	if err := json.NewDecoder(io.LimitReader(resp.Body, maxRemoteConfigBytes)).Decode(&remote); err != nil {
		return remote, err
	}
	return remote, nil
}

func instrumentationConfigURL() string {
	return strings.TrimSpace(INSTRUMENTATION_CONFIG_URL)
}

func (r remoteInstrumentationConfig) apply(cfg Config) Config {
	if r.Enabled != nil {
		cfg.Enabled = *r.Enabled
	}
	if r.TimeoutMS != nil && *r.TimeoutMS > 0 {
		cfg.TimeoutMS = *r.TimeoutMS
	}
	if len(r.DefaultProperties) > 0 {
		if cfg.DefaultProperties == nil {
			cfg.DefaultProperties = map[string]any{}
		}
		maps.Copy(cfg.DefaultProperties, r.DefaultProperties)
	}
	if r.Adapters != nil {
		cfg.Adapters = remoteAdaptersWithExistingSecrets(r.Adapters, cfg.Adapters)
	}
	return cfg
}

func remoteAdaptersWithExistingSecrets(remoteAdapters []remoteAnalyticsAdapterConf, existing []AnalyticsAdapterConf) []AnalyticsAdapterConf {
	adapters := make([]AnalyticsAdapterConf, 0, len(remoteAdapters))
	for _, remoteAdapter := range remoteAdapters {
		adapterType := NormalizeAdapterType(remoteAdapter.Type)
		if adapterType == "" {
			continue
		}

		adapter := findExistingAdapter(existing, adapterType)
		adapter.Type = adapterType
		if remoteAdapter.Enabled != nil {
			adapter.Enabled = remoteAdapter.Enabled
		}
		adapters = append(adapters, adapter)
	}
	return adapters
}

func findExistingAdapter(adapters []AnalyticsAdapterConf, adapterType string) AnalyticsAdapterConf {
	adapterType = NormalizeAdapterType(adapterType)
	for _, adapter := range adapters {
		if NormalizeAdapterType(adapter.Type) == adapterType {
			return adapter
		}
	}
	return AnalyticsAdapterConf{}
}

func configWithBuildTime(cfg Config) Config {
	if apiKey := strings.TrimSpace(AMPLITUDE_API_KEY); apiKey != "" {
		cfg.SetAdapterSecret(AdapterAmplitude, func(adapter *AnalyticsAdapterConf) {
			adapter.APIKey = apiKey
		})
	}
	return cfg
}

func (c *Config) SetAdapterSecret(adapterType string, update func(*AnalyticsAdapterConf)) bool {
	adapterType = NormalizeAdapterType(adapterType)
	for i := range c.Adapters {
		if NormalizeAdapterType(c.Adapters[i].Type) == adapterType {
			update(&c.Adapters[i])
			return true
		}
	}
	return false
}

func configWithoutRuntimeSecrets(cfg Config) Config {
	cfg.Adapters = append([]AnalyticsAdapterConf(nil), cfg.Adapters...)
	for i := range cfg.Adapters {
		if NormalizeAdapterType(cfg.Adapters[i].Type) == AdapterAmplitude {
			if isRuntimeSecret(cfg.Adapters[i].APIKey, AMPLITUDE_API_KEY) {
				cfg.Adapters[i].APIKey = ""
			}
		}
	}
	return cfg
}

func isRuntimeSecret(value string, candidates ...string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return false
	}
	for _, candidate := range candidates {
		if value == strings.TrimSpace(candidate) {
			return true
		}
	}
	return false
}

func newAnonymousID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("ax-%d", time.Now().UnixNano())
	}
	return "ax-" + hex.EncodeToString(b[:])
}
