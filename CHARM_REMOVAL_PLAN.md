# Position - Charm Removal Plan

## Charmbracelet Dependencies

**Direct:**
- `github.com/charmbracelet/charm` (replaced with 2389-research fork)

**Indirect (removed with charm):**
- bubbles, bubbletea, keygen, lipgloss, log, x/ansi, x/term

## Files Importing Charm

| File | Imports |
|------|---------|
| `internal/charm/client.go` | `charm/kv` |
| `internal/charm/items.go` | `charm/kv` |
| `internal/charm/positions.go` | `charm/kv` |
| `internal/charm/wal_test.go` | `charm/kv` |
| `cmd/position/sync.go` | `charm/client`, `charm/kv` |

## Current Architecture

Hybrid SQLite + Charm KV with cloud sync:
- Charm KV wraps SQLite with cloud sync via SSH to `charm.2389.dev`
- Data stored as JSON with type-prefixed keys (`item:UUID`, `position:UUID`)
- Auto-sync on every write
- Stale sync pulls from cloud if data > 1 hour old

## Removal Strategy

### Phase 1: Create Pure SQLite Backend

New `internal/storage/` package (standardized across suite) with direct SQLite access using `modernc.org/sqlite`.

**Schema:**
```sql
CREATE TABLE IF NOT EXISTS items (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS positions (
    id TEXT PRIMARY KEY,
    item_id TEXT NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    latitude REAL NOT NULL,
    longitude REAL NOT NULL,
    label TEXT,
    recorded_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_positions_item_id ON positions(item_id);
CREATE INDEX IF NOT EXISTS idx_positions_recorded_at ON positions(recorded_at);
```

### Phase 2: Remove Sync Commands

Delete `cmd/position/sync.go` entirely.

### Phase 3: Add Export Formats

**Existing GeoJSON:**
```bash
position export [name] --format geojson
```

**New Markdown:**
```bash
position export [name] --format markdown
```
```markdown
# Position History

## harper
| Date | Location | Coordinates |
|------|----------|-------------|
| 2024-12-14 | chicago | (41.8781, -87.6298) |
```

**New YAML:**
```bash
position backup --output positions.yaml
position import positions.yaml
```
```yaml
version: "1.0"
exported_at: "2026-01-31T12:00:00Z"
tool: "position"

items:
  - id: "uuid"
    name: "harper"
    created_at: "2024-12-14T00:00:00Z"

positions:
  - id: "uuid"
    item_id: "uuid"
    latitude: 41.8781
    longitude: -87.6298
    label: "chicago"
    recorded_at: "2024-12-14T00:00:00Z"
```

## Files to Modify

### DELETE:
- `cmd/position/sync.go`
- `internal/charm/client.go`
- `internal/charm/items.go`
- `internal/charm/positions.go`
- `internal/charm/wal_test.go`

### CREATE:
- `internal/storage/sqlite.go` - Connection management
- `internal/storage/schema.go` - Migrations
- `internal/storage/items.go` - Item CRUD
- `internal/storage/positions.go` - Position CRUD
- `cmd/position/backup.go` - YAML backup command
- `cmd/position/import.go` - Import from YAML

### MODIFY:
- `go.mod` - Remove charm deps, add modernc.org/sqlite, gopkg.in/yaml.v3
- `cmd/position/root.go` - Change client init to SQLite
- `cmd/position/export.go` - Add markdown/YAML formats
- `cmd/position/add.go` - Update imports
- `cmd/position/list.go` - Update imports
- `internal/mcp/server.go` - Update imports
- `internal/mcp/tools.go` - Update imports

## Implementation Order

1. Create `internal/storage/` package
2. Implement Repository interface with SQLite
3. Add markdown/YAML export formats
4. Create backup command
5. Update all cmd/ files to use new client
6. Update MCP server
7. Delete `internal/charm/`
8. Delete `cmd/position/sync.go`
9. Update go.mod, run `go mod tidy`

## Data Path

`~/.local/share/position/position.db`
