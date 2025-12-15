package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

type CommandConfig struct {
	Name    string
	Version string
}

func NewCommand(cfg CommandConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   cfg.Name,
		Short: fmt.Sprintf("%s manages AppsInToss Developer eXperience with AI.", cfg.Name),
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}

	cmd.AddCommand(NewMcpCommand())

	return cmd
}
