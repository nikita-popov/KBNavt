package kb

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"strings"

	org "github.com/niklasfasching/go-org/org"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// Parser handles document parsing for different formats
type Parser struct {
	orgConfig *org.Configuration
	mdParser  goldmark.Markdown
}

// NewParser creates a new parser
func NewParser() *Parser {
	md := goldmark.New(
		goldmark.WithParser(
			parser.NewParser(
				parser.WithAutoHeadingID(),
			),
		),
	)

	return &Parser{
		orgConfig: org.New(),
		mdParser:  md,
	}
}

// ParseOrgMode parses an Org-mode document
func (p *Parser) ParseOrgMode(content string) (*Document, error) {
	doc := p.orgConfig.Parse(bytes.NewReader([]byte(content)), "")
	//if err != nil {
	//	return nil, fmt.Errorf("org parsing error: %w", err)
	//}

	headers := p.extractOrgHeaders(doc.Nodes)

	return &Document{
		Content: content,
		Headers: headers,
		Format:  FormatOrg,
	}, nil
}

// ParseMarkdown parses a Markdown document
func (p *Parser) ParseMarkdown(content string) (*Document, error) {
	src := []byte(content)
	// Corrected: use text.NewReader
	reader := text.NewReader(src)

	// Corrected: Parse takes reader and optional parsing options
	doc := p.mdParser.Parser().Parse(reader)

	headers := p.extractMarkdownHeaders(doc, content)

	return &Document{
		Content: content,
		Headers: headers,
		Format:  FormatMarkdown,
	}, nil
}

// ParseText parses a plain text document (minimal parsing)
func (p *Parser) ParseText(content string) (*Document, error) {
	// Extract headers from markdown-like syntax or treat as single section
	lines := strings.Split(content, "\n")
	var headers []Header

	for i, line := range lines {
		if strings.HasPrefix(line, "## ") || strings.HasPrefix(line, "# ") {
			headers = append(headers, Header{
				Title:   strings.TrimSpace(strings.TrimPrefix(line, "# ")),
				LineNum: i + 1,
			})
		}
	}

	return &Document{
		Content: content,
		Headers: headers,
		Format:  FormatText,
	}, nil
}

func (p *Parser) extractOrgHeaders(nodes []org.Node) []Header {
	var headers []Header

	for _, node := range nodes {
		if headline, ok := node.(*org.Headline); ok {
			header := Header{
				Level: headline.Lvl,
				Title: orgNodesToString(headline.Title),
				LineNum: headline.Index,
			}

			// Extract content under this header
			if len(headline.Children) > 0 {
				//contentBuf := new(bytes.Buffer)
				//for _, child := range headline.Children {
				//	if section, ok := child.(*org.Section); ok {
				//		contentBuf.WriteString(section.String())
				//	}
				//}
				header.Children = p.extractOrgHeaders(headline.Children)
				header.Content = orgNodesToString(headline.Children)
			}

			headers = append(headers, header)
		}
	}

	return headers
}

func (p *Parser) extractMarkdownHeaders(root ast.Node, content string) []Header {
	var headers []Header
	scanner := bufio.NewScanner(strings.NewReader(content))
	lineNum := 0

	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()
		lineNum++

		// Detect markdown headers
		if match := regexp.MustCompile(`^(#+)\s+(.*)$`).FindStringSubmatch(line); match != nil {
			level := len(match)
			title := match[0] // TODO: check

			headers = append(headers, Header{
				Level:   level,
				Title:   title,
				LineNum: lineNum,
			})
		}
	}

	return headers
}

// ReadSection reads a specific header section from a document
func (p *Parser) ReadSection(content string, format Format, headerTitle string) (string, error) {
    var headers []Header

    switch format {
    case FormatOrg:
        doc, err := p.ParseOrgMode(content)
        if err != nil {
            return "", err
        }
        headers = doc.Headers
    case FormatMarkdown:
        doc, err := p.ParseMarkdown(content)
        if err != nil {
            return "", err
        }
        headers = doc.Headers
    default:
        return content, nil
    }

    for _, h := range headers {
        if strings.EqualFold(h.Title, headerTitle) {
            return h.Content, nil
        }
    }

    return "", fmt.Errorf("header not found: %s", headerTitle)
}

// Helper to convert Org nodes to string
func orgNodesToString(nodes []org.Node) string {
	var b strings.Builder
	for _, n := range nodes {
		b.WriteString(n.String())
	}
	return b.String()
}
