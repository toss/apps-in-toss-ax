package docs

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"

	"github.com/toss/apps-in-toss-ax/pkg/llms"
)

const (
	llmsUrl           = "https://developers-apps-in-toss.toss.im/llms.txt"
	llmsFullUrl       = "https://developers-apps-in-toss.toss.im/llms-full.txt"
	examplesUrl       = "https://developers-apps-in-toss.toss.im/tutorials/examples.md"
	tdsReactNativeUrl = "https://tossmini-docs.toss.im/tds-react-native/llms-full.txt"
	tdsWebUrl         = "https://tossmini-docs.toss.im/tds-mobile/llms-full.txt"
)

type AppsInTossDocs struct{}

func New() *AppsInTossDocs {
	return &AppsInTossDocs{}
}

func (docs *AppsInTossDocs) GetLLMsRoot(ctx context.Context) ([]LlmDocument, error) {
	llmsRoot, err := llms.NewReader().ReadLlms(ctx, llmsUrl)
	if err != nil {
		return nil, err
	}
	return flattenToDocuments(llmsRoot.Sections), nil
}

func (docs *AppsInTossDocs) GetLLMsFull(ctx context.Context) (*llms.LlmsTxt, error) {
	return llms.NewReader().ReadLlms(ctx, llmsFullUrl)
}

func (docs *AppsInTossDocs) GetExamples(ctx context.Context) ([]LlmDocument, error) {
	examples, err := llms.NewReader().ReadLlms(ctx, examplesUrl)
	if err != nil {
		return nil, err
	}
	return flattenToDocuments(examples.Sections), nil
}

func (docs *AppsInTossDocs) GetTDSReactNative(ctx context.Context) (*llms.LlmsTxt, error) {
	return llms.NewReader().ReadLlms(ctx, tdsReactNativeUrl)
}

func (docs *AppsInTossDocs) GetTDSWeb(ctx context.Context) (*llms.LlmsTxt, error) {
	return llms.NewReader().ReadLlms(ctx, tdsWebUrl)
}

func (docs *AppsInTossDocs) GetDocument(ctx context.Context, docId string) (string, error) {
	docsList, err := docs.GetLLMsRoot(ctx)

	if err != nil {
		return "", err
	}

	for _, doc := range docsList {

		if doc.ID == docId {

			return llms.NewReader().ReadLlmsRaw(ctx, doc.URL)
		}
	}

	return "", errors.New("document not found")
}

func (docs *AppsInTossDocs) GetExample(ctx context.Context, exampleId string) (string, error) {
	examples, err := docs.GetExamples(ctx)
	if err != nil {
		return "", err
	}
	for _, example := range examples {
		if example.ID == exampleId {
			return llms.NewReader().ReadLlmsRaw(ctx, example.URL)
		}
	}

	return "", errors.New("example not found")
}

type LlmDocument struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	URL      string `json:"url,omitempty"`
	Category string `json:"category,omitempty"`
}

func flattenToDocuments(sections []llms.Section) []LlmDocument {
	var docs []LlmDocument
	for _, s := range sections {
		docs = append(docs, flattenSection(s, "")...)
	}
	return docs
}

func flattenSection(s llms.Section, parentCategory string) []LlmDocument {
	category := s.Title
	if parentCategory != "" {
		category = parentCategory + " > " + s.Title
	}

	var docs []LlmDocument
	for _, link := range s.Links {
		doc := LlmDocument{
			Title:    link.Title,
			Content:  link.Description,
			URL:      link.URL,
			Category: category,
		}
		doc.ID = generateID(doc)
		docs = append(docs, doc)
	}

	for _, child := range s.Children {
		docs = append(docs, flattenSection(child, category)...)
	}
	return docs
}

// documentKey는 ID 생성에 사용되는 필드만 포함 (멱등성 보장)
type documentKey struct {
	Title    string `json:"title"`
	URL      string `json:"url,omitempty"`
	Category string `json:"category,omitempty"`
}

func generateID(doc LlmDocument) string {
	key := documentKey{
		Title:    doc.Title,
		URL:      doc.URL,
		Category: doc.Category,
	}
	data, _ := json.Marshal(key)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:8])
}
