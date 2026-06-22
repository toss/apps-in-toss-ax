package instrumentation

import (
	"context"
	"fmt"
	"maps"
	"net/http"
	"runtime"
	"sync"
	"time"
)

type AnalyticsAdapter interface {
	Name() string
	Track(ctx context.Context, event Event) error
}

type flushableAdapter interface {
	Flush(ctx context.Context) error
}

type closeableAdapter interface {
	Close(ctx context.Context) error
}

type Event struct {
	Name        string
	Timestamp   time.Time
	AnonymousID string
	Properties  map[string]any
}

type Analytics struct {
	enabled           bool
	anonymousID       string
	timeout           time.Duration
	defaultProperties map[string]any
	adapters          []AnalyticsAdapter
	wg                sync.WaitGroup
}

func NewAnalytics(cfg Config) (*Analytics, error) {
	return NewAnalyticsWithHTTPClient(cfg, http.DefaultClient)
}

func NewAnalyticsWithHTTPClient(cfg Config, client *http.Client) (*Analytics, error) {
	if client == nil {
		client = http.DefaultClient
	}
	if cfg.AnonymousID == "" {
		cfg.AnonymousID = newAnonymousID()
	}

	defaultProperties := newPropertiesFactory(cfg.DefaultProperties).
		WithRuntime().
		Build()

	analytics := &Analytics{
		enabled:           cfg.Enabled,
		anonymousID:       cfg.AnonymousID,
		timeout:           cfg.Timeout(),
		defaultProperties: defaultProperties,
	}

	if !cfg.Enabled {
		return analytics, nil
	}

	for _, adapterConfig := range cfg.Adapters {
		if !AdapterEnabled(adapterConfig) {
			continue
		}
		adapter, err := newAdapter(adapterConfig, client, analytics.timeout)
		if err != nil {
			continue
		}
		analytics.adapters = append(analytics.adapters, adapter)
	}

	return analytics, nil
}

func NewAdapter(cfg AnalyticsAdapterConf, client *http.Client) (AnalyticsAdapter, error) {
	return newAdapter(cfg, client, defaultTimeout)
}

func newAdapter(cfg AnalyticsAdapterConf, client *http.Client, timeout time.Duration) (AnalyticsAdapter, error) {
	switch NormalizeAdapterType(cfg.Type) {
	case AdapterAmplitude:
		return NewAmplitudeAdapter(cfg, client, timeout)
	default:
		return nil, fmt.Errorf("unsupported analytics adapter: %s", cfg.Type)
	}
}

func (a *Analytics) Enabled() bool {
	return a != nil && a.enabled && len(a.adapters) > 0
}

func (a *Analytics) AdapterNames() []string {
	if a == nil {
		return nil
	}
	names := make([]string, 0, len(a.adapters))
	for _, adapter := range a.adapters {
		names = append(names, adapter.Name())
	}
	return names
}

func (a *Analytics) Track(ctx context.Context, name string, properties map[string]any) {
	if !a.Enabled() || name == "" {
		return
	}

	event := Event{
		Name:        name,
		Timestamp:   time.Now().UTC(),
		AnonymousID: a.anonymousID,
		Properties:  a.eventProperties(properties),
	}

	for _, adapter := range a.adapters {
		adapter := adapter
		a.wg.Add(1)
		go func() {
			defer a.wg.Done()
			sendCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), a.timeout)
			defer cancel()
			_ = adapter.Track(sendCtx, event)
		}()
	}
}

func (a *Analytics) Flush(ctx context.Context) error {
	if a == nil {
		return nil
	}

	if err := a.waitForTracks(ctx); err != nil {
		return err
	}

	for _, adapter := range a.adapters {
		adapter, ok := adapter.(flushableAdapter)
		if !ok {
			continue
		}
		if err := adapter.Flush(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (a *Analytics) Close(ctx context.Context) error {
	if a == nil {
		return nil
	}

	if err := a.Flush(ctx); err != nil {
		return err
	}

	for _, adapter := range a.adapters {
		adapter, ok := adapter.(closeableAdapter)
		if !ok {
			continue
		}
		if err := adapter.Close(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (a *Analytics) waitForTracks(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		a.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (a *Analytics) eventProperties(properties map[string]any) map[string]any {
	return newPropertiesFactory(a.defaultProperties).
		WithProperties(properties).
		Build()
}

func cloneProperties(properties map[string]any) map[string]any {
	clone := make(map[string]any, len(properties))
	maps.Copy(clone, properties)
	return clone
}

type propertiesFactory struct {
	properties map[string]any
}

func newPropertiesFactory(base map[string]any) *propertiesFactory {
	return &propertiesFactory{
		properties: cloneProperties(base),
	}
}

func (f *propertiesFactory) WithRuntime() *propertiesFactory {
	f.properties["os"] = runtime.GOOS
	f.properties["architecture"] = runtime.GOARCH
	return f
}

func (f *propertiesFactory) WithProperties(properties map[string]any) *propertiesFactory {
	maps.Copy(f.properties, properties)
	return f
}

func (f *propertiesFactory) Build() map[string]any {
	return f.properties
}
