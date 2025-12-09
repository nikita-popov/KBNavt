package kb

import (
    "time"
)

// Document represents a single knowledge base entry
type Document struct {
    ID        string    `json:"id"`
    Path      string    `json:"path"`
    Title     string    `json:"title"`
    Content   string    `json:"content"`
    Format    Format    `json:"format"`
    Headers   []Header  `json:"headers,omitempty"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    Size      int64     `json:"size"`
}

// Header represents a section/heading in a document
type Header struct {
    ID       string   `json:"id"`
    Level    int      `json:"level"`
    Title    string   `json:"title"`
    Content  string   `json:"content"`
    Children []Header `json:"children,omitempty"`
    LineNum  int      `json:"line_num"`
}

// Format enum
type Format string

const (
    FormatOrg      Format = "org"
    FormatMarkdown Format = "markdown"
    FormatText     Format = "text"
)

// SearchResult represents a full-text search result
type SearchResult struct {
    DocumentID string  `json:"document_id"`
    DocumentPath string `json:"document_path"`
    Score      float64 `json:"score"`
    Snippet    string  `json:"snippet"`
    Header     *Header `json:"header,omitempty"`
}

// Resource represents an MCP resource URI
type Resource struct {
    URI       string `json:"uri"`
    Name      string `json:"name"`
    MimeType  string `json:"mime_type"`
    DocumentID string `json:"document_id"`
}

// ContentParams for MCP read_resource
type ContentParams struct {
    URI string `json:"uri"`
}

// ReadResourceResult for MCP
type ReadResourceResult struct {
    Contents []interface{} `json:"contents"`
}

// TextContent for MCP resource content
type TextContent struct {
    Type string `json:"type"`
    Text string `json:"text"`
}

// MCPError represents MCP error response
type MCPError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
}
