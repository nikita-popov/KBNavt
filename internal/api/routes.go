package api

import (
	"fmt"
    "log/slog"

    "github.com/labstack/echo/v4"
    "kbnavt/internal/config"
    "kbnavt/pkg/kb"
)

// SetupRoutes configures all API routes
func SetupRoutes(e *echo.Echo, navigator *kb.Navigator, cfg *config.Config, logger *slog.Logger) {
    // Public routes
    e.GET("/health", HealthHandler)

    // Protected routes with Basic Auth
    api := e.Group("")
    api.Use(BasicAuthMiddleware(cfg.API.AuthUser, cfg.API.AuthPass, logger))

    api.GET("/documents", ListDocumentsHandler(navigator, logger))
    api.GET("/documents/:path", ReadDocumentHandler(navigator, logger))
    api.GET("/documents/:path/section/:section", ReadSectionHandler(navigator, logger))
    api.GET("/search", SearchHandler(navigator, logger))
    api.GET("/resources", ListResourcesHandler(navigator, logger))
}

// HealthHandler checks server health
func HealthHandler(c echo.Context) error {
    return c.JSON(200, map[string]string{"status": "ok"})
}

// ListDocumentsHandler lists all documents
func ListDocumentsHandler(navigator *kb.Navigator, logger *slog.Logger) echo.HandlerFunc {
    return func(c echo.Context) error {
        docs, err := navigator.ListDocuments()
        if err != nil {
            logger.Error("failed to list documents", "error", err)
            return c.JSON(500, map[string]string{"error": err.Error()})
        }
        return c.JSON(200, map[string]interface{}{"documents": docs})
    }
}

// ReadDocumentHandler reads a specific document
func ReadDocumentHandler(navigator *kb.Navigator, logger *slog.Logger) echo.HandlerFunc {
    return func(c echo.Context) error {
        path := c.Param("path")
        doc, err := navigator.ReadDocument(path)
        if err != nil {
            logger.Error("failed to read document", "path", path, "error", err)
            return c.JSON(404, map[string]string{"error": "document not found"})
        }
        return c.JSON(200, doc)
    }
}

// ReadSectionHandler reads a section from a document
func ReadSectionHandler(navigator *kb.Navigator, logger *slog.Logger) echo.HandlerFunc {
    return func(c echo.Context) error {
        path := c.Param("path")
        section := c.Param("section")
        content, err := navigator.ReadSection(path, section)
        if err != nil {
            logger.Error("failed to read section", "path", path, "section", section, "error", err)
            return c.JSON(404, map[string]string{"error": "section not found"})
        }
        return c.JSON(200, map[string]string{"content": content})
    }
}

// SearchHandler searches documents
func SearchHandler(navigator *kb.Navigator, logger *slog.Logger) echo.HandlerFunc {
    return func(c echo.Context) error {
        query := c.QueryParam("q")
        limit := 10
        if l := c.QueryParam("limit"); l != "" {
            fmt.Sscanf(l, "%d", &limit)
        }

        results, err := navigator.SearchDocuments(query, limit)
        if err != nil {
            logger.Error("search failed", "query", query, "error", err)
            return c.JSON(500, map[string]string{"error": err.Error()})
        }
        return c.JSON(200, map[string]interface{}{"results": results})
    }
}

// ListResourcesHandler lists all resources
func ListResourcesHandler(navigator *kb.Navigator, logger *slog.Logger) echo.HandlerFunc {
    return func(c echo.Context) error {
        resources, err := navigator.ListResources()
        if err != nil {
            logger.Error("failed to list resources", "error", err)
            return c.JSON(500, map[string]string{"error": err.Error()})
        }
        return c.JSON(200, map[string]interface{}{"resources": resources})
    }
}
