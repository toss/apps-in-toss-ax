package mcp

import _ "embed"

//go:embed instructions.md
var instructionsBody string

func instructions() string {
	return instructionsBody
}
