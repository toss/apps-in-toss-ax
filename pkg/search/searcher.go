package search

import (
	"context"
	"os"
	"path/filepath"

	"github.com/toss/apps-in-toss-ax/internal/httputil"
	"github.com/toss/apps-in-toss-ax/pkg/llms"
)

// ContentIndexer는 llms-full.txt 내용을 IndexDocument 슬라이스로 변환하는 함수 타입입니다
type ContentIndexer func(content string, categoryMap map[string]string) []IndexDocument

type SearchResult struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Content     string  `json:"content"`
	Description string  `json:"description"`
	URL         string  `json:"url"`
	Category    string  `json:"category"`
	Score       float64 `json:"score"`
}

type SearchOptions struct {
	Limit           int
	MaxContentLength int
}

type Searcher struct {
	llmsFullUrl  string
	llmsUrl      string
	cacheManager *CacheManager
	indexManager  *IndexManager
	indexer      ContentIndexer
	urlTransform URLTransformFunc
}

func newSearcher(llmsFullUrl, llmsUrl string, cacheConfig CacheConfig, indexer ContentIndexer, urlTransform URLTransformFunc) (*Searcher, error) {
	cacheManager, err := NewCacheManagerWithConfig(cacheConfig)
	if err != nil {
		return nil, err
	}

	indexManager := NewIndexManager(cacheManager.IndexPath())

	return &Searcher{
		llmsFullUrl:  llmsFullUrl,
		llmsUrl:      llmsUrl,
		cacheManager: cacheManager,
		indexManager:  indexManager,
		indexer:      indexer,
		urlTransform: urlTransform,
	}, nil
}

func New() (*Searcher, error) {
	return newSearcher(
		llmsFullUrl,
		llmsUrl,
		CacheConfig{
			MetadataFileName: defaultMetadataFileName,
			IndexSubDir:      defaultIndexSubDir,
		},
		appsInTossIndexer,
		nil,
	)
}

func NewTDSSearcher() (*Searcher, error) {
	return newSearcher(
		tdsReactNativeLlmsFullUrl,
		tdsReactNativeLlmsUrl,
		CacheConfig{
			MetadataFileName: "tds-cache-metadata.json",
			IndexSubDir:      "tds-search-index",
		},
		tdsIndexer,
		tdsURLTransform,
	)
}

func NewTDSMobileSearcher() (*Searcher, error) {
	return newSearcher(
		tdsMobileLlmsFullUrl,
		tdsMobileLlmsUrl,
		CacheConfig{
			MetadataFileName: "tds-mobile-cache-metadata.json",
			IndexSubDir:      "tds-mobile-search-index",
		},
		tdsIndexer,
		tdsURLTransform,
	)
}

func (s *Searcher) EnsureIndex(ctx context.Context) error {
	indexExists := s.cacheManager.IndexExists()

	if indexExists {
		etag, changed, err := s.cacheManager.CheckETag(ctx, s.llmsFullUrl)
		if err != nil {
			if err := s.indexManager.OpenIndex(); err == nil {
				return nil
			}
			return s.buildIndex(ctx)
		}

		if !changed {
			if err := s.indexManager.OpenIndex(); err == nil {
				return nil
			}
			_ = s.cacheManager.DeleteIndex()
			return s.buildIndex(ctx)
		}

		if err := s.cacheManager.DeleteIndex(); err != nil {
			return err
		}

		return s.buildIndexWithETag(ctx, etag)
	}

	return s.buildIndex(ctx)
}

func (s *Searcher) buildIndex(ctx context.Context) error {
	return s.buildIndexWithETag(ctx, "")
}

func (s *Searcher) buildIndexWithETag(ctx context.Context, etag string) error {
	content, newETag, err := httputil.FetchWithETag(ctx, s.llmsFullUrl, 0)
	if err != nil {
		return err
	}

	if newETag != "" {
		etag = newETag
	}

	categoryMap := s.fetchCategoryMap(ctx)

	if err := s.indexManager.CreateIndex(); err != nil {
		return err
	}

	documents := s.indexer(content, categoryMap)
	if err := s.indexManager.IndexDocuments(documents); err != nil {
		return err
	}

	if etag != "" {
		if err := s.cacheManager.SaveETag(s.llmsFullUrl, etag); err != nil {
			return err
		}
	}

	return nil
}

func (s *Searcher) fetchCategoryMap(ctx context.Context) map[string]string {
	content, _, err := httputil.FetchWithETag(ctx, s.llmsUrl, 0)
	if err != nil {
		return nil
	}

	parser := llms.NewParser()
	llmsTxt, err := parser.Parse(content)
	if err != nil {
		return nil
	}

	return BuildCategoryMapWithURLTransform(llmsTxt, s.urlTransform)
}

const defaultMaxContentLength = 500

func (s *Searcher) Search(ctx context.Context, query string, opts *SearchOptions) ([]SearchResult, error) {
	limit := 10
	maxContentLen := defaultMaxContentLength
	if opts != nil {
		if opts.Limit > 0 {
			limit = opts.Limit
		}
		if opts.MaxContentLength > 0 {
			maxContentLen = opts.MaxContentLength
		}
	}

	docs, scores, err := s.indexManager.Search(query, limit)
	if err != nil {
		return nil, err
	}

	results := make([]SearchResult, len(docs))
	for i, doc := range docs {
		results[i] = SearchResult{
			ID:          doc.ID,
			Title:       doc.Title,
			Content:     truncateContent(doc.Content, maxContentLen),
			Description: doc.Description,
			URL:         doc.URL,
			Category:    doc.Category,
			Score:       scores[i],
		}
	}

	return results, nil
}

// truncateContent는 콘텐츠를 maxLen 룬 이하로 잘라냅니다.
func truncateContent(content string, maxLen int) string {
	runes := []rune(content)
	if len(runes) <= maxLen {
		return content
	}
	return string(runes[:maxLen]) + "..."
}

func (s *Searcher) GetDocument(ctx context.Context, id string) (*SearchResult, error) {
	doc, err := s.indexManager.GetByID(id)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, nil
	}

	return &SearchResult{
		ID:          doc.ID,
		Title:       doc.Title,
		Content:     doc.Content,
		Description: doc.Description,
		URL:         doc.URL,
		Category:    doc.Category,
	}, nil
}

func (s *Searcher) Close() error {
	return s.indexManager.Close()
}

// NewTestSearcher는 테스트용 Searcher를 생성합니다.
// 실제 HTTP 호출 없이 인덱스 파일을 미리 생성하여
// EnsureIndex에서 OpenIndex가 성공하도록 합니다.
func NewTestSearcher() (*Searcher, error) {
	tempDir, err := os.MkdirTemp("", "test-searcher-*")
	if err != nil {
		return nil, err
	}

	indexPath := filepath.Join(tempDir, "test-index")

	// 인덱스를 생성한 뒤 닫아서 파일만 남겨둔다.
	// EnsureIndex → CheckETag 실패 → OpenIndex 경로로 진입하여 성공한다.
	im := NewIndexManager(indexPath)
	if err := im.CreateIndex(); err != nil {
		return nil, err
	}
	im.Close()

	cm := &CacheManager{
		cacheDir:     tempDir,
		metadataPath: filepath.Join(tempDir, "test-metadata.json"),
		indexPath:    indexPath,
	}

	return &Searcher{
		llmsFullUrl:  "https://test.invalid/llms-full.txt",
		cacheManager: cm,
		indexManager:  NewIndexManager(indexPath),
		indexer:      appsInTossIndexer,
	}, nil
}
