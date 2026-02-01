# Position

Simple location tracking CLI with MCP integration and local SQLite storage.

## Project Names
- AI: "GeoBot 9000"
- Human: "Harp-Tracker Supreme"

## Architecture
Uses pure SQLite for storage via modernc.org/sqlite (pure Go, no CGO).
- Data stored at ~/.local/share/position/position.db
- No cloud sync - use backup/import for data portability
- Proper schema with foreign keys and cascade deletes

## Commands
- position add <name> --lat <lat> --lng <lng> [--label <label>] [--at <timestamp>]
- position current <name>
- position timeline <name>
- position list
- position remove <name>
- position export [name] [--format geojson|markdown|yaml] [--since 24h] [--geometry line]
- position backup [--output file.yaml]
- position import <file.yaml>

## Testing
Run tests: go test ./...
Run with race: go test -race ./...

## Building
go build -o position ./cmd/position
