package mcp

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/toss/apps-in-toss-ax/pkg/docs"
)

var getDocs = &mcp.Tool{
	Name:        "get_docs",
	Title:       "Get AppsInToss Docs",
	Description: "Get AppsInToss Document. you must call list_docs first to get the documents list. and then you choose the you want to get.",
	Annotations: &mcp.ToolAnnotations{
		Title:          "Get AppsInToss Docs",
		ReadOnlyHint:   true,
		IdempotentHint: true,
	},
}

func (p *Protocol) getDocsHandler(ctx context.Context, r *mcp.CallToolRequest, input GetDocsInput) (result *mcp.CallToolResult, output GetDocsOutput, err error) {
	docsService := docs.New()
	content, err := docsService.GetDocument(ctx, input.DocumentID)
	if err != nil {
		return nil, GetDocsOutput{}, err
	}

	return nil, GetDocsOutput{Content: content}, nil
}

type GetDocsInput struct {
	DocumentID string `json:"document_id"`
}

type GetDocsOutput struct {
	Content string `json:"content"`
}
