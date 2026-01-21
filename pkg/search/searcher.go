package search

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/toss/apps-in-toss-ax/pkg/llms"
)

const (
	llmsUrl        = "https://developers-apps-in-toss.toss.im/llms.txt"
	llmsFullUrl    = "https://developers-apps-in-toss.toss.im/llms-full.txt"
	defaultTimeout = 30 * time.Second
)

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
	Limit int
}

type Searcher struct {
	cacheManager *CacheManager
	indexManager *IndexManager
}

func New() (*Searcher, error) {
	cacheManager, err := NewCacheManager()
	if err != nil {
		return nil, err
	}

	indexManager := NewIndexManager(cacheManager.IndexPath())

	return &Searcher{
		cacheManager: cacheManager,
		indexManager: indexManager,
	}, nil
}

func (s *Searcher) EnsureIndex(ctx context.Context) error {
	indexExists := s.cacheManager.IndexExists()

	if indexExists {
		etag, changed, err := s.cacheManager.CheckETag(ctx, llmsFullUrl)
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

func (s *Searcher) rebuildIndex(ctx context.Context) error {
	etag, _, err := s.cacheManager.CheckETag(ctx, llmsFullUrl)
	if err != nil {
		etag = ""
	}
	return s.buildIndex(ctx, etag)
}

func (s *Searcher) buildIndex(ctx context.Context, etag string) error {
	// llms-full.txt 가져오기
	content, newETag, err := s.fetchWithETag(ctx, llmsFullUrl)
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

	// llms-full.txt 형식을 직접 파싱하여 인덱싱 (카테고리 매핑 포함)
	if err := s.indexManager.IndexFromLlmsFullContent(content, categoryMap); err != nil {
		return err
	}

	if etag != "" {
		if err := s.cacheManager.SaveETag(llmsFullUrl, etag); err != nil {
			return err
		}
	}

	return nil
}

// fetchCategoryMap은 llms.txt를 가져와서 URL → Category 매핑을 생성합니다
func (s *Searcher) fetchCategoryMap(ctx context.Context) map[string]string {
	content, _, err := s.fetchWithETag(ctx, llmsUrl)
	if err != nil {
		return nil
	}

	parser := llms.NewParser()
	llmsTxt, err := parser.Parse(content)
	if err != nil {
		return nil
	}

	return BuildCategoryMap(llmsTxt)
}

func (s *Searcher) fetchWithETag(ctx context.Context, url string) (content string, etag string, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", "", err
	}

	client := &http.Client{Timeout: defaultTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", &FetchError{URL: url, StatusCode: resp.StatusCode}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	return string(body), resp.Header.Get("ETag"), nil
}

func (s *Searcher) Search(ctx context.Context, query string, opts *SearchOptions) ([]SearchResult, error) {
	limit := 10
	if opts != nil && opts.Limit > 0 {
		limit = opts.Limit
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
			Content:     doc.Content,
			Description: doc.Description,
			URL:         doc.URL,
			Category:    doc.Category,
			Score:       scores[i],
		}
	}

	return results, nil
}

func (s *Searcher) Close() error {
	return s.indexManager.Close()
}

type FetchError struct {
	URL        string
	StatusCode int
}

func (e *FetchError) Error() string {
	return "fetch error: " + e.URL
}
