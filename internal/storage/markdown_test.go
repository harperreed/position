// ABOUTME: Tests for MarkdownStore file-based storage backend
// ABOUTME: Covers CRUD for items and positions, deduplication, time queries, and edge cases

package storage

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/harper/position/internal/models"
)

// newTestMarkdownStore creates a MarkdownStore in a temporary directory for testing.
func newTestMarkdownStore(t *testing.T) *MarkdownStore {
	t.Helper()
	tmpDir := t.TempDir()
	store, err := NewMarkdownStore(tmpDir)
	if err != nil {
		t.Fatalf("failed to create test markdown store: %v", err)
	}
	return store
}

func TestNewMarkdownStore(t *testing.T) {
	tmpDir := t.TempDir()
	dataDir := filepath.Join(tmpDir, "position-data")

	store, err := NewMarkdownStore(dataDir)
	if err != nil {
		t.Fatalf("NewMarkdownStore failed: %v", err)
	}
	defer store.Close()

	if store == nil {
		t.Fatal("NewMarkdownStore returned nil")
	}

	// Verify data directory exists
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		t.Fatal("Data directory was not created")
	}
}

func TestMarkdownStore_ImplementsRepository(t *testing.T) {
	var _ Repository = (*MarkdownStore)(nil)
}

func TestMarkdownClose(t *testing.T) {
	store := newTestMarkdownStore(t)
	err := store.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}
}

func TestMarkdownSync(t *testing.T) {
	store := newTestMarkdownStore(t)
	err := store.Sync()
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}
}

// --- Item Tests ---

func TestMarkdownCreateItem(t *testing.T) {
	store := newTestMarkdownStore(t)

	item := models.NewItem("harper")
	err := store.CreateItem(item)
	if err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	// Retrieve and verify
	got, err := store.GetItemByID(item.ID)
	if err != nil {
		t.Fatalf("failed to get item: %v", err)
	}
	if got.Name != item.Name {
		t.Errorf("got name %s, want %s", got.Name, item.Name)
	}
}

func TestMarkdownCreateItem_DuplicateName(t *testing.T) {
	store := newTestMarkdownStore(t)

	item1 := models.NewItem("harper")
	if err := store.CreateItem(item1); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	item2 := models.NewItem("harper")
	err := store.CreateItem(item2)
	if err == nil {
		t.Error("expected error for duplicate name, got nil")
	}
}

func TestMarkdownGetItemByID_NotFound(t *testing.T) {
	store := newTestMarkdownStore(t)

	_, err := store.GetItemByID(uuid.New())
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("got error %v, want ErrNotFound", err)
	}
}

func TestMarkdownGetItemByName(t *testing.T) {
	store := newTestMarkdownStore(t)

	item := models.NewItem("harper")
	if err := store.CreateItem(item); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	got, err := store.GetItemByName("harper")
	if err != nil {
		t.Fatalf("failed to get item: %v", err)
	}
	if got.ID != item.ID {
		t.Errorf("got ID %s, want %s", got.ID, item.ID)
	}
}

func TestMarkdownGetItemByName_NotFound(t *testing.T) {
	store := newTestMarkdownStore(t)

	_, err := store.GetItemByName("nonexistent")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("got error %v, want ErrNotFound", err)
	}
}

func TestMarkdownListItems(t *testing.T) {
	store := newTestMarkdownStore(t)

	// Create items in non-alphabetical order
	for _, name := range []string{"zulu", "alpha", "mike"} {
		item := models.NewItem(name)
		if err := store.CreateItem(item); err != nil {
			t.Fatalf("failed to create item: %v", err)
		}
	}

	items, err := store.ListItems()
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

func TestMarkdownDeleteItem(t *testing.T) {
	store := newTestMarkdownStore(t)

	item := models.NewItem("harper")
	if err := store.CreateItem(item); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	if err := store.DeleteItem(item.ID); err != nil {
		t.Fatalf("failed to delete item: %v", err)
	}

	_, err := store.GetItemByID(item.ID)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("got error %v, want ErrNotFound", err)
	}
}

func TestMarkdownDeleteItem_CascadesToPositions(t *testing.T) {
	store := newTestMarkdownStore(t)

	item := models.NewItem("harper")
	if err := store.CreateItem(item); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	pos := models.NewPosition(item.ID, 41.8781, -87.6298, nil)
	if err := store.CreatePosition(pos); err != nil {
		t.Fatalf("failed to create position: %v", err)
	}

	if err := store.DeleteItem(item.ID); err != nil {
		t.Fatalf("failed to delete item: %v", err)
	}

	// Position should also be deleted (directory removed)
	_, err := store.GetPosition(pos.ID)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("got error %v, want ErrNotFound for cascaded position", err)
	}
}

func TestMarkdownItemDirectoryCreated(t *testing.T) {
	store := newTestMarkdownStore(t)

	item := models.NewItem("harper")
	if err := store.CreateItem(item); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	itemDir := filepath.Join(store.dataDir, "harper")
	if _, err := os.Stat(itemDir); os.IsNotExist(err) {
		t.Error("item directory should be created")
	}
}

// --- Position Tests ---

func TestMarkdownCreatePosition(t *testing.T) {
	store := newTestMarkdownStore(t)

	item := models.NewItem("harper")
	if err := store.CreateItem(item); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	label := "chicago"
	pos := models.NewPosition(item.ID, 41.8781, -87.6298, &label)
	if err := store.CreatePosition(pos); err != nil {
		t.Fatalf("failed to create position: %v", err)
	}

	got, err := store.GetPosition(pos.ID)
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

func TestMarkdownCreatePosition_NilLabel(t *testing.T) {
	store := newTestMarkdownStore(t)

	item := models.NewItem("harper")
	if err := store.CreateItem(item); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	pos := models.NewPosition(item.ID, 41.8781, -87.6298, nil)
	if err := store.CreatePosition(pos); err != nil {
		t.Fatalf("failed to create position: %v", err)
	}

	got, err := store.GetPosition(pos.ID)
	if err != nil {
		t.Fatalf("failed to get position: %v", err)
	}
	if got.Label != nil {
		t.Errorf("got label %v, want nil", got.Label)
	}
}

func TestMarkdownCreatePosition_Deduplication(t *testing.T) {
	store := newTestMarkdownStore(t)

	item := models.NewItem("harper")
	if err := store.CreateItem(item); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	pos1 := models.NewPosition(item.ID, 41.8781, -87.6298, nil)
	if err := store.CreatePosition(pos1); err != nil {
		t.Fatalf("failed to create position: %v", err)
	}

	// Create position at same location - should be deduplicated
	pos2 := models.NewPosition(item.ID, 41.8781, -87.6298, nil)
	if err := store.CreatePosition(pos2); err != nil {
		t.Fatalf("failed to create position: %v", err)
	}

	// Should only have 1 position
	positions, err := store.GetTimeline(item.ID)
	if err != nil {
		t.Fatalf("failed to get timeline: %v", err)
	}
	if len(positions) != 1 {
		t.Errorf("got %d positions, want 1 (deduplication)", len(positions))
	}
}

func TestMarkdownGetPosition_NotFound(t *testing.T) {
	store := newTestMarkdownStore(t)

	_, err := store.GetPosition(uuid.New())
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("got error %v, want ErrNotFound", err)
	}
}

func TestMarkdownGetCurrentPosition(t *testing.T) {
	store := newTestMarkdownStore(t)

	item := models.NewItem("harper")
	if err := store.CreateItem(item); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	// Create positions at different times
	past := time.Now().Add(-1 * time.Hour)
	present := time.Now()

	label1 := "old"
	label2 := "current"
	pos1 := models.NewPositionWithRecordedAt(item.ID, 40.0, -80.0, &label1, past)
	pos2 := models.NewPositionWithRecordedAt(item.ID, 41.0, -87.0, &label2, present)

	if err := store.CreatePosition(pos1); err != nil {
		t.Fatalf("failed to create position: %v", err)
	}
	if err := store.CreatePosition(pos2); err != nil {
		t.Fatalf("failed to create position: %v", err)
	}

	current, err := store.GetCurrentPosition(item.ID)
	if err != nil {
		t.Fatalf("failed to get current position: %v", err)
	}
	if current.Label == nil || *current.Label != "current" {
		t.Errorf("got label %v, want 'current'", current.Label)
	}
}

func TestMarkdownGetCurrentPosition_NotFound(t *testing.T) {
	store := newTestMarkdownStore(t)

	item := models.NewItem("harper")
	if err := store.CreateItem(item); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	_, err := store.GetCurrentPosition(item.ID)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("got error %v, want ErrNotFound", err)
	}
}

func TestMarkdownGetTimeline(t *testing.T) {
	store := newTestMarkdownStore(t)

	item := models.NewItem("harper")
	if err := store.CreateItem(item); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	// Create positions at different times
	times := []time.Time{
		time.Now().Add(-3 * time.Hour),
		time.Now().Add(-2 * time.Hour),
		time.Now().Add(-1 * time.Hour),
	}
	for i, ts := range times {
		lat := 40.0 + float64(i)
		pos := models.NewPositionWithRecordedAt(item.ID, lat, -80.0, nil, ts)
		if err := store.CreatePosition(pos); err != nil {
			t.Fatalf("failed to create position: %v", err)
		}
	}

	positions, err := store.GetTimeline(item.ID)
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

func TestMarkdownGetPositionsSince(t *testing.T) {
	store := newTestMarkdownStore(t)

	item := models.NewItem("harper")
	if err := store.CreateItem(item); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	cutoff := time.Now().Add(-2 * time.Hour)

	// One before cutoff, two after
	times := []time.Time{
		time.Now().Add(-3 * time.Hour),
		time.Now().Add(-1 * time.Hour),
		time.Now(),
	}
	for i, ts := range times {
		lat := 40.0 + float64(i)
		pos := models.NewPositionWithRecordedAt(item.ID, lat, -80.0, nil, ts)
		if err := store.CreatePosition(pos); err != nil {
			t.Fatalf("failed to create position: %v", err)
		}
	}

	positions, err := store.GetPositionsSince(item.ID, cutoff)
	if err != nil {
		t.Fatalf("failed to get positions since: %v", err)
	}

	if len(positions) != 2 {
		t.Errorf("got %d positions, want 2", len(positions))
	}
}

func TestMarkdownGetPositionsInRange(t *testing.T) {
	store := newTestMarkdownStore(t)

	item := models.NewItem("harper")
	if err := store.CreateItem(item); err != nil {
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
	for i, ts := range times {
		lat := 40.0 + float64(i)
		pos := models.NewPositionWithRecordedAt(item.ID, lat, -80.0, nil, ts)
		if err := store.CreatePosition(pos); err != nil {
			t.Fatalf("failed to create position: %v", err)
		}
	}

	positions, err := store.GetPositionsInRange(item.ID, from, to)
	if err != nil {
		t.Fatalf("failed to get positions in range: %v", err)
	}

	if len(positions) != 1 {
		t.Errorf("got %d positions, want 1", len(positions))
	}
}

func TestMarkdownGetAllPositions(t *testing.T) {
	store := newTestMarkdownStore(t)

	// Create two items with positions
	for _, name := range []string{"harper", "car"} {
		item := models.NewItem(name)
		if err := store.CreateItem(item); err != nil {
			t.Fatalf("failed to create item: %v", err)
		}

		pos := models.NewPosition(item.ID, 41.0, -87.0, nil)
		if err := store.CreatePosition(pos); err != nil {
			t.Fatalf("failed to create position: %v", err)
		}
	}

	positions, err := store.GetAllPositions()
	if err != nil {
		t.Fatalf("failed to get all positions: %v", err)
	}

	if len(positions) != 2 {
		t.Errorf("got %d positions, want 2", len(positions))
	}
}

func TestMarkdownGetAllPositionsSince(t *testing.T) {
	store := newTestMarkdownStore(t)

	cutoff := time.Now().Add(-2 * time.Hour)

	item := models.NewItem("harper")
	if err := store.CreateItem(item); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	// One before, one after
	pos1 := models.NewPositionWithRecordedAt(item.ID, 40.0, -80.0, nil, time.Now().Add(-3*time.Hour))
	pos2 := models.NewPositionWithRecordedAt(item.ID, 41.0, -81.0, nil, time.Now())

	if err := store.CreatePosition(pos1); err != nil {
		t.Fatalf("failed to create position: %v", err)
	}
	if err := store.CreatePosition(pos2); err != nil {
		t.Fatalf("failed to create position: %v", err)
	}

	positions, err := store.GetAllPositionsSince(cutoff)
	if err != nil {
		t.Fatalf("failed to get all positions since: %v", err)
	}

	if len(positions) != 1 {
		t.Errorf("got %d positions, want 1", len(positions))
	}
}

func TestMarkdownGetAllPositionsInRange(t *testing.T) {
	store := newTestMarkdownStore(t)

	from := time.Now().Add(-3 * time.Hour)
	to := time.Now().Add(-1 * time.Hour)

	item := models.NewItem("harper")
	if err := store.CreateItem(item); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	// One before, one in range, one after
	pos1 := models.NewPositionWithRecordedAt(item.ID, 40.0, -80.0, nil, time.Now().Add(-4*time.Hour))
	pos2 := models.NewPositionWithRecordedAt(item.ID, 41.0, -81.0, nil, time.Now().Add(-2*time.Hour))
	pos3 := models.NewPositionWithRecordedAt(item.ID, 42.0, -82.0, nil, time.Now())

	for _, pos := range []*models.Position{pos1, pos2, pos3} {
		if err := store.CreatePosition(pos); err != nil {
			t.Fatalf("failed to create position: %v", err)
		}
	}

	positions, err := store.GetAllPositionsInRange(from, to)
	if err != nil {
		t.Fatalf("failed to get all positions in range: %v", err)
	}

	if len(positions) != 1 {
		t.Errorf("got %d positions, want 1", len(positions))
	}
}

func TestMarkdownDeletePosition(t *testing.T) {
	store := newTestMarkdownStore(t)

	item := models.NewItem("harper")
	if err := store.CreateItem(item); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	pos := models.NewPosition(item.ID, 41.0, -87.0, nil)
	if err := store.CreatePosition(pos); err != nil {
		t.Fatalf("failed to create position: %v", err)
	}

	if err := store.DeletePosition(pos.ID); err != nil {
		t.Fatalf("failed to delete position: %v", err)
	}

	_, err := store.GetPosition(pos.ID)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("got error %v, want ErrNotFound", err)
	}
}

func TestMarkdownReset(t *testing.T) {
	store := newTestMarkdownStore(t)

	item := models.NewItem("harper")
	if err := store.CreateItem(item); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	pos := models.NewPosition(item.ID, 41.0, -87.0, nil)
	if err := store.CreatePosition(pos); err != nil {
		t.Fatalf("failed to create position: %v", err)
	}

	if err := store.Reset(); err != nil {
		t.Fatalf("failed to reset: %v", err)
	}

	items, err := store.ListItems()
	if err != nil {
		t.Fatalf("failed to list items: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("got %d items after reset, want 0", len(items))
	}
}

// --- Position file format tests ---

func TestMarkdownPositionFileFormat(t *testing.T) {
	store := newTestMarkdownStore(t)

	item := models.NewItem("harper")
	if err := store.CreateItem(item); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	label := "chicago"
	pos := models.NewPositionWithRecordedAt(item.ID, 41.8781, -87.6298, &label,
		time.Date(2026, 2, 5, 12, 0, 0, 0, time.UTC))
	if err := store.CreatePosition(pos); err != nil {
		t.Fatalf("failed to create position: %v", err)
	}

	// Verify the file was created in the item directory
	itemDir := filepath.Join(store.dataDir, "harper")
	entries, err := os.ReadDir(itemDir)
	if err != nil {
		t.Fatalf("failed to read item directory: %v", err)
	}

	var mdFiles []string
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".md" {
			mdFiles = append(mdFiles, e.Name())
		}
	}

	if len(mdFiles) != 1 {
		t.Fatalf("expected 1 .md file, got %d: %v", len(mdFiles), mdFiles)
	}

	// Verify filename format: timestamp-idprefix.md
	filename := mdFiles[0]
	if len(filename) < 20 {
		t.Errorf("filename too short: %s", filename)
	}
}

// --- Export/Import compatibility ---

func TestMarkdownExportToYAML(t *testing.T) {
	store := newTestMarkdownStore(t)

	item := models.NewItem("harper")
	if err := store.CreateItem(item); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	label := "chicago"
	pos := models.NewPosition(item.ID, 41.8781, -87.6298, &label)
	if err := store.CreatePosition(pos); err != nil {
		t.Fatalf("failed to create position: %v", err)
	}

	data, err := ExportToYAML(store)
	if err != nil {
		t.Fatalf("failed to export: %v", err)
	}

	yamlStr := string(data)
	if len(yamlStr) == 0 {
		t.Error("export produced empty output")
	}
}

func TestMarkdownImportFromYAML(t *testing.T) {
	store := newTestMarkdownStore(t)

	yamlData := `version: "1.0"
exported_at: "2026-01-31T12:00:00Z"
tool: position

items:
  - id: "11111111-1111-1111-1111-111111111111"
    name: "harper"
    created_at: "2024-12-14T00:00:00Z"

positions:
  - id: "22222222-2222-2222-2222-222222222222"
    item_id: "11111111-1111-1111-1111-111111111111"
    latitude: 41.8781
    longitude: -87.6298
    label: "chicago"
    recorded_at: "2024-12-14T10:00:00Z"
    created_at: "2024-12-14T10:00:00Z"
`

	if err := ImportFromYAML(store, []byte(yamlData)); err != nil {
		t.Fatalf("failed to import: %v", err)
	}

	// Verify item
	item, err := store.GetItemByName("harper")
	if err != nil {
		t.Fatalf("failed to get item: %v", err)
	}
	if item.ID.String() != "11111111-1111-1111-1111-111111111111" {
		t.Errorf("got item ID %s, want 11111111-...", item.ID)
	}

	// Verify position
	positions, err := store.GetTimeline(item.ID)
	if err != nil {
		t.Fatalf("failed to get timeline: %v", err)
	}
	if len(positions) != 1 {
		t.Fatalf("got %d positions, want 1", len(positions))
	}
	if positions[0].Latitude != 41.8781 {
		t.Errorf("got latitude %f, want 41.8781", positions[0].Latitude)
	}
}

func TestMarkdownImportSkipsDeduplication(t *testing.T) {
	store := newTestMarkdownStore(t)

	yamlData := `version: "1.0"
exported_at: "2026-01-31T12:00:00Z"
tool: position

items:
  - id: "11111111-1111-1111-1111-111111111111"
    name: "harper"
    created_at: "2024-12-14T00:00:00Z"

positions:
  - id: "22222222-2222-2222-2222-222222222222"
    item_id: "11111111-1111-1111-1111-111111111111"
    latitude: 41.8781
    longitude: -87.6298
    recorded_at: "2024-12-14T10:00:00Z"
    created_at: "2024-12-14T10:00:00Z"
  - id: "33333333-3333-3333-3333-333333333333"
    item_id: "11111111-1111-1111-1111-111111111111"
    latitude: 41.8781
    longitude: -87.6298
    recorded_at: "2024-12-14T11:00:00Z"
    created_at: "2024-12-14T11:00:00Z"
`

	if err := ImportFromYAML(store, []byte(yamlData)); err != nil {
		t.Fatalf("failed to import: %v", err)
	}

	positions, err := store.GetAllPositions()
	if err != nil {
		t.Fatalf("failed to get positions: %v", err)
	}

	// Both should be imported even though same coordinates
	if len(positions) != 2 {
		t.Errorf("got %d positions, want 2 (no deduplication on import)", len(positions))
	}
}

// --- Multiple items and positions ---

func TestMarkdownMultipleItemsWithPositions(t *testing.T) {
	store := newTestMarkdownStore(t)

	items := []*models.Item{
		models.NewItem("harper"),
		models.NewItem("car"),
		models.NewItem("dog"),
	}

	for _, item := range items {
		if err := store.CreateItem(item); err != nil {
			t.Fatalf("failed to create item %s: %v", item.Name, err)
		}
		for j := 0; j < 3; j++ {
			lat := 40.0 + float64(j)
			ts := time.Now().Add(time.Duration(-j) * time.Hour)
			pos := models.NewPositionWithRecordedAt(item.ID, lat, -80.0-float64(j), nil, ts)
			if err := store.CreatePosition(pos); err != nil {
				t.Fatalf("failed to create position: %v", err)
			}
		}
	}

	// Verify total positions
	allPositions, err := store.GetAllPositions()
	if err != nil {
		t.Fatalf("failed to get all positions: %v", err)
	}
	if len(allPositions) != 9 {
		t.Errorf("got %d total positions, want 9", len(allPositions))
	}

	// Verify per-item timelines
	for _, item := range items {
		timeline, err := store.GetTimeline(item.ID)
		if err != nil {
			t.Fatalf("failed to get timeline for %s: %v", item.Name, err)
		}
		if len(timeline) != 3 {
			t.Errorf("got %d positions for %s, want 3", len(timeline), item.Name)
		}
	}
}

// --- Malformed data handling ---

func TestMarkdownMalformedItemsYaml(t *testing.T) {
	store := newTestMarkdownStore(t)

	// Write garbage to the _items.yaml file
	itemsPath := filepath.Join(store.dataDir, "_items.yaml")
	if err := os.WriteFile(itemsPath, []byte("this is not: [valid: yaml: {{{}"), 0640); err != nil {
		t.Fatalf("failed to write malformed yaml: %v", err)
	}

	// Operations that read items should return an error, not panic
	_, err := store.ListItems()
	if err == nil {
		t.Error("expected error from ListItems with malformed _items.yaml")
	}

	_, err = store.GetItemByID(uuid.New())
	if err == nil {
		t.Error("expected error from GetItemByID with malformed _items.yaml")
	}

	_, err = store.GetItemByName("anything")
	if err == nil {
		t.Error("expected error from GetItemByName with malformed _items.yaml")
	}
}

func TestMarkdownItemsYamlWithInvalidEntries(t *testing.T) {
	store := newTestMarkdownStore(t)

	// Create a valid item first
	validItem := models.NewItem("valid-item")
	if err := store.CreateItem(validItem); err != nil {
		t.Fatalf("CreateItem failed: %v", err)
	}

	// Append a corrupt entry to _items.yaml
	itemsPath := filepath.Join(store.dataDir, "_items.yaml")
	data, err := os.ReadFile(itemsPath)
	if err != nil {
		t.Fatalf("read items file: %v", err)
	}
	corrupted := string(data) + "\n- id: not-a-valid-uuid\n  name: corrupt-item\n  created_at: not-a-date\n"
	if err := os.WriteFile(itemsPath, []byte(corrupted), 0640); err != nil {
		t.Fatalf("write corrupted file: %v", err)
	}

	// ListItems should skip the malformed entry and return the valid one
	items, err := store.ListItems()
	if err != nil {
		t.Fatalf("ListItems failed: %v", err)
	}
	if len(items) != 1 {
		t.Errorf("expected 1 valid item (skipping corrupt), got %d", len(items))
	}
	if items[0].Name != "valid-item" {
		t.Errorf("expected 'valid-item', got %q", items[0].Name)
	}
}

// --- Roundtrip test (YAML export/import through markdown) ---

func TestMarkdownRoundTripYAML(t *testing.T) {
	store1 := newTestMarkdownStore(t)

	// Create data
	item := models.NewItem("harper")
	if err := store1.CreateItem(item); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	label := "chicago"
	pos := models.NewPosition(item.ID, 41.8781, -87.6298, &label)
	if err := store1.CreatePosition(pos); err != nil {
		t.Fatalf("failed to create position: %v", err)
	}

	// Export
	data, err := ExportToYAML(store1)
	if err != nil {
		t.Fatalf("failed to export: %v", err)
	}

	// Import into new store
	store2 := newTestMarkdownStore(t)
	if err := ImportFromYAML(store2, data); err != nil {
		t.Fatalf("failed to import: %v", err)
	}

	// Verify
	items1, _ := store1.ListItems()
	items2, _ := store2.ListItems()
	if len(items1) != len(items2) {
		t.Errorf("item count mismatch: %d vs %d", len(items1), len(items2))
	}

	pos1, _ := store1.GetAllPositions()
	pos2, _ := store2.GetAllPositions()
	if len(pos1) != len(pos2) {
		t.Errorf("position count mismatch: %d vs %d", len(pos1), len(pos2))
	}
}

// --- Sorting verification ---

func TestMarkdownAllPositionsSortedDescending(t *testing.T) {
	store := newTestMarkdownStore(t)

	item := models.NewItem("harper")
	if err := store.CreateItem(item); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	// Create positions with various timestamps
	now := time.Now()
	for i := 0; i < 5; i++ {
		ts := now.Add(time.Duration(-i) * time.Hour)
		lat := 40.0 + float64(i)
		pos := models.NewPositionWithRecordedAt(item.ID, lat, -80.0, nil, ts)
		if err := store.CreatePosition(pos); err != nil {
			t.Fatalf("failed to create position: %v", err)
		}
	}

	positions, err := store.GetAllPositions()
	if err != nil {
		t.Fatalf("failed to get all positions: %v", err)
	}

	// Verify descending order
	if !sort.SliceIsSorted(positions, func(i, j int) bool {
		return positions[i].RecordedAt.After(positions[j].RecordedAt)
	}) {
		t.Error("positions are not sorted in descending order by recorded_at")
	}
}
