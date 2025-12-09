package kb

import (
    "fmt"
    "log/slog"

    "github.com/blevesearch/bleve/v2"
)

// SearchEngine handles full-text indexing and searching
type SearchEngine struct {
    index  bleve.Index
    logger *slog.Logger
}

// NewSearchEngine creates a new search engine
func NewSearchEngine(indexPath string, logger *slog.Logger) (*SearchEngine, error) {
    // Try to open existing index
    index, err := bleve.Open(indexPath)
    if err == nil {
        return &SearchEngine{
            index:  index,
            logger: logger,
        }, nil
    }

    // Create new index if doesn't exist
    mapping := bleve.NewIndexMapping()
    index, err = bleve.NewUsing(indexPath, mapping, "scorch", "mmap", nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create search index: %w", err)
    }

    return &SearchEngine{
        index:  index,
        logger: logger,
    }, nil
}

// IndexDocument adds or updates a document in the index
func (se *SearchEngine) IndexDocument(doc *Document) error {
    err := se.index.Index(doc.Path, map[string]interface{}{
        "title":   doc.Title,
        "content": doc.Content,
        "path":    doc.Path,
        "format":  doc.Format,
    })
    if err != nil {
        se.logger.Error("failed to index document", "path", doc.Path, "error", err)
        return err
    }
    return nil
}

// Search performs a full-text search
func (se *SearchEngine) Search(query string, limit int) ([]*SearchResult, error) {
    searchQuery := bleve.NewQueryStringQuery(query)
    search := bleve.NewSearchRequestOptions(searchQuery, limit, 0, false)
    search.Fields = []string{"title", "content", "path"}

    results, err := se.index.Search(search)
    if err != nil {
        se.logger.Error("search failed", "query", query, "error", err)
        return nil, err
    }

    var searchResults []*SearchResult
    for _, hit := range results.Hits {
        result := &SearchResult{
            DocumentID:   hit.ID,
            DocumentPath: hit.ID,
            Score:        hit.Score,
        }

        if fragments, ok := hit.Fragments["content"]; ok && len(fragments) > 0 {
            result.Snippet = fragments[0] // TODO: check
        }

        searchResults = append(searchResults, result)
    }

    return searchResults, nil
}

// Close closes the search index
func (se *SearchEngine) Close() error {
    return se.index.Close()
}
