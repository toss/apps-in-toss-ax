package search

import (
	"context"
	"strings"

	"github.com/toss/apps-in-toss-ax/internal/httputil"
	"github.com/toss/apps-in-toss-ax/pkg/docid"
	"github.com/toss/apps-in-toss-ax/pkg/llms"
)

const (
	tdsBaseURL = "https://tossmini-docs.toss.im"

	tdsReactNativeLlmsFullUrl = "https://tossmini-docs.toss.im/tds-react-native/llms-full.txt"
	tdsReactNativeLlmsUrl     = "https://tossmini-docs.toss.im/tds-react-native/llms.txt"
	tdsMobileLlmsFullUrl      = "https://tossmini-docs.toss.im/tds-mobile/llms-full.txt"
	tdsMobileLlmsUrl          = "https://tossmini-docs.toss.im/tds-mobile/llms.txt"
)

// TDSSearcher는 TDS 문서 검색을 위한 구조체입니다
type TDSSearcher struct {
	llmsFullUrl  string
	llmsUrl      string
	cacheManager *CacheManager
	indexManager *IndexManager
}

// NewTDSSearcher는 TDS React Native 문서용 Searcher를 생성합니다
func NewTDSSearcher() (*TDSSearcher, error) {
	return newTDSSearcherWithConfig(
		tdsReactNativeLlmsFullUrl,
		tdsReactNativeLlmsUrl,
		CacheConfig{
			MetadataFileName: "tds-cache-metadata.json",
			IndexSubDir:      "tds-search-index",
		},
	)
}

// NewTDSMobileSearcher는 TDS Mobile 문서용 Searcher를 생성합니다
func NewTDSMobileSearcher() (*TDSSearcher, error) {
	return newTDSSearcherWithConfig(
		tdsMobileLlmsFullUrl,
		tdsMobileLlmsUrl,
		CacheConfig{
			MetadataFileName: "tds-mobile-cache-metadata.json",
			IndexSubDir:      "tds-mobile-search-index",
		},
	)
}

func newTDSSearcherWithConfig(llmsFullUrl, llmsUrl string, cacheConfig CacheConfig) (*TDSSearcher, error) {
	cacheManager, err := NewCacheManagerWithConfig(cacheConfig)
	if err != nil {
		return nil, err
	}

	indexManager := NewIndexManager(cacheManager.IndexPath())

	return &TDSSearcher{
		llmsFullUrl:  llmsFullUrl,
		llmsUrl:      llmsUrl,
		cacheManager: cacheManager,
		indexManager: indexManager,
	}, nil
}

// EnsureIndex는 인덱스가 존재하고 최신 상태인지 확인합니다
func (s *TDSSearcher) EnsureIndex(ctx context.Context) error {
	indexExists := s.cacheManager.IndexExists()

	if indexExists {
		etag, changed, err := s.cacheManager.CheckETag(ctx, s.llmsFullUrl)
		if err != nil {
			if err := s.indexManager.OpenIndex(); err == nil {
				return nil
			}
			return s.rebuildIndex(ctx)
		}

		if !changed {
			return s.indexManager.OpenIndex()
		}

		if err := s.cacheManager.DeleteIndex(); err != nil {
			return err
		}

		return s.buildIndex(ctx, etag)
	}

	return s.rebuildIndex(ctx)
}

func (s *TDSSearcher) rebuildIndex(ctx context.Context) error {
	etag, _, err := s.cacheManager.CheckETag(ctx, s.llmsFullUrl)
	if err != nil {
		etag = ""
	}
	return s.buildIndex(ctx, etag)
}

func (s *TDSSearcher) buildIndex(ctx context.Context, etag string) error {
	// TDS llms-full.txt 가져오기
	content, newETag, err := httputil.FetchWithETag(ctx, s.llmsFullUrl, 0)
	if err != nil {
		return err
	}

	if newETag != "" {
		etag = newETag
	}

	// llms.txt에서 카테고리 정보 가져오기
	categoryMap := s.fetchCategoryMap(ctx)

	if err := s.indexManager.CreateIndex(); err != nil {
		return err
	}

	// TDS llms-full.txt 형식을 파싱하여 인덱싱 (카테고리 매핑 포함)
	if err := s.indexFromTdsContent(content, categoryMap); err != nil {
		return err
	}

	if etag != "" {
		if err := s.cacheManager.SaveETag(s.llmsFullUrl, etag); err != nil {
			return err
		}
	}

	return nil
}

// fetchCategoryMap은 llms.txt를 가져와서 URL → Category 매핑을 생성합니다
func (s *TDSSearcher) fetchCategoryMap(ctx context.Context) map[string]string {
	content, _, err := httputil.FetchWithETag(ctx, s.llmsUrl, 0)
	if err != nil {
		return nil
	}

	parser := llms.NewParser()
	llmsTxt, err := parser.Parse(content)
	if err != nil {
		return nil
	}

	return BuildCategoryMapWithURLTransform(llmsTxt, tdsURLTransform)
}

func (s *TDSSearcher) indexFromTdsContent(content string, categoryMap map[string]string) error {
	parsedDocs := ParseTdsLlmsFull(content)

	var documents []IndexDocument
	for _, doc := range parsedDocs {
		// 카테고리 맵에서 먼저 찾고, 없으면 URL 경로에서 추출
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

	return s.indexManager.IndexDocuments(documents)
}

// Search는 TDS 문서를 검색합니다
func (s *TDSSearcher) Search(ctx context.Context, query string, opts *SearchOptions) ([]SearchResult, error) {
	return DoSearch(s.indexManager, query, opts)
}

// Close는 인덱스를 닫습니다
func (s *TDSSearcher) Close() error {
	return s.indexManager.Close()
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
	// base URL 제거
	path := url
	if len(url) > len(tdsBaseURL) {
		path = url[len(tdsBaseURL):]
	}

	// 경로에서 카테고리 추출
	parts := splitPath(path)
	if len(parts) == 0 {
		return "TDS"
	}

	var categories []string
	for i, part := range parts {
		if i >= 2 { // 첫 두 부분만 카테고리로 사용
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

// splitPath는 URL 경로를 분리합니다
func splitPath(path string) []string {
	var parts []string
	for _, part := range strings.Split(path, "/") {
		if part != "" {
			parts = append(parts, part)
		}
	}
	return parts
}

// formatPathPart는 kebab-case 경로를 Title Case로 변환합니다
// 예: "tds-react-native" → "Tds React Native", "components" → "Components"
func formatPathPart(part string) string {
	words := strings.Split(part, "-")
	for i, word := range words {
		if len(word) > 0 && word[0] >= 'a' && word[0] <= 'z' {
			words[i] = strings.ToUpper(string(word[0])) + word[1:]
		}
	}
	return strings.Join(words, " ")
}
