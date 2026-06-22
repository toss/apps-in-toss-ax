package instrumentation

import (
	"runtime"
	"testing"
)

func TestNewAnalyticsSkipsMisconfiguredAdapters(t *testing.T) {
	analytics, err := NewAnalytics(Config{
		Enabled: true,
		Adapters: []AnalyticsAdapterConf{
			{Type: AdapterAmplitude},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if analytics.Enabled() {
		t.Fatal("analytics.Enabled() = true, want false when all adapters are misconfigured")
	}
}

func TestNewAnalyticsAddsRuntimeProperties(t *testing.T) {
	analytics, err := NewAnalytics(Config{Enabled: false})
	if err != nil {
		t.Fatal(err)
	}

	properties := analytics.eventProperties(nil)
	if _, ok := properties["device_os"]; ok {
		t.Error("device_os property is set, want omitted")
	}
	if properties["os"] != runtime.GOOS {
		t.Errorf("os = %v, want %s", properties["os"], runtime.GOOS)
	}
	if properties["architecture"] != runtime.GOARCH {
		t.Errorf("architecture = %v, want %s", properties["architecture"], runtime.GOARCH)
	}
}
