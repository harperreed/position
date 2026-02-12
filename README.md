# Position

Simple location tracking CLI with MCP integration for AI agents.

```
██████╗  ██████╗ ███████╗██╗████████╗██╗ ██████╗ ███╗   ██╗
██╔══██╗██╔═══██╗██╔════╝██║╚══██╔══╝██║██╔═══██╗████╗  ██║
██████╔╝██║   ██║███████╗██║   ██║   ██║██║   ██║██╔██╗ ██║
██╔═══╝ ██║   ██║╚════██║██║   ██║   ██║██║   ██║██║╚██╗██║
██║     ╚██████╔╝███████║██║   ██║   ██║╚██████╔╝██║ ╚████║
╚═╝      ╚═════╝ ╚══════╝╚═╝   ╚═╝   ╚═╝ ╚═════╝ ╚═╝  ╚═══╝
```

Track items (people, vehicles, things) and their locations over time. Query current position or full history.

## Installation

```bash
# Build from source
git clone https://github.com/harper/position
cd position
make build
```

## Quick Start

```bash
# Track someone's location
position add harper --lat 41.8781 --lng -87.6298 --label chicago

# Where are they now?
position current harper
# harper @ chicago (41.8781, -87.6298) - just now

# They moved!
position add harper --lat 40.7128 --lng -74.0060 --label "new york"

# See the journey
position timeline harper
# harper timeline:
#   new york (40.7128, -74.0060) - Dec 14, 3:00 PM
#   chicago (41.8781, -87.6298) - Dec 14, 8:00 AM

# What are we tracking?
position list
# harper - new york (2 hours ago)
# car - garage (1 day ago)

# Stop tracking
position remove harper
```

## Commands

| Command | Alias | Description |
|---------|-------|-------------|
| `position add <name> --lat <lat> --lng <lng>` | `a` | Add a position for an item |
| `position current <name>` | `c` | Get current (most recent) position |
| `position timeline <name>` | `t` | Get position history (newest first) |
| `position list` | `ls` | List all tracked items |
| `position remove <name>` | `rm` | Remove item and all history |
| `position export [name]` | - | Export positions (geojson, markdown, yaml) |
| `position backup [--output file]` | - | Backup all data to YAML |
| `position import <file>` | - | Import data from YAML backup |
| `position migrate --to <backend>` | - | Migrate between storage backends |
| `position mcp` | - | Start MCP server for AI agents |

### Add Options

```bash
# Basic usage (--lat and --lng are required)
position add <name> --lat <latitude> --lng <longitude>

# With location label
position add harper --lat 41.8781 --lng -87.6298 --label chicago
position add harper --lat 41.8781 --lng -87.6298 -l chicago

# Backdate a position (RFC3339 timestamp)
position add harper --lat 41.8781 --lng -87.6298 --at "2024-12-14T08:00:00Z"

# Combined
position add harper --lat 41.8781 --lng -87.6298 -l chicago --at "2024-12-14T08:00:00Z"
```

### Remove Options

```bash
# Interactive confirmation
position remove harper
# Remove 'harper' and all position history? [y/N]

# Skip confirmation
position remove harper --confirm
```

## Data Storage

Position supports pluggable storage backends, configured via `~/.config/position/config.json`:

```json
{
  "backend": "sqlite",
  "data_dir": "~/.local/share/position"
}
```

### Backends

| Backend | Description |
|---------|-------------|
| `sqlite` | Pure Go SQLite via modernc.org/sqlite (default) |
| `markdown` | File-based storage using mdstore (git-friendly) |

Data is stored at `~/.local/share/position/` by default (respects `XDG_DATA_HOME`).

Use `position migrate --to <backend>` to switch between backends.

## MCP Integration

Position includes a Model Context Protocol (MCP) server for AI agent integration.

### Starting the Server

```bash
position mcp
```

### Claude Desktop Configuration

Add to your Claude Desktop config (`~/Library/Application Support/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "position": {
      "command": "/path/to/position",
      "args": ["mcp"]
    }
  }
}
```

### Available Tools

| Tool | Description |
|------|-------------|
| `add_position` | Add a position for an item (creates item if needed) |
| `get_current` | Get current position of an item |
| `get_timeline` | Get position history for an item |
| `list_items` | List all tracked items with positions |
| `remove_item` | Remove an item and all history |

### Available Resources

| URI | Description |
|-----|-------------|
| `position://items` | All items with current positions (JSON) |

### Tool Schemas

**add_position**
```json
{
  "name": "string (required)",
  "latitude": "number (required, -90 to 90)",
  "longitude": "number (required, -180 to 180)",
  "label": "string (optional)",
  "at": "string (optional, RFC3339 timestamp)"
}
```

**get_current / get_timeline / remove_item**
```json
{
  "name": "string (required)"
}
```

**list_items**
```json
{}
```

## Development

### Prerequisites

- Go 1.24+
- Make (optional)

### Building

```bash
# Build binary
make build

# Run tests
make test

# Run tests with race detector
make test-race

# Run linter (requires golangci-lint)
make lint

# Full check (lint + test-race)
make check

# Install to GOPATH/bin
make install

# Clean build artifacts
make clean
```

### Project Structure

```
position/
├── cmd/position/          # CLI entry point
│   ├── main.go           # Entry point
│   ├── root.go           # Root command, config loading
│   ├── add.go            # Add command
│   ├── current.go        # Current command
│   ├── timeline.go       # Timeline command
│   ├── list.go           # List command
│   ├── remove.go         # Remove command
│   ├── export.go         # Export command (geojson, markdown, yaml)
│   ├── backup.go         # Backup command
│   ├── import.go         # Import command
│   ├── migrate.go        # Migrate command
│   ├── mcp.go            # MCP server command
│   ├── skill.go          # Skill install command
│   └── skill/SKILL.md    # MCP skill definition
├── internal/
│   ├── config/           # Configuration management
│   │   └── config.go     # Backend + data dir config
│   ├── storage/          # Storage backends
│   │   ├── repository.go # Storage interface
│   │   ├── sqlite.go     # SQLite backend
│   │   ├── markdown.go   # Markdown/mdstore backend
│   │   ├── migrate.go    # Backend migration
│   │   ├── export.go     # Export logic
│   │   └── errors.go     # Storage errors
│   ├── models/           # Data models
│   │   └── models.go     # Item, Position structs
│   ├── geojson/          # GeoJSON generation
│   │   └── geojson.go    # GeoJSON export support
│   ├── mcp/              # MCP integration
│   │   ├── server.go     # MCP server
│   │   ├── tools.go      # MCP tools
│   │   └── resources.go  # MCP resources
│   └── ui/               # Terminal formatting
│       └── format.go     # Output formatting
├── docs/plans/           # Design documents
├── go.mod
├── Makefile
└── CLAUDE.md
```

### Running Tests

```bash
# All tests
go test ./...

# With verbose output
go test -v ./...

# With race detector
go test -race ./...
```

## Dependencies

- [cobra](https://github.com/spf13/cobra) - CLI framework
- [modernc.org/sqlite](https://modernc.org/sqlite) - Pure Go SQLite
- [harperreed/mdstore](https://github.com/harperreed/mdstore) - Markdown file storage
- [google/uuid](https://github.com/google/uuid) - UUID generation
- [fatih/color](https://github.com/fatih/color) - Terminal colors
- [modelcontextprotocol/go-sdk](https://github.com/modelcontextprotocol/go-sdk) - MCP integration
- [gopkg.in/yaml.v3](https://pkg.go.dev/gopkg.in/yaml.v3) - YAML support

## License

MIT
