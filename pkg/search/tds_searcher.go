package search

import (
	"strings"

	"github.com/toss/apps-in-toss-ax/pkg/docid"
)

const (
	tdsBaseURL = "https://tossmini-docs.toss.im"

	tdsReactNativeLlmsFullUrl = "https://tossmini-docs.toss.im/tds-react-native/llms-full.txt"
	tdsReactNativeLlmsUrl     = "https://tossmini-docs.toss.im/tds-react-native/llms.txt"
	tdsMobileLlmsFullUrl      = "https://tossmini-docs.toss.im/tds-mobile/llms-full.txt"
	tdsMobileLlmsUrl          = "https://tossmini-docs.toss.im/tds-mobile/llms.txt"
)

// tdsIndexer는 TDS llms-full.txt 내용을 IndexDocument로 변환합니다
func tdsIndexer(content string, categoryMap map[string]string) []IndexDocument {
	parsedDocs := ParseTdsLlmsFull(content)

	var documents []IndexDocument
	for _, doc := range parsedDocs {
		category := ""
		if categoryMap != nil {
			category = categoryMap[doc.URL]
		}
		if category == "" {
			category = extractTdsCategory(doc.URL)
		}

		documents = append(documents, IndexDocument{
			ID:       docid.Generate(doc.Title, doc.URL, category),
			Title:    doc.Title,
			Content:  doc.Content,
			URL:      doc.URL,
			Category: category,
		})
	}

	return documents
}

// tdsURLTransform은 TDS 문서의 상대 경로를 절대 경로로 변환합니다
func tdsURLTransform(url string) string {
	if strings.HasPrefix(url, "/") {
		return tdsBaseURL + url
	}
	return url
}

// extractTdsCategory는 URL에서 카테고리를 추출합니다
// 예: https://tossmini-docs.toss.im/tds-react-native/components/button/ -> "TDS React Native > Components"
func extractTdsCategory(url string) string {
	path := url
	if len(url) > len(tdsBaseURL) {
		path = url[len(tdsBaseURL):]
	}

	parts := splitPath(path)
	if len(parts) == 0 {
		return "TDS"
	}

	var categories []string
	for i, part := range parts {
		if i >= 2 {
			break
		}
		categories = append(categories, formatPathPart(part))
	}

	if len(categories) == 0 {
		return "TDS"
	}

	result := categories[0]
	for i := 1; i < len(categories); i++ {
		result += " > " + categories[i]
	}
	return result
}

func splitPath(path string) []string {
	var parts []string
	for _, part := range strings.Split(path, "/") {
		if part != "" {
			parts = append(parts, part)
		}
	}
	return parts
}

func formatPathPart(part string) string {
	words := strings.Split(part, "-")
	for i, word := range words {
		if len(word) > 0 && word[0] >= 'a' && word[0] <= 'z' {
			words[i] = strings.ToUpper(string(word[0])) + word[1:]
		}
	}
	return strings.Join(words, " ")
}
