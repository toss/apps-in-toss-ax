package cmd

import (
	"github.com/spf13/cobra"
	"github.com/toss/apps-in-toss-ax/pkg/features"
	"github.com/toss/apps-in-toss-ax/pkg/mcp"
)

func NewMcpCommand(instrumentation features.InstrumentationFeature) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Manage MCP (Message Context Protocol) servers",
		RunE: func(cmd *cobra.Command, args []string) error {
			return startMcpServer(cmd, args, instrumentation)
		},
	}

	return cmd
}

func startMcpServer(cmd *cobra.Command, _ []string, instrumentation features.InstrumentationFeature) error {
	analytics := instrumentation.Analytics
	if usageStatsDisabled(cmd) {
		analytics = nil
	}
	p := mcp.New(mcp.WithAnalytics(analytics))

	return p.Server.Run(cmd.Context(), p.Transport)
}
