package llms

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

const defaultTimeout = 10 * time.Second

// FetchError는 HTTP 요청 실패를 나타내는 에러입니다
type FetchError struct {
	URL        string
	StatusCode int
	Err        error
}

func (e *FetchError) Error() string {
	if e.StatusCode != 0 {
		return fmt.Sprintf("fetch %s: status %d", e.URL, e.StatusCode)
	}
	return fmt.Sprintf("fetch %s: %v", e.URL, e.Err)
}

func (e *FetchError) Unwrap() error { return e.Err }

type Reader struct {
	httpClient *http.Client
	parser     *Parser
}

func NewReader() *Reader {
	return &Reader{
		httpClient: &http.Client{Timeout: defaultTimeout},
		parser:     NewParser(),
	}
}

func (r *Reader) fetch(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", &FetchError{URL: url, Err: err}
	}

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return "", &FetchError{URL: url, Err: err}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", &FetchError{URL: url, StatusCode: resp.StatusCode}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", &FetchError{URL: url, Err: err}
	}

	return string(body), nil
}

func (r *Reader) fetchAndParse(ctx context.Context, url string) (*LlmsTxt, error) {
	content, err := r.fetch(ctx, url)
	if err != nil {
		return nil, err
	}
	return r.parser.Parse(content)
}

// ReadLlms는 llms.txt를 가져와서 파싱합니다
func (r *Reader) ReadLlms(ctx context.Context, url string) (*LlmsTxt, error) {
	return r.fetchAndParse(ctx, url)
}

func (r *Reader) ReadLlmsRaw(ctx context.Context, url string) (string, error) {
	return r.fetch(ctx, url)
}
