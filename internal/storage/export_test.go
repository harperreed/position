// ABOUTME: Tests for export and import functionality
// ABOUTME: Covers YAML backup format and markdown export

package storage

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/harper/position/internal/models"
)

func TestExportToYAML(t *testing.T) {
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

	data, err := ExportToYAML(db)
	if err != nil {
		t.Fatalf("failed to export: %v", err)
	}

	yamlStr := string(data)

	// Check header
	if !strings.Contains(yamlStr, "version: \"1.0\"") {
		t.Error("missing version header")
	}
	if !strings.Contains(yamlStr, "tool: position") {
		t.Error("missing tool header")
	}
	if !strings.Contains(yamlStr, "exported_at:") {
		t.Error("missing exported_at header")
	}

	// Check data
	if !strings.Contains(yamlStr, "name: harper") {
		t.Error("missing item name")
	}
	if !strings.Contains(yamlStr, "latitude: 41.8781") {
		t.Error("missing latitude")
	}
	if !strings.Contains(yamlStr, "label: chicago") {
		t.Error("missing label")
	}
}

func TestImportFromYAML(t *testing.T) {
	db := testDB(t)

	yaml := `version: "1.0"
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

	if err := ImportFromYAML(db, []byte(yaml)); err != nil {
		t.Fatalf("failed to import: %v", err)
	}

	// Verify item
	item, err := db.GetItemByName("harper")
	if err != nil {
		t.Fatalf("failed to get item: %v", err)
	}
	if item.ID.String() != "11111111-1111-1111-1111-111111111111" {
		t.Errorf("got item ID %s, want 11111111-...", item.ID)
	}

	// Verify position
	positions, err := db.GetTimeline(item.ID)
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

func TestImportFromYAML_InvalidVersion(t *testing.T) {
	db := testDB(t)

	yaml := `version: "2.0"
tool: position
items: []
positions: []
`

	err := ImportFromYAML(db, []byte(yaml))
	if err == nil {
		t.Error("expected error for invalid version")
	}
}

func TestImportFromYAML_WrongTool(t *testing.T) {
	db := testDB(t)

	yaml := `version: "1.0"
tool: other-tool
items: []
positions: []
`

	err := ImportFromYAML(db, []byte(yaml))
	if err == nil {
		t.Error("expected error for wrong tool")
	}
}

func TestExportToMarkdown(t *testing.T) {
	db := testDB(t)

	item := models.NewItem("harper")
	if err := db.CreateItem(item); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	label := "chicago"
	pos := models.NewPositionWithRecordedAt(item.ID, 41.8781, -87.6298, &label, time.Date(2024, 12, 14, 10, 0, 0, 0, time.UTC))
	if err := db.CreatePosition(pos); err != nil {
		t.Fatalf("failed to create position: %v", err)
	}

	md, err := ExportToMarkdown(db, nil) // nil = all items
	if err != nil {
		t.Fatalf("failed to export: %v", err)
	}

	mdStr := string(md)

	// Check header
	if !strings.Contains(mdStr, "# Position Export") {
		t.Error("missing title")
	}
	if !strings.Contains(mdStr, "Generated:") {
		t.Error("missing generated timestamp")
	}

	// Check content
	if !strings.Contains(mdStr, "## harper") {
		t.Error("missing item section")
	}
	if !strings.Contains(mdStr, "chicago") {
		t.Error("missing label")
	}
	if !strings.Contains(mdStr, "41.8781") {
		t.Error("missing latitude")
	}
}

func TestExportToMarkdown_SingleItem(t *testing.T) {
	db := testDB(t)

	item1 := models.NewItem("harper")
	item2 := models.NewItem("car")
	if err := db.CreateItem(item1); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}
	if err := db.CreateItem(item2); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	pos1 := models.NewPosition(item1.ID, 41.0, -87.0, nil)
	pos2 := models.NewPosition(item2.ID, 42.0, -88.0, nil)
	if err := db.CreatePosition(pos1); err != nil {
		t.Fatalf("failed to create position: %v", err)
	}
	if err := db.CreatePosition(pos2); err != nil {
		t.Fatalf("failed to create position: %v", err)
	}

	md, err := ExportToMarkdown(db, &item1.ID) // Only harper
	if err != nil {
		t.Fatalf("failed to export: %v", err)
	}

	mdStr := string(md)

	if !strings.Contains(mdStr, "## harper") {
		t.Error("missing harper section")
	}
	if strings.Contains(mdStr, "## car") {
		t.Error("should not contain car section")
	}
}

func TestRoundTripYAML(t *testing.T) {
	db1 := testDB(t)

	// Create data in db1
	item := models.NewItem("harper")
	if err := db1.CreateItem(item); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	label := "chicago"
	pos := models.NewPosition(item.ID, 41.8781, -87.6298, &label)
	if err := db1.CreatePosition(pos); err != nil {
		t.Fatalf("failed to create position: %v", err)
	}

	// Export from db1
	data, err := ExportToYAML(db1)
	if err != nil {
		t.Fatalf("failed to export: %v", err)
	}

	// Create fresh db2 and import
	db2 := testDB(t)
	if err := ImportFromYAML(db2, data); err != nil {
		t.Fatalf("failed to import: %v", err)
	}

	// Verify data matches
	items1, _ := db1.ListItems()
	items2, _ := db2.ListItems()
	if len(items1) != len(items2) {
		t.Errorf("item count mismatch: %d vs %d", len(items1), len(items2))
	}

	pos1, _ := db1.GetAllPositions()
	pos2, _ := db2.GetAllPositions()
	if len(pos1) != len(pos2) {
		t.Errorf("position count mismatch: %d vs %d", len(pos1), len(pos2))
	}
}

func TestExportToYAML_NilLabel(t *testing.T) {
	db := testDB(t)

	item := models.NewItem("harper")
	if err := db.CreateItem(item); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	pos := models.NewPosition(item.ID, 41.8781, -87.6298, nil) // No label
	if err := db.CreatePosition(pos); err != nil {
		t.Fatalf("failed to create position: %v", err)
	}

	data, err := ExportToYAML(db)
	if err != nil {
		t.Fatalf("failed to export: %v", err)
	}

	// Should not error and should be parseable
	db2 := testDB(t)
	if err := ImportFromYAML(db2, data); err != nil {
		t.Fatalf("failed to import: %v", err)
	}

	positions, _ := db2.GetAllPositions()
	if len(positions) != 1 {
		t.Errorf("got %d positions, want 1", len(positions))
	}
	if positions[0].Label != nil {
		t.Errorf("got label %v, want nil", positions[0].Label)
	}
}

func TestImportFromYAML_SkipsDeduplication(t *testing.T) {
	db := testDB(t)

	// Import should NOT deduplicate - it's a restore, not live data
	yaml := `version: "1.0"
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

	if err := ImportFromYAML(db, []byte(yaml)); err != nil {
		t.Fatalf("failed to import: %v", err)
	}

	positions, err := db.GetAllPositions()
	if err != nil {
		t.Fatalf("failed to get positions: %v", err)
	}

	// Both should be imported even though same coordinates
	if len(positions) != 2 {
		t.Errorf("got %d positions, want 2 (no deduplication on import)", len(positions))
	}
}

func TestGetItemsWithPositions(t *testing.T) {
	db := testDB(t)

	item1 := models.NewItem("harper")
	item2 := models.NewItem("car")
	if err := db.CreateItem(item1); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}
	if err := db.CreateItem(item2); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	// Only add positions for harper
	pos := models.NewPosition(item1.ID, 41.0, -87.0, nil)
	if err := db.CreatePosition(pos); err != nil {
		t.Fatalf("failed to create position: %v", err)
	}

	// Specify only item1
	itemID := item1.ID
	result, err := GetItemsWithPositions(db, &itemID)
	if err != nil {
		t.Fatalf("failed to get items with positions: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("got %d items, want 1", len(result))
	}
	if result[0].Item.Name != "harper" {
		t.Errorf("got item name %s, want harper", result[0].Item.Name)
	}
	if len(result[0].Positions) != 1 {
		t.Errorf("got %d positions, want 1", len(result[0].Positions))
	}
}

func TestGetItemsWithPositions_All(t *testing.T) {
	db := testDB(t)

	item1 := models.NewItem("harper")
	item2 := models.NewItem("car")
	if err := db.CreateItem(item1); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}
	if err := db.CreateItem(item2); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	pos1 := models.NewPosition(item1.ID, 41.0, -87.0, nil)
	pos2 := models.NewPosition(item2.ID, 42.0, -88.0, nil)
	if err := db.CreatePosition(pos1); err != nil {
		t.Fatalf("failed to create position: %v", err)
	}
	if err := db.CreatePosition(pos2); err != nil {
		t.Fatalf("failed to create position: %v", err)
	}

	// nil = all items
	result, err := GetItemsWithPositions(db, nil)
	if err != nil {
		t.Fatalf("failed to get items with positions: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("got %d items, want 2", len(result))
	}
}

func TestExportToYAML_Empty(t *testing.T) {
	db := testDB(t)

	data, err := ExportToYAML(db)
	if err != nil {
		t.Fatalf("failed to export empty db: %v", err)
	}

	yamlStr := string(data)
	if !strings.Contains(yamlStr, "items: []") {
		t.Error("expected empty items array")
	}
	if !strings.Contains(yamlStr, "positions: []") {
		t.Error("expected empty positions array")
	}
}

func TestExportToMarkdown_Empty(t *testing.T) {
	db := testDB(t)

	md, err := ExportToMarkdown(db, nil)
	if err != nil {
		t.Fatalf("failed to export empty db: %v", err)
	}

	mdStr := string(md)
	if !strings.Contains(mdStr, "No items tracked") {
		t.Error("expected 'no items' message")
	}
}

func TestExportBackup(t *testing.T) {
	db := testDB(t)

	item := models.NewItem("harper")
	if err := db.CreateItem(item); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	data, err := ExportBackup(db)
	if err != nil {
		t.Fatalf("failed to create backup: %v", err)
	}

	// Backup is just YAML export with different name
	if !strings.Contains(string(data), "version: \"1.0\"") {
		t.Error("backup should be valid YAML")
	}
}

func TestImportBackup(t *testing.T) {
	db := testDB(t)

	yaml := `version: "1.0"
exported_at: "2026-01-31T12:00:00Z"
tool: position

items:
  - id: "` + uuid.New().String() + `"
    name: "test"
    created_at: "2024-12-14T00:00:00Z"

positions: []
`

	// ImportBackup is just ImportFromYAML with different name
	if err := ImportBackup(db, []byte(yaml)); err != nil {
		t.Fatalf("failed to import backup: %v", err)
	}

	items, _ := db.ListItems()
	if len(items) != 1 {
		t.Errorf("got %d items, want 1", len(items))
	}
}
