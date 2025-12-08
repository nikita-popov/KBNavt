package config

import (
	"os"
	//o"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create temp config file
	tmpFile, err := os.CreateTemp("", "config*.ini")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	configContent := `[server]
addr = :9090
data_dir = ./test_data
username = testuser
password = testpass

[mcp]
api_url = http://localhost:9090
mode = stdio
sse_port = 3002
`

	if _, err := tmpFile.WriteString(configContent); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	cfg, err := Load(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Server.Addr != ":9090" {
		t.Errorf("Expected port :9090, got %s", cfg.Server.Addr)
	}
	if cfg.Server.DataDir != "./test_data" {
		t.Errorf("Expected data_dir ./test_data, got %s", cfg.Server.DataDir)
	}
	if cfg.Server.Username != "testuser" {
		t.Errorf("Expected username testuser, got %s", cfg.Server.Username)
	}
	if cfg.Server.Password != "testpass" {
		t.Errorf("Expected password testpass, got %s", cfg.Server.Password)
	}
	if cfg.MCP.ApiURL != "http://localhost:9090" {
		t.Errorf("Expected api_url http://localhost:9090, got %s", cfg.MCP.ApiURL)
	}
	if cfg.MCP.Mode != "stdio" {
		t.Errorf("Expected mode stdio, got %s", cfg.MCP.Mode)
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	// Create minimal config file
	tmpFile, err := os.CreateTemp("", "config*.ini")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	configContent := `[server]
[mcp]
`

	if _, err := tmpFile.WriteString(configContent); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	cfg, err := Load(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Check defaults
	if cfg.Server.Addr != ":8080" {
		t.Errorf("Expected default addr :8080, got %s", cfg.Server.Addr)
	}
	if cfg.Server.DataDir != "./" {
		t.Errorf("Expected default data_dir ./, got %s", cfg.Server.DataDir)
	}
	if cfg.MCP.ApiURL != "http://localhost:8080" {
		t.Errorf("Expected default api_url, got %s", cfg.MCP.ApiURL)
	}
	if cfg.MCP.Mode != "stdio" {
		t.Errorf("Expected default mode stdio, got %s", cfg.MCP.Mode)
	}
}

func TestLoadConfigFileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/config.ini")
	if err == nil {
		t.Error("Expected error for nonexistent config file")
	}
}
