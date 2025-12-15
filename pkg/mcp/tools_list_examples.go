package mcp

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/toss/apps-in-toss-ax/pkg/docs"
)

var listExamples = &mcp.Tool{
	Name:        "list.examples",
	Title:       "List AppsInToss Examples",
	Description: "List AppsInToss Examples. you can get the AppsInToss examples projects list.",
	Annotations: &mcp.ToolAnnotations{
		Title:          "List AppsInToss Examples",
		ReadOnlyHint:   true,
		IdempotentHint: true,
	},
}

func (p *Protocol) listExamplesHandler(ctx context.Context, r *mcp.CallToolRequest, input ListExamplesInput) (result *mcp.CallToolResult, output ListExamplesOutput, err error) {
	docsService := docs.New()
	examples, err := docsService.GetExamples(ctx)
	if err != nil {
		return nil, ListExamplesOutput{}, err
	}

	return nil, ListExamplesOutput{Examples: examples}, nil
}

type ListExamplesInput struct {
}

type ListExamplesOutput struct {
	Examples []docs.LlmDocument `json:"examples"`
}
