package main_test

import (
	"bytes"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBrewFormulaTemplate(t *testing.T) {
	// Read template file
	templatePath := filepath.Join("templates", "ax.rb")
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		t.Fatalf("failed to read template file: %v", err)
	}

	// Parse template
	tmpl, err := template.New("test").Parse(string(templateContent))
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	// Mock
	config := struct {
		Version     string
		Checksum    map[string]string
		Description string
		FormulaName string
	}{
		Version:     "1.0.0",
		Description: "Test CLI",
		FormulaName: "TestFormula",
		Checksum: map[string]string{
			"ax_darwin_arm64": "def456",
			"ax_darwin_amd64": "ghi789",
			"ax_linux_arm64":  "jkl012",
			"ax_linux_amd64":  "mno345",
		},
	}

	// Render template
	var buf bytes.Buffer
	if err := tmpl.Option("missingkey=error").Execute(&buf, config); err != nil {
		t.Fatalf("failed to render template: %v", err)
	}

	result := buf.String()

	// Verify
	if !strings.Contains(result, config.Version) {
		t.Errorf("result does not contain version: %s", config.Version)
	}
	if !strings.Contains(result, config.Description) {
		t.Errorf("result does not contain description: %s", config.Description)
	}
	if !strings.Contains(result, config.FormulaName) {
		t.Errorf("result does not contain formula name: %s", config.FormulaName)
	}

	if !strings.Contains(result, config.Checksum["ax_darwin_amd64"]) {
		t.Errorf("result does not contain ax_darwin_amd64 checksum: %s", config.Checksum["ax_darwin_amd64"])
	}
	if !strings.Contains(result, config.Checksum["ax_linux_arm64"]) {
		t.Errorf("result does not contain ax_linux_arm64 checksum: %s", config.Checksum["ax_linux_arm64"])
	}
	if !strings.Contains(result, config.Checksum["ax_linux_amd64"]) {
		t.Errorf("result does not contain ax_linux_amd64 checksum: %s", config.Checksum["ax_linux_amd64"])
	}
}

func TestScoopBucketTemplate(t *testing.T) {
	templatePath := filepath.Join("templates", "ax.json")
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		t.Fatalf("failed to read template file: %v", err)
	}

	tmpl, err := template.New("test").Parse(string(templateContent))
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	// Mock
	config := struct {
		Version     string
		Checksum    map[string]string
		Description string
		FormulaName string
	}{
		Version:     "1.0.0",
		Description: "Test CLI",
		FormulaName: "TestFormula",
		Checksum: map[string]string{
			"ax_windows_amd64": "abc123",
			"ax_windows_arm64": "def456",
		},
	}

	var buf bytes.Buffer
	if err := tmpl.Option("missingkey=error").Execute(&buf, config); err != nil {
		t.Fatalf("failed to render template: %v", err)
	}

	result := buf.String()

	// Verify
	if !strings.Contains(result, config.Version) {
		t.Errorf("result does not contain version: %s", config.Version)
	}
	if !strings.Contains(result, config.Description) {
		t.Errorf("result does not contain description: %s", config.Description)
	}
	if !strings.Contains(result, config.Checksum["ax_windows_amd64"]) {
		t.Errorf("result does not contain Windows AMD64 checksum: %s", config.Checksum["ax_windows_amd64"])
	}
	if !strings.Contains(result, config.Checksum["ax_windows_arm64"]) {
		t.Errorf("result does not contain Windows ARM64 checksum: %s", config.Checksum["ax_windows_arm64"])
	}
}

func TestTemplateRenderingWithMissingKey(t *testing.T) {
	templatePath := filepath.Join("templates", "ax.json")
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		t.Fatalf("failed to read template file: %v", err)
	}

	tmpl, err := template.New("test").Parse(string(templateContent))
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	// Mock
	config := struct {
		Version     string
		Description string
		FormulaName string
	}{
		Version:     "1.0.0",
		Description: "Test CLI",
		FormulaName: "TestFormula",
	}

	var buf bytes.Buffer
	err = tmpl.Option("missingkey=error").Execute(&buf, config)
	if err == nil {
		t.Error("expected error for missing key")
	}
}
