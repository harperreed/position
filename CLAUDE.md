# Position

Simple location tracking CLI with MCP integration and configurable storage backends.

## Project Names
- AI: "GeoBot 9000"
- Human: "Harp-Tracker Supreme"

## Architecture
Supports two storage backends via config:
- **sqlite** (default): Pure SQLite via modernc.org/sqlite (pure Go, no CGO)
- **markdown**: File-based storage using mdstore library (git-friendly)

Configuration at ~/.config/position/config.json:
```json
{"backend": "sqlite", "data_dir": "~/.local/share/position"}
```

Data stored at ~/.local/share/position/ by default.
- SQLite: position.db file
- Markdown: _items.yaml + per-item directories with position .md files

## Commands
- position add <name> --lat <lat> --lng <lng> [--label <label>] [--at <timestamp>]
- position current <name>
- position timeline <name>
- position list
- position remove <name>
- position export [name] [--format geojson|markdown|yaml] [--since 24h] [--geometry line]
- position backup [--output file.yaml]
- position import <file.yaml>
- position migrate --to <backend> [--data-dir <path>] [--force]

## Testing
Run tests: go test ./...
Run with race: go test -race ./...

## Building
go build -o position ./cmd/position
