package mcp

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/toss/apps-in-toss-ax/pkg/search"
)

var searchTdsWebDocs = &mcp.Tool{
	Name:        "search_tds_web_docs",
	Title:       "Search TDS Web Documents",
	Description: "Search TDS (Toss Design System) Web documentation using full-text search. Returns matching documents ranked by relevance.",
	Annotations: &mcp.ToolAnnotations{
		Title:          "Search TDS Web Documents",
		ReadOnlyHint:   true,
		IdempotentHint: true,
	},
}

func (p *Protocol) searchTdsWebDocsHandler(ctx context.Context, r *mcp.CallToolRequest, input SearchInput) (result *mcp.CallToolResult, output SearchOutput, err error) {
	searcher, err := search.NewTDSMobileSearcher()
	if err != nil {
		return nil, SearchOutput{}, err
	}
	defer searcher.Close()

	if err := searcher.EnsureIndex(ctx); err != nil {
		return nil, SearchOutput{}, err
	}

	limit := input.Limit
	if limit <= 0 {
		limit = 10
	}

	results, err := searcher.Search(ctx, input.Query, &search.SearchOptions{
		Limit: limit,
	})
	if err != nil {
		return nil, SearchOutput{}, err
	}

	return nil, SearchOutput{
		Results: results,
		Total:   len(results),
	}, nil
}
