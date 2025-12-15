package llms

import (
	"testing"
)

func TestParser_Parse(t *testing.T) {
	content := `# 앱인토스 개발자센터

> 앱인토스 개발자센터

## Table of Contents

### 시작

- [앱인토스 Unity 적용 가이드](https://developers-apps-in-toss.toss.im/unity/intro/overview.md)
- [포팅 순서](https://developers-apps-in-toss.toss.im/unity/intro/migration-guide.md)
- [Unity 포팅](https://developers-apps-in-toss.toss.im/unity/porting_tutorials/unity.md): Unity 게임을 앱인토스 미니앱으로 포팅하는 가이드입니다.

### 준비

- [동작 방식](https://developers-apps-in-toss.toss.im/unity/guide/runtime-structure.md)
- [전환 점검](https://developers-apps-in-toss.toss.im/unity/guide/precheck.md)

## 디자인

- [토스 디자인 시스템 (TDS)](https://developers-apps-in-toss.toss.im/design/components.md): 앱인토스 미니앱 개발을 위한 토스 디자인 시스템(TDS) 가이드입니다.
`

	parser := NewParser()
	result, err := parser.Parse(content)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Test title
	if result.Title != "앱인토스 개발자센터" {
		t.Errorf("Expected title '앱인토스 개발자센터', got '%s'", result.Title)
	}

	// Test summary
	if result.Summary != "앱인토스 개발자센터" {
		t.Errorf("Expected summary '앱인토스 개발자센터', got '%s'", result.Summary)
	}

	// Test sections count
	if len(result.Sections) != 2 {
		t.Errorf("Expected 2 top-level sections, got %d", len(result.Sections))
	}

	// Test first section (Table of Contents)
	if result.Sections[0].Title != "Table of Contents" {
		t.Errorf("Expected first section 'Table of Contents', got '%s'", result.Sections[0].Title)
	}

	// Test children of Table of Contents
	if len(result.Sections[0].Children) != 2 {
		t.Errorf("Expected 2 children in Table of Contents, got %d", len(result.Sections[0].Children))
	}

	// Test 시작 section
	startSection := result.Sections[0].Children[0]
	if startSection.Title != "시작" {
		t.Errorf("Expected child section '시작', got '%s'", startSection.Title)
	}
	if len(startSection.Links) != 3 {
		t.Errorf("Expected 3 links in 시작 section, got %d", len(startSection.Links))
	}

	// Test link with description
	unityLink := startSection.Links[2]
	if unityLink.Title != "Unity 포팅" {
		t.Errorf("Expected link title 'Unity 포팅', got '%s'", unityLink.Title)
	}
	if unityLink.Description != "Unity 게임을 앱인토스 미니앱으로 포팅하는 가이드입니다." {
		t.Errorf("Expected description, got '%s'", unityLink.Description)
	}

	// Test second top-level section (디자인)
	if result.Sections[1].Title != "디자인" {
		t.Errorf("Expected second section '디자인', got '%s'", result.Sections[1].Title)
	}
	if len(result.Sections[1].Links) != 1 {
		t.Errorf("Expected 1 link in 디자인 section, got %d", len(result.Sections[1].Links))
	}
}

func TestParser_ParseLinks(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		wantTitle string
		wantURL   string
		wantDesc  string
	}{
		{
			name: "Link with description",
			content: `# Test
## Section
- [Title](https://example.com): Description text
`,
			wantTitle: "Title",
			wantURL:   "https://example.com",
			wantDesc:  "Description text",
		},
		{
			name: "Link without description",
			content: `# Test
## Section
- [Title](https://example.com)
`,
			wantTitle: "Title",
			wantURL:   "https://example.com",
			wantDesc:  "",
		},
		{
			name: "Link with Korean text",
			content: `# Test
## Section
- [한국어 제목](https://example.com/path.md): 한국어 설명입니다.
`,
			wantTitle: "한국어 제목",
			wantURL:   "https://example.com/path.md",
			wantDesc:  "한국어 설명입니다.",
		},
	}

	parser := NewParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(tt.content)
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			if len(result.Sections) == 0 {
				t.Fatal("Expected at least one section")
			}

			section := result.Sections[0]
			if len(section.Links) == 0 {
				t.Fatal("Expected at least one link")
			}

			link := section.Links[0]
			if link.Title != tt.wantTitle {
				t.Errorf("Expected title '%s', got '%s'", tt.wantTitle, link.Title)
			}
			if link.URL != tt.wantURL {
				t.Errorf("Expected URL '%s', got '%s'", tt.wantURL, link.URL)
			}
			if link.Description != tt.wantDesc {
				t.Errorf("Expected description '%s', got '%s'", tt.wantDesc, link.Description)
			}
		})
	}
}

func TestLlmsTxt_GetAllLinks(t *testing.T) {
	content := `# Test

> Summary

## Section 1

- [Link 1](https://example.com/1)
- [Link 2](https://example.com/2)

### Subsection 1.1

- [Link 3](https://example.com/3)

## Section 2

- [Link 4](https://example.com/4)
`

	parser := NewParser()
	result, _ := parser.Parse(content)

	links := result.GetAllLinks()
	if len(links) != 4 {
		t.Errorf("Expected 4 links, got %d", len(links))
	}
}

func TestLlmsTxt_FindSection(t *testing.T) {
	content := `# Test

> Summary

## Section 1

### Subsection 1.1

## Section 2
`

	parser := NewParser()
	result, _ := parser.Parse(content)

	// Test finding top-level section
	section := result.FindSection("Section 1")
	if section == nil {
		t.Fatal("Expected to find 'Section 1'")
	}
	if section.Title != "Section 1" {
		t.Errorf("Expected 'Section 1', got '%s'", section.Title)
	}

	// Test finding nested section
	subsection := result.FindSection("Subsection 1.1")
	if subsection == nil {
		t.Fatal("Expected to find 'Subsection 1.1'")
	}
	if subsection.Title != "Subsection 1.1" {
		t.Errorf("Expected 'Subsection 1.1', got '%s'", subsection.Title)
	}

	// Test case-insensitive search
	section = result.FindSection("section 2")
	if section == nil {
		t.Fatal("Expected to find 'section 2' (case-insensitive)")
	}

	// Test not found
	section = result.FindSection("Non-existent")
	if section != nil {
		t.Error("Expected nil for non-existent section")
	}
}

func TestLlmsTxt_GetSectionTitles(t *testing.T) {
	content := `# Test

> Summary

## Section 1

### Subsection 1.1

## Section 2
`

	parser := NewParser()
	result, _ := parser.Parse(content)

	titles := result.GetSectionTitles()
	expected := []string{"Section 1", "Subsection 1.1", "Section 2"}

	if len(titles) != len(expected) {
		t.Errorf("Expected %d titles, got %d", len(expected), len(titles))
	}

	for i, title := range expected {
		if titles[i] != title {
			t.Errorf("Expected title '%s' at index %d, got '%s'", title, i, titles[i])
		}
	}
}
