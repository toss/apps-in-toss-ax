package mcp

import (
	"strings"

	"github.com/toss/apps-in-toss-ax/pkg/search"
)

// classifyPlatform determines the platform of a document based on title and URL.
// Returns "web", "rn", or "common".
func classifyPlatform(r search.SearchResult) string {
	title := strings.ToLower(r.Title)
	url := strings.ToLower(r.URL)

	// Check golden file overrides first
	if platform, ok := platformOverrides[url]; ok {
		return platform
	}

	// Rule-based classification
	hasRN := strings.Contains(title, "react native") ||
		strings.HasPrefix(url, "https://developers-apps-in-toss.toss.im/development/client/") ||
		strings.Contains(url, "/rn-") || strings.HasSuffix(url, "-rn.md")
	hasWeb := strings.Contains(title, "webview") && !hasRN

	if hasRN && !hasWeb {
		return "rn"
	}
	if hasWeb && !hasRN {
		return "web"
	}
	return "common"
}

// filterByPlatform removes documents that belong exclusively to the opposite platform.
// - platform="web" → exclude "rn" docs (keep "common" + "web")
// - platform="rn"  → exclude "web" docs (keep "common" + "rn")
func filterByPlatform(results []search.SearchResult, platform string) []search.SearchResult {
	filtered := make([]search.SearchResult, 0, len(results))
	for _, r := range results {
		docPlatform := classifyPlatform(r)
		if platform == "web" && docPlatform == "rn" {
			continue
		}
		if platform == "rn" && docPlatform == "web" {
			continue
		}
		filtered = append(filtered, r)
	}
	return filtered
}

// platformOverrides is a golden file for documents that are misclassified by rules.
// Key: lowercase URL, Value: "web" | "rn" | "common"
var platformOverrides = map[string]string{
	// WebView props is common (RN also uses WebView)
	"https://developers-apps-in-toss.toss.im/bedrock/reference/framework/속성 제어/webview-props.md": "common",
	// Banner ads are platform-specific (keys must be lowercase to match strings.ToLower)
	"https://developers-apps-in-toss.toss.im/bedrock/reference/framework/광고/bannerad.md":    "web",
	"https://developers-apps-in-toss.toss.im/bedrock/reference/framework/광고/rn-bannerad.md": "rn",
	// Environment variables (React Native)
	"https://developers-apps-in-toss.toss.im/bedrock/reference/framework/환경 변수/env.md": "rn",
	// Getting started tutorials
	"https://developers-apps-in-toss.toss.im/tutorials/webview.md":       "web",
	"https://developers-apps-in-toss.toss.im/tutorials/react-native.md": "rn",
	// Client setup (Android/iOS) is RN-only
	"https://developers-apps-in-toss.toss.im/development/client/android.md": "rn",
	"https://developers-apps-in-toss.toss.im/development/client/ios.md":     "rn",
}
