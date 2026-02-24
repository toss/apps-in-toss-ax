package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/toss/apps-in-toss-ax/pkg/docs"
)

func NewListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List AppsInToss resources",
	}

	cmd.AddCommand(newListExamplesCommand())

	return cmd
}

func newListExamplesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "examples",
		Short: "List AppsInToss examples",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			examples, err := docs.New().GetExamples(ctx)
			if err != nil {
				return err
			}

			output, err := json.MarshalIndent(examples, "", "  ")
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), string(output))
			return nil
		},
	}
}
