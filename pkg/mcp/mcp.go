package mcp

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	name    = "ax"
	title   = "ax"
	version = "0.1.0"
)

type Protocol struct {
	OnInit    func(context.Context)
	Transport mcp.Transport
	Server    *mcp.Server
}

type Option func(*Protocol)

func WithTransport(transport mcp.Transport) Option {
	return func(s *Protocol) {
		s.Transport = transport
	}
}

func New(options ...Option) *Protocol {
	p := &Protocol{
		Transport: &mcp.StdioTransport{},
		OnInit:    func(_ context.Context) {},
	}

	for _, o := range options {
		o(p)
	}

	i := mcp.NewServer(
		&mcp.Implementation{
			Name:    name,
			Title:   title,
			Version: version,
		},
		&mcp.ServerOptions{
			Instructions: instructions(),
			HasPrompts:   true,
			HasResources: true,
			HasTools:     true,
		})

	mcp.AddTool(i, listDocs, p.listDocsHandler)
	mcp.AddTool(i, getDocs, p.getDocsHandler)
	mcp.AddTool(i, listExamples, p.listExamplesHandler)
	mcp.AddTool(i, getExample, p.getExampleHandler)
	mcp.AddTool(i, searchDocs, p.searchDocsHandler)

	p.Server = i
	return p
}
