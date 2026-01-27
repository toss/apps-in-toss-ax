package mcp

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/toss/apps-in-toss-ax/pkg/search"
)

var searchTdsRnDocs = &mcp.Tool{
	Name:        "search_tds_rn_docs",
	Title:       "Search TDS React Native Documents",
	Description: "Search TDS (Toss Design System) React Native documentation using full-text search. Returns matching documents ranked by relevance.",
	Annotations: &mcp.ToolAnnotations{
		Title:          "Search TDS React Native Documents",
		ReadOnlyHint:   true,
		IdempotentHint: true,
	},
}

func (p *Protocol) searchTdsRnDocsHandler(ctx context.Context, r *mcp.CallToolRequest, input SearchInput) (result *mcp.CallToolResult, output SearchOutput, err error) {
	searcher, err := search.NewTDSSearcher()
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
