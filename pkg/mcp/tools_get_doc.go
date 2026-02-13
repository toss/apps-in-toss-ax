package mcp

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var getDoc = &mcp.Tool{
	Name:        "get_doc",
	Title:       "Get AppsInToss Document",
	Description: "Retrieve the full content of an AppsInToss document by its ID. Use this after search_docs to get the complete document content.",
	Annotations: &mcp.ToolAnnotations{
		Title:          "Get AppsInToss Document",
		ReadOnlyHint:   true,
		IdempotentHint: true,
	},
}

var getTdsRnDoc = &mcp.Tool{
	Name:        "get_tds_rn_doc",
	Title:       "Get TDS React Native Document",
	Description: "Retrieve the full content of a TDS React Native document by its ID. Use this after search_tds_rn_docs to get the complete document content.",
	Annotations: &mcp.ToolAnnotations{
		Title:          "Get TDS React Native Document",
		ReadOnlyHint:   true,
		IdempotentHint: true,
	},
}

var getTdsWebDoc = &mcp.Tool{
	Name:        "get_tds_web_doc",
	Title:       "Get TDS Web Document",
	Description: "Retrieve the full content of a TDS Web document by its ID. Use this after search_tds_web_docs to get the complete document content.",
	Annotations: &mcp.ToolAnnotations{
		Title:          "Get TDS Web Document",
		ReadOnlyHint:   true,
		IdempotentHint: true,
	},
}

func (p *Protocol) getDocHandler(ctx context.Context, r *mcp.CallToolRequest, input GetDocInput) (result *mcp.CallToolResult, output GetDocOutput, err error) {
	return p.getDocFromSearcher(ctx, p.docSearcher, input.ID)
}

func (p *Protocol) getTdsRnDocHandler(ctx context.Context, r *mcp.CallToolRequest, input GetDocInput) (result *mcp.CallToolResult, output GetDocOutput, err error) {
	return p.getDocFromSearcher(ctx, p.tdsRn, input.ID)
}

func (p *Protocol) getTdsWebDocHandler(ctx context.Context, r *mcp.CallToolRequest, input GetDocInput) (result *mcp.CallToolResult, output GetDocOutput, err error) {
	return p.getDocFromSearcher(ctx, p.tdsWeb, input.ID)
}

func (p *Protocol) getDocFromSearcher(ctx context.Context, ls *lazySearcher, id string) (*mcp.CallToolResult, GetDocOutput, error) {
	searcher, err := ls.get(ctx)
	if err != nil {
		return nil, GetDocOutput{}, err
	}

	doc, err := searcher.GetDocument(ctx, id)
	if err != nil {
		return nil, GetDocOutput{}, err
	}
	if doc == nil {
		return nil, GetDocOutput{}, fmt.Errorf("document not found: %s", id)
	}

	return nil, GetDocOutput{Document: doc}, nil
}
