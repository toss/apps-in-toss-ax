package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/toss/apps-in-toss-ax/pkg/docs"
	"github.com/toss/apps-in-toss-ax/pkg/search"
)

func NewGetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get AppsInToss document or example by ID",
	}

	cmd.AddCommand(newGetDocCommand())
	cmd.AddCommand(newGetTdsRnCommand())
	cmd.AddCommand(newGetTdsWebCommand())
	cmd.AddCommand(newGetExampleCommand())

	return cmd
}

func newGetDocCommand() *cobra.Command {
	var id string

	cmd := &cobra.Command{
		Use:   "doc",
		Short: "Get an AppsInToss document by ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGetDoc(cmd, search.New, id)
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "Document ID (required)")
	cmd.MarkFlagRequired("id")

	return cmd
}

func newGetTdsRnCommand() *cobra.Command {
	var id string

	cmd := &cobra.Command{
		Use:   "tds-rn",
		Short: "Get a TDS React Native document by ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGetDoc(cmd, search.NewTDSSearcher, id)
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "Document ID (required)")
	cmd.MarkFlagRequired("id")

	return cmd
}

func newGetTdsWebCommand() *cobra.Command {
	var id string

	cmd := &cobra.Command{
		Use:   "tds-web",
		Short: "Get a TDS Web document by ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGetDoc(cmd, search.NewTDSMobileSearcher, id)
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "Document ID (required)")
	cmd.MarkFlagRequired("id")

	return cmd
}

func newGetExampleCommand() *cobra.Command {
	var id string

	cmd := &cobra.Command{
		Use:   "example",
		Short: "Get an AppsInToss example by ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			content, err := docs.New().GetExample(ctx, id)
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), content)
			return nil
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "Example ID (required)")
	cmd.MarkFlagRequired("id")

	return cmd
}

func runGetDoc(cmd *cobra.Command, factory searcherFactory, id string) error {
	ctx := cmd.Context()

	s, err := factory()
	if err != nil {
		return err
	}
	defer s.Close()

	if err := s.EnsureIndex(ctx); err != nil {
		return err
	}

	doc, err := s.GetDocument(ctx, id)
	if err != nil {
		return err
	}
	if doc == nil {
		return fmt.Errorf("document not found: %s", id)
	}

	output, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}

	fmt.Fprintln(cmd.OutOrStdout(), string(output))
	return nil
}
