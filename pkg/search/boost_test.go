package search

import (
	"context"
	"math"
	"os"
	"testing"
)

func floatPtr(v float64) *float64 {
	return &v
}

func TestBoostOverridesResolve(t *testing.T) {
	// 재정의 없음: 전부 기본값
	boosts := BoostOverrides{}.resolve()
	if boosts != DefaultFieldBoosts() {
		t.Errorf("Expected default boosts, got %+v", boosts)
	}

	// 부분 재정의: 지정한 필드만 바뀌고 나머지는 기본값 유지
	boosts = BoostOverrides{Title: floatPtr(2.0), Content: floatPtr(10.0)}.resolve()
	if boosts.Title != 2.0 {
		t.Errorf("Expected title boost 2.0, got %v", boosts.Title)
	}
	if boosts.Content != 10.0 {
		t.Errorf("Expected content boost 10.0, got %v", boosts.Content)
	}
	if boosts.Description != DefaultDescriptionBoost {
		t.Errorf("Expected default description boost, got %v", boosts.Description)
	}
	if boosts.Category != DefaultCategoryBoost {
		t.Errorf("Expected default category boost, got %v", boosts.Category)
	}

	// 0으로 재정의: 필드 비활성화도 허용
	boosts = BoostOverrides{Category: floatPtr(0)}.resolve()
	if boosts.Category != 0 {
		t.Errorf("Expected category boost 0, got %v", boosts.Category)
	}
}

func TestFieldBoostsValidate(t *testing.T) {
	valid := []FieldBoosts{
		DefaultFieldBoosts(),
		{Title: 3.0},                             // 나머지 0은 허용 (하나라도 양수면 됨)
		{Title: 0, Description: 0, Content: 1.0}, // 개별 0 허용
		{Title: MaxFieldBoost, Content: 1.0},     // 상한값은 허용
	}
	for _, fb := range valid {
		if err := fb.validate(); err != nil {
			t.Errorf("Expected no error for %+v, got: %v", fb, err)
		}
	}

	invalid := []FieldBoosts{
		{},                                 // 전부 0: queryNorm이 0으로 나눠져 NaN 점수 발생
		{Title: -1, Content: 1.0},          // 음수
		{Title: math.NaN(), Content: 1.0},  // NaN은 < 0 비교를 통과하므로 명시적으로 거부해야 함
		{Title: math.Inf(1), Content: 1.0}, // +Inf도 점수를 NaN으로 오염시킴
		{Title: 1e308, Content: 1.0},       // 유한해도 과대하면 bleve 가중치 계산이 오버플로됨
		{Title: MaxFieldBoost + 1},         // 상한 초과
	}
	for _, fb := range invalid {
		if err := fb.validate(); err == nil {
			t.Errorf("Expected error for %+v", fb)
		}
	}
}

func TestSearcher_SearchWithBoostOverrides(t *testing.T) {
	s, tempDir := testSearcher(t, appsInTossIndexer, nil)
	defer os.RemoveAll(tempDir)

	if err := s.indexManager.CreateIndex(); err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}
	defer s.Close()

	// title-doc은 제목에만, content-doc은 본문에만 검색어가 등장한다
	docs := []IndexDocument{
		{
			ID:      "title-doc",
			Title:   "결제 연동 가이드",
			Content: "미니앱 개발 문서입니다",
			URL:     "https://example.com/title-doc",
		},
		{
			ID:      "content-doc",
			Title:   "미니앱 시작하기",
			Content: "토스페이 결제 연동에 대한 상세한 설명입니다",
			URL:     "https://example.com/content-doc",
		},
	}
	if err := s.indexManager.IndexDocuments(docs); err != nil {
		t.Fatalf("Failed to index: %v", err)
	}

	ctx := context.Background()

	// 기본 부스트: 제목 매칭 문서가 먼저 나와야 한다
	results, err := s.Search(ctx, "결제", nil)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) < 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}
	if results[0].ID != "title-doc" {
		t.Errorf("Expected title-doc first with default boosts, got %s", results[0].ID)
	}

	// content 부스트를 크게 올리고 title을 낮추면 본문 매칭 문서가 먼저 나와야 한다
	results, err = s.Search(ctx, "결제", &SearchOptions{
		Boosts: BoostOverrides{
			Title:   floatPtr(0.1),
			Content: floatPtr(50.0),
		},
	})
	if err != nil {
		t.Fatalf("Search with boost overrides failed: %v", err)
	}
	if len(results) < 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}
	if results[0].ID != "content-doc" {
		t.Errorf("Expected content-doc first with boosted content, got %s", results[0].ID)
	}
}

func TestSearcher_SearchInvalidBoostsRejected(t *testing.T) {
	s, tempDir := testSearcher(t, appsInTossIndexer, nil)
	defer os.RemoveAll(tempDir)

	if err := s.indexManager.CreateIndex(); err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}
	defer s.Close()

	cases := []struct {
		name   string
		boosts BoostOverrides
	}{
		{"negative", BoostOverrides{Title: floatPtr(-1)}},
		{"NaN", BoostOverrides{Title: floatPtr(math.NaN())}},
		{"+Inf", BoostOverrides{Content: floatPtr(math.Inf(1))}},
		{"huge finite", BoostOverrides{Title: floatPtr(1e308)}},
		{"all zero", BoostOverrides{
			Title:       floatPtr(0),
			Description: floatPtr(0),
			Content:     floatPtr(0),
			Category:    floatPtr(0),
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := s.Search(context.Background(), "결제", &SearchOptions{Boosts: tc.boosts})
			if err == nil {
				t.Errorf("Expected error for %s boost", tc.name)
			}
		})
	}
}
