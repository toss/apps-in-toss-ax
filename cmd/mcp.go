package cmd

import (
	"github.com/spf13/cobra"
	"github.com/toss/apps-in-toss-ax/pkg/mcp"
)

func NewMcpCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Manage MCP (Message Context Protocol) servers",
		Run: func(cmd *cobra.Command, args []string) {
			startMcpServer(cmd, args)
		},
	}
	return cmd
}

func startMcpServer(cmd *cobra.Command, _ []string) error {
	p := mcp.New()

	return p.Server.Run(cmd.Context(), p.Transport)
}
