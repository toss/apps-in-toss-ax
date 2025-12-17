package mcp

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/toss/apps-in-toss-ax/pkg/docs"
)

var getExample = &mcp.Tool{
	Name:        "get_example",
	Title:       "Get AppsInToss Example",
	Description: "Get a AppsInToss Example. you must call list.examples first to get the examples list. and then you choose the you want to get.",
	Annotations: &mcp.ToolAnnotations{
		Title:          "Get AppsInToss Example",
		ReadOnlyHint:   true,
		IdempotentHint: true,
	},
}

func (p *Protocol) getExampleHandler(ctx context.Context, r *mcp.CallToolRequest, input GetExampleInput) (result *mcp.CallToolResult, output GetExampleOutput, err error) {
	docsService := docs.New()
	content, err := docsService.GetExample(ctx, input.ExampleID)
	if err != nil {
		return nil, GetExampleOutput{}, err
	}

	return nil, GetExampleOutput{Content: content}, nil
}

type GetExampleInput struct {
	ExampleID string `json:"example_id"`
}

type GetExampleOutput struct {
	Content string `json:"content"`
}
