package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "kbnavt/docs" // Swagger docs
	"kbnavt/internal/api"
	"kbnavt/internal/config"
	"kbnavt/internal/logger"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           KBNav API
// @version         1.0
// @description     Simple HTTP JSON API for navigating a local knowledge base (Markdown/Org/TXT).
// @BasePath        /
// @securityDefinitions.basic  BasicAuth
func main() {
	cfg, err := config.Load("config.ini")
	if err != nil {
		logger.Fatal("Failed to load config: %v", err)
	}

	h := &api.Handler{
		DataDir: cfg.Server.DataDir,
		User:    cfg.Server.Username,
		Pass:    cfg.Server.Password,
	}

	//r := gin.Default()
	router := gin.New()
	router.Use(ginLogger(), gin.Recovery())

	// Swagger endpoint
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Auth-protected group
	authorized := router.Group("/", h.BasicAuthMiddleware())
	v1 := authorized.Group("/v1")
	{
		v1.GET("/search", h.SearchFiles)
		v1.GET("/read", h.ReadFile)
	}

	//log.Printf("KBnavT API running on :%s serving %s", cfg.Server.Port, cfg.Server.DataDir)
	//if err := r.Run(":" + cfg.Server.Port); err != nil {
	//	log.Fatal(err)
	//}

	// Create HTTP server
	srv := &http.Server{
		Addr:    cfg.Server.Addr,
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		logger.Info("Starting server on %s", cfg.Server.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Info("Server forced to shutdown: %v", err)
	}
}

// ginLogger returns a gin middleware for logging
func ginLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		end := time.Now()
		latency := end.Sub(start)
		statusCode := c.Writer.Status()
		method := c.Request.Method
		clientIP := c.ClientIP()

		if query != "" {
			path = path + "?" + query
		}

		// Log based on status code
		if statusCode >= 500 {
			logger.Error("%s %s %d %v %s", method, path, statusCode, latency, clientIP)
		} else if statusCode >= 400 {
			logger.Warn("%s %s %d %v %s", method, path, statusCode, latency, clientIP)
		} else {
			logger.Info("%s %s %d %v %s", method, path, statusCode, latency, clientIP)
		}
	}
}
