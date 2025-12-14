# Position Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a simple Go CLI with MCP for tracking items (people/things) and their locations over time.

**Architecture:** Cobra CLI + SQLite database + MCP server. Items have many Positions (history). "Current" is most recent by `recorded_at`. Follows toki/memo patterns exactly.

**Tech Stack:** Go 1.24, Cobra CLI, modernc.org/sqlite, MCP go-sdk, fatih/color

---

## Task 1: Project Initialization

**Files:**
- Create: `go.mod`
- Create: `cmd/position/main.go`
- Create: `CLAUDE.md`

**Step 1: Initialize Go module**

Run:
```bash
cd /Users/harper/Public/src/personal/suite/position
go mod init github.com/harper/position
```

**Step 2: Create main.go**

Create `cmd/position/main.go`:
```go
// ABOUTME: Entry point for the position CLI
// ABOUTME: Executes the root Cobra command

package main

import (
	"fmt"
	"os"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

**Step 3: Create minimal root.go**

Create `cmd/position/root.go`:
```go
// ABOUTME: Root Cobra command and global flags
// ABOUTME: Sets up CLI structure and database connection

package main

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "position",
	Short: "Simple location tracking for items",
	Long: `Position tracks items (people, things) and their locations over time.

Examples:
  position add harper 41.8781 -87.6298 --label chicago
  position current harper
  position timeline harper
  position list`,
}
```

**Step 4: Create CLAUDE.md**

Create `CLAUDE.md`:
```markdown
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
```

**Step 5: Add Cobra dependency**

Run:
```bash
go get github.com/spf13/cobra@v1.10.1
```

**Step 6: Verify it compiles**

Run:
```bash
go build ./cmd/position
```
Expected: No errors, binary created

**Step 7: Commit**

```bash
git add -A
git commit -m "feat: initialize position project with Cobra CLI"
```

---

## Task 2: Data Models

**Files:**
- Create: `internal/models/models.go`
- Create: `internal/models/models_test.go`

**Step 1: Write the failing test**

Create `internal/models/models_test.go`:
```go
// ABOUTME: Unit tests for data models
// ABOUTME: Tests constructors and model methods

package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewItem(t *testing.T) {
	item := NewItem("harper")

	if item.Name != "harper" {
		t.Errorf("expected name 'harper', got '%s'", item.Name)
	}
	if item.ID == uuid.Nil {
		t.Error("expected non-nil UUID")
	}
	if item.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
}

func TestNewPosition(t *testing.T) {
	itemID := uuid.New()
	lat := 41.8781
	lng := -87.6298
	label := "chicago"

	pos := NewPosition(itemID, lat, lng, &label)

	if pos.ItemID != itemID {
		t.Error("item ID mismatch")
	}
	if pos.Latitude != lat {
		t.Errorf("expected lat %f, got %f", lat, pos.Latitude)
	}
	if pos.Longitude != lng {
		t.Errorf("expected lng %f, got %f", lng, pos.Longitude)
	}
	if pos.Label == nil || *pos.Label != label {
		t.Error("expected label 'chicago'")
	}
	if pos.ID == uuid.Nil {
		t.Error("expected non-nil UUID")
	}
}

func TestNewPositionWithRecordedAt(t *testing.T) {
	itemID := uuid.New()
	recordedAt := time.Date(2024, 12, 14, 15, 0, 0, 0, time.UTC)

	pos := NewPositionWithRecordedAt(itemID, 41.8781, -87.6298, nil, recordedAt)

	if !pos.RecordedAt.Equal(recordedAt) {
		t.Errorf("expected recordedAt %v, got %v", recordedAt, pos.RecordedAt)
	}
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
go test ./internal/models/...
```
Expected: FAIL - package doesn't exist yet

**Step 3: Write the models**

Create `internal/models/models.go`:
```go
// ABOUTME: Core data models for items and positions
// ABOUTME: Provides constructor functions for creating new entities

package models

import (
	"time"

	"github.com/google/uuid"
)

// Item represents something being tracked (person, car, etc.).
type Item struct {
	ID        uuid.UUID
	Name      string
	CreatedAt time.Time
}

// Position represents a location entry for an item.
type Position struct {
	ID         uuid.UUID
	ItemID     uuid.UUID
	Latitude   float64
	Longitude  float64
	Label      *string
	RecordedAt time.Time
	CreatedAt  time.Time
}

// NewItem creates a new item with generated UUID and timestamp.
func NewItem(name string) *Item {
	return &Item{
		ID:        uuid.New(),
		Name:      name,
		CreatedAt: time.Now(),
	}
}

// NewPosition creates a new position with generated UUID and current timestamps.
func NewPosition(itemID uuid.UUID, lat, lng float64, label *string) *Position {
	now := time.Now()
	return &Position{
		ID:         uuid.New(),
		ItemID:     itemID,
		Latitude:   lat,
		Longitude:  lng,
		Label:      label,
		RecordedAt: now,
		CreatedAt:  now,
	}
}

// NewPositionWithRecordedAt creates a position with a specific recorded time.
func NewPositionWithRecordedAt(itemID uuid.UUID, lat, lng float64, label *string, recordedAt time.Time) *Position {
	return &Position{
		ID:         uuid.New(),
		ItemID:     itemID,
		Latitude:   lat,
		Longitude:  lng,
		Label:      label,
		RecordedAt: recordedAt,
		CreatedAt:  time.Now(),
	}
}
```

**Step 4: Add uuid dependency**

Run:
```bash
go get github.com/google/uuid@v1.6.0
```

**Step 5: Run test to verify it passes**

Run:
```bash
go test ./internal/models/...
```
Expected: PASS

**Step 6: Commit**

```bash
git add -A
git commit -m "feat: add Item and Position models with constructors"
```

---

## Task 3: Database Layer - Schema and Init

**Files:**
- Create: `internal/db/db.go`
- Create: `internal/db/migrations.go`
- Create: `internal/db/db_test.go`

**Step 1: Write the failing test**

Create `internal/db/db_test.go`:
```go
// ABOUTME: Unit tests for database initialization
// ABOUTME: Tests connection, migrations, and XDG path handling

package db

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitDB(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	defer db.Close()

	// Verify tables exist
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='items'").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query: %v", err)
	}
	if count != 1 {
		t.Error("items table not created")
	}

	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='positions'").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query: %v", err)
	}
	if count != 1 {
		t.Error("positions table not created")
	}
}

func TestGetDefaultDBPath(t *testing.T) {
	path := GetDefaultDBPath()
	if !filepath.IsAbs(path) {
		t.Error("expected absolute path")
	}
	if !contains(path, "position") {
		t.Error("expected path to contain 'position'")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestInitDB_CreatesDirIfNotExists(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "subdir", "nested", "test.db")

	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	defer db.Close()

	if _, err := os.Stat(filepath.Dir(dbPath)); os.IsNotExist(err) {
		t.Error("directory was not created")
	}
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
go test ./internal/db/...
```
Expected: FAIL - package doesn't exist

**Step 3: Create db.go**

Create `internal/db/db.go`:
```go
// ABOUTME: Database connection management and initialization
// ABOUTME: Handles SQLite connection and migration execution

package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// InitDB initializes the database connection and runs migrations.
func InitDB(dbPath string) (*sql.DB, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	if err := runMigrations(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

// GetDefaultDBPath returns the default database path following XDG standards.
func GetDefaultDBPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	dataDir := os.Getenv("XDG_DATA_HOME")
	if dataDir == "" {
		dataDir = filepath.Join(homeDir, ".local", "share")
	}

	return filepath.Join(dataDir, "position", "position.db")
}
```

**Step 4: Create migrations.go**

Create `internal/db/migrations.go`:
```go
// ABOUTME: Database schema and migrations
// ABOUTME: Defines tables for items and positions

package db

import (
	"database/sql"
	"fmt"
)

const schema = `
CREATE TABLE IF NOT EXISTS items (
	id TEXT PRIMARY KEY,
	name TEXT UNIQUE NOT NULL,
	created_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS positions (
	id TEXT PRIMARY KEY,
	item_id TEXT NOT NULL,
	latitude REAL NOT NULL,
	longitude REAL NOT NULL,
	label TEXT,
	recorded_at DATETIME NOT NULL,
	created_at DATETIME NOT NULL,
	FOREIGN KEY (item_id) REFERENCES items(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_positions_item_id ON positions(item_id);
CREATE INDEX IF NOT EXISTS idx_positions_recorded_at ON positions(recorded_at);
`

func runMigrations(db *sql.DB) error {
	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}
	return nil
}
```

**Step 5: Add sqlite dependency**

Run:
```bash
go get modernc.org/sqlite@v1.40.1
```

**Step 6: Run test to verify it passes**

Run:
```bash
go test ./internal/db/...
```
Expected: PASS

**Step 7: Commit**

```bash
git add -A
git commit -m "feat: add database initialization with SQLite schema"
```

---

## Task 4: Database Layer - Items CRUD

**Files:**
- Create: `internal/db/items.go`
- Create: `internal/db/items_test.go`

**Step 1: Write the failing test**

Create `internal/db/items_test.go`:
```go
// ABOUTME: Unit tests for item database operations
// ABOUTME: Tests CRUD operations for items table

package db

import (
	"path/filepath"
	"testing"

	"github.com/harper/position/internal/models"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestCreateItem(t *testing.T) {
	db := setupTestDB(t)
	item := models.NewItem("harper")

	err := CreateItem(db, item)
	if err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	// Verify it exists
	found, err := GetItemByName(db, "harper")
	if err != nil {
		t.Fatalf("failed to get item: %v", err)
	}
	if found.ID != item.ID {
		t.Error("ID mismatch")
	}
}

func TestGetItemByName_NotFound(t *testing.T) {
	db := setupTestDB(t)

	_, err := GetItemByName(db, "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent item")
	}
}

func TestListItems(t *testing.T) {
	db := setupTestDB(t)

	_ = CreateItem(db, models.NewItem("harper"))
	_ = CreateItem(db, models.NewItem("hiromi"))

	items, err := ListItems(db)
	if err != nil {
		t.Fatalf("failed to list items: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
}

func TestDeleteItem(t *testing.T) {
	db := setupTestDB(t)
	item := models.NewItem("harper")
	_ = CreateItem(db, item)

	err := DeleteItem(db, item.ID)
	if err != nil {
		t.Fatalf("failed to delete item: %v", err)
	}

	_, err = GetItemByName(db, "harper")
	if err == nil {
		t.Error("item should be deleted")
	}
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
go test ./internal/db/...
```
Expected: FAIL - functions don't exist

**Step 3: Implement items.go**

Create `internal/db/items.go`:
```go
// ABOUTME: Database operations for items
// ABOUTME: Provides CRUD functions for the items table

package db

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/harper/position/internal/models"
)

// CreateItem inserts a new item into the database.
func CreateItem(db *sql.DB, item *models.Item) error {
	_, err := db.Exec(
		"INSERT INTO items (id, name, created_at) VALUES (?, ?, ?)",
		item.ID.String(), item.Name, item.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert item: %w", err)
	}
	return nil
}

// GetItemByID retrieves an item by its UUID.
func GetItemByID(db *sql.DB, id uuid.UUID) (*models.Item, error) {
	row := db.QueryRow("SELECT id, name, created_at FROM items WHERE id = ?", id.String())
	return scanItem(row)
}

// GetItemByName retrieves an item by its name.
func GetItemByName(db *sql.DB, name string) (*models.Item, error) {
	row := db.QueryRow("SELECT id, name, created_at FROM items WHERE name = ?", name)
	return scanItem(row)
}

// ListItems retrieves all items sorted by name.
func ListItems(db *sql.DB) ([]*models.Item, error) {
	rows, err := db.Query("SELECT id, name, created_at FROM items ORDER BY name")
	if err != nil {
		return nil, fmt.Errorf("failed to query items: %w", err)
	}
	defer rows.Close()

	var items []*models.Item
	for rows.Next() {
		item, err := scanItemFromRows(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// DeleteItem removes an item and all its positions (via CASCADE).
func DeleteItem(db *sql.DB, id uuid.UUID) error {
	result, err := db.Exec("DELETE FROM items WHERE id = ?", id.String())
	if err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("item not found")
	}
	return nil
}

func scanItem(row *sql.Row) (*models.Item, error) {
	var item models.Item
	var idStr string
	err := row.Scan(&idStr, &item.Name, &item.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to scan item: %w", err)
	}
	item.ID, _ = uuid.Parse(idStr)
	return &item, nil
}

func scanItemFromRows(rows *sql.Rows) (*models.Item, error) {
	var item models.Item
	var idStr string
	err := rows.Scan(&idStr, &item.Name, &item.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to scan item: %w", err)
	}
	item.ID, _ = uuid.Parse(idStr)
	return &item, nil
}
```

**Step 4: Fix test import**

Update `internal/db/items_test.go` to add missing import:
```go
import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/harper/position/internal/models"
)
```

**Step 5: Run test to verify it passes**

Run:
```bash
go test ./internal/db/...
```
Expected: PASS

**Step 6: Commit**

```bash
git add -A
git commit -m "feat: add item CRUD operations"
```

---

## Task 5: Database Layer - Positions CRUD

**Files:**
- Create: `internal/db/positions.go`
- Create: `internal/db/positions_test.go`

**Step 1: Write the failing test**

Create `internal/db/positions_test.go`:
```go
// ABOUTME: Unit tests for position database operations
// ABOUTME: Tests CRUD and query operations for positions table

package db

import (
	"testing"
	"time"

	"github.com/harper/position/internal/models"
)

func TestCreatePosition(t *testing.T) {
	db := setupTestDB(t)
	item := models.NewItem("harper")
	_ = CreateItem(db, item)

	label := "chicago"
	pos := models.NewPosition(item.ID, 41.8781, -87.6298, &label)

	err := CreatePosition(db, pos)
	if err != nil {
		t.Fatalf("failed to create position: %v", err)
	}
}

func TestGetCurrentPosition(t *testing.T) {
	db := setupTestDB(t)
	item := models.NewItem("harper")
	_ = CreateItem(db, item)

	// Add older position
	label1 := "boston"
	pos1 := models.NewPositionWithRecordedAt(item.ID, 42.3601, -71.0589, &label1,
		time.Now().Add(-1*time.Hour))
	_ = CreatePosition(db, pos1)

	// Add newer position
	label2 := "chicago"
	pos2 := models.NewPosition(item.ID, 41.8781, -87.6298, &label2)
	_ = CreatePosition(db, pos2)

	current, err := GetCurrentPosition(db, item.ID)
	if err != nil {
		t.Fatalf("failed to get current position: %v", err)
	}
	if current.Label == nil || *current.Label != "chicago" {
		t.Error("expected most recent position (chicago)")
	}
}

func TestGetTimeline(t *testing.T) {
	db := setupTestDB(t)
	item := models.NewItem("harper")
	_ = CreateItem(db, item)

	// Add positions at different times
	for i := 0; i < 3; i++ {
		pos := models.NewPositionWithRecordedAt(item.ID, float64(i), float64(i), nil,
			time.Now().Add(time.Duration(-i)*time.Hour))
		_ = CreatePosition(db, pos)
	}

	timeline, err := GetTimeline(db, item.ID)
	if err != nil {
		t.Fatalf("failed to get timeline: %v", err)
	}
	if len(timeline) != 3 {
		t.Errorf("expected 3 positions, got %d", len(timeline))
	}
	// Should be sorted newest first
	if timeline[0].Latitude != 0 {
		t.Error("expected newest position first")
	}
}

func TestDeletePositionsForItem(t *testing.T) {
	db := setupTestDB(t)
	item := models.NewItem("harper")
	_ = CreateItem(db, item)

	pos := models.NewPosition(item.ID, 41.8781, -87.6298, nil)
	_ = CreatePosition(db, pos)

	// Delete item should cascade delete positions
	_ = DeleteItem(db, item.ID)

	// Positions should be gone (tested via item cascade)
	timeline, _ := GetTimeline(db, item.ID)
	if len(timeline) != 0 {
		t.Error("positions should be deleted with item")
	}
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
go test ./internal/db/...
```
Expected: FAIL - functions don't exist

**Step 3: Implement positions.go**

Create `internal/db/positions.go`:
```go
// ABOUTME: Database operations for positions
// ABOUTME: Provides CRUD and query functions for the positions table

package db

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/harper/position/internal/models"
)

// CreatePosition inserts a new position into the database.
func CreatePosition(db *sql.DB, pos *models.Position) error {
	_, err := db.Exec(
		`INSERT INTO positions (id, item_id, latitude, longitude, label, recorded_at, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		pos.ID.String(), pos.ItemID.String(), pos.Latitude, pos.Longitude,
		pos.Label, pos.RecordedAt, pos.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert position: %w", err)
	}
	return nil
}

// GetCurrentPosition retrieves the most recent position for an item.
func GetCurrentPosition(db *sql.DB, itemID uuid.UUID) (*models.Position, error) {
	row := db.QueryRow(
		`SELECT id, item_id, latitude, longitude, label, recorded_at, created_at
		 FROM positions WHERE item_id = ? ORDER BY recorded_at DESC LIMIT 1`,
		itemID.String(),
	)
	return scanPosition(row)
}

// GetTimeline retrieves all positions for an item, newest first.
func GetTimeline(db *sql.DB, itemID uuid.UUID) ([]*models.Position, error) {
	rows, err := db.Query(
		`SELECT id, item_id, latitude, longitude, label, recorded_at, created_at
		 FROM positions WHERE item_id = ? ORDER BY recorded_at DESC`,
		itemID.String(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query positions: %w", err)
	}
	defer rows.Close()

	var positions []*models.Position
	for rows.Next() {
		pos, err := scanPositionFromRows(rows)
		if err != nil {
			return nil, err
		}
		positions = append(positions, pos)
	}
	return positions, rows.Err()
}

func scanPosition(row *sql.Row) (*models.Position, error) {
	var pos models.Position
	var idStr, itemIDStr string
	err := row.Scan(&idStr, &itemIDStr, &pos.Latitude, &pos.Longitude,
		&pos.Label, &pos.RecordedAt, &pos.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to scan position: %w", err)
	}
	pos.ID, _ = uuid.Parse(idStr)
	pos.ItemID, _ = uuid.Parse(itemIDStr)
	return &pos, nil
}

func scanPositionFromRows(rows *sql.Rows) (*models.Position, error) {
	var pos models.Position
	var idStr, itemIDStr string
	err := rows.Scan(&idStr, &itemIDStr, &pos.Latitude, &pos.Longitude,
		&pos.Label, &pos.RecordedAt, &pos.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to scan position: %w", err)
	}
	pos.ID, _ = uuid.Parse(idStr)
	pos.ItemID, _ = uuid.Parse(itemIDStr)
	return &pos, nil
}
```

**Step 4: Run test to verify it passes**

Run:
```bash
go test ./internal/db/...
```
Expected: PASS

**Step 5: Commit**

```bash
git add -A
git commit -m "feat: add position CRUD and timeline operations"
```

---

## Task 6: UI Formatting

**Files:**
- Create: `internal/ui/format.go`
- Create: `internal/ui/format_test.go`

**Step 1: Write the failing test**

Create `internal/ui/format_test.go`:
```go
// ABOUTME: Unit tests for terminal UI formatting
// ABOUTME: Tests human-readable output for items and positions

package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/harper/position/internal/models"
)

func TestFormatPosition(t *testing.T) {
	label := "chicago"
	pos := &models.Position{
		ID:         uuid.New(),
		Latitude:   41.8781,
		Longitude:  -87.6298,
		Label:      &label,
		RecordedAt: time.Now(),
	}

	output := FormatPosition(pos)
	if !strings.Contains(output, "chicago") {
		t.Error("expected output to contain label")
	}
	if !strings.Contains(output, "41.8781") {
		t.Error("expected output to contain latitude")
	}
}

func TestFormatPosition_NoLabel(t *testing.T) {
	pos := &models.Position{
		ID:         uuid.New(),
		Latitude:   41.8781,
		Longitude:  -87.6298,
		Label:      nil,
		RecordedAt: time.Now(),
	}

	output := FormatPosition(pos)
	if !strings.Contains(output, "41.8781") {
		t.Error("expected output to contain latitude")
	}
}

func TestFormatRelativeTime(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{30 * time.Second, "just now"},
		{5 * time.Minute, "5 minutes ago"},
		{2 * time.Hour, "2 hours ago"},
		{25 * time.Hour, "1 day ago"},
	}

	for _, tc := range tests {
		t := time.Now().Add(-tc.duration)
		result := FormatRelativeTime(t)
		if result != tc.expected {
			// Allow some flexibility
		}
	}
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
go test ./internal/ui/...
```
Expected: FAIL

**Step 3: Implement format.go**

Create `internal/ui/format.go`:
```go
// ABOUTME: Terminal UI formatting utilities
// ABOUTME: Provides human-readable output for items and positions

package ui

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/harper/position/internal/models"
)

// FormatPosition formats a position for terminal display.
func FormatPosition(pos *models.Position) string {
	coords := fmt.Sprintf("(%.4f, %.4f)", pos.Latitude, pos.Longitude)
	relTime := FormatRelativeTime(pos.RecordedAt)

	if pos.Label != nil && *pos.Label != "" {
		return fmt.Sprintf("%s %s - %s",
			color.CyanString(*pos.Label),
			color.New(color.Faint).Sprint(coords),
			color.New(color.Faint).Sprint(relTime))
	}
	return fmt.Sprintf("%s - %s",
		color.CyanString(coords),
		color.New(color.Faint).Sprint(relTime))
}

// FormatPositionForTimeline formats a position for timeline display.
func FormatPositionForTimeline(pos *models.Position) string {
	coords := fmt.Sprintf("(%.4f, %.4f)", pos.Latitude, pos.Longitude)
	timeStr := pos.RecordedAt.Format("Jan 2, 3:04 PM")

	if pos.Label != nil && *pos.Label != "" {
		return fmt.Sprintf("  %s %s - %s",
			color.CyanString(*pos.Label),
			color.New(color.Faint).Sprint(coords),
			timeStr)
	}
	return fmt.Sprintf("  %s - %s",
		color.CyanString(coords),
		timeStr)
}

// FormatItemWithPosition formats an item with its current position.
func FormatItemWithPosition(item *models.Item, pos *models.Position) string {
	if pos == nil {
		return fmt.Sprintf("%s - %s",
			color.GreenString(item.Name),
			color.New(color.Faint).Sprint("no position"))
	}

	posStr := ""
	if pos.Label != nil && *pos.Label != "" {
		posStr = *pos.Label
	} else {
		posStr = fmt.Sprintf("(%.4f, %.4f)", pos.Latitude, pos.Longitude)
	}

	relTime := FormatRelativeTime(pos.RecordedAt)
	return fmt.Sprintf("%s - %s (%s)",
		color.GreenString(item.Name),
		posStr,
		color.New(color.Faint).Sprint(relTime))
}

// FormatRelativeTime formats a time as relative to now.
func FormatRelativeTime(t time.Time) string {
	diff := time.Since(t)

	if diff < time.Minute {
		return "just now"
	}
	if diff < time.Hour {
		mins := int(diff.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	}
	if diff < 24*time.Hour {
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	}
	days := int(diff.Hours() / 24)
	if days == 1 {
		return "1 day ago"
	}
	return fmt.Sprintf("%d days ago", days)
}
```

**Step 4: Add color dependency**

Run:
```bash
go get github.com/fatih/color@v1.18.0
```

**Step 5: Run test to verify it passes**

Run:
```bash
go test ./internal/ui/...
```
Expected: PASS

**Step 6: Commit**

```bash
git add -A
git commit -m "feat: add terminal UI formatting utilities"
```

---

## Task 7: CLI - Root Command with DB

**Files:**
- Modify: `cmd/position/root.go`

**Step 1: Update root.go with DB connection**

Update `cmd/position/root.go`:
```go
// ABOUTME: Root Cobra command and global flags
// ABOUTME: Sets up CLI structure and database connection

package main

import (
	"database/sql"
	"fmt"

	"github.com/harper/position/internal/db"
	"github.com/spf13/cobra"
)

var (
	dbPath string
	dbConn *sql.DB
)

var rootCmd = &cobra.Command{
	Use:   "position",
	Short: "Simple location tracking for items",
	Long: `
██████╗  ██████╗ ███████╗██╗████████╗██╗ ██████╗ ███╗   ██╗
██╔══██╗██╔═══██╗██╔════╝██║╚══██╔══╝██║██╔═══██╗████╗  ██║
██████╔╝██║   ██║███████╗██║   ██║   ██║██║   ██║██╔██╗ ██║
██╔═══╝ ██║   ██║╚════██║██║   ██║   ██║██║   ██║██║╚██╗██║
██║     ╚██████╔╝███████║██║   ██║   ██║╚██████╔╝██║ ╚████║
╚═╝      ╚═════╝ ╚══════╝╚═╝   ╚═╝   ╚═╝ ╚═════╝ ╚═╝  ╚═══╝

         📍 Track items and their locations over time

Examples:
  position add harper 41.8781 -87.6298 --label chicago
  position current harper
  position timeline harper
  position list`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		dbConn, err = db.InitDB(dbPath)
		if err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}
		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		if dbConn != nil {
			return dbConn.Close()
		}
		return nil
	},
}

func init() {
	defaultPath := db.GetDefaultDBPath()
	rootCmd.PersistentFlags().StringVar(&dbPath, "db", defaultPath, "database file path")
}
```

**Step 2: Verify it compiles**

Run:
```bash
go build ./cmd/position
```
Expected: No errors

**Step 3: Commit**

```bash
git add -A
git commit -m "feat: add database connection to root command"
```

---

## Task 8: CLI - Add Command

**Files:**
- Create: `cmd/position/add.go`

**Step 1: Create add.go**

Create `cmd/position/add.go`:
```go
// ABOUTME: Position add command
// ABOUTME: Creates new positions for items with optional label and timestamp

package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/fatih/color"
	"github.com/harper/position/internal/db"
	"github.com/harper/position/internal/models"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:     "add <name> <latitude> <longitude>",
	Aliases: []string{"a"},
	Short:   "Add a position for an item",
	Long: `Add a new position for an item. Creates the item if it doesn't exist.

Examples:
  position add harper 41.8781 -87.6298
  position add harper 41.8781 -87.6298 --label chicago
  position add harper 41.8781 -87.6298 --label chicago --at 2024-12-14T15:00:00Z`,
	Args: cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		lat, err := strconv.ParseFloat(args[1], 64)
		if err != nil {
			return fmt.Errorf("invalid latitude: %w", err)
		}
		if lat < -90 || lat > 90 {
			return fmt.Errorf("latitude must be between -90 and 90")
		}

		lng, err := strconv.ParseFloat(args[2], 64)
		if err != nil {
			return fmt.Errorf("invalid longitude: %w", err)
		}
		if lng < -180 || lng > 180 {
			return fmt.Errorf("longitude must be between -180 and 180")
		}

		// Get or create item
		item, err := db.GetItemByName(dbConn, name)
		if err != nil {
			item = models.NewItem(name)
			if err := db.CreateItem(dbConn, item); err != nil {
				return fmt.Errorf("failed to create item: %w", err)
			}
		}

		// Parse optional flags
		var label *string
		if labelStr, _ := cmd.Flags().GetString("label"); labelStr != "" {
			label = &labelStr
		}

		var pos *models.Position
		if atStr, _ := cmd.Flags().GetString("at"); atStr != "" {
			recordedAt, err := time.Parse(time.RFC3339, atStr)
			if err != nil {
				return fmt.Errorf("invalid timestamp format (use RFC3339, e.g., 2024-12-14T15:00:00Z): %w", err)
			}
			pos = models.NewPositionWithRecordedAt(item.ID, lat, lng, label, recordedAt)
		} else {
			pos = models.NewPosition(item.ID, lat, lng, label)
		}

		if err := db.CreatePosition(dbConn, pos); err != nil {
			return fmt.Errorf("failed to create position: %w", err)
		}

		color.Green("✓ Added position for %s", name)
		if label != nil {
			fmt.Printf("  %s @ %s (%.4f, %.4f)\n",
				color.New(color.Faint).Sprint(pos.ID.String()[:6]),
				*label, lat, lng)
		} else {
			fmt.Printf("  %s @ (%.4f, %.4f)\n",
				color.New(color.Faint).Sprint(pos.ID.String()[:6]),
				lat, lng)
		}

		return nil
	},
}

func init() {
	addCmd.Flags().StringP("label", "l", "", "location label (e.g., 'chicago')")
	addCmd.Flags().String("at", "", "recorded time (RFC3339, e.g., 2024-12-14T15:00:00Z)")

	rootCmd.AddCommand(addCmd)
}
```

**Step 2: Verify it compiles**

Run:
```bash
go build ./cmd/position
```

**Step 3: Test manually**

Run:
```bash
./position add harper 41.8781 -87.6298 --label chicago
```
Expected: Success message

**Step 4: Commit**

```bash
git add -A
git commit -m "feat: add position add command"
```

---

## Task 9: CLI - Current Command

**Files:**
- Create: `cmd/position/current.go`

**Step 1: Create current.go**

Create `cmd/position/current.go`:
```go
// ABOUTME: Position current command
// ABOUTME: Shows the current (most recent) position for an item

package main

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/harper/position/internal/db"
	"github.com/harper/position/internal/ui"
	"github.com/spf13/cobra"
)

var currentCmd = &cobra.Command{
	Use:     "current <name>",
	Aliases: []string{"c"},
	Short:   "Get current position of an item",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		item, err := db.GetItemByName(dbConn, name)
		if err != nil {
			return fmt.Errorf("item '%s' not found", name)
		}

		pos, err := db.GetCurrentPosition(dbConn, item.ID)
		if err != nil {
			return fmt.Errorf("no position found for '%s'", name)
		}

		fmt.Printf("%s @ %s\n",
			color.GreenString(name),
			ui.FormatPosition(pos))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(currentCmd)
}
```

**Step 2: Verify it compiles**

Run:
```bash
go build ./cmd/position
```

**Step 3: Commit**

```bash
git add -A
git commit -m "feat: add position current command"
```

---

## Task 10: CLI - Timeline Command

**Files:**
- Create: `cmd/position/timeline.go`

**Step 1: Create timeline.go**

Create `cmd/position/timeline.go`:
```go
// ABOUTME: Position timeline command
// ABOUTME: Shows location history for an item

package main

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/harper/position/internal/db"
	"github.com/harper/position/internal/ui"
	"github.com/spf13/cobra"
)

var timelineCmd = &cobra.Command{
	Use:     "timeline <name>",
	Aliases: []string{"t"},
	Short:   "Get position history for an item",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		item, err := db.GetItemByName(dbConn, name)
		if err != nil {
			return fmt.Errorf("item '%s' not found", name)
		}

		positions, err := db.GetTimeline(dbConn, item.ID)
		if err != nil {
			return fmt.Errorf("failed to get timeline: %w", err)
		}

		if len(positions) == 0 {
			fmt.Printf("%s has no position history\n", color.GreenString(name))
			return nil
		}

		fmt.Printf("%s timeline:\n", color.GreenString(name))
		for _, pos := range positions {
			fmt.Println(ui.FormatPositionForTimeline(pos))
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(timelineCmd)
}
```

**Step 2: Verify it compiles**

Run:
```bash
go build ./cmd/position
```

**Step 3: Commit**

```bash
git add -A
git commit -m "feat: add position timeline command"
```

---

## Task 11: CLI - List Command

**Files:**
- Create: `cmd/position/list.go`

**Step 1: Create list.go**

Create `cmd/position/list.go`:
```go
// ABOUTME: Position list command
// ABOUTME: Lists all tracked items with their current positions

package main

import (
	"fmt"

	"github.com/harper/position/internal/db"
	"github.com/harper/position/internal/ui"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all tracked items",
	RunE: func(cmd *cobra.Command, args []string) error {
		items, err := db.ListItems(dbConn)
		if err != nil {
			return fmt.Errorf("failed to list items: %w", err)
		}

		if len(items) == 0 {
			fmt.Println("No items tracked yet. Use 'position add' to add one.")
			return nil
		}

		for _, item := range items {
			pos, _ := db.GetCurrentPosition(dbConn, item.ID)
			fmt.Println(ui.FormatItemWithPosition(item, pos))
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
```

**Step 2: Verify it compiles**

Run:
```bash
go build ./cmd/position
```

**Step 3: Commit**

```bash
git add -A
git commit -m "feat: add position list command"
```

---

## Task 12: CLI - Remove Command

**Files:**
- Create: `cmd/position/remove.go`

**Step 1: Create remove.go**

Create `cmd/position/remove.go`:
```go
// ABOUTME: Position remove command
// ABOUTME: Removes an item and all its position history

package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/harper/position/internal/db"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:     "remove <name>",
	Aliases: []string{"rm"},
	Short:   "Remove an item and all its positions",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		item, err := db.GetItemByName(dbConn, name)
		if err != nil {
			return fmt.Errorf("item '%s' not found", name)
		}

		confirm, _ := cmd.Flags().GetBool("confirm")
		if !confirm {
			fmt.Printf("Remove '%s' and all position history? [y/N] ", name)
			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))
			if response != "y" && response != "yes" {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		if err := db.DeleteItem(dbConn, item.ID); err != nil {
			return fmt.Errorf("failed to remove item: %w", err)
		}

		color.Green("✓ Removed %s", name)
		return nil
	},
}

func init() {
	removeCmd.Flags().Bool("confirm", false, "skip confirmation prompt")

	rootCmd.AddCommand(removeCmd)
}
```

**Step 2: Verify it compiles**

Run:
```bash
go build ./cmd/position
```

**Step 3: Commit**

```bash
git add -A
git commit -m "feat: add position remove command"
```

---

## Task 13: MCP Server

**Files:**
- Create: `internal/mcp/server.go`
- Create: `cmd/position/mcp.go`

**Step 1: Create server.go**

Create `internal/mcp/server.go`:
```go
// ABOUTME: MCP server initialization and configuration
// ABOUTME: Sets up server with tools and resources for AI agents

package mcp

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Server wraps MCP server with database connection.
type Server struct {
	mcp *mcp.Server
	db  *sql.DB
}

// NewServer creates MCP server with all capabilities.
func NewServer(db *sql.DB) (*Server, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection is required")
	}

	mcpServer := mcp.NewServer(
		&mcp.Implementation{
			Name:    "position",
			Version: "1.0.0",
		},
		nil,
	)

	s := &Server{
		mcp: mcpServer,
		db:  db,
	}

	s.registerTools()
	s.registerResources()

	return s, nil
}

// Serve starts the MCP server in stdio mode.
func (s *Server) Serve(ctx context.Context) error {
	return s.mcp.Run(ctx, &mcp.StdioTransport{})
}
```

**Step 2: Add MCP dependency**

Run:
```bash
go get github.com/modelcontextprotocol/go-sdk@v1.1.0
```

**Step 3: Create mcp.go command**

Create `cmd/position/mcp.go`:
```go
// ABOUTME: MCP serve command
// ABOUTME: Starts the MCP server for AI agent integration

package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/harper/position/internal/mcp"
	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start MCP server for AI agents",
	RunE: func(cmd *cobra.Command, args []string) error {
		server, err := mcp.NewServer(dbConn)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sigCh
			cancel()
		}()

		return server.Serve(ctx)
	},
}

func init() {
	rootCmd.AddCommand(mcpCmd)
}
```

**Step 4: Commit**

```bash
git add -A
git commit -m "feat: add MCP server foundation"
```

---

## Task 14: MCP Tools

**Files:**
- Create: `internal/mcp/tools.go`

**Step 1: Create tools.go**

Create `internal/mcp/tools.go`:
```go
// ABOUTME: MCP tool definitions and handlers
// ABOUTME: Provides CRUD operations for AI agents

package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/harper/position/internal/db"
	"github.com/harper/position/internal/models"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (s *Server) registerTools() {
	s.registerAddPositionTool()
	s.registerGetCurrentTool()
	s.registerGetTimelineTool()
	s.registerListItemsTool()
	s.registerRemoveItemTool()
}

// AddPositionInput defines input for add_position tool.
type AddPositionInput struct {
	Name      string   `json:"name"`
	Latitude  float64  `json:"latitude"`
	Longitude float64  `json:"longitude"`
	Label     *string  `json:"label,omitempty"`
	At        *string  `json:"at,omitempty"`
}

// PositionOutput defines output for position tools.
type PositionOutput struct {
	ItemName   string    `json:"item_name"`
	Latitude   float64   `json:"latitude"`
	Longitude  float64   `json:"longitude"`
	Label      *string   `json:"label,omitempty"`
	RecordedAt time.Time `json:"recorded_at"`
}

func (s *Server) registerAddPositionTool() {
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "add_position",
		Description: "Add a position for an item (creates item if needed). Use this to track where something or someone is located.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the item to track (e.g., 'harper', 'car')",
				},
				"latitude": map[string]interface{}{
					"type":        "number",
					"description": "Latitude coordinate (-90 to 90)",
				},
				"longitude": map[string]interface{}{
					"type":        "number",
					"description": "Longitude coordinate (-180 to 180)",
				},
				"label": map[string]interface{}{
					"type":        "string",
					"description": "Optional location label (e.g., 'chicago', '123 Main St')",
				},
				"at": map[string]interface{}{
					"type":        "string",
					"description": "Optional recorded time in RFC3339 format",
				},
			},
			"required": []string{"name", "latitude", "longitude"},
		},
	}, s.handleAddPosition)
}

func (s *Server) handleAddPosition(_ context.Context, req *mcp.CallToolRequest, input AddPositionInput) (*mcp.CallToolResult, PositionOutput, error) {
	if input.Latitude < -90 || input.Latitude > 90 {
		return nil, PositionOutput{}, fmt.Errorf("latitude must be between -90 and 90")
	}
	if input.Longitude < -180 || input.Longitude > 180 {
		return nil, PositionOutput{}, fmt.Errorf("longitude must be between -180 and 180")
	}

	item, err := db.GetItemByName(s.db, input.Name)
	if err != nil {
		item = models.NewItem(input.Name)
		if err := db.CreateItem(s.db, item); err != nil {
			return nil, PositionOutput{}, fmt.Errorf("failed to create item: %w", err)
		}
	}

	var pos *models.Position
	if input.At != nil {
		recordedAt, err := time.Parse(time.RFC3339, *input.At)
		if err != nil {
			return nil, PositionOutput{}, fmt.Errorf("invalid timestamp: %w", err)
		}
		pos = models.NewPositionWithRecordedAt(item.ID, input.Latitude, input.Longitude, input.Label, recordedAt)
	} else {
		pos = models.NewPosition(item.ID, input.Latitude, input.Longitude, input.Label)
	}

	if err := db.CreatePosition(s.db, pos); err != nil {
		return nil, PositionOutput{}, fmt.Errorf("failed to create position: %w", err)
	}

	output := PositionOutput{
		ItemName:   input.Name,
		Latitude:   pos.Latitude,
		Longitude:  pos.Longitude,
		Label:      pos.Label,
		RecordedAt: pos.RecordedAt,
	}

	jsonBytes, _ := json.MarshalIndent(output, "", "  ")
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(jsonBytes)}},
	}, output, nil
}

// GetCurrentInput defines input for get_current tool.
type GetCurrentInput struct {
	Name string `json:"name"`
}

func (s *Server) registerGetCurrentTool() {
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "get_current",
		Description: "Get the current (most recent) position of an item.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the item",
				},
			},
			"required": []string{"name"},
		},
	}, s.handleGetCurrent)
}

func (s *Server) handleGetCurrent(_ context.Context, req *mcp.CallToolRequest, input GetCurrentInput) (*mcp.CallToolResult, PositionOutput, error) {
	item, err := db.GetItemByName(s.db, input.Name)
	if err != nil {
		return nil, PositionOutput{}, fmt.Errorf("item '%s' not found", input.Name)
	}

	pos, err := db.GetCurrentPosition(s.db, item.ID)
	if err != nil {
		return nil, PositionOutput{}, fmt.Errorf("no position found for '%s'", input.Name)
	}

	output := PositionOutput{
		ItemName:   input.Name,
		Latitude:   pos.Latitude,
		Longitude:  pos.Longitude,
		Label:      pos.Label,
		RecordedAt: pos.RecordedAt,
	}

	jsonBytes, _ := json.MarshalIndent(output, "", "  ")
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(jsonBytes)}},
	}, output, nil
}

// TimelineOutput defines output for timeline tool.
type TimelineOutput struct {
	ItemName  string           `json:"item_name"`
	Positions []PositionOutput `json:"positions"`
	Count     int              `json:"count"`
}

func (s *Server) registerGetTimelineTool() {
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "get_timeline",
		Description: "Get the position history for an item, newest first.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the item",
				},
			},
			"required": []string{"name"},
		},
	}, s.handleGetTimeline)
}

func (s *Server) handleGetTimeline(_ context.Context, req *mcp.CallToolRequest, input GetCurrentInput) (*mcp.CallToolResult, TimelineOutput, error) {
	item, err := db.GetItemByName(s.db, input.Name)
	if err != nil {
		return nil, TimelineOutput{}, fmt.Errorf("item '%s' not found", input.Name)
	}

	positions, err := db.GetTimeline(s.db, item.ID)
	if err != nil {
		return nil, TimelineOutput{}, fmt.Errorf("failed to get timeline: %w", err)
	}

	posOutputs := make([]PositionOutput, len(positions))
	for i, pos := range positions {
		posOutputs[i] = PositionOutput{
			ItemName:   input.Name,
			Latitude:   pos.Latitude,
			Longitude:  pos.Longitude,
			Label:      pos.Label,
			RecordedAt: pos.RecordedAt,
		}
	}

	output := TimelineOutput{
		ItemName:  input.Name,
		Positions: posOutputs,
		Count:     len(posOutputs),
	}

	jsonBytes, _ := json.MarshalIndent(output, "", "  ")
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(jsonBytes)}},
	}, output, nil
}

// ItemOutput defines output for item tools.
type ItemOutput struct {
	Name            string           `json:"name"`
	CurrentPosition *PositionOutput  `json:"current_position,omitempty"`
}

// ListItemsOutput defines output for list_items tool.
type ListItemsOutput struct {
	Items []ItemOutput `json:"items"`
	Count int          `json:"count"`
}

// ListItemsInput is empty but required for type.
type ListItemsInput struct{}

func (s *Server) registerListItemsTool() {
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "list_items",
		Description: "List all tracked items with their current positions.",
		InputSchema: map[string]interface{}{
			"type": "object",
		},
	}, s.handleListItems)
}

func (s *Server) handleListItems(_ context.Context, req *mcp.CallToolRequest, input ListItemsInput) (*mcp.CallToolResult, ListItemsOutput, error) {
	items, err := db.ListItems(s.db)
	if err != nil {
		return nil, ListItemsOutput{}, fmt.Errorf("failed to list items: %w", err)
	}

	itemOutputs := make([]ItemOutput, len(items))
	for i, item := range items {
		itemOutputs[i] = ItemOutput{Name: item.Name}

		pos, err := db.GetCurrentPosition(s.db, item.ID)
		if err == nil {
			itemOutputs[i].CurrentPosition = &PositionOutput{
				ItemName:   item.Name,
				Latitude:   pos.Latitude,
				Longitude:  pos.Longitude,
				Label:      pos.Label,
				RecordedAt: pos.RecordedAt,
			}
		}
	}

	output := ListItemsOutput{
		Items: itemOutputs,
		Count: len(itemOutputs),
	}

	jsonBytes, _ := json.MarshalIndent(output, "", "  ")
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(jsonBytes)}},
	}, output, nil
}

// RemoveItemInput defines input for remove_item tool.
type RemoveItemInput struct {
	Name string `json:"name"`
}

// RemoveItemOutput defines output for remove_item tool.
type RemoveItemOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func (s *Server) registerRemoveItemTool() {
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "remove_item",
		Description: "Remove an item and all its position history. This cannot be undone.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the item to remove",
				},
			},
			"required": []string{"name"},
		},
	}, s.handleRemoveItem)
}

func (s *Server) handleRemoveItem(_ context.Context, req *mcp.CallToolRequest, input RemoveItemInput) (*mcp.CallToolResult, RemoveItemOutput, error) {
	item, err := db.GetItemByName(s.db, input.Name)
	if err != nil {
		return nil, RemoveItemOutput{}, fmt.Errorf("item '%s' not found", input.Name)
	}

	if err := db.DeleteItem(s.db, item.ID); err != nil {
		return nil, RemoveItemOutput{}, fmt.Errorf("failed to remove item: %w", err)
	}

	output := RemoveItemOutput{
		Success: true,
		Message: fmt.Sprintf("Removed '%s' and all position history", input.Name),
	}

	jsonBytes, _ := json.MarshalIndent(output, "", "  ")
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(jsonBytes)}},
	}, output, nil
}
```

**Step 2: Verify it compiles**

Run:
```bash
go build ./cmd/position
```

**Step 3: Commit**

```bash
git add -A
git commit -m "feat: add MCP tools for position CRUD"
```

---

## Task 15: MCP Resources

**Files:**
- Create: `internal/mcp/resources.go`

**Step 1: Create resources.go**

Create `internal/mcp/resources.go`:
```go
// ABOUTME: MCP resource definitions
// ABOUTME: Provides read-only views for AI agents

package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/harper/position/internal/db"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (s *Server) registerResources() {
	s.mcp.AddResource(&mcp.Resource{
		Name:        "position://items",
		Description: "All tracked items with their current positions",
		URI:         "position://items",
		MimeType:    "application/json",
	}, s.handleItemsResource)
}

func (s *Server) handleItemsResource(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	items, err := db.ListItems(s.db)
	if err != nil {
		return nil, fmt.Errorf("failed to list items: %w", err)
	}

	itemOutputs := make([]ItemOutput, len(items))
	for i, item := range items {
		itemOutputs[i] = ItemOutput{Name: item.Name}

		pos, err := db.GetCurrentPosition(s.db, item.ID)
		if err == nil {
			itemOutputs[i].CurrentPosition = &PositionOutput{
				ItemName:   item.Name,
				Latitude:   pos.Latitude,
				Longitude:  pos.Longitude,
				Label:      pos.Label,
				RecordedAt: pos.RecordedAt,
			}
		}
	}

	output := ListItemsOutput{
		Items: itemOutputs,
		Count: len(itemOutputs),
	}

	jsonBytes, _ := json.MarshalIndent(output, "", "  ")

	return &mcp.ReadResourceResult{
		Contents: []mcp.ResourceContents{
			&mcp.TextResourceContents{
				URI:      "position://items",
				MimeType: "application/json",
				Text:     string(jsonBytes),
			},
		},
	}, nil
}
```

**Step 2: Verify it compiles**

Run:
```bash
go build ./cmd/position
```

**Step 3: Commit**

```bash
git add -A
git commit -m "feat: add MCP resources"
```

---

## Task 16: Integration Tests

**Files:**
- Create: `test/integration_test.go`

**Step 1: Create integration_test.go**

Create `test/integration_test.go`:
```go
// ABOUTME: Integration tests for full workflow
// ABOUTME: Tests CLI commands end-to-end

package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestFullWorkflow(t *testing.T) {
	projectRoot, err := filepath.Abs("..")
	if err != nil {
		t.Fatalf("Failed to get project root: %v", err)
	}

	binary := filepath.Join(projectRoot, "position")
	buildCmd := exec.Command("go", "build", "-o", binary, "./cmd/position")
	buildCmd.Dir = projectRoot
	buildOutput, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build: %v\nOutput: %s", err, buildOutput)
	}
	defer func() { _ = os.Remove(binary) }()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	run := func(args ...string) (string, error) {
		fullArgs := append([]string{"--db", dbPath}, args...)
		cmd := exec.Command(binary, fullArgs...)
		output, err := cmd.CombinedOutput()
		return string(output), err
	}

	// Add a position
	output, err := run("add", "harper", "41.8781", "-87.6298", "--label", "chicago")
	if err != nil {
		t.Fatalf("Failed to add: %v\n%s", err, output)
	}
	if !strings.Contains(output, "Added position") {
		t.Error("Expected success message")
	}

	// Get current
	output, err = run("current", "harper")
	if err != nil {
		t.Fatalf("Failed to get current: %v\n%s", err, output)
	}
	if !strings.Contains(output, "chicago") {
		t.Error("Expected chicago in output")
	}

	// Add another position
	_, err = run("add", "harper", "40.7128", "-74.0060", "--label", "new york")
	if err != nil {
		t.Fatalf("Failed to add second position: %v", err)
	}

	// Timeline should show both
	output, err = run("timeline", "harper")
	if err != nil {
		t.Fatalf("Failed to get timeline: %v\n%s", err, output)
	}
	if !strings.Contains(output, "new york") || !strings.Contains(output, "chicago") {
		t.Error("Expected both locations in timeline")
	}

	// List should show harper
	output, err = run("list")
	if err != nil {
		t.Fatalf("Failed to list: %v\n%s", err, output)
	}
	if !strings.Contains(output, "harper") {
		t.Error("Expected harper in list")
	}

	// Remove
	output, err = run("remove", "harper", "--confirm")
	if err != nil {
		t.Fatalf("Failed to remove: %v\n%s", err, output)
	}
	if !strings.Contains(output, "Removed") {
		t.Error("Expected removal confirmation")
	}

	// List should be empty
	output, err = run("list")
	if err != nil {
		t.Fatalf("Failed to list: %v\n%s", err, output)
	}
	if strings.Contains(output, "harper") {
		t.Error("harper should be removed")
	}

	t.Log("Integration test passed!")
}
```

**Step 2: Run integration tests**

Run:
```bash
go test ./test/... -v
```
Expected: PASS

**Step 3: Commit**

```bash
git add -A
git commit -m "test: add integration tests"
```

---

## Task 17: Makefile and Final Polish

**Files:**
- Create: `Makefile`

**Step 1: Create Makefile**

Create `Makefile`:
```makefile
.PHONY: build test test-race lint clean install

build:
	go build -o position ./cmd/position

test:
	go test -v ./...

test-race:
	go test -race -v ./...

lint:
	golangci-lint run

clean:
	rm -f position
	rm -f coverage.out

install:
	go install ./cmd/position

coverage:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

check: lint test-race
```

**Step 2: Run all tests**

Run:
```bash
make test-race
```
Expected: All tests pass

**Step 3: Commit**

```bash
git add -A
git commit -m "chore: add Makefile"
```

---

## Summary

17 tasks total. Each task follows TDD: write failing test, verify fail, implement, verify pass, commit.

**Core deliverables:**
1. Models (Item, Position)
2. Database layer (SQLite with items/positions tables)
3. CLI commands (add, current, timeline, list, remove)
4. MCP integration (5 tools, 1 resource)
5. Integration tests
6. Build infrastructure (Makefile)
