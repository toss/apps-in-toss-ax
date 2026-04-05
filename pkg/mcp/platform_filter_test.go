package mcp

import (
	"testing"

	"github.com/toss/apps-in-toss-ax/pkg/search"
)

func TestClassifyPlatform(t *testing.T) {
	tests := []struct {
		name     string
		result   search.SearchResult
		expected string
	}{
		{
			name:     "React Native in title → rn",
			result:   search.SearchResult{Title: "React Native", URL: "https://developers-apps-in-toss.toss.im/tutorials/react-native.md"},
			expected: "rn",
		},
		{
			name:     "WebView in title → web",
			result:   search.SearchResult{Title: "WebView", URL: "https://developers-apps-in-toss.toss.im/tutorials/webview.md"},
			expected: "web",
		},
		{
			name:     "Android client setup → rn",
			result:   search.SearchResult{Title: "Android 환경설정", URL: "https://developers-apps-in-toss.toss.im/development/client/android.md"},
			expected: "rn",
		},
		{
			name:     "iOS client setup → rn",
			result:   search.SearchResult{Title: "iOS 환경설정", URL: "https://developers-apps-in-toss.toss.im/development/client/ios.md"},
			expected: "rn",
		},
		{
			name:     "Storage API → common",
			result:   search.SearchResult{Title: "네이티브 저장소 이용하기", URL: "https://developers-apps-in-toss.toss.im/bedrock/reference/framework/저장소/Storage.md"},
			expected: "common",
		},
		{
			name:     "RN BannerAd golden override → rn",
			result:   search.SearchResult{Title: "인앱 광고", URL: "https://developers-apps-in-toss.toss.im/bedrock/reference/framework/광고/rn-bannerad.md"},
			expected: "rn",
		},
		{
			name:     "Web BannerAd golden override → web",
			result:   search.SearchResult{Title: "인앱 광고", URL: "https://developers-apps-in-toss.toss.im/bedrock/reference/framework/광고/BannerAd.md"},
			expected: "web",
		},
		{
			name:     "환경 변수 설정 (React Native) → rn",
			result:   search.SearchResult{Title: "환경 변수 설정 (React Native)", URL: "https://developers-apps-in-toss.toss.im/bedrock/reference/framework/환경 변수/env.md"},
			expected: "rn",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifyPlatform(tt.result)
			if got != tt.expected {
				t.Errorf("classifyPlatform(%q, %q) = %q, want %q", tt.result.Title, tt.result.URL, got, tt.expected)
			}
		})
	}
}

func TestFilterByPlatform(t *testing.T) {
	docs := []search.SearchResult{
		{Title: "네이티브 저장소 이용하기", URL: "https://developers-apps-in-toss.toss.im/bedrock/reference/framework/저장소/Storage.md"},     // common
		{Title: "React Native", URL: "https://developers-apps-in-toss.toss.im/tutorials/react-native.md"},                                      // rn
		{Title: "WebView", URL: "https://developers-apps-in-toss.toss.im/tutorials/webview.md"},                                                 // web
		{Title: "인앱 결제", URL: "https://developers-apps-in-toss.toss.im/bedrock/reference/framework/결제/payment.md"},                       // common
		{Title: "Android 환경설정", URL: "https://developers-apps-in-toss.toss.im/development/client/android.md"},                                // rn
	}

	t.Run("platform=web excludes rn-only docs", func(t *testing.T) {
		filtered := filterByPlatform(docs, "web")
		// Should keep: Storage(common), WebView(web), 인앱결제(common) = 3
		if len(filtered) != 3 {
			t.Errorf("expected 3 results, got %d", len(filtered))
			for _, d := range filtered {
				t.Logf("  kept: %s", d.Title)
			}
		}
		for _, d := range filtered {
			if d.Title == "React Native" || d.Title == "Android 환경설정" {
				t.Errorf("should not include RN doc: %s", d.Title)
			}
		}
	})

	t.Run("platform=rn excludes web-only docs", func(t *testing.T) {
		filtered := filterByPlatform(docs, "rn")
		// Should keep: Storage(common), React Native(rn), 인앱결제(common), Android(rn) = 4
		if len(filtered) != 4 {
			t.Errorf("expected 4 results, got %d", len(filtered))
			for _, d := range filtered {
				t.Logf("  kept: %s", d.Title)
			}
		}
		for _, d := range filtered {
			if d.Title == "WebView" {
				t.Errorf("should not include Web doc: %s", d.Title)
			}
		}
	})

	t.Run("no platform returns all", func(t *testing.T) {
		filtered := filterByPlatform(docs, "")
		if len(filtered) != 5 {
			t.Errorf("expected 5 results, got %d", len(filtered))
		}
	})
}
