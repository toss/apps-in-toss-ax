# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is the **ax** CLI tool, an MCP (Model Context Protocol) server that provides AppsInToss documentation and examples to AI assistants. It serves as a bridge between AI agents and AppsInToss Developer Center resources, enabling context-aware assistance for mini-app development on the Toss platform.

## Build and Development Commands

### Building
```bash
make build          # Builds the binary to ./ax
go build -o ax      # Alternative build command
```

### Testing
```bash
go test ./...                           # Run all tests
go test ./pkg/llms/parser_test.go       # Run specific test file
```

### Running the MCP Server
```bash
./ax mcp            # Start MCP server in stdio mode
```

### Release Process
The project uses GoReleaser for multi-platform releases:
- Releases are created via GitHub Actions (`.github/workflows/release.yml`)
- Binary distributions for darwin, linux, windows (amd64/arm64)
- Package managers: Homebrew (tap), Scoop (bucket), npm
- The `tools/publish/main.go` script updates package manager repositories after releases

## Architecture

### Entry Point Flow
1. `main.go` → `pkg/app/app.go:Run()` → `cmd/root.go:NewCommand()`
2. The root command initializes with `cmd/mcp.go:NewMcpCommand()`
3. MCP server starts via `pkg/mcp/mcp.go:New()`

### MCP Server Structure
The MCP server (`pkg/mcp/`) implements the Model Context Protocol SDK:
- **Server Registration**: `mcp.go` initializes the server with instructions, prompts, resources, and tools
- **Tool Handlers**: Each tool has two files:
  - `tools_list_*.go` - Tool definition and handler
  - Implements input/output types and handler functions
- **Available Tools**:
  - `list_docs` - Lists AppsInToss documentation
  - `get_docs` - Retrieves specific document by ID
  - `list.examples` - Lists code examples
  - `get.example` - Retrieves specific example by ID

### Documentation Fetching System
The `pkg/docs/` and `pkg/llms/` packages form the documentation retrieval layer:
- **docs.go**: Manages multiple documentation sources
  - AppsInToss Developer Center (llms.txt format)
  - TDS React Native documentation
  - TDS Web documentation
  - Code examples
- **llms/reader.go**: HTTP client for fetching documentation
- **llms/parser.go**: Parses `llms.txt` markdown format into structured data
  - Uses goldmark for markdown parsing
  - Extracts sections, links, and hierarchies
  - Generates stable document IDs via SHA-256 hashing

### Key Data Structures
- `LlmsTxt`: Root structure with title, summary, and sections
- `Section`: Hierarchical sections with levels, links, and children
- `Link`: Individual documentation links with title, URL, and description
- `LlmDocument`: Flattened document representation with generated ID

## Important Patterns

### Document ID Generation
Document IDs are deterministic SHA-256 hashes based on title, URL, and category (see `pkg/docs/docs.go:generateID`). This ensures stable references across fetches.

### LLMS.txt Format
The system parses a specialized markdown format:
- `#` Title
- `> ` Summary blockquote
- `##` Top-level sections
- `###` Nested sections
- `- [Link Title](url)` Documentation links
- `- [Link Title](url): Description` Links with descriptions

### Error Handling
The `pkg/fetcher/http.go` provides a centralized HTTP client with:
- Request editors for auth/headers
- Expected status code validation
- Detailed error responses with status codes

## AppsInToss Context

**AppsInToss** is Toss's mini-app platform for React Native and WebView apps. Key concepts:
- **Granite/Bedrock**: Framework for mini-app development (Granite is the current name for 1.0+)
- **TDS**: Toss Design System, required for non-game mini-apps
- **Unity Support**: WebGL-based Unity games can run as mini-apps

The MCP server's instructions (`pkg/mcp/instructions.md`) provide detailed guidance for AI assistants on terminology and tool usage.

## Dependencies
- `github.com/modelcontextprotocol/go-sdk` - MCP protocol implementation
- `github.com/spf13/cobra` - CLI framework
- `github.com/yuin/goldmark` - Markdown parsing
- `github.com/google/go-github` - GitHub API (for release publishing)

## Publishing Workflow
1. Tag a new version (e.g., `v0.1.1`)
2. GitHub Actions triggers GoReleaser
3. Binaries are built for all platforms
4. `tools/publish/main.go` runs to update package managers:
   - Downloads checksums from GitHub release
   - Renders templates (`tools/publish/templates/`)
   - Creates PRs to `toss/homebrew-tap` and `toss/scoop-bucket`
5. npm package downloads binary via `scripts/postinstall.js`
