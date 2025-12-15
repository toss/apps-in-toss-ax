package llms

import (
	"bytes"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// LlmsTxt represents the parsed structure of an llms.txt file
type LlmsTxt struct {
	Title    string    `json:"title"`
	Summary  string    `json:"summary"`
	Sections []Section `json:"sections"`
}

// Section represents a section in the llms.txt file
type Section struct {
	Title    string    `json:"title"`
	Level    int       `json:"level"` // 2 for ##, 3 for ###, etc.
	Links    []Link    `json:"links"`
	Children []Section `json:"children,omitempty"`
}

// Link represents a link entry in a section
type Link struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
}

// Parser parses llms.txt content using goldmark
type Parser struct {
	md goldmark.Markdown
}

// NewParser creates a new llms.txt parser
func NewParser() *Parser {
	return &Parser{
		md: goldmark.New(),
	}
}

// Parse parses the llms.txt content and returns the structured representation
func (p *Parser) Parse(content string) (*LlmsTxt, error) {
	source := []byte(content)
	reader := text.NewReader(source)
	doc := p.md.Parser().Parse(reader)

	result := &LlmsTxt{
		Sections: []Section{},
	}

	var sectionStack []*Section

	// Walk through the AST
	err := ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch n := node.(type) {
		case *ast.Heading:
			title := p.extractText(n, source)

			if n.Level == 1 {
				// Title
				result.Title = title
			} else {
				// Section (## or deeper)
				newSection := Section{
					Title:    title,
					Level:    n.Level,
					Links:    []Link{},
					Children: []Section{},
				}

				// Pop sections with same or higher level
				for len(sectionStack) > 0 && sectionStack[len(sectionStack)-1].Level >= n.Level {
					sectionStack = sectionStack[:len(sectionStack)-1]
				}

				if len(sectionStack) == 0 {
					// Top-level section
					result.Sections = append(result.Sections, newSection)
					sectionStack = append(sectionStack, &result.Sections[len(result.Sections)-1])
				} else {
					// Child section
					parent := sectionStack[len(sectionStack)-1]
					parent.Children = append(parent.Children, newSection)
					sectionStack = append(sectionStack, &parent.Children[len(parent.Children)-1])
				}
			}

		case *ast.Blockquote:
			// Summary (> text)
			if result.Summary == "" {
				result.Summary = strings.TrimSpace(p.extractText(n, source))
			}

		case *ast.ListItem:
			// Parse link from list item
			link := p.extractLink(n, source)
			if link != nil && len(sectionStack) > 0 {
				currentSection := sectionStack[len(sectionStack)-1]
				currentSection.Links = append(currentSection.Links, *link)
			}
			return ast.WalkSkipChildren, nil
		}

		return ast.WalkContinue, nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// extractText extracts plain text from a node
func (p *Parser) extractText(node ast.Node, source []byte) string {
	var buf bytes.Buffer
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		p.extractTextRecursive(child, source, &buf)
	}
	return strings.TrimSpace(buf.String())
}

// extractTextRecursive recursively extracts text from nodes
func (p *Parser) extractTextRecursive(node ast.Node, source []byte, buf *bytes.Buffer) {
	switch n := node.(type) {
	case *ast.Text:
		buf.Write(n.Segment.Value(source))
	case *ast.String:
		buf.Write(n.Value)
	default:
		for child := node.FirstChild(); child != nil; child = child.NextSibling() {
			p.extractTextRecursive(child, source, buf)
		}
	}
}

// extractLink extracts a Link from a list item node
func (p *Parser) extractLink(listItem ast.Node, source []byte) *Link {
	var link *Link
	var description strings.Builder

	// Find the TextBlock or Paragraph inside the list item
	for child := listItem.FirstChild(); child != nil; child = child.NextSibling() {
		// Check for both TextBlock and Paragraph
		var container ast.Node
		if _, ok := child.(*ast.TextBlock); ok {
			container = child
		} else if _, ok := child.(*ast.Paragraph); ok {
			container = child
		}

		if container != nil {
			foundLink := false

			for pChild := container.FirstChild(); pChild != nil; pChild = pChild.NextSibling() {
				switch n := pChild.(type) {
				case *ast.Link:
					if !foundLink {
						link = &Link{
							Title: p.extractText(n, source),
							URL:   string(n.Destination),
						}
						foundLink = true
					}
				case *ast.Text:
					if foundLink {
						text := string(n.Segment.Value(source))
						// Remove leading colon and space
						text = strings.TrimLeft(text, ": ")
						if text != "" {
							if description.Len() > 0 {
								description.WriteString(" ")
							}
							description.WriteString(text)
						}
					}
				}
			}

			if link != nil {
				link.Description = strings.TrimSpace(description.String())
			}
		}
	}

	return link
}

// GetAllLinks returns all links from the parsed llms.txt as a flat list
func (l *LlmsTxt) GetAllLinks() []Link {
	var links []Link
	for _, section := range l.Sections {
		links = append(links, section.getAllLinks()...)
	}
	return links
}

// getAllLinks recursively gets all links from a section and its children
func (s *Section) getAllLinks() []Link {
	var links []Link
	links = append(links, s.Links...)
	for _, child := range s.Children {
		links = append(links, child.getAllLinks()...)
	}
	return links
}

// FindSection finds a section by title (case-insensitive)
func (l *LlmsTxt) FindSection(title string) *Section {
	lowerTitle := strings.ToLower(title)
	for i := range l.Sections {
		if found := l.Sections[i].findSection(lowerTitle); found != nil {
			return found
		}
	}
	return nil
}

// findSection recursively finds a section by title
func (s *Section) findSection(lowerTitle string) *Section {
	if strings.ToLower(s.Title) == lowerTitle {
		return s
	}
	for i := range s.Children {
		if found := s.Children[i].findSection(lowerTitle); found != nil {
			return found
		}
	}
	return nil
}

// GetSectionTitles returns all section titles as a flat list
func (l *LlmsTxt) GetSectionTitles() []string {
	var titles []string
	for _, section := range l.Sections {
		titles = append(titles, section.getSectionTitles()...)
	}
	return titles
}

// getSectionTitles recursively gets all section titles
func (s *Section) getSectionTitles() []string {
	titles := []string{s.Title}
	for _, child := range s.Children {
		titles = append(titles, child.getSectionTitles()...)
	}
	return titles
}
