---
name: position
description: Location tracking - log positions, view timelines, track entities. Use when the user mentions locations, coordinates, or wants to track where something/someone is.
---

# position - Location Tracking

Track geographic positions for entities over time with labels and notes.

## When to use position

- User mentions a location or coordinates
- User wants to track where something is
- User asks about location history
- User discusses travel or movement

## Available MCP tools

| Tool | Purpose |
|------|---------|
| `mcp__position__add_position` | Log a position |
| `mcp__position__get_current` | Get latest position |
| `mcp__position__get_timeline` | Get position history |
| `mcp__position__list_items` | List tracked items |
| `mcp__position__remove_item` | Remove an item |

## Common patterns

### Log a position
```
mcp__position__add_position(
  name="harper",
  latitude=37.7749,
  longitude=-122.4194,
  label="San Francisco Office"
)
```

### Get current location
```
mcp__position__get_current(name="harper")
```

### Get timeline
```
mcp__position__get_timeline(name="harper")
```

### List all tracked entities
```
mcp__position__list_items()
```

## CLI commands (if MCP unavailable)

```bash
position add harper --lat 37.7749 --lng -122.4194 --label "SF Office"
position current harper           # Latest position
position timeline harper          # History
position list                     # All entities
position export --format geojson  # GeoJSON export
position export --format markdown # Markdown table
```

## Data location

Configurable via `~/.config/position/config.json`. Supports `sqlite` (default) and `markdown` backends. Data stored at `~/.local/share/position/` (respects XDG_DATA_HOME).
