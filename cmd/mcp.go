package cmd

import (
	"github.com/spf13/cobra"
	"github.com/toss/apps-in-toss-ax/pkg/mcp"
)

func NewMcpCommand() *cobra.Command {
	var platform string

	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Manage MCP (Message Context Protocol) servers",
		Run: func(cmd *cobra.Command, args []string) {
			startMcpServer(cmd, platform)
		},
	}

	cmd.Flags().StringVar(&platform, "platform", "", "Filter by platform: web, rn (react-native)")

	return cmd
}

func startMcpServer(cmd *cobra.Command, platform string) error {
	p := mcp.New(mcp.WithPlatform(platform))

	return p.Server.Run(cmd.Context(), p.Transport)
}
