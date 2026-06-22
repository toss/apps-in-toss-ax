package cmd

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/toss/apps-in-toss-ax/pkg/features"
)

const disableUsageStatsFlag = "disable-usage-stats"

type CommandConfig struct {
	Name            string
	Instrumentation features.InstrumentationFeature
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
	cmd.PersistentFlags().Bool(disableUsageStatsFlag, false, "Disable usage statistics for this command")
	cmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		trackCommand(cfg.Instrumentation, cmd, args, "cli.command_started", nil)
	}
	cmd.PersistentPostRun = func(cmd *cobra.Command, args []string) {
		trackCommand(cfg.Instrumentation, cmd, args, "cli.command_completed", map[string]any{
			"duration_ms": commandDurationMS(cmd),
		})
	}

	defaultHelp := cmd.HelpFunc()
	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		if cmd.Root() == cmd {
			printBanner()
		}
		defaultHelp(cmd, args)
	})

	cmd.AddCommand(NewMcpCommand(cfg.Instrumentation))
	cmd.AddCommand(NewVersionCommand())
	cmd.AddCommand(NewSearchCommand())
	cmd.AddCommand(NewGetCommand())
	cmd.AddCommand(NewListCommand())

	return cmd
}

func trackCommand(instrumentation features.InstrumentationFeature, cmd *cobra.Command, args []string, eventName string, properties map[string]any) {
	if instrumentation.Analytics == nil {
		return
	}
	if usageStatsDisabled(cmd) {
		return
	}
	if properties == nil {
		properties = map[string]any{}
	}
	properties["command"] = cmd.CommandPath()
	properties["leaf_command"] = cmd.Name()
	properties["args_count"] = len(args)
	if eventName == "cli.command_started" {
		cmd.SetContext(context.WithValue(cmd.Context(), commandStartKey{}, time.Now()))
	}
	instrumentation.Analytics.Track(cmd.Context(), eventName, properties)
}

func usageStatsDisabled(cmd *cobra.Command) bool {
	for current := cmd; current != nil; current = current.Parent() {
		if flagSetBool(current.Flags(), disableUsageStatsFlag) ||
			flagSetBool(current.PersistentFlags(), disableUsageStatsFlag) ||
			flagSetBool(current.InheritedFlags(), disableUsageStatsFlag) {
			return true
		}
	}
	return false
}

func flagSetBool(flags *pflag.FlagSet, name string) bool {
	flag := flags.Lookup(name)
	if flag == nil {
		return false
	}
	value, err := strconv.ParseBool(flag.Value.String())
	return err == nil && value
}

type commandStartKey struct{}

func commandDurationMS(cmd *cobra.Command) int64 {
	startedAt, ok := cmd.Context().Value(commandStartKey{}).(time.Time)
	if !ok {
		return 0
	}
	return time.Since(startedAt).Milliseconds()
}
