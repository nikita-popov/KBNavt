package kb

import (
	"errors"
	"fmt"
	//"io/ioutil"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	//"time"
)

// Navigator handles knowledge base operations
type Navigator struct {
	rootPath  string
	security *SecurityManager
	parser   *Parser
	search   *SearchEngine
	logger   *slog.Logger
}

// NewNavigator creates a new navigator
func NewNavigator(rootPath string, logger *slog.Logger) (*Navigator, error) {
	nav := &Navigator{rootPath: rootPath}
    se, err := NewSearchEngine(filepath.Join(rootPath, ".search-index"), logger)
    if err != nil {
        return nil, fmt.Errorf("search engine init: %w", err)
    }
    nav.search = se

    docs, err := nav.ListDocuments()
    if err != nil {
        return nil, err
    }
    for _, doc := range docs {
        content, err := os.ReadFile(filepath.Join(rootPath, doc.Path))
        if err != nil {
            continue
        }
        //_ = se.Index(doc.Path, string(content))
		_ = se.IndexDocument(doc.Path, string(content))
    }

    return nav, nil
}

func (n *Navigator) Close() error {
    if n.search != nil {
        return n.search.index.Close()
    }
    return nil
}

// ListDocuments returns all documents in the KB
func (n *Navigator) ListDocuments() ([]Document, error) {
	var documents []Document

	err := filepath.Walk(n.rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !n.security.IsAllowedFile(info.Name()) {
			return nil
		}

		relPath, _ := filepath.Rel(n.rootPath, path)
		format := detectFormat(info.Name())

		doc := Document{
			Path:      relPath,
			Title:     strings.TrimSuffix(info.Name(), filepath.Ext(info.Name())),
			Format:    format,
			CreatedAt: info.ModTime(),
			UpdatedAt: info.ModTime(),
			Size:      info.Size(),
		}

		documents = append(documents, doc)
		return nil
	})

	return documents, err
}

// ReadDocument reads a full document
func (n *Navigator) ReadDocument(relativePath string) (*Document, error) {
	// Security check
	fullPath, err := n.security.ValidatePath(relativePath)
	if err != nil {
		n.logger.Warn("path validation failed", "path", relativePath, "error", err)
		return nil, err
	}

	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("document not found: %s", relativePath)
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read document: %w", err)
	}

	format := detectFormat(info.Name())

	var doc *Document
	switch format {
	case FormatOrg:
		doc, err = n.parser.ParseOrgMode(string(content))
	case FormatMarkdown:
		doc, err = n.parser.ParseMarkdown(string(content))
	default:
		doc, err = n.parser.ParseText(string(content))
	}

	if err != nil {
		return nil, fmt.Errorf("parsing error: %w", err)
	}

	doc.Path = relativePath
	doc.Title = strings.TrimSuffix(info.Name(), filepath.Ext(info.Name()))
	doc.CreatedAt = info.ModTime()
	doc.UpdatedAt = info.ModTime()
	doc.Size = info.Size()

	return doc, nil
}

// ReadSection reads a specific section from a document
func (n *Navigator) ReadSection(relativePath, sectionTitle string) (string, error) {
	fullPath, err := n.security.ValidatePath(relativePath)
	if err != nil {
		return "", err
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to read document: %w", err)
	}

	format := detectFormat(filepath.Base(relativePath))
	return n.parser.ReadSection(string(content), format, sectionTitle)
}

// ListResources returns all resources as MCP-compatible URIs
func (n *Navigator) ListResources() ([]Resource, error) {
	docs, err := n.ListDocuments()
	if err != nil {
		return nil, err
	}

	var resources []Resource
	for _, doc := range docs {
		res := Resource{
			URI:        fmt.Sprintf("kb://documents/%s", strings.ReplaceAll(doc.Path, "\\", "/")),
			Name:       doc.Title,
			MimeType:   "text/plain",
			DocumentID: doc.Path,
		}
		resources = append(resources, res)
	}

	return resources, nil
}

// SearchDocuments performs keyword search
/*func (n *Navigator) SearchDocuments(query string, limit int) ([]SearchResult, error) {
	var results []SearchResult
	docs, err := n.ListDocuments()
	if err != nil {
		return nil, err
	}

	for _, doc := range docs {
		fullDoc, err := n.ReadDocument(doc.Path)
		if err != nil {
			n.logger.Debug("failed to read document", "path", doc.Path, "error", err)
			continue
		}

		if score := calculateRelevance(fullDoc.Content, query); score > 0 {
			results = append(results, SearchResult{
				DocumentID:   doc.Path,
				DocumentPath: doc.Path,
				Score:        score,
				Snippet:      extractSnippet(fullDoc.Content, query, 150),
			})
		}
	}

	// Sort by score (simplified)
	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
    }*/
/*func (n *Navigator) SearchDocuments(query string, limit int) ([]SearchResult, error) {
	if n.search == nil {
		return nil, errors.New("search engine not initialized")
	}
	return n.search.Search(query, limit)
    }
*/
func (n *Navigator) SearchDocuments(query string, limit int) ([]*SearchResult, error) {
    if n.search == nil {
        return nil, errors.New("search engine not initialized")
    }
    return n.search.Search(query, limit)
}

func detectFormat(filename string) Format {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".org":
		return FormatOrg
	case ".md", ".markdown":
		return FormatMarkdown
	default:
		return FormatText
	}
}

func calculateRelevance(content, query string) float64 {
	contentLower := strings.ToLower(content)
	queryLower := strings.ToLower(query)

	if !strings.Contains(contentLower, queryLower) {
		return 0
	}

	// Simple scoring: count occurrences
	count := strings.Count(contentLower, queryLower)
	return float64(count) / float64(len(content)) * 100
}

func extractSnippet(content, query string, length int) string {
	idx := strings.Index(strings.ToLower(content), strings.ToLower(query))
	if idx == -1 {
		if len(content) > length {
			return content[:length] + "..."
		}
		return content
	}

	start := idx
	if start > 0 {
		start -= 50
	}

	end := idx + len(query) + 100
	if end > len(content) {
		end = len(content)
	}

	return "..." + strings.TrimSpace(content[start:end]) + "..."
}
