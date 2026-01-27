package search

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/toss/apps-in-toss-ax/internal/httputil"
)

const (
	cacheSubDir = "ax"

	// 기본 캐시 설정 (AppsInToss 문서용)
	defaultMetadataFileName = "cache-metadata.json"
	defaultIndexSubDir      = "search-index"
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

// CacheConfig는 CacheManager 설정입니다
type CacheConfig struct {
	MetadataFileName string
	IndexSubDir      string
}

// NewCacheManager는 기본 설정으로 CacheManager를 생성합니다
func NewCacheManager() (*CacheManager, error) {
	return NewCacheManagerWithConfig(CacheConfig{
		MetadataFileName: defaultMetadataFileName,
		IndexSubDir:      defaultIndexSubDir,
	})
}

// NewCacheManagerWithConfig는 설정을 지정하여 CacheManager를 생성합니다
func NewCacheManagerWithConfig(config CacheConfig) (*CacheManager, error) {
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
		metadataPath: filepath.Join(cacheDir, config.MetadataFileName),
		indexPath:    filepath.Join(cacheDir, config.IndexSubDir),
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

	return httputil.CheckETag(ctx, url, cachedETag)
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
