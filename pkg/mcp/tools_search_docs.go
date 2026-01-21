package mcp

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/toss/apps-in-toss-ax/pkg/search"
)

var searchDocs = &mcp.Tool{
	Name:        "search_document",
	Title:       "Search AppsInToss Documents",
	Description: "Search AppsInToss documentation using full-text search. Returns matching documents ranked by relevance.",
	Annotations: &mcp.ToolAnnotations{
		Title:          "Search AppsInToss Documents",
		ReadOnlyHint:   true,
		IdempotentHint: true,
	},
}

func (p *Protocol) searchDocsHandler(ctx context.Context, r *mcp.CallToolRequest, input SearchDocsInput) (result *mcp.CallToolResult, output SearchDocsOutput, err error) {
	searcher, err := search.New()
	if err != nil {
		return nil, SearchDocsOutput{}, err
	}
	defer searcher.Close()

	if err := searcher.EnsureIndex(ctx); err != nil {
		return nil, SearchDocsOutput{}, err
	}

	limit := input.Limit
	if limit <= 0 {
		limit = 10
	}

	results, err := searcher.Search(ctx, input.Query, &search.SearchOptions{
		Limit: limit,
	})
	if err != nil {
		return nil, SearchDocsOutput{}, err
	}

	return nil, SearchDocsOutput{
		Results: results,
		Total:   len(results),
	}, nil
}

type SearchDocsInput struct {
	Query string `json:"query"`
	Limit int    `json:"limit,omitempty"`
}

type SearchDocsOutput struct {
	Results []search.SearchResult `json:"results"`
	Total   int                   `json:"total"`
}
