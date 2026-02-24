package mcp

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func newTestRegistry() *CompletionRegistry {
	r := NewCompletionRegistry()
	r.RegisterAll([]Completion{
		{Ref: PromptRef("my-prompt"), Arg: "platform", Values: []string{"react-native", "web"}},
		{Ref: PromptRef("my-prompt"), Arg: "lang", Values: []string{"ko", "en", "ja"}},
		{Ref: ResourceRef("file:///{path}"), Arg: "path", Values: []string{"src/", "pkg/", "cmd/"}},
	})
	return r
}

func TestComplete_PromptRef_ExactMatch(t *testing.T) {
	r := newTestRegistry()

	got := r.Complete(PromptRef("my-prompt"), "platform", "web")
	if len(got) != 1 || got[0] != "web" {
		t.Errorf("expected [web], got %v", got)
	}
}

func TestComplete_PromptRef_PrefixMatch(t *testing.T) {
	r := newTestRegistry()

	got := r.Complete(PromptRef("my-prompt"), "platform", "re")
	if len(got) != 1 || got[0] != "react-native" {
		t.Errorf("expected [react-native], got %v", got)
	}
}

func TestComplete_PromptRef_EmptyPrefix(t *testing.T) {
	r := newTestRegistry()

	got := r.Complete(PromptRef("my-prompt"), "lang", "")
	if len(got) != 3 {
		t.Errorf("expected 3 values, got %v", got)
	}
}

func TestComplete_PromptRef_NoMatch(t *testing.T) {
	r := newTestRegistry()

	got := r.Complete(PromptRef("my-prompt"), "platform", "python")
	if len(got) != 0 {
		t.Errorf("expected no matches, got %v", got)
	}
}

func TestComplete_UnknownPrompt(t *testing.T) {
	r := newTestRegistry()

	got := r.Complete(PromptRef("unknown"), "platform", "")
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestComplete_UnknownArg(t *testing.T) {
	r := newTestRegistry()

	got := r.Complete(PromptRef("my-prompt"), "unknown", "")
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestComplete_ResourceRef(t *testing.T) {
	r := newTestRegistry()

	got := r.Complete(ResourceRef("file:///{path}"), "path", "s")
	if len(got) != 1 || got[0] != "src/" {
		t.Errorf("expected [src/], got %v", got)
	}
}

func TestComplete_ResourceRef_DoesNotMatchPrompt(t *testing.T) {
	r := newTestRegistry()

	// PromptRef와 ResourceRef는 같은 이름이라도 별개
	got := r.Complete(ResourceRef("my-prompt"), "platform", "")
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestHandler_PromptRef(t *testing.T) {
	r := newTestRegistry()

	result, err := r.Handler(context.Background(), &mcp.CompleteRequest{
		Params: &mcp.CompleteParams{
			Ref: &mcp.CompleteReference{
				Type: "ref/prompt",
				Name: "my-prompt",
			},
			Argument: mcp.CompleteParamsArgument{
				Name:  "platform",
				Value: "r",
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Completion.Values) != 1 || result.Completion.Values[0] != "react-native" {
		t.Errorf("expected [react-native], got %v", result.Completion.Values)
	}
	if result.Completion.Total != 1 {
		t.Errorf("expected total 1, got %d", result.Completion.Total)
	}
}

func TestHandler_ResourceRef(t *testing.T) {
	r := newTestRegistry()

	result, err := r.Handler(context.Background(), &mcp.CompleteRequest{
		Params: &mcp.CompleteParams{
			Ref: &mcp.CompleteReference{
				Type: "ref/resource",
				URI:  "file:///{path}",
			},
			Argument: mcp.CompleteParamsArgument{
				Name:  "path",
				Value: "p",
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Completion.Values) != 1 || result.Completion.Values[0] != "pkg/" {
		t.Errorf("expected [pkg/], got %v", result.Completion.Values)
	}
}

func TestHandler_NilRef(t *testing.T) {
	r := newTestRegistry()

	result, err := r.Handler(context.Background(), &mcp.CompleteRequest{
		Params: &mcp.CompleteParams{
			Ref: nil,
			Argument: mcp.CompleteParamsArgument{
				Name:  "platform",
				Value: "",
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Completion.Values) != 0 {
		t.Errorf("expected empty values, got %v", result.Completion.Values)
	}
}

func TestRegister_OverwritesExisting(t *testing.T) {
	r := NewCompletionRegistry()

	r.Register(Completion{Ref: PromptRef("p"), Arg: "a", Values: []string{"old"}})
	r.Register(Completion{Ref: PromptRef("p"), Arg: "a", Values: []string{"new"}})

	got := r.Complete(PromptRef("p"), "a", "")
	if len(got) != 1 || got[0] != "new" {
		t.Errorf("expected [new], got %v", got)
	}
}
