package main

import (
    "bufio"
    "context"
    "encoding/json"
    "flag"
    "fmt"
    "log/slog"
    "os"

    "kbnavt/internal/config"
    "kbnavt/internal/mcp"
    "kbnavt/pkg/kb"
)

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
    logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
        Level: parseLogLevel(cfg.Logging.Level),
    }))

    // Initialize navigator
    navigator, err := kb.NewNavigator(cfg.KB.BaseDir, logger)
    if err != nil {
        logger.Error("Failed to initialize navigator", "error", err)
        os.Exit(1)
    }

    // Create MCP server
    mcpServer := mcp.NewMCPServer(navigator, logger)

    logger.Info("KBNavt MCP Server started", "transport", cfg.MCP.Transport)

    // Run MCP server via stdio
    runStdioServer(mcpServer, logger)
}

func runStdioServer(mcpServer *mcp.MCPServer, logger *slog.Logger) {
    scanner := bufio.NewScanner(os.Stdin)

    for scanner.Scan() {
        line := scanner.Bytes()

        // Handle JSON-RPC request
        response, err := mcpServer.HandleRequest(context.Background(), line)

        var respBytes []byte
        if err != nil {
            respBytes, _ = json.Marshal(map[string]interface{}{
                "error": map[string]interface{}{
                    "code":    -32603,
                    "message": err.Error(),
                },
            })
        } else {
            respBytes, _ = json.Marshal(map[string]interface{}{
                "result": response,
            })
        }

        fmt.Println(string(respBytes))
    }

    if err := scanner.Err(); err != nil {
        logger.Error("Scanner error", "error", err)
    }
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
