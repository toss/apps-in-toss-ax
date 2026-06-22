package app

import (
	"testing"

	"github.com/toss/apps-in-toss-ax/cmd"
)

func TestLoadInstrumentationConfigAddsVersionProperties(t *testing.T) {
	cfg := loadInstrumentationConfig(cmd.VersionInfo{
		Version: "1.2.3",
		Hash:    "abc123",
	})

	if cfg.DefaultProperties["ax_version"] != "1.2.3" {
		t.Errorf("ax_version = %v, want 1.2.3", cfg.DefaultProperties["ax_version"])
	}
	if cfg.DefaultProperties["ax_commit"] != "abc123" {
		t.Errorf("ax_commit = %v, want abc123", cfg.DefaultProperties["ax_commit"])
	}
}
