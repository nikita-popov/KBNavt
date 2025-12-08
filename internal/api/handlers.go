package api

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"kbnavt/internal/models"
)

type Handler struct {
	DataDir string
	User    string
	Pass    string
}

// basicAuthMiddleware – simple BasicAuth using config values.
func (h *Handler) BasicAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, pass, ok := c.Request.BasicAuth()
		if !ok || user != h.User || pass != h.Pass {
			c.Header("WWW-Authenticate", `Basic realm="Restricted"`)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Next()
	}
}

// SearchFiles godoc
// @Summary      Search files
// @Description  Search for files in the knowledge base by (case-insensitive) substring in relative path.
// @Tags         search
// @Accept       json
// @Produce      json
// @Param        q   query     string  true  "Search query (substring of path)"
// @Success      200 {array}   domain.SearchResult
// @Failure      400 {object}  gin.H
// @Failure      500 {object}  gin.H
// @Security     BasicAuth
// @Router       /search [get]
func (h *Handler) SearchFiles(c *gin.Context) {
	query := strings.ToLower(c.Query("q"))
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing query parameter 'q'"})
		return
	}

	results := make([]models.SearchResult, 0)

	err := filepath.Walk(h.DataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".md" && ext != ".txt" && ext != ".org" {
			return nil
		}

		relPath, _ := filepath.Rel(h.DataDir, path)
		if strings.Contains(strings.ToLower(relPath), query) {
			results = append(results, models.SearchResult{Path: relPath})
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, results)
}

// ReadFile godoc
// @Summary      Read file
// @Description  Read content of a specific file from the knowledge base.
// @Tags         files
// @Accept       json
// @Produce      json
// @Param        path  query     string  true  "Relative path to file (within data_dir)"
// @Success      200   {object}  domain.FileContent
// @Failure      400   {object}  gin.H
// @Failure      403   {object}  gin.H
// @Failure      404   {object}  gin.H
// @Security     BasicAuth
// @Router       /read [get]
func (h *Handler) ReadFile(c *gin.Context) {
	targetPath := c.Query("path")
	if targetPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing query parameter 'path'"})
		return
	}

	fullPath := filepath.Join(h.DataDir, targetPath)

	// Path traversal protection
	dataDirClean := filepath.Clean(h.DataDir)
	fullClean := filepath.Clean(fullPath)
	if !strings.HasPrefix(fullClean, dataDirClean) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	content, err := os.ReadFile(fullClean)
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	resp := models.FileContent{
		Path:    targetPath,
		Content: string(content),
		Type:    filepath.Ext(targetPath),
	}
	c.JSON(http.StatusOK, resp)
}
