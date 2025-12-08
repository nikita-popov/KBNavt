# KBNavt

Knowledge Base Navigator is a lightweight bridge between your local knowledge base (Markdown, Org-mode, TXT files) and Large Language Models (LLMs).

It consists of two main components:

1. **Core API**: A simple HTTP JSON API that indexes and serves your local files with Basic Auth protection.
2. **MCP Server**: An adapter implementing the Model Context Protocol (stdio/SSE), allowing AI assistants like Claude or ChatGPT to "navigate" and read your notes via the Core API.

## Features

- **Format Agnostic**: Works with .md, .org, and .txt files.
- **Secure**: The Core API is protected by Basic Auth; the MCP server acts as a client.
- **Dual Mode MCP**: Supports both stdio (for local desktop apps like Claude Desktop) and sse (Server-Sent Events for remote connections).
- **Configurable**: Simple INI configuration for ports, paths, and credentials.
- **Test Coverage**: Includes unit tests for critical paths.

## Getting Started

### Prerequisites

- Go 1.24+
- Make

### Installation

1. Clone the repository:

```bash
git clone https://github.com/nikita-popov/kbnavt.git
cd kbnavt
```

2. Build:

```bash
make
```

3. Configure the application:

Create a `config.ini` file in the root directory (see [Configuration](#configuration)).

### Configuration

Create a `config.ini` file:

```ini
[server]
addr = :8080
# Absolute or relative path to your notes directory
data_dir = ./my_notes
username = admin
password = your_secure_password

[mcp]
# URL where the Core API is running
api_url = http://localhost:8080
# Mode: "stdio" (for Claude Desktop) or "sse"
mode = stdio
addr = :3001
```

### Usage

1. Start the Core API

This service must be running for the MCP server to work.

```bash
./kbnavt-api
```

The API will start at http://localhost:8080.

2. Connect to an LLM (Claude Desktop)

To use KBNavt with the Claude Desktop App, add the following to your `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "kbnavt": {
      "command": "go",
      "args": ["run", "cmd/mcp/main.go"],
      "env": {
        "KBNAVT_CONFIG_PATH": "/absolute/path/to/kbnavt/config.ini"
      }
    }
  }
}
```

Note: Ensure the args point to the correct path of the `cmd/mcp/main.go` file.

3. Manual Testing via HTTP

You can test the Core API directly using curl:

```bash
# Search for files containing "linux" in the filename
curl -u admin:your_secure_password "http://localhost:8080/search?q=linux"

# Read a specific file
curl -u admin:your_secure_password "http://localhost:8080/read?path=linux/commands.md"
```

### Development

#### Running Tests

```bash
make check
```

#### Architecture Note

We intentionally separated the Core API and the MCP Server.

- Security: The Core API can be deployed on a remote server (VPS) behind a reverse proxy or VPN.
- Flexibility: The MCP Server can run locally on your machine, connecting to the remote API securely. This allows you to use local LLM tools with a remote knowledge base.

## License

MIT License. See [LICENSE](LICENSE.md) for details.
