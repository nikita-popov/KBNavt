package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"kbnavt/internal/config"
	"kbnavt/internal/models"
)

// Клиент к нашему API
type APIClient struct {
	BaseURL  string
	Username string
	Password string
	Client   *http.Client
}

func (c *APIClient) Request(endpoint string) (*http.Response, error) {
	req, err := http.NewRequest("GET", c.BaseURL+endpoint, nil)
	if err != nil {
		return nil, err
	}
	auth := base64.StdEncoding.EncodeToString([]byte(c.Username + ":" + c.Password))
	req.Header.Add("Authorization", "Basic "+auth)
	return c.Client.Do(req)
}

func main() {
	cfg, err := config.Load("config.ini")
	if err != nil {
		log.Fatal(err)
	}

	apiClient := &APIClient{
		BaseURL:  cfg.MCP.ApiURL,
		Username: cfg.Server.Username,
		Password: cfg.Server.Password,
		Client:   &http.Client{Timeout: 10 * time.Second},
	}

	// Создаем MCP сервер
	s := server.NewMCPServer(
		"KBNav Wrapper",
		"1.0.0",
	)

	// Инструмент 1: Поиск
	s.AddTool(mcp.NewTool("search_files",
		mcp.WithDescription("Search for files in the knowledge base by name"),
		mcp.WithString("query", mcp.Required(), mcp.Description("Search query string")),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		//query := request.Params.Arguments["query"].(string)
		argsMap, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return mcp.NewToolResultError("Invalid arguments type"), nil
		}
		query, ok := argsMap["query"].(string)
		if !ok {
			return mcp.NewToolResultError("Missing or invalid 'query' parameter"), nil
		}
		resp, err := apiClient.Request(fmt.Sprintf("/search?q=%s", query))
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("API error: %v", err)), nil
		}
		defer resp.Body.Close()

		var results []models.SearchResult
		if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
			return mcp.NewToolResultError("Failed to decode response"), nil
		}

		jsonResult, _ := json.Marshal(results)
		return mcp.NewToolResultText(string(jsonResult)), nil
	})

	// Инструмент 2: Чтение
	s.AddTool(mcp.NewTool("read_file",
		mcp.WithDescription("Read content of a specific file"),
		mcp.WithString("path", mcp.Required(), mcp.Description("Relative path to file")),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		//path := request.Params.Arguments["path"].(string)
		argsMap, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return mcp.NewToolResultError("Invalid arguments type"), nil
		}
		path, ok := argsMap["path"].(string)
		if !ok {
			return mcp.NewToolResultError("Missing or invalid 'path' parameter"), nil
		}
		resp, err := apiClient.Request(fmt.Sprintf("/read?path=%s", path))
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("API error: %v", err)), nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return mcp.NewToolResultError("File not found or access denied"), nil
		}

		var content models.FileContent
		if err := json.NewDecoder(resp.Body).Decode(&content); err != nil {
			return mcp.NewToolResultError("Failed to decode response"), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("File: %s\n\n%s", content.Path, content.Content)), nil
	})

	// Запуск
	if cfg.MCP.Mode == "sse" {
		log.Printf("Starting MCP SSE server on :%s", cfg.MCP.Addr)
		//sseServer := server.NewSSEServer(s, "http://localhost:"+cfg.MCP.Addr)
		sseServer := server.NewSSEServer(s)
		log.Fatal(sseServer.Start(":" + cfg.MCP.Addr))
	} else {
		// Stdio (по умолчанию для Claude Desktop)
		log.Println("Starting MCP Stdio server")
		if err := server.ServeStdio(s); err != nil {
			log.Fatal(err)
		}
	}
}
