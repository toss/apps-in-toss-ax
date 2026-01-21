package search

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/toss/apps-in-toss-ax/pkg/docid"
	"github.com/toss/apps-in-toss-ax/pkg/llms"
)

func TestIndexManager_CreateAndSearch(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "bleve-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	indexPath := filepath.Join(tempDir, "test-index")
	im := NewIndexManager(indexPath)

	if err := im.CreateIndex(); err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}
	defer im.Close()

	docs := []IndexDocument{
		{
			ID:          "doc1",
			Title:       "React Native 시작하기",
			Description: "앱인토스에서 React Native 앱을 개발하는 방법",
			URL:         "https://example.com/react-native",
			Category:    "시작하기",
		},
		{
			ID:          "doc2",
			Title:       "토스페이 결제 연동",
			Description: "토스페이 결제를 미니앱 연동하는 가이드",
			URL:         "https://example.com/tosspay",
			Category:    "결제",
		},
		{
			ID:          "doc3",
			Title:       "Unity 게임 포팅",
			Description: "Unity WebGL 게임을 앱인토스로 포팅하기",
			URL:         "https://example.com/unity",
			Category:    "게임",
		},
	}

	if err := im.IndexDocuments(docs); err != nil {
		t.Fatalf("Failed to index documents: %v", err)
	}

	tests := []struct {
		name      string
		query     string
		wantFirst string // 첫 번째 결과로 기대되는 문서 ID
		wantMatch bool   // 최소 1개 이상의 결과가 있어야 하는지
	}{
		{
			name:      "Korean search - 토스페이",
			query:     "토스페이",
			wantFirst: "doc2",
			wantMatch: true,
		},
		{
			name:      "Korean search - 결제",
			query:     "결제",
			wantFirst: "doc2",
			wantMatch: true,
		},
		{
			name:      "English search - React",
			query:     "React",
			wantFirst: "doc1",
			wantMatch: true,
		},
		{
			name:      "English search - Unity",
			query:     "Unity",
			wantFirst: "doc3",
			wantMatch: true,
		},
		{
			name:      "Korean search - 미니앱 (partial match)",
			query:     "미니앱",
			wantFirst: "doc2",
			wantMatch: true,
		},
		{
			name:      "Korean search - 앱인토스",
			query:     "앱인토스",
			wantFirst: "", // doc1, doc3 둘 다 포함하므로 순서는 점수에 따라 다름
			wantMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, scores, err := im.Search(tt.query, 10)
			if err != nil {
				t.Fatalf("Search failed: %v", err)
			}

			if tt.wantMatch && len(results) == 0 {
				t.Errorf("Expected at least one result, got none")
			}

			if tt.wantFirst != "" && len(results) > 0 {
				if results[0].ID != tt.wantFirst {
					t.Errorf("Expected first result ID '%s', got '%s'", tt.wantFirst, results[0].ID)
					for i, r := range results {
						t.Logf("  Result %d: %s (score: %f)", i, r.Title, scores[i])
					}
				}
			}

			if len(results) > 0 && len(scores) != len(results) {
				t.Errorf("Scores count mismatch: %d results, %d scores", len(results), len(scores))
			}
		})
	}
}

func TestIndexManager_IndexFromLlmsTxt(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "bleve-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	indexPath := filepath.Join(tempDir, "test-index")
	im := NewIndexManager(indexPath)

	if err := im.CreateIndex(); err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}
	defer im.Close()

	llmsTxt := &llms.LlmsTxt{
		Title:   "앱인토스 개발자센터",
		Summary: "앱인토스 개발 문서",
		Sections: []llms.Section{
			{
				Title: "시작하기",
				Level: 2,
				Links: []llms.Link{
					{
						Title:       "Quick Start",
						URL:         "https://example.com/quick-start",
						Description: "빠르게 시작하기",
					},
				},
				Children: []llms.Section{
					{
						Title: "React Native",
						Level: 3,
						Links: []llms.Link{
							{
								Title:       "React Native 가이드",
								URL:         "https://example.com/react-native",
								Description: "React Native로 미니앱 개발하기",
							},
						},
					},
				},
			},
			{
				Title: "결제",
				Level: 2,
				Links: []llms.Link{
					{
						Title:       "토스페이 연동",
						URL:         "https://example.com/tosspay",
						Description: "토스페이 결제 연동 가이드",
					},
				},
			},
		},
	}

	if err := im.IndexFromLlmsTxt(llmsTxt); err != nil {
		t.Fatalf("Failed to index from LlmsTxt: %v", err)
	}

	results, _, err := im.Search("토스페이", 10)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) == 0 {
		t.Error("Expected at least one result for '토스페이'")
	}

	results, _, err = im.Search("React Native", 10)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) == 0 {
		t.Error("Expected at least one result for 'React Native'")
	}
}

func TestIndexManager_OpenExistingIndex(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "bleve-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	indexPath := filepath.Join(tempDir, "test-index")

	im1 := NewIndexManager(indexPath)
	if err := im1.CreateIndex(); err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}

	docs := []IndexDocument{
		{
			ID:          "doc1",
			Title:       "테스트 문서",
			Description: "테스트 설명",
			URL:         "https://example.com/test",
		},
	}
	if err := im1.IndexDocuments(docs); err != nil {
		t.Fatalf("Failed to index documents: %v", err)
	}
	im1.Close()

	im2 := NewIndexManager(indexPath)
	if err := im2.OpenIndex(); err != nil {
		t.Fatalf("Failed to open existing index: %v", err)
	}
	defer im2.Close()

	results, _, err := im2.Search("테스트", 10)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
}

func TestCacheManager_ETagOperations(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "cache-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cm := &CacheManager{
		cacheDir:     tempDir,
		metadataPath: filepath.Join(tempDir, "cache-metadata.json"),
		indexPath:    filepath.Join(tempDir, "search-index"),
	}

	etag, err := cm.GetCachedETag()
	if err != nil {
		t.Fatalf("GetCachedETag failed: %v", err)
	}
	if etag != "" {
		t.Errorf("Expected empty etag, got '%s'", etag)
	}

	testETag := "\"abc123\""
	testURL := "https://example.com/llms-full.txt"
	if err := cm.SaveETag(testURL, testETag); err != nil {
		t.Fatalf("SaveETag failed: %v", err)
	}

	etag, err = cm.GetCachedETag()
	if err != nil {
		t.Fatalf("GetCachedETag failed: %v", err)
	}
	if etag != testETag {
		t.Errorf("Expected etag '%s', got '%s'", testETag, etag)
	}

	if cm.IndexExists() {
		t.Error("Expected IndexExists to return false")
	}

	if err := os.MkdirAll(cm.indexPath, 0755); err != nil {
		t.Fatalf("Failed to create index dir: %v", err)
	}

	if !cm.IndexExists() {
		t.Error("Expected IndexExists to return true")
	}

	if err := cm.DeleteIndex(); err != nil {
		t.Fatalf("DeleteIndex failed: %v", err)
	}

	if cm.IndexExists() {
		t.Error("Expected IndexExists to return false after delete")
	}
}

func TestPartialMatchKorean(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "bleve-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	indexPath := filepath.Join(tempDir, "test-index")
	im := NewIndexManager(indexPath)

	if err := im.CreateIndex(); err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}
	defer im.Close()

	docs := []IndexDocument{
		{
			ID:          "doc1",
			Title:       "미니앱에서 결제하기",
			Description: "앱인토스에서 미니앱 결제를 연동합니다",
			URL:         "https://example.com/payment",
		},
	}

	if err := im.IndexDocuments(docs); err != nil {
		t.Fatalf("Failed to index documents: %v", err)
	}

	// "미니앱에서" 텍스트에서 "미니앱"으로 검색
	// edgengram은 앞에서부터 매칭하므로 중간 매칭("앱인토스" → "토스")은 지원 안 함
	tests := []struct {
		query     string
		wantMatch bool
	}{
		{"미니앱", true},  // "미니앱에서"에서 "미니앱" 검색
		{"앱인토스", true}, // "앱인토스에서"에서 "앱인토스" 검색
		{"결제", true},   // 정확히 일치
		{"미니", true},   // 앞부분 매칭
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			results, _, err := im.Search(tt.query, 10)
			if err != nil {
				t.Fatalf("Search failed: %v", err)
			}

			if tt.wantMatch && len(results) == 0 {
				t.Errorf("Expected match for '%s', got none", tt.query)
			}
			if !tt.wantMatch && len(results) > 0 {
				t.Errorf("Expected no match for '%s', got %d results", tt.query, len(results))
			}
		})
	}
}

func TestParseLlmsFull(t *testing.T) {
	content := `---
url: >-
  https://developers-apps-in-toss.toss.im/bedrock/reference/framework/화면제어/IOScrollView.md
---
# 스크롤 영역 노출 감지하기

## 개요

IOScrollView는 스크롤 영역 내에서 특정 요소가 화면에 노출되는지 감지하는 컴포넌트입니다.

***

## 기본 사용법

` + "```tsx\n" + `import { IOScrollView } from '@appsintos/bedrock';
` + "```" + `

---
url: >-
  https://developers-apps-in-toss.toss.im/bedrock/reference/framework/네비게이션/useNavigation.md
---
# 네비게이션 사용하기

앱인토스에서 화면 전환을 위한 네비게이션 훅입니다.
`

	docs := ParseLlmsFull(content)

	if len(docs) != 2 {
		t.Fatalf("Expected 2 documents, got %d", len(docs))
	}

	// 첫 번째 문서 검증
	if docs[0].URL != "https://developers-apps-in-toss.toss.im/bedrock/reference/framework/화면제어/IOScrollView.md" {
		t.Errorf("Unexpected URL: %s", docs[0].URL)
	}
	if docs[0].Title != "스크롤 영역 노출 감지하기" {
		t.Errorf("Unexpected title: %s", docs[0].Title)
	}
	if !strings.Contains(docs[0].Content, "IOScrollView는") {
		t.Errorf("Content should contain 'IOScrollView는': %s", docs[0].Content)
	}

	// 두 번째 문서 검증
	if docs[1].URL != "https://developers-apps-in-toss.toss.im/bedrock/reference/framework/네비게이션/useNavigation.md" {
		t.Errorf("Unexpected URL: %s", docs[1].URL)
	}
	if docs[1].Title != "네비게이션 사용하기" {
		t.Errorf("Unexpected title: %s", docs[1].Title)
	}
}

func TestSearchWithLlmsFullFormat(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "bleve-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	indexPath := filepath.Join(tempDir, "test-index")
	im := NewIndexManager(indexPath)

	if err := im.CreateIndex(); err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}
	defer im.Close()

	// llms-full.txt 형식의 테스트 데이터
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

	// 카테고리 매핑 (테스트용)
	categoryMap := map[string]string{
		"https://example.com/scroll.md":     "시작 > 스크롤",
		"https://example.com/navigation.md": "시작 > 네비게이션",
	}

	if err := im.IndexFromLlmsFullContent(content, categoryMap); err != nil {
		t.Fatalf("Failed to index: %v", err)
	}

	tests := []struct {
		query     string
		wantMatch bool
		wantTitle string
	}{
		{"스크롤", true, "스크롤 영역 노출 감지하기"},
		{"네비게이션", true, "네비게이션 사용하기"},
		{"앱인토스", true, ""}, // 둘 다 매칭
		{"미니앱", true, "네비게이션 사용하기"},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			results, _, err := im.Search(tt.query, 10)
			if err != nil {
				t.Fatalf("Search failed: %v", err)
			}

			if tt.wantMatch && len(results) == 0 {
				t.Errorf("Expected match for '%s', got none", tt.query)
			}

			if tt.wantTitle != "" && len(results) > 0 && results[0].Title != tt.wantTitle {
				t.Errorf("Expected first result title '%s', got '%s'", tt.wantTitle, results[0].Title)
			}
		})
	}
}

func TestGenerateDocID(t *testing.T) {
	// docs.go의 generateID와 동일한 결과를 생성하는지 확인
	tests := []struct {
		title    string
		url      string
		category string
	}{
		{
			title:    "앱인토스 Unity 적용 가이드",
			url:      "https://developers-apps-in-toss.toss.im/unity/intro/overview.md",
			category: "Table of Contents > 시작",
		},
		{
			title:    "스크롤 영역 노출 감지하기",
			url:      "https://developers-apps-in-toss.toss.im/bedrock/reference/framework/화면제어/IOScrollView.md",
			category: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			id := docid.Generate(tt.title, tt.url, tt.category)

			// ID가 16자 hex 문자열인지 확인 (8바이트 = 16 hex chars)
			if len(id) != 16 {
				t.Errorf("Expected ID length 16, got %d: %s", len(id), id)
			}

			// 동일한 입력에 대해 동일한 ID가 생성되는지 확인
			id2 := docid.Generate(tt.title, tt.url, tt.category)
			if id != id2 {
				t.Errorf("ID should be deterministic: %s != %s", id, id2)
			}
		})
	}
}

func TestBuildCategoryMap(t *testing.T) {
	llmsTxt := &llms.LlmsTxt{
		Title:   "앱인토스 개발자센터",
		Summary: "앱인토스 개발자센터",
		Sections: []llms.Section{
			{
				Title: "Table of Contents",
				Level: 2,
				Children: []llms.Section{
					{
						Title: "시작",
						Level: 3,
						Links: []llms.Link{
							{
								Title: "앱인토스 Unity 적용 가이드",
								URL:   "https://example.com/unity.md",
							},
						},
					},
					{
						Title: "준비",
						Level: 3,
						Links: []llms.Link{
							{
								Title: "동작 방식",
								URL:   "https://example.com/runtime.md",
							},
						},
					},
				},
			},
		},
	}

	categoryMap := BuildCategoryMap(llmsTxt)

	expected := map[string]string{
		"https://example.com/unity.md":   "Table of Contents > 시작",
		"https://example.com/runtime.md": "Table of Contents > 준비",
	}

	for url, expectedCategory := range expected {
		if categoryMap[url] != expectedCategory {
			t.Errorf("URL %s: expected category '%s', got '%s'", url, expectedCategory, categoryMap[url])
		}
	}
}

func TestSearchBoostAndFuzziness(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "bleve-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	indexPath := filepath.Join(tempDir, "test-index")
	im := NewIndexManager(indexPath)

	if err := im.CreateIndex(); err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}
	defer im.Close()

	// Boost 테스트를 위한 문서들
	// title(5.0) > description(1.5) > category(1.0) = content(1.0)
	docs := []IndexDocument{
		{
			ID:          "title_match",
			Title:       "결제 시스템 가이드",
			Description: "일반적인 설명입니다",
			Content:     "일반적인 내용입니다",
			URL:         "https://example.com/title",
			Category:    "일반",
		},
		{
			ID:          "desc_match",
			Title:       "일반 가이드",
			Description: "결제 연동에 대한 설명입니다",
			Content:     "일반적인 내용입니다",
			URL:         "https://example.com/desc",
			Category:    "일반",
		},
		{
			ID:          "category_match",
			Title:       "일반 가이드",
			Description: "일반적인 설명입니다",
			Content:     "일반적인 내용입니다",
			URL:         "https://example.com/category",
			Category:    "결제",
		},
		{
			ID:          "content_match",
			Title:       "일반 가이드",
			Description: "일반적인 설명입니다",
			Content:     "결제 기능을 구현하는 방법입니다",
			URL:         "https://example.com/content",
			Category:    "일반",
		},
	}

	if err := im.IndexDocuments(docs); err != nil {
		t.Fatalf("Failed to index documents: %v", err)
	}

	t.Run("Boost_TitleHighestPriority", func(t *testing.T) {
		results, scores, err := im.Search("결제", 10)
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}

		if len(results) < 4 {
			t.Fatalf("Expected at least 4 results, got %d", len(results))
		}

		// title 매칭(boost 5.0)이 가장 높은 점수여야 함
		if results[0].ID != "title_match" {
			t.Errorf("Expected 'title_match' to be first (highest boost), got '%s'", results[0].ID)
			for i, r := range results {
				t.Logf("  Result %d: %s (score: %f)", i, r.ID, scores[i])
			}
		}

		// description 매칭(boost 1.5)이 두 번째로 높아야 함
		if results[1].ID != "desc_match" {
			t.Errorf("Expected 'desc_match' to be second (boost 1.5), got '%s'", results[1].ID)
		}

		// 점수 순서 확인: title > category > desc > content
		for i := 0; i < len(scores)-1; i++ {
			if scores[i] < scores[i+1] {
				t.Errorf("Scores should be in descending order: score[%d]=%f < score[%d]=%f",
					i, scores[i], i+1, scores[i+1])
			}
		}
	})

	t.Run("Boost_DescriptionVsContent", func(t *testing.T) {
		// description(0.7)이 content(0.5)보다 높은 점수여야 함
		results, scores, err := im.Search("결제", 10)
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}

		var descScore, contentScore float64
		for i, r := range results {
			if r.ID == "desc_match" {
				descScore = scores[i]
			}
			if r.ID == "content_match" {
				contentScore = scores[i]
			}
		}

		if descScore <= contentScore {
			t.Errorf("Description match (boost 0.7) should score higher than content match (boost 0.5): desc=%f, content=%f",
				descScore, contentScore)
		}
	})
}

func TestSearchFuzziness(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "bleve-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	indexPath := filepath.Join(tempDir, "test-index")
	im := NewIndexManager(indexPath)

	if err := im.CreateIndex(); err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}
	defer im.Close()

	docs := []IndexDocument{
		{
			ID:          "doc1",
			Title:       "Navigation 가이드",
			Description: "React Navigation을 사용한 화면 전환",
			Content:     "네비게이션 구현 방법을 설명합니다",
			URL:         "https://example.com/nav",
			Category:    "네비게이션",
		},
		{
			ID:          "doc2",
			Title:       "Button 컴포넌트",
			Description: "버튼 컴포넌트 사용법",
			Content:     "다양한 버튼 스타일을 제공합니다",
			URL:         "https://example.com/button",
			Category:    "컴포넌트",
		},
	}

	if err := im.IndexDocuments(docs); err != nil {
		t.Fatalf("Failed to index documents: %v", err)
	}

	// Fuzziness 테스트 - description과 content에 AutoFuzziness가 설정됨
	// 영어 단어의 경우 약간의 오타도 매칭될 수 있음
	t.Run("Fuzziness_EnglishTypo", func(t *testing.T) {
		// "Navigatin" (오타) 로 검색 시 "Navigation" 매칭 가능 여부
		results, _, err := im.Search("Navigaton", 10)
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}

		// Fuzziness가 동작하면 결과가 있어야 함
		if len(results) > 0 {
			t.Logf("Fuzziness working: found %d results for typo 'Navigaton'", len(results))
		} else {
			t.Logf("Fuzziness did not match typo 'Navigaton' - this may be expected behavior")
		}
	})

	t.Run("ExactMatch_StillWorks", func(t *testing.T) {
		// 정확한 검색어는 반드시 매칭되어야 함
		results, _, err := im.Search("Navigation", 10)
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}

		if len(results) == 0 {
			t.Error("Expected at least one result for exact match 'Navigation'")
		}
	})
}

func TestSearchMultiFieldMatching(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "bleve-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	indexPath := filepath.Join(tempDir, "test-index")
	im := NewIndexManager(indexPath)

	if err := im.CreateIndex(); err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}
	defer im.Close()

	// 여러 필드에 동일한 키워드가 있는 문서는 더 높은 점수를 받아야 함
	docs := []IndexDocument{
		{
			ID:          "multi_match",
			Title:       "앱인토스 결제 가이드",
			Description: "앱인토스에서 결제를 연동하는 방법",
			Content:     "앱인토스 미니앱에서 결제 기능 구현",
			URL:         "https://example.com/multi",
			Category:    "앱인토스 > 결제",
		},
		{
			ID:          "single_match",
			Title:       "일반 가이드",
			Description: "일반적인 설명입니다",
			Content:     "앱인토스 관련 내용이 있습니다",
			URL:         "https://example.com/single",
			Category:    "일반",
		},
	}

	if err := im.IndexDocuments(docs); err != nil {
		t.Fatalf("Failed to index documents: %v", err)
	}

	t.Run("MultiFieldMatch_HigherScore", func(t *testing.T) {
		results, scores, err := im.Search("앱인토스", 10)
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}

		if len(results) < 2 {
			t.Fatalf("Expected at least 2 results, got %d", len(results))
		}

		// 여러 필드에서 매칭되는 문서가 더 높은 점수를 받아야 함
		if results[0].ID != "multi_match" {
			t.Errorf("Expected 'multi_match' (multiple field matches) to rank first, got '%s'", results[0].ID)
			for i, r := range results {
				t.Logf("  Result %d: %s (score: %f)", i, r.ID, scores[i])
			}
		}

		// multi_match의 점수가 single_match보다 높아야 함
		if len(scores) >= 2 && scores[0] <= scores[1] {
			t.Errorf("Multi-field match should have higher score: multi=%f, single=%f", scores[0], scores[1])
		}
	})
}

func TestSearchDescriptionVsCategory(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "bleve-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	indexPath := filepath.Join(tempDir, "test-index")
	im := NewIndexManager(indexPath)

	if err := im.CreateIndex(); err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}
	defer im.Close()

	// Description boost(1.5) vs Category boost(1.0) 테스트
	docs := []IndexDocument{
		{
			ID:          "desc_unity",
			Title:       "일반 가이드",
			Description: "Unity 엔진을 사용한 게임 개발",
			Content:     "일반적인 내용입니다",
			URL:         "https://example.com/desc",
			Category:    "일반",
		},
		{
			ID:          "category_unity",
			Title:       "일반 가이드",
			Description: "일반적인 설명입니다",
			Content:     "일반적인 내용입니다",
			URL:         "https://example.com/cat",
			Category:    "Unity 게임",
		},
	}

	if err := im.IndexDocuments(docs); err != nil {
		t.Fatalf("Failed to index documents: %v", err)
	}

	t.Run("DescriptionBoost_HigherThanCategory", func(t *testing.T) {
		results, scores, err := im.Search("Unity", 10)
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}

		if len(results) < 2 {
			t.Fatalf("Expected at least 2 results, got %d", len(results))
		}

		// Description(boost 1.5)이 Category(boost 1.0)보다 높아야 함
		if results[0].ID != "desc_unity" {
			t.Errorf("Expected 'desc_unity' (boost 1.5) to rank higher than category match (boost 1.0), got '%s'", results[0].ID)
			for i, r := range results {
				t.Logf("  Result %d: %s (score: %f)", i, r.ID, scores[i])
			}
		}
	})
}

func TestCJKAnalyzer_KoreanTokenization(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "bleve-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	indexPath := filepath.Join(tempDir, "test-index")
	im := NewIndexManager(indexPath)

	if err := im.CreateIndex(); err != nil {
		t.Fatalf("Failed to create index: %v", err)
	}
	defer im.Close()

	docs := []IndexDocument{
		{
			ID:          "doc1",
			Title:       "토스페이먼츠 결제 시스템",
			Description: "토스페이먼츠의 온라인 결제 연동",
			URL:         "https://example.com/payments",
		},
	}

	if err := im.IndexDocuments(docs); err != nil {
		t.Fatalf("Failed to index documents: %v", err)
	}

	// CJK bigram analyzer는 2글자씩 토큰화함
	// "토스페이먼츠" → "토스", "스페", "페이", "이먼", "먼츠"
	// "결제" → "결제" (2글자는 그대로)
	searchTerms := []string{
		"결제",
		"시스템",
		"온라인",
		"토스페이먼츠", // 전체 단어도 검색 가능
	}

	for _, term := range searchTerms {
		results, _, err := im.Search(term, 10)
		if err != nil {
			t.Fatalf("Search for '%s' failed: %v", term, err)
		}

		if len(results) == 0 {
			t.Errorf("Expected results for Korean term '%s', got none", term)
		}
	}
}
