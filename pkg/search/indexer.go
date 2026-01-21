package search

import (
	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/analysis/analyzer/custom"
	"github.com/blevesearch/bleve/v2/analysis/lang/cjk"
	"github.com/blevesearch/bleve/v2/analysis/token/edgengram"
	"github.com/blevesearch/bleve/v2/analysis/token/lowercase"
	"github.com/blevesearch/bleve/v2/analysis/tokenizer/unicode"
	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/toss/apps-in-toss-ax/pkg/docid"
	"github.com/toss/apps-in-toss-ax/pkg/llms"
)

const cjkAnalyzerName = "cjk_analyzer"

type IndexDocument struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Content     string `json:"content"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Category    string `json:"category"`
}

type IndexManager struct {
	indexPath string
	index     bleve.Index
}

func NewIndexManager(indexPath string) *IndexManager {
	return &IndexManager{
		indexPath: indexPath,
	}
}

func (im *IndexManager) createIndexMapping() (mapping.IndexMapping, error) {
	indexMapping := bleve.NewIndexMapping()

	// Edge N-gram token filter 등록 (1-10글자)
	err := indexMapping.AddCustomTokenFilter("ko_edgengram", map[string]interface{}{
		"type": edgengram.Name,
		"min":  1.0,
		"max":  10.0,
		"side": "front",
	})
	if err != nil {
		return nil, err
	}

	// 인덱싱용 analyzer (cjk_bigram + edgengram)
	err = indexMapping.AddCustomAnalyzer(cjkAnalyzerName, map[string]interface{}{
		"type":      custom.Name,
		"tokenizer": unicode.Name,
		"token_filters": []string{
			lowercase.Name,
			cjk.BigramName,
			"ko_edgengram",
		},
	})
	if err != nil {
		return nil, err
	}

	// 검색용 analyzer (edgengram 제외, cjk_bigram만)
	err = indexMapping.AddCustomAnalyzer("cjk_search", map[string]interface{}{
		"type":      custom.Name,
		"tokenizer": unicode.Name,
		"token_filters": []string{
			lowercase.Name,
			cjk.BigramName,
		},
	})
	if err != nil {
		return nil, err
	}

	docMapping := bleve.NewDocumentMapping()

	textFieldMapping := bleve.NewTextFieldMapping()
	textFieldMapping.Analyzer = cjkAnalyzerName

	docMapping.AddFieldMappingsAt("title", textFieldMapping)
	docMapping.AddFieldMappingsAt("content", textFieldMapping)
	docMapping.AddFieldMappingsAt("description", textFieldMapping)
	docMapping.AddFieldMappingsAt("category", textFieldMapping)

	keywordMapping := bleve.NewTextFieldMapping()
	keywordMapping.Analyzer = "keyword"
	docMapping.AddFieldMappingsAt("url", keywordMapping)

	indexMapping.AddDocumentMapping("document", docMapping)
	indexMapping.DefaultMapping = docMapping

	return indexMapping, nil
}

func (im *IndexManager) CreateIndex() error {
	indexMapping, err := im.createIndexMapping()
	if err != nil {
		return err
	}

	index, err := bleve.New(im.indexPath, indexMapping)
	if err != nil {
		return err
	}

	im.index = index
	return nil
}

func (im *IndexManager) OpenIndex() error {
	index, err := bleve.Open(im.indexPath)
	if err != nil {
		return err
	}

	im.index = index
	return nil
}

func (im *IndexManager) IndexDocuments(documents []IndexDocument) error {
	batch := im.index.NewBatch()

	for _, doc := range documents {
		if err := batch.Index(doc.ID, doc); err != nil {
			return err
		}
	}

	return im.index.Batch(batch)
}

func (im *IndexManager) IndexFromLlmsTxt(llmsTxt *llms.LlmsTxt) error {
	var documents []IndexDocument

	links := llmsTxt.GetAllLinks()
	for i, link := range links {
		doc := IndexDocument{
			ID:          generateDocID(i, link.URL),
			Title:       link.Title,
			Content:     "",
			Description: link.Description,
			URL:         link.URL,
			Category:    "",
		}
		documents = append(documents, doc)
	}

	indexDocsFromSections(llmsTxt.Sections, "", &documents)

	return im.IndexDocuments(documents)
}

// IndexFromLlmsFullContent는 llms-full.txt 내용을 직접 파싱하여 인덱싱합니다
// categoryMap은 URL → Category 매핑입니다 (llms.txt에서 생성)
func (im *IndexManager) IndexFromLlmsFullContent(content string, categoryMap map[string]string) error {
	parsedDocs := ParseLlmsFull(content)

	var documents []IndexDocument
	for _, doc := range parsedDocs {
		category := ""
		if categoryMap != nil {
			category = categoryMap[doc.URL]
		}

		documents = append(documents, IndexDocument{
			ID:       docid.Generate(doc.Title, doc.URL, category),
			Title:    doc.Title,
			Content:  doc.Content,
			URL:      doc.URL,
			Category: category,
		})
	}

	return im.IndexDocuments(documents)
}

// BuildCategoryMap은 llms.txt 파싱 결과에서 URL → Category 매핑을 생성합니다
func BuildCategoryMap(llmsTxt *llms.LlmsTxt) map[string]string {
	categoryMap := make(map[string]string)
	buildCategoryMapFromSections(llmsTxt.Sections, "", categoryMap)
	return categoryMap
}

func buildCategoryMapFromSections(sections []llms.Section, parentCategory string, categoryMap map[string]string) {
	for _, section := range sections {
		category := section.Title
		if parentCategory != "" {
			category = parentCategory + " > " + section.Title
		}

		for _, link := range section.Links {
			categoryMap[link.URL] = category
		}

		if len(section.Children) > 0 {
			buildCategoryMapFromSections(section.Children, category, categoryMap)
		}
	}
}

func indexDocsFromSections(sections []llms.Section, parentCategory string, documents *[]IndexDocument) {
	for _, section := range sections {
		category := section.Title
		if parentCategory != "" {
			category = parentCategory + " > " + section.Title
		}

		for i, link := range section.Links {
			found := false
			for j := range *documents {
				if (*documents)[j].URL == link.URL {
					(*documents)[j].Category = category
					found = true
					break
				}
			}
			if !found {
				doc := IndexDocument{
					ID:          generateDocID(len(*documents)+i, link.URL),
					Title:       link.Title,
					Description: link.Description,
					URL:         link.URL,
					Category:    category,
				}
				*documents = append(*documents, doc)
			}
		}

		if len(section.Children) > 0 {
			indexDocsFromSections(section.Children, category, documents)
		}
	}
}

func generateDocID(index int, url string) string {
	return url
}

func (im *IndexManager) Search(query string, limit int) ([]IndexDocument, []float64, error) {
	// 여러 필드에서 검색하기 위해 DisjunctionQuery 사용
	// 검색용 analyzer 사용 (edgengram 제외)
	titleQuery := bleve.NewMatchQuery(query)
	titleQuery.SetField("title")
	titleQuery.Analyzer = "cjk_search"
	titleQuery.SetBoost(5.0)

	descQuery := bleve.NewMatchQuery(query)
	descQuery.SetField("description")
	descQuery.Analyzer = "cjk_search"
	descQuery.SetAutoFuzziness(true)
	descQuery.SetPrefix(1)
	descQuery.SetBoost(1.5)

	contentQuery := bleve.NewMatchQuery(query)
	contentQuery.SetField("content")
	contentQuery.Analyzer = "cjk_search"
	contentQuery.SetAutoFuzziness(true)
	contentQuery.SetPrefix(1)
	contentQuery.SetBoost(1.0)

	categoryQuery := bleve.NewMatchQuery(query)
	categoryQuery.SetField("category")
	categoryQuery.Analyzer = "cjk_search"
	categoryQuery.SetBoost(1.0)

	searchQuery := bleve.NewDisjunctionQuery(titleQuery, descQuery, contentQuery, categoryQuery)
	searchRequest := bleve.NewSearchRequestOptions(searchQuery, limit, 0, false)
	searchRequest.Fields = []string{"title", "content", "description", "url", "category"}

	searchResult, err := im.index.Search(searchRequest)
	if err != nil {
		return nil, nil, err
	}

	var results []IndexDocument
	var scores []float64

	for _, hit := range searchResult.Hits {
		doc := IndexDocument{
			ID: hit.ID,
		}

		if title, ok := hit.Fields["title"].(string); ok {
			doc.Title = title
		}
		if content, ok := hit.Fields["content"].(string); ok {
			doc.Content = content
		}
		if description, ok := hit.Fields["description"].(string); ok {
			doc.Description = description
		}
		if url, ok := hit.Fields["url"].(string); ok {
			doc.URL = url
		}
		if category, ok := hit.Fields["category"].(string); ok {
			doc.Category = category
		}

		results = append(results, doc)
		scores = append(scores, hit.Score)
	}

	return results, scores, nil
}

func (im *IndexManager) Close() error {
	if im.index != nil {
		return im.index.Close()
	}
	return nil
}
