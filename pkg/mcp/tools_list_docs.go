package mcp

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/toss/apps-in-toss-ax/pkg/docs"
)

var listDocs = &mcp.Tool{
	Name:        "list.docs",
	Title:       "List AppsInToss Docs",
	Description: "List AppsInToss Documents",
	Annotations: &mcp.ToolAnnotations{
		Title:          "List AppsInToss Docs",
		ReadOnlyHint:   true,
		IdempotentHint: true,
	},
}

func (p *Protocol) listDocsHandler(ctx context.Context, r *mcp.CallToolRequest, input ListDocsInput) (result *mcp.CallToolResult, output ListDocsOutput, err error) {
	docsService := docs.New()
	docs, err := docsService.GetLLMsRoot(ctx)
	if err != nil {
		return nil, ListDocsOutput{}, err
	}

	return nil, ListDocsOutput{
		Documents: docs,
	}, nil
}

type ListDocsInput struct{}

type ListDocsOutput struct {
	Documents []docs.LlmDocument `json:"documents"`
}
