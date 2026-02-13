package mcp

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/toss/apps-in-toss-ax/pkg/search"
)

var searchDocs = &mcp.Tool{
	Name:        "search_docs",
	Title:       "Search AppsInToss Documents",
	Description: "Search AppsInToss documentation using full-text search. Returns matching documents ranked by relevance.",
	Annotations: &mcp.ToolAnnotations{
		Title:          "Search AppsInToss Documents",
		ReadOnlyHint:   true,
		IdempotentHint: true,
	},
}

func (p *Protocol) searchDocsHandler(ctx context.Context, r *mcp.CallToolRequest, input SearchInput) (result *mcp.CallToolResult, output SearchOutput, err error) {
	searcher, err := p.docSearcher.get(ctx)
	if err != nil {
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
