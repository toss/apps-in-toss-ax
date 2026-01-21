package search

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	cacheSubDir      = "ax"
	metadataFileName = "cache-metadata.json"
	indexSubDir      = "search-index"
)

type CacheMetadata struct {
	ETag        string `json:"etag"`
	LastFetched string `json:"last_fetched"`
	URL         string `json:"url"`
}

type CacheManager struct {
	cacheDir     string
	metadataPath string
	indexPath    string
}

func NewCacheManager() (*CacheManager, error) {
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}

	cacheDir := filepath.Join(userCacheDir, cacheSubDir)
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, err
	}

	return &CacheManager{
		cacheDir:     cacheDir,
		metadataPath: filepath.Join(cacheDir, metadataFileName),
		indexPath:    filepath.Join(cacheDir, indexSubDir),
	}, nil
}

func (cm *CacheManager) IndexPath() string {
	return cm.indexPath
}

func (cm *CacheManager) GetCachedETag() (string, error) {
	data, err := os.ReadFile(cm.metadataPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	var metadata CacheMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return "", nil
	}

	return metadata.ETag, nil
}

func (cm *CacheManager) SaveETag(url, etag string) error {
	metadata := CacheMetadata{
		ETag:        etag,
		LastFetched: time.Now().UTC().Format(time.RFC3339),
		URL:         url,
	}

	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cm.metadataPath, data, 0644)
}

func (cm *CacheManager) CheckETag(ctx context.Context, url string) (etag string, changed bool, err error) {
	cachedETag, err := cm.GetCachedETag()
	if err != nil {
		return "", true, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return "", true, err
	}

	if cachedETag != "" {
		req.Header.Set("If-None-Match", cachedETag)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", true, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotModified {
		return cachedETag, false, nil
	}

	newETag := resp.Header.Get("ETag")
	return newETag, true, nil
}

func (cm *CacheManager) IndexExists() bool {
	info, err := os.Stat(cm.indexPath)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func (cm *CacheManager) DeleteIndex() error {
	return os.RemoveAll(cm.indexPath)
}
