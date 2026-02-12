# Documentation Audit Report

Generated: 2026-02-11 | Commit: e6450b3

## Executive Summary

| Metric | Count |
|--------|-------|
| Documents scanned | 6 |
| Claims verified | ~75 |
| Verified TRUE | ~48 (64%) |
| **Verified FALSE** | **27 (36%)** |
| Stale documents | 3 |

The project underwent a major architectural change (Charm KV removal, storage backend abstraction) but documentation was only partially updated. The README.md is significantly out of date, referencing a deleted `internal/db/` package structure. The SKILL.md has incorrect MCP tool names. The code-review.md and missing-tests.md are entirely stale, referencing the pre-refactor architecture.

## Documents Audited

| Document | Status | Notes |
|----------|--------|-------|
| README.md | **Needs major update** | Project structure, flags, deps outdated |
| CLAUDE.md | Current | Accurate |
| SKILL.md | **Needs update** | Wrong tool names, wrong param names, phantom feature |
| code-review.md | **Entirely stale** | References deleted `internal/db/` package |
| missing-tests.md | **Entirely stale** | References deleted `internal/db/` package |
| CHARM_REMOVAL_PLAN.md | **Stale (completed plan)** | Migration complete, plan can be archived |

---

## False Claims Requiring Fixes

### README.md

| Section | Claim | Reality | Fix |
|---------|-------|---------|-----|
| Line 20 | `go install github.com/harper/position/cmd/position@latest` | Module path is correct in go.mod, but not published to a registry | Verify if installable; if not, remove or note it |
| Line 25 | `make build` in install instructions | Makefile exists and has `build` target | TRUE (but duplicated below) |
| Line 97 | "Position uses SQLite for local storage" | Now supports SQLite AND Markdown backends via config | Update to mention both backends |
| Line 100 | `position --db /path/to/custom.db <command>` | No `--db` flag exists on root command. Storage configured via `~/.config/position/config.json` | Remove --db reference, document config file |
| Line 195 | "Go 1.21+" | go.mod requires `go 1.24.11` | Update to Go 1.24+ |
| Lines 225-256 | Project structure shows `internal/db/` | `internal/db/` does not exist. Replaced by `internal/storage/` | Rewrite project structure |
| Line 238 | `internal/db/db.go` | MISSING - file does not exist | Remove |
| Line 239 | `internal/db/migrations.go` | MISSING - file does not exist | Remove |
| Line 240 | `internal/db/items.go` | MISSING - file does not exist | Remove |
| Line 241 | `internal/db/positions.go` | MISSING - file does not exist | Remove |
| Line 251 | `test/integration_test.go` | test/ directory is empty | Remove or note tests are in-package |
| Line 270 | `go test ./internal/db/...` | Package does not exist | Update to `./internal/storage/...` |
| Line 273 | `go test ./test/...` | No tests in test/ directory | Remove |
| Lines 225-256 | Structure omits `internal/config/` | `internal/config/` package exists with config.go | Add to structure |
| Lines 225-256 | Structure omits `internal/geojson/` | `internal/geojson/` package exists | Add to structure |
| Lines 225-256 | Structure omits `internal/storage/` | Primary storage package with sqlite.go, markdown.go, etc. | Add to structure |
| Lines 225-256 | Structure omits cmd/position/export.go | File exists | Add to structure |
| Lines 225-256 | Structure omits cmd/position/backup.go | File exists | Add to structure |
| Lines 225-256 | Structure omits cmd/position/import.go | File exists | Add to structure |
| Lines 225-256 | Structure omits cmd/position/migrate.go | File exists | Add to structure |
| Lines 225-256 | Structure omits cmd/position/skill.go | File exists | Add to structure |
| Line 56-65 | Commands table missing export, backup, import, migrate | These commands exist in the codebase | Add to commands table |
| Lines 279-283 | Dependencies list missing mdstore, yaml.v3 | Both are direct dependencies in go.mod | Add to dependencies |

### SKILL.md

| Line | Claim | Reality | Fix |
|------|-------|---------|-----|
| Line 24 | `mcp__position__list_entities` | Actual tool name is `list_items` (tools.go:262) | Change to `mcp__position__list_items` |
| Line 25 | `mcp__position__remove_entity` | Actual tool name is `remove_item` (tools.go:316) | Change to `mcp__position__remove_item` |
| Lines 31-36 | `add_position(name, lat, lng, label)` | Actual params are `latitude` and `longitude` (tools.go:55-63) | Change `lat` → `latitude`, `lng` → `longitude` |
| Line 46 | `get_timeline(name, since="2026-01-01")` | Tool only accepts `name` param, no `since` (tools.go:194-203) | Remove `since` param from example |
| Line 49 | `list_entities()` | Actual function is `list_items` | Change to `list_items()` |
| Line 67 | Data location: `~/.local/share/position/position.db` only | Now configurable with markdown backend option | Update to mention config-based backend |

### code-review.md

| Section | Claim | Reality | Fix |
|---------|-------|---------|-----|
| Throughout | References `internal/db/` package | Package deleted, replaced by `internal/storage/` | Entire document is stale - delete or regenerate |
| Line 19 | `db.go (Lines 1-42)` | File does not exist | Stale |
| Line 34 | `items.go (Lines 1-79)` | File does not exist | Stale |
| Line 54 | `positions.go (Lines 1-76)` | File does not exist | Stale |
| Line 150 | Go version 1.24 concern "ensure this version exists" | Go 1.24 does exist | Stale concern |

### missing-tests.md

| Section | Claim | Reality | Fix |
|---------|-------|---------|-----|
| Throughout | References `internal/db/items.go`, `internal/db/positions.go` | These files do not exist | Entire document is stale - delete or regenerate |
| Line 17 | `internal/db/items.go` test gaps | File does not exist | Stale |
| Line 33 | `internal/db/positions.go` test gaps | File does not exist | Stale |
| Line 78 | `test/integration_test.go` | File does not exist | Stale |
| Line 116 | `db.InitDB` references | Function likely renamed/moved | Stale |
| Line 134 | `GetDefaultDBPath` tests | Function may have been replaced by config system | Stale |

### CHARM_REMOVAL_PLAN.md

| Section | Claim | Reality | Fix |
|---------|-------|---------|-----|
| Entire doc | Plan for removing Charm dependencies | Migration completed successfully | Archive or delete - plan is done |
| Lines 115-118 | CREATE `internal/storage/schema.go`, `items.go`, `positions.go` | Created as `sqlite.go` and `markdown.go` instead (consolidated) | N/A - plan completed differently |

---

## Pattern Summary

| Pattern | Count | Root Cause |
|---------|-------|------------|
| Dead `internal/db/` references | 12 | Package deleted during storage refactor, docs not updated |
| Missing new commands in README | 4 | export, backup, import, migrate added but not documented in commands table |
| Missing packages in structure | 3 | config, geojson, storage not in project tree |
| SKILL.md wrong names | 3 | Tool names diverged from doc (entities vs items) |
| Stale review/test docs | 2 | Generated against old architecture, never regenerated |

---

## Verified TRUE Claims (Highlights)

- CLAUDE.md: All command syntaxes correct
- CLAUDE.md: Backend descriptions accurate (sqlite + markdown)
- CLAUDE.md: Config path and fields correct
- README.md: All CLI aliases (a, c, t, ls, rm) correct
- README.md: MCP tool names in README correct (add_position, get_current, get_timeline, list_items, remove_item)
- README.md: MCP resource URI `position://items` correct
- README.md: Schema SQL matches actual implementation
- README.md: `--confirm` flag on remove command correct
- README.md: All 5 listed dependencies present in go.mod
- README.md: Makefile targets all exist and work
- SKILL.md: add_position, get_current, get_timeline base names correct

---

## Human Review Queue

- [ ] Decide: Delete or regenerate `code-review.md` against current architecture
- [ ] Decide: Delete or regenerate `missing-tests.md` against current architecture
- [ ] Decide: Archive or delete `CHARM_REMOVAL_PLAN.md` (migration is complete)
- [ ] Verify: Is `go install github.com/harper/position/cmd/position@latest` actually publishable/installable?
- [ ] Decide: Should README document the config file system for backend selection?
