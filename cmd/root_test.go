package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestUsageStatsDisabled(t *testing.T) {
	root := &cobra.Command{Use: "ax"}
	root.PersistentFlags().Bool(disableUsageStatsFlag, false, "")
	parent := &cobra.Command{Use: "mcp"}
	child := &cobra.Command{Use: "start"}
	root.AddCommand(parent)
	parent.AddCommand(child)

	if usageStatsDisabled(child) {
		t.Fatal("usageStatsDisabled = true before flag is set")
	}

	if err := root.PersistentFlags().Set(disableUsageStatsFlag, "true"); err != nil {
		t.Fatal(err)
	}
	if !usageStatsDisabled(child) {
		t.Fatal("usageStatsDisabled = false after root flag is set")
	}
}

func TestUsageStatsDisabledLocalFlag(t *testing.T) {
	cmd := &cobra.Command{Use: "start"}
	cmd.Flags().Bool(disableUsageStatsFlag, false, "")

	if usageStatsDisabled(cmd) {
		t.Fatal("usageStatsDisabled = true before flag is set")
	}

	if err := cmd.Flags().Set(disableUsageStatsFlag, "true"); err != nil {
		t.Fatal(err)
	}
	if !usageStatsDisabled(cmd) {
		t.Fatal("usageStatsDisabled = false after local flag is set")
	}
}
