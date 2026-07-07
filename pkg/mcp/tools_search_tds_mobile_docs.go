package mcp

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var searchTdsWebDocs = &mcp.Tool{
	Name:        "search_tds_web_docs",
	Title:       "Search TDS Web Documents",
	Description: "Search TDS (Toss Design System) Web documentation using full-text search. Returns matching documents ranked by relevance. Per-field relevance weights can be tuned via the optional *_boost parameters (defaults: title=5.0, description=1.5, content=1.0, category=1.0).",
	Annotations: &mcp.ToolAnnotations{
		Title:          "Search TDS Web Documents",
		ReadOnlyHint:   true,
		IdempotentHint: true,
	},
}

func (p *Protocol) searchTdsWebDocsHandler(ctx context.Context, r *mcp.CallToolRequest, input SearchInput) (result *mcp.CallToolResult, output SearchOutput, err error) {
	searcher, err := p.tdsWeb.get(ctx)
	if err != nil {
		return nil, SearchOutput{}, err
	}

	results, err := searcher.Search(ctx, input.Query, input.searchOptions())
	if err != nil {
		return nil, SearchOutput{}, err
	}

	return nil, SearchOutput{
		Results: results,
		Total:   len(results),
	}, nil
}
