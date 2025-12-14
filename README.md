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
# From source
go install github.com/harper/position/cmd/position@latest

# Or build locally
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

Position uses SQLite for local storage, following XDG standards:

- **Default path:** `~/.local/share/position/position.db`
- **Custom path:** `position --db /path/to/custom.db <command>`

### Schema

```sql
-- Items being tracked
CREATE TABLE items (
    id TEXT PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    created_at DATETIME NOT NULL
);

-- Location history
CREATE TABLE positions (
    id TEXT PRIMARY KEY,
    item_id TEXT NOT NULL,
    latitude REAL NOT NULL,
    longitude REAL NOT NULL,
    label TEXT,
    recorded_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL,
    FOREIGN KEY (item_id) REFERENCES items(id) ON DELETE CASCADE
);
```

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

- Go 1.21+
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
│   ├── root.go           # Root command, DB connection
│   ├── add.go            # Add command
│   ├── current.go        # Current command
│   ├── timeline.go       # Timeline command
│   ├── list.go           # List command
│   ├── remove.go         # Remove command
│   └── mcp.go            # MCP server command
├── internal/
│   ├── db/               # Database layer
│   │   ├── db.go         # Connection management
│   │   ├── migrations.go # Schema
│   │   ├── items.go      # Item CRUD
│   │   └── positions.go  # Position CRUD
│   ├── models/           # Data models
│   │   └── models.go     # Item, Position structs
│   ├── mcp/              # MCP integration
│   │   ├── server.go     # MCP server
│   │   ├── tools.go      # MCP tools
│   │   └── resources.go  # MCP resources
│   └── ui/               # Terminal formatting
│       └── format.go     # Output formatting
├── test/
│   └── integration_test.go
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

# Specific package
go test ./internal/db/...

# Integration tests only
go test ./test/...
```

## Dependencies

- [cobra](https://github.com/spf13/cobra) - CLI framework
- [modernc.org/sqlite](https://modernc.org/sqlite) - Pure Go SQLite
- [google/uuid](https://github.com/google/uuid) - UUID generation
- [fatih/color](https://github.com/fatih/color) - Terminal colors
- [modelcontextprotocol/go-sdk](https://github.com/modelcontextprotocol/go-sdk) - MCP integration

## License

MIT
