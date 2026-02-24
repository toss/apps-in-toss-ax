package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

type CommandConfig struct {
	Name string
}

func NewCommand(cfg CommandConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   cfg.Name,
		Short: fmt.Sprintf("%s manages AppsInToss Developer eXperience with AI.", cfg.Name),
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: false,
		},
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	defaultHelp := cmd.HelpFunc()
	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		if cmd.Root() == cmd {
			printBanner()
		}
		defaultHelp(cmd, args)
	})

	cmd.AddCommand(NewMcpCommand())
	cmd.AddCommand(NewVersionCommand())
	cmd.AddCommand(NewSearchCommand())
	cmd.AddCommand(NewGetCommand())
	cmd.AddCommand(NewListCommand())

	return cmd
}
