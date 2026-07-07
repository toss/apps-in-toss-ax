package mcp

import "github.com/toss/apps-in-toss-ax/pkg/search"

// SearchInput은 모든 검색 도구의 공통 입력 타입입니다
type SearchInput struct {
	Query string `json:"query"`
	Limit int    `json:"limit,omitempty"`

	// 필드별 부스트 재정의. 생략하면 기본값이 적용되며,
	// 검색 결과가 만족스럽지 않을 때 호출자(LLM)가 직접 조정할 수 있습니다.
	TitleBoost       *float64 `json:"title_boost,omitempty" jsonschema:"Relevance boost for title matches (default 5.0, valid range 0 to 1000000; at least one of the four boosts must stay > 0). Raise it when the query names a specific document or component; lower it to surface documents that only mention the term in the body."`
	DescriptionBoost *float64 `json:"description_boost,omitempty" jsonschema:"Relevance boost for description matches (default 1.5, valid range 0 to 1000000; at least one of the four boosts must stay > 0)."`
	ContentBoost     *float64 `json:"content_boost,omitempty" jsonschema:"Relevance boost for body content matches (default 1.0, valid range 0 to 1000000; at least one of the four boosts must stay > 0). Raise it when searching for error messages or code identifiers that appear in document bodies rather than titles."`
	CategoryBoost    *float64 `json:"category_boost,omitempty" jsonschema:"Relevance boost for category matches (default 1.0, valid range 0 to 1000000; at least one of the four boosts must stay > 0)."`
}

// searchOptions는 SearchInput을 search.SearchOptions로 변환합니다
func (in SearchInput) searchOptions() *search.SearchOptions {
	limit := in.Limit
	if limit <= 0 {
		limit = 10
	}
	return &search.SearchOptions{
		Limit: limit,
		Boosts: search.BoostOverrides{
			Title:       in.TitleBoost,
			Description: in.DescriptionBoost,
			Content:     in.ContentBoost,
			Category:    in.CategoryBoost,
		},
	}
}

// SearchOutput은 모든 검색 도구의 공통 출력 타입입니다
type SearchOutput struct {
	Results []search.SearchResult `json:"results"`
	Total   int                   `json:"total"`
}

// GetDocInput은 문서 조회 도구의 입력 타입입니다
type GetDocInput struct {
	ID string `json:"id"`
}

// GetDocOutput은 문서 조회 도구의 출력 타입입니다
type GetDocOutput struct {
	Document *search.SearchResult `json:"document,omitempty"`
}
