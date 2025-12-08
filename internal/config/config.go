package config

import (
	"gopkg.in/ini.v1"
)

type Config struct {
	Server struct {
		Addr     string
		DataDir  string
		Username string
		Password string
	}
	MCP struct {
		ApiURL string
		Mode   string
		Addr   string
	}
}

func Load(path string) (*Config, error) {
	cfg, err := ini.Load(path)
	if err != nil {
		return nil, err
	}
	var c Config
	c.Server.Addr = cfg.Section("server").Key("addr").MustString(":8080")
	c.Server.DataDir = cfg.Section("server").Key("data_dir").MustString("./")
	c.Server.Username = cfg.Section("server").Key("username").String()
	c.Server.Password = cfg.Section("server").Key("password").String()

	c.MCP.ApiURL = cfg.Section("mcp").Key("api_url").MustString("http://localhost:8080")
	c.MCP.Mode = cfg.Section("mcp").Key("mode").MustString("stdio")
	c.MCP.Addr = cfg.Section("mcp").Key("addr").MustString(":3001")

	return &c, nil
}
