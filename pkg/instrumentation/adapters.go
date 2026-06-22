package instrumentation

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	amplitude "github.com/amplitude/analytics-go/amplitude"
)

type AmplitudeAdapter struct {
	client amplitude.Client
}

func NewAmplitudeAdapter(cfg AnalyticsAdapterConf, _ *http.Client, timeout time.Duration) (*AmplitudeAdapter, error) {
	if strings.TrimSpace(cfg.APIKey) == "" {
		return nil, fmt.Errorf("amplitude analytics adapter requires api_key")
	}

	amplitudeConfig := amplitude.NewConfig(cfg.APIKey)
	if endpoint := strings.TrimSpace(cfg.Endpoint); endpoint != "" {
		amplitudeConfig.ServerURL = endpoint
	}
	amplitudeConfig.FlushQueueSize = 1
	amplitudeConfig.ConnectionTimeout = timeout
	amplitudeConfig.Logger = noopAmplitudeLogger{}

	return &AmplitudeAdapter{
		client: amplitude.NewClient(amplitudeConfig),
	}, nil
}

func (a *AmplitudeAdapter) Name() string {
	return AdapterAmplitude
}

func (a *AmplitudeAdapter) Track(ctx context.Context, event Event) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		a.client.Track(amplitude.Event{
			EventType: event.Name,
			EventOptions: amplitude.EventOptions{
				DeviceID: event.AnonymousID,
				Time:     event.Timestamp.UnixMilli(),
			},
			EventProperties: event.Properties,
		})
		return nil
	}
}

func (a *AmplitudeAdapter) Flush(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		a.client.Flush()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (a *AmplitudeAdapter) Close(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		a.client.Shutdown()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

type noopAmplitudeLogger struct{}

func (noopAmplitudeLogger) Debugf(string, ...interface{}) {}
func (noopAmplitudeLogger) Infof(string, ...interface{})  {}
func (noopAmplitudeLogger) Warnf(string, ...interface{})  {}
func (noopAmplitudeLogger) Errorf(string, ...interface{}) {}
