package mcp

import "github.com/toss/apps-in-toss-ax/pkg/search"

// SearchInput은 모든 검색 도구의 공통 입력 타입입니다
type SearchInput struct {
	Query string `json:"query"`
	Limit int    `json:"limit,omitempty"`
}

// SearchOutput은 모든 검색 도구의 공통 출력 타입입니다
type SearchOutput struct {
	Results []search.SearchResult `json:"results"`
	Total   int                   `json:"total"`
}
