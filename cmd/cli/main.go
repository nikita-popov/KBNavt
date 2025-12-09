package main

import (
    "bufio"
    "flag"
    "fmt"
    "log/slog"
    "os"
    //"path/filepath"
    "strings"

    "kbnavt/internal/config"
    "kbnavt/pkg/kb"
)

func main() {
    configPath := flag.String("config", "", "Path to config file")
    flag.Parse()

    args := flag.Args()

    if len(args) == 0 {
        printUsage()
        os.Exit(0)
    }

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

    command := args[0]
    cmdArgs := args[1:]

    switch command {
    case "list":
        cmdList(navigator)
    case "read":
        cmdRead(navigator, cmdArgs)
    case "search":
        cmdSearch(navigator, cmdArgs)
    case "repl":
        cmdREPL(navigator)
    default:
        fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
        os.Exit(1)
    }
}

func cmdList(navigator *kb.Navigator) {
    docs, err := navigator.ListDocuments()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }

    if len(docs) == 0 {
        fmt.Println("No documents found")
        return
    }

    fmt.Printf("%-30s %-10s %-20s\n", "Path", "Format", "Size")
    fmt.Println(strings.Repeat("-", 70))

    for _, doc := range docs {
        fmt.Printf("%-30s %-10s %-20d\n", doc.Path, doc.Format, doc.Size)
    }
}

func cmdRead(navigator *kb.Navigator, args []string) {
    if len(args) < 1 {
        fmt.Fprintf(os.Stderr, "Usage: kbnavt read <path> [section]\n")
        os.Exit(1)
    }

    path := args[0]

    if len(args) > 1 {
        section := strings.Join(args[1:], " ")
        content, err := navigator.ReadSection(path, section)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error: %v\n", err)
            os.Exit(1)
        }
        fmt.Println(content)
    } else {
        doc, err := navigator.ReadDocument(path)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error: %v\n", err)
            os.Exit(1)
        }
        fmt.Println(doc.Content)
    }
}

func cmdSearch(navigator *kb.Navigator, args []string) {
    if len(args) < 1 {
        fmt.Fprintf(os.Stderr, "Usage: kbnavt search <query> [limit]\n")
        os.Exit(1)
    }

    query := args[0]
    limit := 10

    if len(args) > 1 {
        fmt.Sscanf(args[1], "%d", &limit)
    }

    results, err := navigator.SearchDocuments(query, limit)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }

    if len(results) == 0 {
        fmt.Println("No results found")
        return
    }

    fmt.Printf("Found %d results for: %s\n\n", len(results), query)

    for _, result := range results {
        fmt.Printf("Path: %s (Score: %.2f)\n", result.DocumentPath, result.Score)
        fmt.Printf("Snippet: %s\n", result.Snippet)
        fmt.Println(strings.Repeat("-", 70))
    }
}

func cmdREPL(navigator *kb.Navigator) {
    fmt.Println("KBNavt Interactive REPL")
    fmt.Println("Commands: list, read <path>, search <query>, headers <path>, exit")
    fmt.Println()

    reader := bufio.NewReader(os.Stdin)

    for {
        fmt.Print("> ")
        input, _ := reader.ReadString('\n')
        input = strings.TrimSpace(input)

        if input == "" {
            continue
        }

        parts := strings.Fields(input)
        cmd := parts[0]
        args := parts[1:]

        switch cmd {
        case "exit", "quit":
            return
        case "list":
            cmdList(navigator)
        case "read":
            cmdRead(navigator, args)
        case "search":
            cmdSearch(navigator, args)
        case "headers":
            if len(args) < 1 {
                fmt.Println("Usage: headers <path>")
                continue
            }
            doc, err := navigator.ReadDocument(args[0])
            if err != nil {
                fmt.Printf("Error: %v\n", err)
                continue
            }
            printHeaders(doc.Headers)
        default:
            fmt.Printf("Unknown command: %s\n", cmd)
        }
        fmt.Println()
    }
}

func printHeaders(headers []kb.Header) {
    if len(headers) == 0 {
        fmt.Println("No headers found")
        return
    }

    for _, h := range headers {
        indent := strings.Repeat("  ", h.Level-1)
        fmt.Printf("%s%s\n", indent, h.Title)
    }
}

func printUsage() {
    fmt.Println(`KBNavt CLI - Knowledge Base Navigator

Usage:
  kbnavt [flags] <command> [args...]

Commands:
  list                    List all documents
  read <path> [section]   Read document or section
  search <query>          Search documents
  repl                    Interactive REPL

Flags:
  -config string         Path to config file

Examples:
  kbnavt list
  kbnavt read notes/2025/daily.org
  kbnavt search "golang tips"
  kbnavt repl
`)
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
