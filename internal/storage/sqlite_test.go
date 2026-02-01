// ABOUTME: Tests for SQLite storage implementation
// ABOUTME: Covers all repository interface methods with real database

package storage

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/harper/position/internal/models"
)

// testDB creates a temporary database for testing.
func testDB(t *testing.T) *SQLiteDB {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := NewSQLiteDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})
	return db
}

func TestNewSQLiteDB(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := NewSQLiteDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	// Verify file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("database file was not created")
	}
}

func TestNewSQLiteDB_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	nestedDir := filepath.Join(tmpDir, "nested", "path")
	dbPath := filepath.Join(nestedDir, "test.db")

	db, err := NewSQLiteDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	defer db.Close()

	// Verify directory was created
	if _, err := os.Stat(nestedDir); os.IsNotExist(err) {
		t.Error("nested directory was not created")
	}
}

func TestCreateItem(t *testing.T) {
	db := testDB(t)

	item := models.NewItem("harper")
	err := db.CreateItem(item)
	if err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	// Retrieve and verify
	got, err := db.GetItemByID(item.ID)
	if err != nil {
		t.Fatalf("failed to get item: %v", err)
	}
	if got.Name != item.Name {
		t.Errorf("got name %s, want %s", got.Name, item.Name)
	}
}

func TestCreateItem_DuplicateName(t *testing.T) {
	db := testDB(t)

	item1 := models.NewItem("harper")
	if err := db.CreateItem(item1); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	item2 := models.NewItem("harper")
	err := db.CreateItem(item2)
	if err == nil {
		t.Error("expected error for duplicate name, got nil")
	}
}

func TestGetItemByID_NotFound(t *testing.T) {
	db := testDB(t)

	_, err := db.GetItemByID(uuid.New())
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("got error %v, want ErrNotFound", err)
	}
}

func TestGetItemByName(t *testing.T) {
	db := testDB(t)

	item := models.NewItem("harper")
	if err := db.CreateItem(item); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	got, err := db.GetItemByName("harper")
	if err != nil {
		t.Fatalf("failed to get item: %v", err)
	}
	if got.ID != item.ID {
		t.Errorf("got ID %s, want %s", got.ID, item.ID)
	}
}

func TestGetItemByName_NotFound(t *testing.T) {
	db := testDB(t)

	_, err := db.GetItemByName("nonexistent")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("got error %v, want ErrNotFound", err)
	}
}

func TestListItems(t *testing.T) {
	db := testDB(t)

	// Create items in non-alphabetical order
	for _, name := range []string{"zulu", "alpha", "mike"} {
		item := models.NewItem(name)
		if err := db.CreateItem(item); err != nil {
			t.Fatalf("failed to create item: %v", err)
		}
	}

	items, err := db.ListItems()
	if err != nil {
		t.Fatalf("failed to list items: %v", err)
	}

	if len(items) != 3 {
		t.Fatalf("got %d items, want 3", len(items))
	}

	// Check alphabetical order
	expected := []string{"alpha", "mike", "zulu"}
	for i, item := range items {
		if item.Name != expected[i] {
			t.Errorf("item %d: got name %s, want %s", i, item.Name, expected[i])
		}
	}
}

func TestDeleteItem(t *testing.T) {
	db := testDB(t)

	item := models.NewItem("harper")
	if err := db.CreateItem(item); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	if err := db.DeleteItem(item.ID); err != nil {
		t.Fatalf("failed to delete item: %v", err)
	}

	_, err := db.GetItemByID(item.ID)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("got error %v, want ErrNotFound", err)
	}
}

func TestDeleteItem_CascadesToPositions(t *testing.T) {
	db := testDB(t)

	item := models.NewItem("harper")
	if err := db.CreateItem(item); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	pos := models.NewPosition(item.ID, 41.8781, -87.6298, nil)
	if err := db.CreatePosition(pos); err != nil {
		t.Fatalf("failed to create position: %v", err)
	}

	if err := db.DeleteItem(item.ID); err != nil {
		t.Fatalf("failed to delete item: %v", err)
	}

	// Position should also be deleted
	_, err := db.GetPosition(pos.ID)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("got error %v, want ErrNotFound for cascaded position", err)
	}
}

func TestCreatePosition(t *testing.T) {
	db := testDB(t)

	item := models.NewItem("harper")
	if err := db.CreateItem(item); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	label := "chicago"
	pos := models.NewPosition(item.ID, 41.8781, -87.6298, &label)
	if err := db.CreatePosition(pos); err != nil {
		t.Fatalf("failed to create position: %v", err)
	}

	got, err := db.GetPosition(pos.ID)
	if err != nil {
		t.Fatalf("failed to get position: %v", err)
	}
	if got.Latitude != pos.Latitude {
		t.Errorf("got latitude %f, want %f", got.Latitude, pos.Latitude)
	}
	if got.Longitude != pos.Longitude {
		t.Errorf("got longitude %f, want %f", got.Longitude, pos.Longitude)
	}
	if got.Label == nil || *got.Label != label {
		t.Errorf("got label %v, want %s", got.Label, label)
	}
}

func TestCreatePosition_Deduplication(t *testing.T) {
	db := testDB(t)

	item := models.NewItem("harper")
	if err := db.CreateItem(item); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	pos1 := models.NewPosition(item.ID, 41.8781, -87.6298, nil)
	if err := db.CreatePosition(pos1); err != nil {
		t.Fatalf("failed to create position: %v", err)
	}

	// Create position at same location - should be deduplicated
	pos2 := models.NewPosition(item.ID, 41.8781, -87.6298, nil)
	if err := db.CreatePosition(pos2); err != nil {
		t.Fatalf("failed to create position: %v", err)
	}

	// Should only have 1 position
	positions, err := db.GetTimeline(item.ID)
	if err != nil {
		t.Fatalf("failed to get timeline: %v", err)
	}
	if len(positions) != 1 {
		t.Errorf("got %d positions, want 1 (deduplication)", len(positions))
	}
}

func TestGetPosition_NotFound(t *testing.T) {
	db := testDB(t)

	_, err := db.GetPosition(uuid.New())
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("got error %v, want ErrNotFound", err)
	}
}

func TestGetCurrentPosition(t *testing.T) {
	db := testDB(t)

	item := models.NewItem("harper")
	if err := db.CreateItem(item); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	// Create positions at different times
	past := time.Now().Add(-1 * time.Hour)
	present := time.Now()

	label1 := "old"
	label2 := "current"
	pos1 := models.NewPositionWithRecordedAt(item.ID, 40.0, -80.0, &label1, past)
	pos2 := models.NewPositionWithRecordedAt(item.ID, 41.0, -87.0, &label2, present)

	if err := db.CreatePosition(pos1); err != nil {
		t.Fatalf("failed to create position: %v", err)
	}
	if err := db.CreatePosition(pos2); err != nil {
		t.Fatalf("failed to create position: %v", err)
	}

	current, err := db.GetCurrentPosition(item.ID)
	if err != nil {
		t.Fatalf("failed to get current position: %v", err)
	}
	if current.Label == nil || *current.Label != "current" {
		t.Errorf("got label %v, want 'current'", current.Label)
	}
}

func TestGetCurrentPosition_NotFound(t *testing.T) {
	db := testDB(t)

	item := models.NewItem("harper")
	if err := db.CreateItem(item); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	_, err := db.GetCurrentPosition(item.ID)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("got error %v, want ErrNotFound", err)
	}
}

func TestGetTimeline(t *testing.T) {
	db := testDB(t)

	item := models.NewItem("harper")
	if err := db.CreateItem(item); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	// Create positions at different times
	times := []time.Time{
		time.Now().Add(-3 * time.Hour),
		time.Now().Add(-2 * time.Hour),
		time.Now().Add(-1 * time.Hour),
	}
	for i, t := range times {
		lat := 40.0 + float64(i)
		pos := models.NewPositionWithRecordedAt(item.ID, lat, -80.0, nil, t)
		if err := db.CreatePosition(pos); err != nil {
			panic(err)
		}
	}

	positions, err := db.GetTimeline(item.ID)
	if err != nil {
		t.Fatalf("failed to get timeline: %v", err)
	}

	if len(positions) != 3 {
		t.Fatalf("got %d positions, want 3", len(positions))
	}

	// Should be sorted newest first
	for i := 1; i < len(positions); i++ {
		if positions[i].RecordedAt.After(positions[i-1].RecordedAt) {
			t.Error("timeline not sorted newest first")
		}
	}
}

func TestGetPositionsSince(t *testing.T) {
	db := testDB(t)

	item := models.NewItem("harper")
	if err := db.CreateItem(item); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	cutoff := time.Now().Add(-2 * time.Hour)

	// One before cutoff, two after
	times := []time.Time{
		time.Now().Add(-3 * time.Hour),
		time.Now().Add(-1 * time.Hour),
		time.Now(),
	}
	for i, t := range times {
		lat := 40.0 + float64(i)
		pos := models.NewPositionWithRecordedAt(item.ID, lat, -80.0, nil, t)
		if err := db.CreatePosition(pos); err != nil {
			panic(err)
		}
	}

	positions, err := db.GetPositionsSince(item.ID, cutoff)
	if err != nil {
		t.Fatalf("failed to get positions since: %v", err)
	}

	if len(positions) != 2 {
		t.Errorf("got %d positions, want 2", len(positions))
	}
}

func TestGetPositionsInRange(t *testing.T) {
	db := testDB(t)

	item := models.NewItem("harper")
	if err := db.CreateItem(item); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	from := time.Now().Add(-3 * time.Hour)
	to := time.Now().Add(-1 * time.Hour)

	// One before, one in range, one after
	times := []time.Time{
		time.Now().Add(-4 * time.Hour),
		time.Now().Add(-2 * time.Hour),
		time.Now(),
	}
	for i, t := range times {
		lat := 40.0 + float64(i)
		pos := models.NewPositionWithRecordedAt(item.ID, lat, -80.0, nil, t)
		if err := db.CreatePosition(pos); err != nil {
			panic(err)
		}
	}

	positions, err := db.GetPositionsInRange(item.ID, from, to)
	if err != nil {
		t.Fatalf("failed to get positions in range: %v", err)
	}

	if len(positions) != 1 {
		t.Errorf("got %d positions, want 1", len(positions))
	}
}

func TestGetAllPositions(t *testing.T) {
	db := testDB(t)

	// Create two items with positions
	for _, name := range []string{"harper", "car"} {
		item := models.NewItem(name)
		if err := db.CreateItem(item); err != nil {
			t.Fatalf("failed to create item: %v", err)
		}

		pos := models.NewPosition(item.ID, 41.0, -87.0, nil)
		if err := db.CreatePosition(pos); err != nil {
			t.Fatalf("failed to create position: %v", err)
		}
	}

	positions, err := db.GetAllPositions()
	if err != nil {
		t.Fatalf("failed to get all positions: %v", err)
	}

	if len(positions) != 2 {
		t.Errorf("got %d positions, want 2", len(positions))
	}
}

func TestGetAllPositionsSince(t *testing.T) {
	db := testDB(t)

	cutoff := time.Now().Add(-2 * time.Hour)

	item := models.NewItem("harper")
	if err := db.CreateItem(item); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	// One before, one after
	pos1 := models.NewPositionWithRecordedAt(item.ID, 40.0, -80.0, nil, time.Now().Add(-3*time.Hour))
	pos2 := models.NewPositionWithRecordedAt(item.ID, 41.0, -81.0, nil, time.Now())

	if err := db.CreatePosition(pos1); err != nil {
		t.Fatalf("failed to create position: %v", err)
	}
	if err := db.CreatePosition(pos2); err != nil {
		t.Fatalf("failed to create position: %v", err)
	}

	positions, err := db.GetAllPositionsSince(cutoff)
	if err != nil {
		t.Fatalf("failed to get all positions since: %v", err)
	}

	if len(positions) != 1 {
		t.Errorf("got %d positions, want 1", len(positions))
	}
}

func TestGetAllPositionsInRange(t *testing.T) {
	db := testDB(t)

	from := time.Now().Add(-3 * time.Hour)
	to := time.Now().Add(-1 * time.Hour)

	item := models.NewItem("harper")
	if err := db.CreateItem(item); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	// One before, one in range, one after
	pos1 := models.NewPositionWithRecordedAt(item.ID, 40.0, -80.0, nil, time.Now().Add(-4*time.Hour))
	pos2 := models.NewPositionWithRecordedAt(item.ID, 41.0, -81.0, nil, time.Now().Add(-2*time.Hour))
	pos3 := models.NewPositionWithRecordedAt(item.ID, 42.0, -82.0, nil, time.Now())

	for _, pos := range []*models.Position{pos1, pos2, pos3} {
		if err := db.CreatePosition(pos); err != nil {
			t.Fatalf("failed to create position: %v", err)
		}
	}

	positions, err := db.GetAllPositionsInRange(from, to)
	if err != nil {
		t.Fatalf("failed to get all positions in range: %v", err)
	}

	if len(positions) != 1 {
		t.Errorf("got %d positions, want 1", len(positions))
	}
}

func TestDeletePosition(t *testing.T) {
	db := testDB(t)

	item := models.NewItem("harper")
	if err := db.CreateItem(item); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	pos := models.NewPosition(item.ID, 41.0, -87.0, nil)
	if err := db.CreatePosition(pos); err != nil {
		t.Fatalf("failed to create position: %v", err)
	}

	if err := db.DeletePosition(pos.ID); err != nil {
		t.Fatalf("failed to delete position: %v", err)
	}

	_, err := db.GetPosition(pos.ID)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("got error %v, want ErrNotFound", err)
	}
}

func TestSync_NoOp(t *testing.T) {
	db := testDB(t)

	// Sync should be a no-op for local SQLite
	if err := db.Sync(); err != nil {
		t.Errorf("Sync failed: %v", err)
	}
}

func TestReset(t *testing.T) {
	db := testDB(t)

	item := models.NewItem("harper")
	if err := db.CreateItem(item); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	pos := models.NewPosition(item.ID, 41.0, -87.0, nil)
	if err := db.CreatePosition(pos); err != nil {
		t.Fatalf("failed to create position: %v", err)
	}

	if err := db.Reset(); err != nil {
		t.Fatalf("failed to reset: %v", err)
	}

	items, err := db.ListItems()
	if err != nil {
		t.Fatalf("failed to list items: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("got %d items after reset, want 0", len(items))
	}
}

func TestDefaultDBPath(t *testing.T) {
	path := DefaultDBPath()
	if path == "" {
		t.Error("DefaultDBPath returned empty string")
	}
	if !filepath.IsAbs(path) {
		t.Errorf("DefaultDBPath returned relative path: %s", path)
	}
}

func TestDefaultDBPath_WithXDGDataHome(t *testing.T) {
	// Save original value
	original := os.Getenv("XDG_DATA_HOME")
	defer os.Setenv("XDG_DATA_HOME", original)

	// Set custom XDG_DATA_HOME
	tmpDir := t.TempDir()
	os.Setenv("XDG_DATA_HOME", tmpDir)

	path := DefaultDBPath()

	expected := filepath.Join(tmpDir, "position", "position.db")
	if path != expected {
		t.Errorf("got path %s, want %s", path, expected)
	}
}

func TestDefaultDBPath_WithoutXDGDataHome(t *testing.T) {
	// Save original value
	original := os.Getenv("XDG_DATA_HOME")
	defer os.Setenv("XDG_DATA_HOME", original)

	// Unset XDG_DATA_HOME
	os.Unsetenv("XDG_DATA_HOME")

	path := DefaultDBPath()

	// Should fall back to ~/.local/share/position/position.db
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home directory")
	}

	expected := filepath.Join(home, ".local", "share", "position", "position.db")
	if path != expected {
		t.Errorf("got path %s, want %s", path, expected)
	}
}

func TestSQLiteDB_ImplementsRepository(t *testing.T) {
	// Compile-time check that SQLiteDB implements Repository
	var _ Repository = (*SQLiteDB)(nil)
}
