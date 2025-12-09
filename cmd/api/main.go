package main

import (
    "flag"
    "fmt"
    "log/slog"
    "os"

    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"


	_ "kbnavt/docs" // Swagger docs
	"kbnavt/internal/api"
	"kbnavt/internal/config"
	//"kbnavt/internal/logger"
    "kbnavt/pkg/kb"

	//"github.com/gin-gonic/gin"
	//swaggerFiles "github.com/swaggo/files"
	//ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           KBNav API
// @version         1.0
// @description     Simple HTTP JSON API for navigating a local knowledge base (Markdown/Org/TXT).
// @BasePath        /
// @securityDefinitions.basic  BasicAuth
func main() {
    configPath := flag.String("config", "", "Path to config file")
    flag.Parse()

    // Load configuration
    cfg, err := config.Load(*configPath)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
        os.Exit(1)
    }

    // Setup logging
    logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
        Level: parseLogLevel(cfg.Logging.Level),
    }))

    // Initialize navigator
    navigator, err := kb.NewNavigator(cfg.KB.BaseDir, logger)
    if err != nil {
        logger.Error("Failed to initialize navigator", "error", err)
        os.Exit(1)
    }

    // Create Echo app
    e := echo.New()
    e.Use(middleware.Logger())
    e.Use(middleware.Recover())

    // Setup routes
    api.SetupRoutes(e, navigator, cfg, logger)

    addr := fmt.Sprintf("%s:%d", cfg.API.Host, cfg.API.Port)
    logger.Info("Starting API server", "addr", addr)
    e.Logger.Fatal(e.Start(addr))
}

func parseLogLevel(level string) slog.Level {
    switch level {
    case "debug":
        return slog.LevelDebug
    case "info":
        return slog.LevelInfo
    case "warn":
        return slog.LevelWarn
    case "error":
        return slog.LevelError
    default:
        return slog.LevelInfo
    }
}
