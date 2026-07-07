package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/toss/apps-in-toss-ax/pkg/search"
)

type searcherFactory func() (*search.Searcher, error)

type searchFlags struct {
	query            string
	limit            int
	titleBoost       float64
	descriptionBoost float64
	contentBoost     float64
	categoryBoost    float64
}

func (f *searchFlags) register(cmd *cobra.Command) {
	cmd.Flags().StringVar(&f.query, "query", "", "Search query (required)")
	cmd.Flags().IntVar(&f.limit, "limit", 10, "Maximum number of results")
	cmd.Flags().Float64Var(&f.titleBoost, "title-boost", search.DefaultTitleBoost, "Relevance boost for title matches")
	cmd.Flags().Float64Var(&f.descriptionBoost, "description-boost", search.DefaultDescriptionBoost, "Relevance boost for description matches")
	cmd.Flags().Float64Var(&f.contentBoost, "content-boost", search.DefaultContentBoost, "Relevance boost for content matches")
	cmd.Flags().Float64Var(&f.categoryBoost, "category-boost", search.DefaultCategoryBoost, "Relevance boost for category matches")
	cmd.MarkFlagRequired("query")
}

func (f *searchFlags) options() *search.SearchOptions {
	return &search.SearchOptions{
		Limit: f.limit,
		Boosts: search.BoostOverrides{
			Title:       &f.titleBoost,
			Description: &f.descriptionBoost,
			Content:     &f.contentBoost,
			Category:    &f.categoryBoost,
		},
	}
}

func NewSearchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search AppsInToss documentation",
	}

	cmd.AddCommand(newSearchSubCommand("docs", "Search AppsInToss documentation", search.New))
	cmd.AddCommand(newSearchSubCommand("tds-rn", "Search TDS React Native documentation", search.NewTDSSearcher))
	cmd.AddCommand(newSearchSubCommand("tds-web", "Search TDS Web documentation", search.NewTDSMobileSearcher))

	return cmd
}

func newSearchSubCommand(use, short string, factory searcherFactory) *cobra.Command {
	var flags searchFlags

	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSearch(cmd, factory, &flags)
		},
	}

	flags.register(cmd)

	return cmd
}

func runSearch(cmd *cobra.Command, factory searcherFactory, flags *searchFlags) error {
	ctx := cmd.Context()

	s, err := factory()
	if err != nil {
		return err
	}
	defer s.Close()

	if err := s.EnsureIndex(ctx); err != nil {
		return err
	}

	results, err := s.Search(ctx, flags.query, flags.options())
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
