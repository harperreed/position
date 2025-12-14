# Position: Simple Location Tracking CLI with MCP

A minimal Go CLI for tracking items (people, things) and their locations over time.

## Data Model

### Items
The thing being tracked (person, car, etc.)
```
Item:
  - id: UUID
  - name: string (unique, e.g., "harper", "car")
  - created_at: timestamp
```

### Positions
Location entries with history
```
Position:
  - id: UUID
  - item_id: UUID (foreign key)
  - latitude: float64
  - longitude: float64
  - label: string (optional, e.g., "chicago", "123 Main St")
  - recorded_at: timestamp (when this position was recorded)
  - created_at: timestamp (when the row was inserted)
```

### Relationships
- One Item -> Many Positions (history)
- "Current" = most recent position by `recorded_at`

### Storage
SQLite via `modernc.org/sqlite`, stored at `~/.local/share/position/position.db`

### Why Two Timestamps?
`recorded_at` is when the person/thing was actually there. `created_at` is when you logged it. Supports backfilling: "harper was in NYC yesterday" - `recorded_at` is yesterday, `created_at` is now.

## CLI Commands

```bash
# Add a new position for an item (creates item if doesn't exist)
position add harper 41.8781 -87.6298
position add harper 41.8781 -87.6298 --label "chicago"
position add harper 41.8781 -87.6298 --label "chicago" --at "2024-12-14T15:00:00"

# Get current location
position current harper
# Output: harper @ chicago (41.8781, -87.6298) - 3 hours ago

# Get location history
position timeline harper
# Output:
# harper timeline:
#   chicago (41.8781, -87.6298) - Dec 14, 3:00 PM
#   new york (40.7128, -74.0060) - Dec 14, 8:00 AM
#   boston (42.3601, -71.0589) - Dec 13, 6:00 PM

# List all tracked items
position list
# Output:
#   harper - chicago (3 hours ago)
#   hiromi - boston (1 day ago)
#   car - 123 main st (2 hours ago)

# Remove an item (and all its positions)
position remove harper
position remove harper --confirm  # skip prompt

# Config (optional, for geocoding)
position config set geocode.api_key "xxx"
position config get geocode.api_key
```

### Short Aliases
- `position a` (add)
- `position c` (current)
- `position t` (timeline)
- `position ls` (list)
- `position rm` (remove)

## MCP Integration

### Tools

| Tool | Description |
|------|-------------|
| `add_position` | Add a position for an item (creates item if needed) |
| `get_current` | Get current location of an item |
| `get_timeline` | Get location history for an item |
| `list_items` | List all tracked items with their current positions |
| `remove_item` | Delete an item and all its positions |

### Resources

| URI | Description |
|-----|-------------|
| `position://items` | All items with current positions |
| `position://item/{name}` | Specific item's current position |
| `position://timeline/{name}` | Full history for an item |

### Prompts

| Prompt | Description |
|--------|-------------|
| `where-is` | "Where is {name}?" - returns current position with context |
| `track-journey` | Log a series of positions for a trip |

## Project Structure

```
position/
├── cmd/
│   └── position/
│       ├── main.go          # Entry point
│       ├── root.go          # Cobra root, DB init
│       ├── add.go           # position add
│       ├── current.go       # position current
│       ├── timeline.go      # position timeline
│       ├── list.go          # position list
│       ├── remove.go        # position remove
│       └── config.go        # position config
├── internal/
│   ├── db/
│   │   ├── db.go            # Init, migrations
│   │   ├── items.go         # Item CRUD
│   │   └── positions.go     # Position CRUD
│   ├── models/
│   │   └── models.go        # Item, Position structs
│   ├── mcp/
│   │   ├── server.go        # MCP server setup
│   │   ├── tools.go         # Tool handlers
│   │   ├── resources.go     # Resource handlers
│   │   └── prompts.go       # Prompt templates
│   └── ui/
│       └── format.go        # Terminal output formatting
├── test/
│   └── integration_test.go  # End-to-end tests
├── go.mod
├── Makefile
├── .goreleaser.yml
└── CLAUDE.md
```

## Config & Geocoding (Future Feature)

### Config File
`~/.config/position/config.yaml`

```yaml
geocode:
  enabled: false
  provider: "nominatim"  # free, no key needed, rate-limited
  api_key: ""            # for paid providers like Google/Mapbox
```

### Geocoding Behavior (when enabled)
```bash
# Instead of coords, you can use an address
position add harper --address "123 Main St, Chicago, IL"
# Auto-geocodes to lat/lng, stores label as the address

# Or just a city
position add harper --address "chicago"
```

### For v1
Skip geocoding entirely. Just lat/lng + optional label. Geocoding is a clean add-on later - the data model already supports it.

## Dependencies

- `github.com/spf13/cobra` - CLI
- `modernc.org/sqlite` - Database
- `github.com/google/uuid` - IDs
- `github.com/modelcontextprotocol/go-sdk` - MCP
- `github.com/fatih/color` - Terminal colors
