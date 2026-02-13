package search

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/toss/apps-in-toss-ax/pkg/llms"
)

// testSearcher는 테스트용 Searcher를 생성하는 헬퍼입니다.
// 실제 HTTP 요청 없이 ContentIndexer와 인덱스를 직접 구성합니다.
func testSearcher(t *testing.T, indexer ContentIndexer, urlTransform URLTransformFunc) (*Searcher, string) {
	t.Helper()

	tempDir, err := os.MkdirTemp("", "searcher-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	indexPath := filepath.Join(tempDir, "test-index")
	im := NewIndexManager(indexPath)

	return &Searcher{
		indexManager: im,
		indexer:      indexer,
		urlTransform: urlTransform,
	}, tempDir
}

func TestAppsInTossIndexer(t *testing.T) {
	content := `---
url: >-
  https://example.com/scroll.md
---
# 스크롤 영역 노출 감지하기

앱인토스에서 스크롤 뷰를 사용하는 방법입니다.

---
url: >-
  https://example.com/navigation.md
---
# 네비게이션 사용하기

앱인토스 미니앱에서 화면 전환하기
`

	categoryMap := map[string]string{
		"https://example.com/scroll.md":     "시작 > 스크롤",
		"https://example.com/navigation.md": "시작 > 네비게이션",
	}

	docs := appsInTossIndexer(content, categoryMap)

	if len(docs) != 2 {
		t.Fatalf("Expected 2 documents, got %d", len(docs))
	}

	if docs[0].Title != "스크롤 영역 노출 감지하기" {
		t.Errorf("Unexpected title: %s", docs[0].Title)
	}
	if docs[0].Category != "시작 > 스크롤" {
		t.Errorf("Expected category '시작 > 스크롤', got '%s'", docs[0].Category)
	}
	if docs[0].ID == "" {
		t.Error("Expected non-empty ID")
	}
}

func TestAppsInTossIndexer_NilCategoryMap(t *testing.T) {
	content := `---
url: >-
  https://example.com/doc.md
---
# 테스트 문서

테스트 내용
`

	docs := appsInTossIndexer(content, nil)

	if len(docs) != 1 {
		t.Fatalf("Expected 1 document, got %d", len(docs))
	}
	if docs[0].Category != "" {
		t.Errorf("Expected empty category with nil map, got '%s'", docs[0].Category)
	}
}

func TestTdsIndexer(t *testing.T) {
	content := `# Button (/tds-react-native/components/button/)

버튼 컴포넌트입니다.

---
# Toast (/tds-react-native/components/toast/)

토스트 컴포넌트입니다.
`

	categoryMap := map[string]string{
		"https://tossmini-docs.toss.im/tds-react-native/components/button/": "Components",
	}

	docs := tdsIndexer(content, categoryMap)

	if len(docs) != 2 {
		t.Fatalf("Expected 2 documents, got %d", len(docs))
	}

	// 첫 번째 문서: 카테고리 맵에 존재
	if docs[0].Title != "Button" {
		t.Errorf("Unexpected title: %s", docs[0].Title)
	}
	if docs[0].Category != "Components" {
		t.Errorf("Expected category 'Components', got '%s'", docs[0].Category)
	}

	// 두 번째 문서: 카테고리 맵에 없으므로 URL에서 추출
	if docs[1].Title != "Toast" {
		t.Errorf("Unexpected title: %s", docs[1].Title)
	}
	if docs[1].Category == "" {
		t.Error("Expected non-empty fallback category from URL")
	}
}

func TestTdsIndexer_FallbackCategory(t *testing.T) {
	content := `# Component (/tds-react-native/components/test/)

테스트
`

	docs := tdsIndexer(content, nil)

	if len(docs) != 1 {
		t.Fatalf("Expected 1 document, got %d", len(docs))
	}
	// nil 카테고리 맵이면 URL에서 카테고리를 추출해야 함
	if docs[0].Category == "" {
		t.Error("Expected fallback category extracted from URL")
	}
}

func TestSearcher_SearchWithCustomIndexer(t *testing.T) {
	s, tempDir := testSearcher(t, appsInTossIndexer, nil)
	defer os.RemoveAll(tempDir)

	if err := s.indexManager.CreateIndex(); err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}
	defer s.Close()

	content := `---
url: >-
  https://example.com/payment.md
---
# 결제 연동 가이드

토스페이 결제를 연동하는 방법입니다.
`

	docs := s.indexer(content, nil)
	if err := s.indexManager.IndexDocuments(docs); err != nil {
		t.Fatalf("Failed to index: %v", err)
	}

	results, err := s.Search(context.Background(), "결제", &SearchOptions{Limit: 10})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) == 0 {
		t.Error("Expected at least one result")
	}
	if results[0].Score <= 0 {
		t.Error("Expected positive score")
	}
}

func TestSearcher_SearchContentTruncation(t *testing.T) {
	s, tempDir := testSearcher(t, appsInTossIndexer, nil)
	defer os.RemoveAll(tempDir)

	if err := s.indexManager.CreateIndex(); err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}
	defer s.Close()

	// 500자 이상의 긴 콘텐츠 생성
	longContent := ""
	for i := 0; i < 100; i++ {
		longContent += "토스페이 결제를 연동하는 방법에 대한 아주 긴 설명입니다. "
	}

	docs := []IndexDocument{{
		ID:      "long-doc",
		Title:   "긴 문서",
		Content: longContent,
		URL:     "https://example.com/long",
	}}
	if err := s.indexManager.IndexDocuments(docs); err != nil {
		t.Fatalf("Failed to index: %v", err)
	}

	// 기본 검색: Content가 500자로 잘려야 함
	results, err := s.Search(context.Background(), "결제", nil)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("Expected at least one result")
	}
	contentRunes := []rune(results[0].Content)
	if len(contentRunes) > 500+3 { // 500 + "..."
		t.Errorf("Expected content truncated to ~500 runes, got %d", len(contentRunes))
	}

	// MaxContentLength 지정: 100자로 잘려야 함
	results, err = s.Search(context.Background(), "결제", &SearchOptions{MaxContentLength: 100})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("Expected at least one result")
	}
	contentRunes = []rune(results[0].Content)
	if len(contentRunes) > 100+3 {
		t.Errorf("Expected content truncated to ~100 runes, got %d", len(contentRunes))
	}
}

func TestSearcher_GetDocument(t *testing.T) {
	s, tempDir := testSearcher(t, appsInTossIndexer, nil)
	defer os.RemoveAll(tempDir)

	if err := s.indexManager.CreateIndex(); err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}
	defer s.Close()

	// 긴 콘텐츠 인덱싱
	longContent := ""
	for i := 0; i < 100; i++ {
		longContent += "토스페이 결제를 연동하는 방법입니다. "
	}

	docs := []IndexDocument{{
		ID:      "test-doc-1",
		Title:   "결제 가이드",
		Content: longContent,
		URL:     "https://example.com/payment",
	}}
	if err := s.indexManager.IndexDocuments(docs); err != nil {
		t.Fatalf("Failed to index: %v", err)
	}

	ctx := context.Background()

	// GetDocument는 전체 Content를 반환해야 함
	result, err := s.GetDocument(ctx, "test-doc-1")
	if err != nil {
		t.Fatalf("GetDocument failed: %v", err)
	}
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	if result.Title != "결제 가이드" {
		t.Errorf("Unexpected title: %s", result.Title)
	}
	if result.Content != longContent {
		t.Errorf("Expected full content (len=%d), got len=%d", len(longContent), len(result.Content))
	}

	// 존재하지 않는 ID
	result, err = s.GetDocument(ctx, "nonexistent-id")
	if err != nil {
		t.Fatalf("Expected nil result without error, got: %v", err)
	}
	if result != nil {
		t.Error("Expected nil result for nonexistent ID")
	}
}

func TestSearcher_SearchWithTdsIndexer(t *testing.T) {
	s, tempDir := testSearcher(t, tdsIndexer, tdsURLTransform)
	defer os.RemoveAll(tempDir)

	if err := s.indexManager.CreateIndex(); err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}
	defer s.Close()

	content := `# Button (/tds-react-native/components/button/)

버튼 컴포넌트입니다. 다양한 스타일을 지원합니다.

---
# Typography (/tds-react-native/components/typography/)

타이포그래피 컴포넌트입니다.
`

	docs := s.indexer(content, nil)
	if err := s.indexManager.IndexDocuments(docs); err != nil {
		t.Fatalf("Failed to index: %v", err)
	}

	results, err := s.Search(context.Background(), "Button", &SearchOptions{Limit: 10})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) == 0 {
		t.Error("Expected at least one result for 'Button'")
	}
}

func TestSearcher_SearchOptionsDefault(t *testing.T) {
	s, tempDir := testSearcher(t, appsInTossIndexer, nil)
	defer os.RemoveAll(tempDir)

	if err := s.indexManager.CreateIndex(); err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}
	defer s.Close()

	docs := []IndexDocument{{ID: "1", Title: "테스트", URL: "https://example.com"}}
	if err := s.indexManager.IndexDocuments(docs); err != nil {
		t.Fatalf("Failed to index: %v", err)
	}

	// nil opts는 기본 limit 10을 사용
	results, err := s.Search(context.Background(), "테스트", nil)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) == 0 {
		t.Error("Expected at least one result")
	}
}

func TestCreateIndex_OverwritesExistingPath(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "bleve-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	indexPath := filepath.Join(tempDir, "test-index")
	im := NewIndexManager(indexPath)

	// 첫 번째 인덱스 생성
	if err := im.CreateIndex(); err != nil {
		t.Fatalf("First CreateIndex failed: %v", err)
	}
	docs := []IndexDocument{{ID: "old", Title: "Old Doc", URL: "https://example.com/old"}}
	if err := im.IndexDocuments(docs); err != nil {
		t.Fatalf("Failed to index: %v", err)
	}
	im.Close()

	// 같은 경로에 다시 생성 — 기존 인덱스를 덮어써야 함
	im2 := NewIndexManager(indexPath)
	if err := im2.CreateIndex(); err != nil {
		t.Fatalf("Second CreateIndex should overwrite, got: %v", err)
	}
	defer im2.Close()

	// 새 인덱스에 새 문서 추가
	newDocs := []IndexDocument{{ID: "new", Title: "New Doc", URL: "https://example.com/new"}}
	if err := im2.IndexDocuments(newDocs); err != nil {
		t.Fatalf("Failed to index new docs: %v", err)
	}

	// 이전 문서는 없어야 함
	results, _, err := im2.Search("Old", 10)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) > 0 {
		t.Error("Expected no results for old document after overwrite")
	}

	// 새 문서만 있어야 함
	results, _, err = im2.Search("New", 10)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 result for new document, got %d", len(results))
	}
}

func TestCreateIndex_ConcurrentDifferentPaths(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "bleve-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	const goroutines = 10
	var wg sync.WaitGroup
	errs := make([]error, goroutines)

	for i := range goroutines {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			indexPath := filepath.Join(tempDir, fmt.Sprintf("index-%d", idx))
			im := NewIndexManager(indexPath)

			if err := im.CreateIndex(); err != nil {
				errs[idx] = err
				return
			}
			defer im.Close()

			docs := []IndexDocument{{
				ID:    fmt.Sprintf("doc-%d", idx),
				Title: "테스트 문서",
				URL:   fmt.Sprintf("https://example.com/test-%d", idx),
			}}
			if err := im.IndexDocuments(docs); err != nil {
				errs[idx] = err
				return
			}

			results, _, err := im.Search("테스트", 10)
			if err != nil {
				errs[idx] = err
				return
			}
			if len(results) == 0 {
				errs[idx] = fmt.Errorf("expected results for goroutine %d", idx)
			}
		}(i)
	}

	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("goroutine %d: %v", i, err)
		}
	}
}

func createTestLlmsTxt() *llms.LlmsTxt {
	return &llms.LlmsTxt{
		Title:   "Test",
		Summary: "Test",
		Sections: []llms.Section{
			{
				Title: "Components",
				Level: 2,
				Links: []llms.Link{
					{Title: "Button", URL: "/components/button"},
				},
			},
		},
	}
}

func TestBuildCategoryMapWithURLTransform(t *testing.T) {
	llmsTxt := createTestLlmsTxt()

	// nil transform
	catMap := BuildCategoryMapWithURLTransform(llmsTxt, nil)
	if catMap["/components/button"] != "Components" {
		t.Errorf("Expected 'Components', got '%s'", catMap["/components/button"])
	}

	// URL transform
	catMap = BuildCategoryMapWithURLTransform(llmsTxt, func(url string) string {
		return "https://example.com" + url
	})
	if catMap["https://example.com/components/button"] != "Components" {
		t.Errorf("Expected 'Components' with transformed URL, got '%s'", catMap["https://example.com/components/button"])
	}
}

func TestExtractTdsCategory(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{
			url:  "https://tossmini-docs.toss.im/tds-react-native/components/button/",
			want: "Tds React Native > Components",
		},
		{
			url:  "https://tossmini-docs.toss.im/tds-mobile/hooks/use-toast/",
			want: "Tds Mobile > Hooks",
		},
		{
			url:  "https://tossmini-docs.toss.im/",
			want: "TDS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			got := extractTdsCategory(tt.url)
			if got != tt.want {
				t.Errorf("extractTdsCategory(%s) = %s, want %s", tt.url, got, tt.want)
			}
		})
	}
}

func TestTdsURLTransform(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"/tds-react-native/components/button/", "https://tossmini-docs.toss.im/tds-react-native/components/button/"},
		{"https://external.com/page", "https://external.com/page"},
	}

	for _, tt := range tests {
		got := tdsURLTransform(tt.input)
		if got != tt.want {
			t.Errorf("tdsURLTransform(%s) = %s, want %s", tt.input, got, tt.want)
		}
	}
}
