package mcp

import (
	"context"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// CompletionRef identifies the target of a completion (prompt or resource).
type CompletionRef struct {
	Type string // "ref/prompt" or "ref/resource"
	Name string // prompt name (for ref/prompt)
	URI  string // resource URI (for ref/resource)
}

// Completion defines allowed autocompletion values for a prompt or resource argument.
type Completion struct {
	Ref    CompletionRef
	Arg    string
	Values []string
}

// key returns the registry lookup key for this ref.
func (ref CompletionRef) key() string {
	if ref.Type == "ref/resource" {
		return ref.Type + ":" + ref.URI
	}
	return ref.Type + ":" + ref.Name
}

// PromptRef creates a CompletionRef for a prompt.
func PromptRef(name string) CompletionRef {
	return CompletionRef{Type: "ref/prompt", Name: name}
}

// ResourceRef creates a CompletionRef for a resource.
func ResourceRef(uri string) CompletionRef {
	return CompletionRef{Type: "ref/resource", URI: uri}
}

// CompletionRegistry manages autocompletion values for prompt and resource arguments.
type CompletionRegistry struct {
	// entries maps ref key -> argName -> allowed values
	entries map[string]map[string][]string
}

// NewCompletionRegistry creates an empty CompletionRegistry.
func NewCompletionRegistry() *CompletionRegistry {
	return &CompletionRegistry{
		entries: make(map[string]map[string][]string),
	}
}

// RegisterAll adds multiple completions at once.
func (r *CompletionRegistry) RegisterAll(completions []Completion) {
	for _, c := range completions {
		r.Register(c)
	}
}

// Register adds allowed completion values for an argument.
func (r *CompletionRegistry) Register(c Completion) {
	key := c.Ref.key()
	if r.entries[key] == nil {
		r.entries[key] = make(map[string][]string)
	}
	r.entries[key][c.Arg] = c.Values
}

// Complete returns values matching the given prefix for an argument.
func (r *CompletionRegistry) Complete(ref CompletionRef, arg, prefix string) []string {
	args, ok := r.entries[ref.key()]
	if !ok {
		return nil
	}
	values, ok := args[arg]
	if !ok {
		return nil
	}

	var matched []string
	for _, v := range values {
		if strings.HasPrefix(v, prefix) {
			matched = append(matched, v)
		}
	}
	return matched
}

// Handler is an MCP CompletionHandler that resolves argument completions.
func (r *CompletionRegistry) Handler(_ context.Context, req *mcp.CompleteRequest) (*mcp.CompleteResult, error) {
	if req.Params.Ref == nil {
		return &mcp.CompleteResult{}, nil
	}

	ref := CompletionRef{
		Type: req.Params.Ref.Type,
		Name: req.Params.Ref.Name,
		URI:  req.Params.Ref.URI,
	}
	matched := r.Complete(ref, req.Params.Argument.Name, req.Params.Argument.Value)

	return &mcp.CompleteResult{
		Completion: mcp.CompletionResultDetails{
			Values: matched,
			Total:  len(matched),
		},
	}, nil
}
