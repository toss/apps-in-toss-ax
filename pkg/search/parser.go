package search

import (
	"regexp"
	"strings"
)

// LlmsFullDocument는 llms-full.txt의 개별 문서를 나타냅니다
type LlmsFullDocument struct {
	URL     string
	Title   string
	Content string
}

// ParseLlmsFull은 llms-full.txt 형식을 파싱합니다
// 형식:
// ---
// url: >-
//
//	https://example.com/doc.md
//
// ---
// # 제목
// (내용...)
func ParseLlmsFull(content string) []LlmsFullDocument {
	var documents []LlmsFullDocument

	// 문서 구분자로 분리 (--- 로 시작하는 YAML frontmatter)
	// 각 문서는 ---\nurl: 로 시작
	docPattern := regexp.MustCompile(`(?m)^---\s*\nurl:`)
	indices := docPattern.FindAllStringIndex(content, -1)

	if len(indices) == 0 {
		return documents
	}

	for i, idx := range indices {
		var docContent string
		if i < len(indices)-1 {
			docContent = content[idx[0]:indices[i+1][0]]
		} else {
			docContent = content[idx[0]:]
		}

		doc := parseDocument(docContent)
		if doc.URL != "" && doc.Title != "" {
			documents = append(documents, doc)
		}
	}

	return documents
}

func parseDocument(content string) LlmsFullDocument {
	doc := LlmsFullDocument{}

	// YAML frontmatter 끝 찾기
	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		return doc
	}

	frontmatter := parts[1]
	body := parts[2]

	// URL 추출
	doc.URL = extractURL(frontmatter)

	// 제목과 내용 추출
	doc.Title, doc.Content = extractTitleAndContent(body)

	return doc
}

func extractURL(frontmatter string) string {
	lines := strings.Split(frontmatter, "\n")

	for i, line := range lines {
		line = strings.TrimSpace(line)

		// url: https://... 형식
		if strings.HasPrefix(line, "url:") {
			urlPart := strings.TrimPrefix(line, "url:")
			urlPart = strings.TrimSpace(urlPart)

			// url: >- 형식 (다음 줄에 URL)
			if urlPart == ">-" || urlPart == ">" || urlPart == "|" {
				if i+1 < len(lines) {
					return strings.TrimSpace(lines[i+1])
				}
			}

			// url: "https://..." 또는 url: https://...
			urlPart = strings.Trim(urlPart, "\"'")
			return urlPart
		}
	}

	return ""
}

func extractTitleAndContent(body string) (title, content string) {
	body = strings.TrimSpace(body)
	lines := strings.Split(body, "\n")

	titleFound := false
	var contentLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// 제목 찾기 (# 로 시작)
		if !titleFound && strings.HasPrefix(trimmed, "# ") {
			title = strings.TrimPrefix(trimmed, "# ")
			title = strings.TrimSpace(title)
			titleFound = true
			continue
		}

		if titleFound {
			contentLines = append(contentLines, line)
		}
	}

	content = strings.TrimSpace(strings.Join(contentLines, "\n"))
	return title, content
}

// TdsLlmsFullDocument는 TDS llms-full.txt의 개별 문서를 나타냅니다
type TdsLlmsFullDocument struct {
	URL     string
	Title   string
	Content string
}

// ParseTdsLlmsFull은 TDS llms-full.txt 형식을 파싱합니다
// 형식:
// # Title (/path/)
// (내용...)
// ---
// # Title2 (/path2/)
// (내용...)
func ParseTdsLlmsFull(content string) []TdsLlmsFullDocument {
	var documents []TdsLlmsFullDocument

	// BOM 제거
	content = strings.TrimPrefix(content, "\ufeff")

	// --- 로 문서 분리
	parts := strings.Split(content, "\n---\n")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		doc := parseTdsDocument(part)
		if doc.URL != "" && doc.Title != "" {
			documents = append(documents, doc)
		}
	}

	return documents
}

// parseTdsDocument는 개별 TDS 문서를 파싱합니다
// 첫 줄: # Title (/path/)
// 나머지: 내용
func parseTdsDocument(content string) TdsLlmsFullDocument {
	doc := TdsLlmsFullDocument{}

	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return doc
	}

	// 첫 줄에서 제목과 경로 추출
	firstLine := strings.TrimSpace(lines[0])
	if !strings.HasPrefix(firstLine, "# ") {
		return doc
	}

	// "# Title (/path/)" 형식 파싱
	titleLine := strings.TrimPrefix(firstLine, "# ")

	// 괄호 안의 경로 찾기
	pathStart := strings.LastIndex(titleLine, "(")
	pathEnd := strings.LastIndex(titleLine, ")")

	if pathStart == -1 || pathEnd == -1 || pathStart >= pathEnd {
		return doc
	}

	doc.Title = strings.TrimSpace(titleLine[:pathStart])
	path := strings.TrimSpace(titleLine[pathStart+1 : pathEnd])

	// 상대 경로에 base URL 추가
	if strings.HasPrefix(path, "/") {
		doc.URL = tdsBaseURL + path
	} else {
		doc.URL = path
	}

	// 나머지 내용 추출
	if len(lines) > 1 {
		contentLines := lines[1:]
		doc.Content = strings.TrimSpace(strings.Join(contentLines, "\n"))
	}

	return doc
}
