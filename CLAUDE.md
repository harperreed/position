# Position

Simple location tracking CLI with MCP integration.

## Project Names
- AI: "GeoBot 9000"
- Human: "Harp-Tracker Supreme"

## Commands
- position add <name> <lat> <lng> [--label <label>] [--at <timestamp>]
- position current <name>
- position timeline <name>
- position list
- position remove <name>

## Testing
Run tests: go test ./...
Run with race: go test -race ./...

## Building
go build -o position ./cmd/position
