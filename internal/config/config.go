package config

import (
    "log/slog"
    "os"
    "path/filepath"

    "github.com/knadh/koanf/v2"
    "github.com/knadh/koanf/parsers/yaml"
    "github.com/knadh/koanf/providers/env"
    "github.com/knadh/koanf/providers/file"
)

// Config represents application configuration
type Config struct {
    KB struct {
        BaseDir string `koanf:"base_dir"`
        MaxSize int64  `koanf:"max_size"`
    } `koanf:"kb"`

    API struct {
        Host     string `koanf:"host"`
        Port     int    `koanf:"port"`
        AuthUser string `koanf:"auth_user"`
        AuthPass string `koanf:"auth_pass"`
    } `koanf:"api"`

    MCP struct {
        Transport string `koanf:"transport"` // "stdio", "sse"
    } `koanf:"mcp"`

    Logging struct {
        Level string `koanf:"level"` // "debug", "info", "warn", "error"
    } `koanf:"logging"`
}

// Load loads configuration from file and environment
func Load(configPath string) (*Config, error) {
    k := koanf.New(".")

    // Load from file if exists
    if configPath != "" {
        if err := k.Load(file.Provider(configPath), yaml.Parser()); err != nil {
            slog.Warn("failed to load config file", "path", configPath, "error", err)
        }
    }

    // Load from environment (with KB_ prefix)
    if err := k.Load(env.Provider("KB_", ".", nil), nil); err != nil {
        return nil, err
    }

    cfg := &Config{}
    if err := k.Unmarshal("", cfg); err != nil {
        return nil, err
    }

    // Apply defaults
    if cfg.KB.BaseDir == "" {
        cfg.KB.BaseDir = filepath.Join(os.Getenv("HOME"), ".kb")
    }
    if cfg.API.Host == "" {
        cfg.API.Host = "localhost"
    }
    if cfg.API.Port == 0 {
        cfg.API.Port = 8080
    }
    if cfg.Logging.Level == "" {
        cfg.Logging.Level = "info"
    }
    if cfg.MCP.Transport == "" {
        cfg.MCP.Transport = "stdio"
    }

    return cfg, nil
}
