package mcp

import (
    "context"
    "encoding/json"
    "fmt"
    "log/slog"
	"strings"

    "kbnavt/pkg/kb"
)

// MCPServer implements the Model Context Protocol
type MCPServer struct {
    navigator *kb.Navigator
    logger    *slog.Logger
    version   string
}

// NewMCPServer creates a new MCP server
func NewMCPServer(navigator *kb.Navigator, logger *slog.Logger) *MCPServer {
    return &MCPServer{
        navigator: navigator,
        logger:    logger,
        version:   "1.0.0",
    }
}

// InitializeRequest handles the initialize request
type InitializeRequest struct {
    ClientInfo struct {
        Name    string `json:"name"`
        Version string `json:"version"`
    } `json:"clientInfo"`
}

// InitializeResponse is the initialize response
type InitializeResponse struct {
    ProtocolVersion string `json:"protocolVersion"`
    Capabilities    struct {
        Tools     bool `json:"tools"`
        Resources bool `json:"resources"`
        Prompts   bool `json:"prompts"`
    } `json:"capabilities"`
    ServerInfo struct {
        Name    string `json:"name"`
        Version string `json:"version"`
    } `json:"serverInfo"`
}

// HandleRequest routes incoming JSON-RPC requests
func (s *MCPServer) HandleRequest(ctx context.Context, data []byte) (interface{}, error) {
    var msg map[string]interface{}
    if err := json.Unmarshal(data, &msg); err != nil {
        return nil, fmt.Errorf("invalid JSON: %w", err)
    }

    method, ok := msg["method"].(string)
    if !ok {
        return nil, fmt.Errorf("missing method")
    }

    switch method {
    case "initialize":
        return s.handleInitialize(ctx)
    case "resources/list":
        return s.handleListResources(ctx)
    case "resources/read":
        return s.handleReadResource(ctx, msg["params"])
    case "tools/list":
        return s.handleListTools(ctx)
    case "tools/call":
        return s.handleCallTool(ctx, msg["params"])
    case "prompts/list":
        return s.handleListPrompts(ctx)
    case "prompts/get":
        return s.handleGetPrompt(ctx, msg["params"])
    default:
        return nil, fmt.Errorf("unknown method: %s", method)
    }
}

func (s *MCPServer) handleInitialize(ctx context.Context) (interface{}, error) {
    return InitializeResponse{
        ProtocolVersion: "2024-11-05",
        Capabilities: struct {
            Tools     bool `json:"tools"`
            Resources bool `json:"resources"`
            Prompts   bool `json:"prompts"`
        }{
            Tools:     true,
            Resources: true,
            Prompts:   true,
        },
        ServerInfo: struct {
            Name    string `json:"name"`
            Version string `json:"version"`
        }{
            Name:    "KBNavt MCP Server",
            Version: s.version,
        },
    }, nil
}

func (s *MCPServer) handleListResources(ctx context.Context) (interface{}, error) {
    resources, err := s.navigator.ListResources()
    if err != nil {
        s.logger.Error("failed to list resources", "error", err)
        return nil, err
    }

    return map[string]interface{}{
        "resources": resources,
    }, nil
}

func (s *MCPServer) handleReadResource(ctx context.Context, params interface{}) (interface{}, error) {
    paramMap, ok := params.(map[string]interface{})
    if !ok {
        return nil, fmt.Errorf("invalid params")
    }

    uri, ok := paramMap["uri"].(string)
    if !ok {
        return nil, fmt.Errorf("missing uri")
    }

    // Parse URI: kb://documents/path/to/doc
    docPath := extractDocPathFromURI(uri)
    if docPath == "" {
        return nil, fmt.Errorf("invalid resource URI: %s", uri)
    }

    doc, err := s.navigator.ReadDocument(docPath)
    if err != nil {
        s.logger.Error("failed to read resource", "uri", uri, "error", err)
        return nil, err
    }

    return map[string]interface{}{
        "contents": []map[string]string{
            {
                "type": "text",
                "text": doc.Content,
            },
        },
    }, nil
}

func (s *MCPServer) handleListTools(ctx context.Context) (interface{}, error) {
    tools := []map[string]interface{}{
        {
            "name":        "list_documents",
            "description": "List all documents in the knowledge base",
            "inputSchema": map[string]interface{}{
                "type":       "object",
                "properties": map[string]interface{}{},
                "required":   []string{},
            },
        },
        {
            "name":        "read_document",
            "description": "Read a specific document from the knowledge base",
            "inputSchema": map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "path": map[string]interface{}{
                        "type":        "string",
                        "description": "Path to the document (relative to KB root)",
                    },
                },
                "required": []string{"path"},
            },
        },
        {
            "name":        "read_section",
            "description": "Read a specific section/header from a document",
            "inputSchema": map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "path": map[string]interface{}{
                        "type":        "string",
                        "description": "Path to the document",
                    },
                    "section": map[string]interface{}{
                        "type":        "string",
                        "description": "Section/header title to read",
                    },
                },
                "required": []string{"path", "section"},
            },
        },
        {
            "name":        "search_documents",
            "description": "Search documents by keyword",
            "inputSchema": map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "query": map[string]interface{}{
                        "type":        "string",
                        "description": "Search query",
                    },
                    "limit": map[string]interface{}{
                        "type":        "integer",
                        "description": "Maximum results",
                        "default":     10,
                    },
                },
                "required": []string{"query"},
            },
        },
    }

    return map[string]interface{}{
        "tools": tools,
    }, nil
}

func (s *MCPServer) handleCallTool(ctx context.Context, params interface{}) (interface{}, error) {
    paramMap, ok := params.(map[string]interface{})
    if !ok {
        return nil, fmt.Errorf("invalid params")
    }

    toolName, ok := paramMap["name"].(string)
    if !ok {
        return nil, fmt.Errorf("missing tool name")
    }

    args, ok := paramMap["arguments"].(map[string]interface{})
    if !ok {
        args = make(map[string]interface{})
    }

    switch toolName {
    case "list_documents":
        docs, err := s.navigator.ListDocuments()
        if err != nil {
            return nil, err
        }
        return map[string]interface{}{
            "content": []map[string]interface{}{
                {
                    "type": "text",
                    "text": fmt.Sprintf("Found %d documents", len(docs)),
                },
            },
        }, nil

    case "read_document":
        path, ok := args["path"].(string)
        if !ok {
            return nil, fmt.Errorf("missing path parameter")
        }
        doc, err := s.navigator.ReadDocument(path)
        if err != nil {
            return nil, err
        }
        return map[string]interface{}{
            "content": []map[string]string{
                {
                    "type": "text",
                    "text": doc.Content,
                },
            },
        }, nil

    case "read_section":
        path, ok := args["path"].(string)
        if !ok {
            return nil, fmt.Errorf("missing path parameter")
        }
        section, ok := args["section"].(string)
        if !ok {
            return nil, fmt.Errorf("missing section parameter")
        }
        content, err := s.navigator.ReadSection(path, section)
        if err != nil {
            return nil, err
        }
        return map[string]interface{}{
            "content": []map[string]string{
                {
                    "type": "text",
                    "text": content,
                },
            },
        }, nil

    case "search_documents":
        query, ok := args["query"].(string)
        if !ok {
            return nil, fmt.Errorf("missing query parameter")
        }
        limit := 10
        if l, ok := args["limit"].(float64); ok {
            limit = int(l)
        }
        results, err := s.navigator.SearchDocuments(query, limit)
        if err != nil {
            return nil, err
        }
        return map[string]interface{}{
            "content": []map[string]interface{}{
                {
                    "type": "text",
                    "text": fmt.Sprintf("Found %d results for: %s", len(results), query),
                },
            },
        }, nil

    default:
        return nil, fmt.Errorf("unknown tool: %s", toolName)
    }
}

func (s *MCPServer) handleListPrompts(ctx context.Context) (interface{}, error) {
    prompts := []map[string]interface{}{
        {
            "name":        "summarize_daily",
            "description": "Summarize today's notes",
            "arguments": []map[string]interface{}{},
        },
        {
            "name":        "find_related",
            "description": "Find related notes about a topic",
            "arguments": []map[string]interface{}{
                {
                    "name":        "topic",
                    "description": "Topic to search for",
                    "required":    true,
                },
            },
        },
    }
    return map[string]interface{}{
        "prompts": prompts,
    }, nil
}

func (s *MCPServer) handleGetPrompt(ctx context.Context, params interface{}) (interface{}, error) {
    paramMap, ok := params.(map[string]interface{})
    if !ok {
        return nil, fmt.Errorf("invalid params")
    }

    promptName, ok := paramMap["name"].(string)
    if !ok {
        return nil, fmt.Errorf("missing prompt name")
    }

    switch promptName {
    case "summarize_daily":
        return map[string]interface{}{
            "messages": []map[string]interface{}{
                {
                    "role": "user",
                    "content": "Summarize my notes from today. Focus on: 1) What I accomplished, 2) Open tasks, 3) Key insights",
                },
            },
        }, nil
    case "find_related":
        topic, _ := paramMap["topic"].(string)
        return map[string]interface{}{
            "messages": []map[string]interface{}{
                {
                    "role": "user",
                    "content": fmt.Sprintf("Find and summarize all my notes related to: %s. Include connections between them.", topic),
                },
            },
        }, nil
    default:
        return nil, fmt.Errorf("unknown prompt: %s", promptName)
    }
}

func extractDocPathFromURI(uri string) string {
    // Parse kb://documents/path/to/doc -> path/to/doc
    if !strings.HasPrefix(uri, "kb://documents/") {
        return ""
    }
    return strings.TrimPrefix(uri, "kb://documents/")
}
