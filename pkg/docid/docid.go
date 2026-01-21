package docid

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

// DocumentKey는 ID 생성에 사용되는 필드 (멱등성 보장)
type DocumentKey struct {
	Title    string `json:"title"`
	URL      string `json:"url,omitempty"`
	Category string `json:"category,omitempty"`
}

// Generate는 문서의 고유 ID를 생성합니다
// Title, URL, Category를 JSON으로 직렬화한 후 SHA-256 해시의 첫 8바이트를 hex로 반환
func Generate(title, url, category string) string {
	key := DocumentKey{
		Title:    title,
		URL:      url,
		Category: category,
	}
	data, _ := json.Marshal(key)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:8])
}
