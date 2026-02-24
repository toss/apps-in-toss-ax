package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/toss/apps-in-toss-ax/pkg/search"
)

type searcherFactory func() (*search.Searcher, error)

func NewSearchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search AppsInToss documentation",
	}

	cmd.AddCommand(newSearchDocsCommand())
	cmd.AddCommand(newSearchTdsRnCommand())
	cmd.AddCommand(newSearchTdsWebCommand())

	return cmd
}

func newSearchDocsCommand() *cobra.Command {
	var query string
	var limit int

	cmd := &cobra.Command{
		Use:   "docs",
		Short: "Search AppsInToss documentation",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSearch(cmd, search.New, query, limit)
		},
	}

	cmd.Flags().StringVar(&query, "query", "", "Search query (required)")
	cmd.Flags().IntVar(&limit, "limit", 10, "Maximum number of results")
	cmd.MarkFlagRequired("query")

	return cmd
}

func newSearchTdsRnCommand() *cobra.Command {
	var query string
	var limit int

	cmd := &cobra.Command{
		Use:   "tds-rn",
		Short: "Search TDS React Native documentation",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSearch(cmd, search.NewTDSSearcher, query, limit)
		},
	}

	cmd.Flags().StringVar(&query, "query", "", "Search query (required)")
	cmd.Flags().IntVar(&limit, "limit", 10, "Maximum number of results")
	cmd.MarkFlagRequired("query")

	return cmd
}

func newSearchTdsWebCommand() *cobra.Command {
	var query string
	var limit int

	cmd := &cobra.Command{
		Use:   "tds-web",
		Short: "Search TDS Web documentation",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSearch(cmd, search.NewTDSMobileSearcher, query, limit)
		},
	}

	cmd.Flags().StringVar(&query, "query", "", "Search query (required)")
	cmd.Flags().IntVar(&limit, "limit", 10, "Maximum number of results")
	cmd.MarkFlagRequired("query")

	return cmd
}

func runSearch(cmd *cobra.Command, factory searcherFactory, query string, limit int) error {
	ctx := cmd.Context()

	s, err := factory()
	if err != nil {
		return err
	}
	defer s.Close()

	if err := s.EnsureIndex(ctx); err != nil {
		return err
	}

	results, err := s.Search(ctx, query, &search.SearchOptions{Limit: limit})
	if err != nil {
		return err
	}

	output, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}

	fmt.Fprintln(cmd.OutOrStdout(), string(output))
	return nil
}
