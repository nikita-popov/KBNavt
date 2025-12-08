package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"

	"kbnavt/internal/models"
)

func setupTestData(t *testing.T) string {
	tmpDir := t.TempDir()

	// Create test files
	testFiles := map[string]string{
		"notes.md":          "# My Notes\nLinux tips",
		"linux/commands.md": "# Linux Commands\nls, cd, grep",
		"linux/bash.org":    "* Bash scripting\necho, read, for",
		"docs/readme.txt":   "README\nProject documentation",
	}

	for relPath, content := range testFiles {
		fullPath := filepath.Join(tmpDir, relPath)
		os.MkdirAll(filepath.Dir(fullPath), 0755)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	return tmpDir
}

func TestSearchFilesSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dataDir := setupTestData(t)

	handler := &Handler{
		DataDir: dataDir,
		User:    "testuser",
		Pass:    "testpass",
	}

	router := gin.New()
	router.GET("/search", handler.BasicAuthMiddleware(), handler.SearchFiles)

	req, _ := http.NewRequest("GET", "/search?q=linux", nil)
	req.SetBasicAuth("testuser", "testpass")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var results []models.SearchResult
	if err := json.Unmarshal(w.Body.Bytes(), &results); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(results) < 2 {
		t.Errorf("Expected at least 2 results for 'linux', got %d", len(results))
	}

	// Check that results contain expected paths
	found := false
	for _, r := range results {
		if r.Path == "linux/commands.md" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected 'linux/commands.md' in results")
	}
}

func TestSearchFilesMissingQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dataDir := setupTestData(t)

	handler := &Handler{
		DataDir: dataDir,
		User:    "testuser",
		Pass:    "testpass",
	}

	router := gin.New()
	router.GET("/search", handler.BasicAuthMiddleware(), handler.SearchFiles)

	req, _ := http.NewRequest("GET", "/search", nil)
	req.SetBasicAuth("testuser", "testpass")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestSearchFilesUnauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dataDir := setupTestData(t)

	handler := &Handler{
		DataDir: dataDir,
		User:    "testuser",
		Pass:    "testpass",
	}

	router := gin.New()
	router.GET("/search", handler.BasicAuthMiddleware(), handler.SearchFiles)

	req, _ := http.NewRequest("GET", "/search?q=test", nil)
	req.SetBasicAuth("wronguser", "wrongpass")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestReadFileSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dataDir := setupTestData(t)

	handler := &Handler{
		DataDir: dataDir,
		User:    "testuser",
		Pass:    "testpass",
	}

	router := gin.New()
	router.GET("/read", handler.BasicAuthMiddleware(), handler.ReadFile)

	req, _ := http.NewRequest("GET", "/read?path=notes.md", nil)
	req.SetBasicAuth("testuser", "testpass")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result models.FileContent
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result.Path != "notes.md" {
		t.Errorf("Expected path notes.md, got %s", result.Path)
	}
	if result.Content == "" {
		t.Error("Expected non-empty content")
	}
	if result.Type != ".md" {
		t.Errorf("Expected type .md, got %s", result.Type)
	}
}

func TestReadFileNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dataDir := setupTestData(t)

	handler := &Handler{
		DataDir: dataDir,
		User:    "testuser",
		Pass:    "testpass",
	}

	router := gin.New()
	router.GET("/read", handler.BasicAuthMiddleware(), handler.ReadFile)

	req, _ := http.NewRequest("GET", "/read?path=nonexistent.md", nil)
	req.SetBasicAuth("testuser", "testpass")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestReadFilePathTraversal(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dataDir := setupTestData(t)

	handler := &Handler{
		DataDir: dataDir,
		User:    "testuser",
		Pass:    "testpass",
	}

	router := gin.New()
	router.GET("/read", handler.BasicAuthMiddleware(), handler.ReadFile)

	// Try to read file outside data_dir
	req, _ := http.NewRequest("GET", "/read?path=../../../etc/passwd", nil)
	req.SetBasicAuth("testuser", "testpass")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403 for path traversal, got %d", w.Code)
	}
}

func TestReadFileMissingPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dataDir := setupTestData(t)

	handler := &Handler{
		DataDir: dataDir,
		User:    "testuser",
		Pass:    "testpass",
	}

	router := gin.New()
	router.GET("/read", handler.BasicAuthMiddleware(), handler.ReadFile)

	req, _ := http.NewRequest("GET", "/read", nil)
	req.SetBasicAuth("testuser", "testpass")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestReadFileWithSubdirectory(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dataDir := setupTestData(t)

	handler := &Handler{
		DataDir: dataDir,
		User:    "testuser",
		Pass:    "testpass",
	}

	router := gin.New()
	router.GET("/read", handler.BasicAuthMiddleware(), handler.ReadFile)

	req, _ := http.NewRequest("GET", "/read?path=linux/commands.md", nil)
	req.SetBasicAuth("testuser", "testpass")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var result models.FileContent
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result.Path != "linux/commands.md" {
		t.Errorf("Expected path linux/commands.md, got %s", result.Path)
	}
}
