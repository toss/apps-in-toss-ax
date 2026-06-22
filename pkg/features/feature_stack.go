package features

import (
	"context"
	"maps"

	"github.com/toss/apps-in-toss-ax/pkg/instrumentation"
)

type InstrumentationFeature struct {
	Analytics *instrumentation.Analytics
}

type FeatureStack struct {
	Instrumentation InstrumentationFeature
}

func NewFeatureStack(cfg instrumentation.Config) (*FeatureStack, error) {
	analytics, err := instrumentation.NewAnalytics(cfg)
	if err != nil {
		disabled := instrumentation.DefaultConfig()
		disabled.Enabled = false
		mergeDefaultProperties(&disabled, cfg.DefaultProperties)
		analytics, _ = instrumentation.NewAnalytics(disabled)
	}

	return &FeatureStack{
		Instrumentation: InstrumentationFeature{
			Analytics: analytics,
		},
	}, nil
}

func (s *FeatureStack) Close(ctx context.Context) error {
	if s == nil || s.Instrumentation.Analytics == nil {
		return nil
	}
	return s.Instrumentation.Analytics.Close(ctx)
}

func mergeDefaultProperties(cfg *instrumentation.Config, properties map[string]any) {
	if len(properties) == 0 {
		return
	}
	if cfg.DefaultProperties == nil {
		cfg.DefaultProperties = map[string]any{}
	}
	maps.Copy(cfg.DefaultProperties, properties)
}
