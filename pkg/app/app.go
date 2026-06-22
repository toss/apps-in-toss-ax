package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/toss/apps-in-toss-ax/cmd"
	"github.com/toss/apps-in-toss-ax/pkg/features"
	"github.com/toss/apps-in-toss-ax/pkg/instrumentation"
)

func Run() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		cancel()
		<-sigs
		os.Exit(137) // sigkill
	}()

	instrumentationConfig := loadInstrumentationConfig(cmd.GetVersion())
	featureStack, err := features.NewFeatureStack(instrumentationConfig)
	if err != nil {
		fmt.Printf("Err: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer shutdownCancel()
		_ = featureStack.Close(shutdownCtx)
	}()

	cfg := cmd.CommandConfig{
		Name:            "ax",
		Instrumentation: featureStack.Instrumentation,
	}

	if err := cmd.NewCommand(cfg).ExecuteContext(ctx); err != nil {
		fmt.Printf("Err: %v\n", err)
		os.Exit(1)
	}
}

func loadInstrumentationConfig(version cmd.VersionInfo) instrumentation.Config {
	cfg, err := instrumentation.LoadConfig()
	if err != nil {
		cfg = instrumentation.DefaultConfig()
		cfg.Enabled = false
	}
	if cfg.DefaultProperties == nil {
		cfg.DefaultProperties = map[string]any{}
	}
	cfg.DefaultProperties["ax_version"] = version.Version
	cfg.DefaultProperties["ax_commit"] = version.Hash
	return cfg
}
