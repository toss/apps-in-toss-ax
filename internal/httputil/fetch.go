package httputil

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

const DefaultTimeout = 30 * time.Second

// FetchError는 HTTP 요청 실패 시 반환되는 에러입니다
type FetchError struct {
	URL        string
	StatusCode int
	Err        error
}

func (e *FetchError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("fetch error: %s: %v", e.URL, e.Err)
	}
	return fmt.Sprintf("fetch error: %s (status %d)", e.URL, e.StatusCode)
}

func (e *FetchError) Unwrap() error {
	return e.Err
}

// FetchWithETag는 URL에서 콘텐츠를 가져오고 ETag를 반환합니다
func FetchWithETag(ctx context.Context, url string, timeout time.Duration) (content string, etag string, err error) {
	if timeout == 0 {
		timeout = DefaultTimeout
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", "", &FetchError{URL: url, Err: err}
	}

	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", &FetchError{URL: url, Err: err}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", &FetchError{URL: url, StatusCode: resp.StatusCode}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", &FetchError{URL: url, Err: err}
	}

	return string(body), resp.Header.Get("ETag"), nil
}

// CheckETag는 URL의 ETag가 변경되었는지 확인합니다
// cachedETag가 비어있으면 항상 changed=true를 반환합니다
func CheckETag(ctx context.Context, url string, cachedETag string) (etag string, changed bool, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return "", true, &FetchError{URL: url, Err: err}
	}

	if cachedETag != "" {
		req.Header.Set("If-None-Match", cachedETag)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", true, &FetchError{URL: url, Err: err}
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotModified {
		return cachedETag, false, nil
	}

	newETag := resp.Header.Get("ETag")
	return newETag, true, nil
}
