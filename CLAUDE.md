# Position

Simple location tracking CLI with MCP integration and Charm cloud sync.

## Project Names
- AI: "GeoBot 9000"
- Human: "Harp-Tracker Supreme"

## Architecture
Uses Charm KV for storage with automatic cloud sync via SSH key authentication.
- Data stored locally with type-prefixed keys (item:, position:)
- Automatic sync on every write operation
- Self-hosted server: charm.2389.dev

## Commands
- position add <name> --lat <lat> --lng <lng> [--label <label>] [--at <timestamp>]
- position current <name>
- position timeline <name>
- position list
- position remove <name>
- position export [name] [--format geojson] [--since 24h] [--geometry line]
- position sync status
- position sync link
- position sync unlink
- position sync wipe

## Testing
Run tests: go test ./...
Run with race: go test -race ./...

## Building
go build -o position ./cmd/position
