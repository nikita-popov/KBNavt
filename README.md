# Lightweight KB Navigator & MCP Bridge

A lightweight, high-performance knowledge base navigator with Model Context Protocol (MCP) support. Access your Org-mode, Markdown, and text files through HTTP API or directly integrate with LLMs via MCP.

## Features

## Features

✨ **Multi-Format Support**
- Org-mode (`.org`) with full header parsing
- Markdown (`.md`) with semantic structure
- Plain text (`.txt`)

🔍 **Smart Search**
- Full-text search across all documents
- Relevance scoring
- Snippet extraction

🔐 **Security First**
- Path traversal protection
- Sandboxed file access
- Basic Authentication for API
- Configurable access roots

🤖 **MCP Integration**
- Native Model Context Protocol support
- Resources: `kb://documents/...` URIs
- Tools: `list_documents`, `read_document`, `read_section`, `search_documents`
- Prompts: Pre-built templates for common tasks

🏗️ **Clean Architecture**
- Unified core library (`pkg/kb`)
- Dual interface: HTTP API + MCP server
- Configurable deployment modes
- Structured logging with slog

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

Create a `config.yaml` file in the root directory (see [Configuration](#configuration)).

### Configuration

Create a `config.yaml` file:

```yaml
kb:
  base_dir: ~/Documents/kb
  max_size: 10485760  # 10MB

api:
  host: localhost
  port: 8080
  auth_user: admin
  auth_pass: changeme

mcp:
  transport: stdio  # stdio or sse

logging:
  level: info  # debug, info, warn, error
```

Or use environment variables with KB_ prefix:

```bash
export KB_KB_BASE_DIR=~/my-kb
export KB_LOGGING_LEVEL=debug
```

### Usage

#### HTTP API Server

```bash
./bin/kbnavt-api -config config.yaml
```

API endpoints:

```bash
# Health check
curl http://localhost:8080/health

# List documents
curl -u admin:changeme http://localhost:8080/documents

# Read document
curl -u admin:changeme http://localhost:8080/documents/2025/notes.org

# Read specific section
curl -u admin:changeme "http://localhost:8080/documents/2025/notes.org/section/Today"

# Search
curl -u admin:changeme "http://localhost:8080/search?q=golang&limit=10"

# List resources
curl -u admin:changeme http://localhost:8080/resources
```

#### MCP Server (Embedded)

```bash
./bin/kbnavt-mcp -config config.yaml
```

Configure in Claude Desktop (`~/.config/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "kbnavt": {
      "command": "/path/to/kbnavt-mcp",
      "args": ["-config=/path/to/config.yaml"]
    }
  }
}
```

#### Interactive CLI

```bash
# List all documents
./bin/kbnavt list

# Read document
./bin/kbnavt read notes/2025/daily.org

# Read section
./bin/kbnavt read notes/2025/daily.org "Morning Review"

# Search
./bin/kbnavt search "golang patterns" 5

# Interactive REPL
./bin/kbnavt repl
```

## MCP Protocol

KBNavt implements the full Model Context Protocol with:

### Resources

Access your KB documents via URIs:

```
kb://documents/path/to/file.org
kb://documents/2025/daily.md
```

LLM clients can list all available resources and read them without exposing filesystem paths.

### Tools

Available MCP tools:

| Tool               | Description            | Parameters              |
|--------------------|------------------------|-------------------------|
| `list_documents`   | List all KB documents  | -                       |
| `read_document`    | Read full document     | path (string)           |
| `read_section`     | Read section by header | path, section           |
| `search_documents` | Full-text search       | query, limit (optional) |

### Prompts

Pre-configured prompt templates:

- `summarize_daily` - Summarize today's notes
- `find_related` - Find notes about a topic

### Security

- Path Validation: All file paths are validated against allowed roots
- Path Traversal Protection: Prevents ../ attacks
- Sandboxing: Only configured KB directories are accessible
- Authentication: Basic Auth for HTTP API
- File Type Filtering: Only .org, .md, .txt files

### Performance

- Lazy loading: Documents read on demand
- Streaming: Large documents handled efficiently
- Caching: Headers cached per-document
- Minimal dependencies: Only essential libraries

## Development

### Running Tests

```bash
# Run tests
make check

# Build
make build

# Clean
make clean
```

## License

MIT License. See [LICENSE](LICENSE) for details.
