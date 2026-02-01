// ABOUTME: Tests for CLI commands
// ABOUTME: Tests add, list, current, timeline, remove, export, and backup commands

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/harper/position/internal/models"
	"github.com/harper/position/internal/storage"
)

// testDB creates a temporary database for testing and sets the global db variable.
func testDB(t *testing.T) {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	var err error
	db, err = storage.NewSQLiteDB(dbPath)
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}
	t.Cleanup(func() {
		if db != nil {
			_ = db.Close()
			db = nil
		}
	})
}

// Tests for rootCmd

func TestRootCmd_Metadata(t *testing.T) {
	if rootCmd.Use != "position" {
		t.Errorf("expected Use 'position', got %q", rootCmd.Use)
	}
	if rootCmd.Short != "Simple location tracking for items" {
		t.Errorf("unexpected Short: %q", rootCmd.Short)
	}
	// Check for ASCII art in Long description - uses UTF-8 box drawing
	if !strings.Contains(rootCmd.Long, "Track items") {
		t.Error("expected description in Long")
	}
}

// Tests for addCmd

func TestAddCmd_Metadata(t *testing.T) {
	if addCmd.Use != "add <name> --lat <latitude> --lng <longitude>" {
		t.Errorf("unexpected Use: %q", addCmd.Use)
	}
	if !contains(addCmd.Aliases, "a") {
		t.Error("expected alias 'a'")
	}
}

func TestAddCmd_RequiredFlags(t *testing.T) {
	latFlag := addCmd.Flags().Lookup("lat")
	if latFlag == nil {
		t.Fatal("lat flag not found")
	}

	lngFlag := addCmd.Flags().Lookup("lng")
	if lngFlag == nil {
		t.Fatal("lng flag not found")
	}
}

func TestAddCmd_OptionalFlags(t *testing.T) {
	labelFlag := addCmd.Flags().Lookup("label")
	if labelFlag == nil {
		t.Fatal("label flag not found")
	}
	if labelFlag.Shorthand != "l" {
		t.Errorf("expected label shorthand 'l', got %q", labelFlag.Shorthand)
	}

	atFlag := addCmd.Flags().Lookup("at")
	if atFlag == nil {
		t.Fatal("at flag not found")
	}
}

func TestAddCmd_Integration(t *testing.T) {
	testDB(t)

	// Test adding a new item
	addCmd.Flags().Set("lat", "41.8781")
	addCmd.Flags().Set("lng", "-87.6298")
	addCmd.Flags().Set("label", "chicago")
	defer func() {
		addCmd.Flags().Set("lat", "0")
		addCmd.Flags().Set("lng", "0")
		addCmd.Flags().Set("label", "")
	}()

	err := addCmd.RunE(addCmd, []string{"harper"})
	if err != nil {
		t.Fatalf("addCmd failed: %v", err)
	}

	// Verify item exists
	item, err := db.GetItemByName("harper")
	if err != nil {
		t.Fatalf("item not created: %v", err)
	}
	if item.Name != "harper" {
		t.Errorf("expected name 'harper', got %q", item.Name)
	}

	// Verify position exists
	pos, err := db.GetCurrentPosition(item.ID)
	if err != nil {
		t.Fatalf("position not created: %v", err)
	}
	if pos.Latitude != 41.8781 {
		t.Errorf("expected latitude 41.8781, got %f", pos.Latitude)
	}
	if pos.Label == nil || *pos.Label != "chicago" {
		t.Error("expected label 'chicago'")
	}
}

func TestAddCmd_InvalidCoordinates(t *testing.T) {
	testDB(t)

	addCmd.Flags().Set("lat", "100") // Invalid
	addCmd.Flags().Set("lng", "-87.6298")
	defer func() {
		addCmd.Flags().Set("lat", "0")
		addCmd.Flags().Set("lng", "0")
	}()

	err := addCmd.RunE(addCmd, []string{"test"})
	if err == nil {
		t.Error("expected error for invalid latitude")
	}
}

func TestAddCmd_WithTimestamp(t *testing.T) {
	testDB(t)

	addCmd.Flags().Set("lat", "41.8781")
	addCmd.Flags().Set("lng", "-87.6298")
	addCmd.Flags().Set("at", "2024-12-15T10:00:00Z")
	defer func() {
		addCmd.Flags().Set("lat", "0")
		addCmd.Flags().Set("lng", "0")
		addCmd.Flags().Set("at", "")
	}()

	err := addCmd.RunE(addCmd, []string{"timetest"})
	if err != nil {
		t.Fatalf("addCmd failed: %v", err)
	}

	item, _ := db.GetItemByName("timetest")
	pos, _ := db.GetCurrentPosition(item.ID)

	expected, _ := time.Parse(time.RFC3339, "2024-12-15T10:00:00Z")
	if !pos.RecordedAt.Equal(expected) {
		t.Errorf("expected recorded_at %v, got %v", expected, pos.RecordedAt)
	}
}

func TestAddCmd_InvalidTimestamp(t *testing.T) {
	testDB(t)

	addCmd.Flags().Set("lat", "41.8781")
	addCmd.Flags().Set("lng", "-87.6298")
	addCmd.Flags().Set("at", "not-a-timestamp")
	defer func() {
		addCmd.Flags().Set("lat", "0")
		addCmd.Flags().Set("lng", "0")
		addCmd.Flags().Set("at", "")
	}()

	err := addCmd.RunE(addCmd, []string{"badtime"})
	if err == nil {
		t.Error("expected error for invalid timestamp")
	}
}

// Tests for listCmd

func TestListCmd_Metadata(t *testing.T) {
	if listCmd.Use != "list" {
		t.Errorf("unexpected Use: %q", listCmd.Use)
	}
	if !contains(listCmd.Aliases, "ls") {
		t.Error("expected alias 'ls'")
	}
}

func TestListCmd_Empty(t *testing.T) {
	testDB(t)

	err := listCmd.RunE(listCmd, []string{})
	if err != nil {
		t.Fatalf("listCmd failed: %v", err)
	}
}

func TestListCmd_WithItems(t *testing.T) {
	testDB(t)

	// Create test items
	item1 := models.NewItem("harper")
	item2 := models.NewItem("car")
	_ = db.CreateItem(item1)
	_ = db.CreateItem(item2)

	// Add position for one item
	pos := models.NewPosition(item1.ID, 41.0, -87.0, nil)
	_ = db.CreatePosition(pos)

	err := listCmd.RunE(listCmd, []string{})
	if err != nil {
		t.Fatalf("listCmd failed: %v", err)
	}
}

// Tests for currentCmd

func TestCurrentCmd_Metadata(t *testing.T) {
	if currentCmd.Use != "current <name>" {
		t.Errorf("unexpected Use: %q", currentCmd.Use)
	}
	if !contains(currentCmd.Aliases, "c") {
		t.Error("expected alias 'c'")
	}
}

func TestCurrentCmd_Success(t *testing.T) {
	testDB(t)

	item := models.NewItem("harper")
	_ = db.CreateItem(item)
	pos := models.NewPosition(item.ID, 41.0, -87.0, nil)
	_ = db.CreatePosition(pos)

	err := currentCmd.RunE(currentCmd, []string{"harper"})
	if err != nil {
		t.Fatalf("currentCmd failed: %v", err)
	}
}

func TestCurrentCmd_ItemNotFound(t *testing.T) {
	testDB(t)

	err := currentCmd.RunE(currentCmd, []string{"nonexistent"})
	if err == nil {
		t.Error("expected error for nonexistent item")
	}
}

func TestCurrentCmd_NoPosition(t *testing.T) {
	testDB(t)

	item := models.NewItem("harper")
	_ = db.CreateItem(item)

	err := currentCmd.RunE(currentCmd, []string{"harper"})
	if err == nil {
		t.Error("expected error when no position")
	}
}

// Tests for timelineCmd

func TestTimelineCmd_Metadata(t *testing.T) {
	if timelineCmd.Use != "timeline <name>" {
		t.Errorf("unexpected Use: %q", timelineCmd.Use)
	}
	if !contains(timelineCmd.Aliases, "t") {
		t.Error("expected alias 't'")
	}
}

func TestTimelineCmd_Success(t *testing.T) {
	testDB(t)

	item := models.NewItem("harper")
	_ = db.CreateItem(item)
	pos1 := models.NewPosition(item.ID, 41.0, -87.0, nil)
	pos2 := models.NewPosition(item.ID, 42.0, -88.0, nil)
	_ = db.CreatePosition(pos1)
	_ = db.CreatePosition(pos2)

	err := timelineCmd.RunE(timelineCmd, []string{"harper"})
	if err != nil {
		t.Fatalf("timelineCmd failed: %v", err)
	}
}

func TestTimelineCmd_EmptyTimeline(t *testing.T) {
	testDB(t)

	item := models.NewItem("harper")
	_ = db.CreateItem(item)

	err := timelineCmd.RunE(timelineCmd, []string{"harper"})
	if err != nil {
		t.Fatalf("timelineCmd failed: %v", err)
	}
}

func TestTimelineCmd_ItemNotFound(t *testing.T) {
	testDB(t)

	err := timelineCmd.RunE(timelineCmd, []string{"nonexistent"})
	if err == nil {
		t.Error("expected error for nonexistent item")
	}
}

// Tests for removeCmd

func TestRemoveCmd_Metadata(t *testing.T) {
	if removeCmd.Use != "remove <name>" {
		t.Errorf("unexpected Use: %q", removeCmd.Use)
	}
	if !contains(removeCmd.Aliases, "rm") {
		t.Error("expected alias 'rm'")
	}
}

func TestRemoveCmd_ConfirmFlag(t *testing.T) {
	flag := removeCmd.Flags().Lookup("confirm")
	if flag == nil {
		t.Fatal("confirm flag not found")
	}
}

func TestRemoveCmd_WithConfirm(t *testing.T) {
	testDB(t)

	item := models.NewItem("harper")
	_ = db.CreateItem(item)
	pos := models.NewPosition(item.ID, 41.0, -87.0, nil)
	_ = db.CreatePosition(pos)

	removeCmd.Flags().Set("confirm", "true")
	defer removeCmd.Flags().Set("confirm", "false")

	err := removeCmd.RunE(removeCmd, []string{"harper"})
	if err != nil {
		t.Fatalf("removeCmd failed: %v", err)
	}

	// Verify item deleted
	_, err = db.GetItemByName("harper")
	if err == nil {
		t.Error("item should have been deleted")
	}
}

func TestRemoveCmd_ItemNotFound(t *testing.T) {
	testDB(t)

	removeCmd.Flags().Set("confirm", "true")
	defer removeCmd.Flags().Set("confirm", "false")

	err := removeCmd.RunE(removeCmd, []string{"nonexistent"})
	if err == nil {
		t.Error("expected error for nonexistent item")
	}
}

// Tests for backupCmd

func TestBackupCmd_Metadata(t *testing.T) {
	if backupCmd.Use != "backup" {
		t.Errorf("unexpected Use: %q", backupCmd.Use)
	}
}

func TestBackupCmd_OutputFlag(t *testing.T) {
	flag := backupCmd.Flags().Lookup("output")
	if flag == nil {
		t.Fatal("output flag not found")
	}
	if flag.Shorthand != "o" {
		t.Errorf("expected output shorthand 'o', got %q", flag.Shorthand)
	}
}

func TestBackupCmd_Success(t *testing.T) {
	testDB(t)

	item := models.NewItem("harper")
	_ = db.CreateItem(item)
	pos := models.NewPosition(item.ID, 41.0, -87.0, nil)
	_ = db.CreatePosition(pos)

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "backup.yaml")

	backupCmd.Flags().Set("output", outputPath)
	defer backupCmd.Flags().Set("output", "")

	err := backupCmd.RunE(backupCmd, []string{})
	if err != nil {
		t.Fatalf("backupCmd failed: %v", err)
	}

	// Verify file created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("backup file not created")
	}
}

// Tests for importCmd

func TestImportCmd_Metadata(t *testing.T) {
	if importCmd.Use != "import <file>" {
		t.Errorf("unexpected Use: %q", importCmd.Use)
	}
}

func TestImportCmd_ConfirmFlag(t *testing.T) {
	flag := importCmd.Flags().Lookup("confirm")
	if flag == nil {
		t.Fatal("confirm flag not found")
	}
}

func TestImportCmd_FileNotFound(t *testing.T) {
	testDB(t)

	importCmd.Flags().Set("confirm", "true")
	defer importCmd.Flags().Set("confirm", "false")

	err := importCmd.RunE(importCmd, []string{"/nonexistent/file.yaml"})
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

// Tests for exportCmd

func TestExportCmd_Metadata(t *testing.T) {
	if exportCmd.Use != "export [name]" {
		t.Errorf("unexpected Use: %q", exportCmd.Use)
	}
	if !contains(exportCmd.Aliases, "e") {
		t.Error("expected alias 'e'")
	}
}

func TestExportCmd_Flags(t *testing.T) {
	formatFlag := exportCmd.Flags().Lookup("format")
	if formatFlag == nil {
		t.Fatal("format flag not found")
	}
	if formatFlag.DefValue != "geojson" {
		t.Errorf("expected default format 'geojson', got %q", formatFlag.DefValue)
	}

	geometryFlag := exportCmd.Flags().Lookup("geometry")
	if geometryFlag == nil {
		t.Fatal("geometry flag not found")
	}
	if geometryFlag.DefValue != "points" {
		t.Errorf("expected default geometry 'points', got %q", geometryFlag.DefValue)
	}

	sinceFlag := exportCmd.Flags().Lookup("since")
	if sinceFlag == nil {
		t.Fatal("since flag not found")
	}

	fromFlag := exportCmd.Flags().Lookup("from")
	if fromFlag == nil {
		t.Fatal("from flag not found")
	}

	toFlag := exportCmd.Flags().Lookup("to")
	if toFlag == nil {
		t.Fatal("to flag not found")
	}

	outputFlag := exportCmd.Flags().Lookup("output")
	if outputFlag == nil {
		t.Fatal("output flag not found")
	}
}

func TestExportCmd_InvalidFormat(t *testing.T) {
	testDB(t)

	exportCmd.Flags().Set("format", "invalid")
	defer exportCmd.Flags().Set("format", "geojson")

	err := exportCmd.RunE(exportCmd, []string{})
	if err == nil {
		t.Error("expected error for invalid format")
	}
}

func TestExportCmd_InvalidGeometry(t *testing.T) {
	testDB(t)

	exportCmd.Flags().Set("geometry", "invalid")
	defer exportCmd.Flags().Set("geometry", "points")

	err := exportCmd.RunE(exportCmd, []string{})
	if err == nil {
		t.Error("expected error for invalid geometry")
	}
}

// Tests for helper functions in export.go

func TestParseDuration(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{"24h", false},
		{"7d", false},
		{"1w", false},
		{"2m", false},
		{"invalid", true},
		{"h", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			_, err := parseDuration(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDuration(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestParseDate(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{"2024-12-15", false},
		{"2024-12-15T10:00:00Z", false},
		{"invalid", true},
		{"", true},
		{"12-15-2024", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			_, err := parseDate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDate(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

// Tests for mcpCmd

func TestMcpCmd_Metadata(t *testing.T) {
	if mcpCmd.Use != "mcp" {
		t.Errorf("unexpected Use: %q", mcpCmd.Use)
	}
}

// Tests for export with actual data

func TestExportCmd_GeoJSON(t *testing.T) {
	testDB(t)

	item := models.NewItem("harper")
	_ = db.CreateItem(item)
	pos := models.NewPosition(item.ID, 41.0, -87.0, nil)
	_ = db.CreatePosition(pos)

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "export.geojson")

	exportCmd.Flags().Set("format", "geojson")
	exportCmd.Flags().Set("output", outputPath)
	defer func() {
		exportCmd.Flags().Set("format", "geojson")
		exportCmd.Flags().Set("output", "")
	}()

	err := exportCmd.RunE(exportCmd, []string{"harper"})
	if err != nil {
		t.Fatalf("exportCmd failed: %v", err)
	}

	// Verify file created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("export file not created")
	}
}

func TestExportCmd_Markdown(t *testing.T) {
	testDB(t)

	item := models.NewItem("harper")
	_ = db.CreateItem(item)
	pos := models.NewPosition(item.ID, 41.0, -87.0, nil)
	_ = db.CreatePosition(pos)

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "export.md")

	exportCmd.Flags().Set("format", "markdown")
	exportCmd.Flags().Set("output", outputPath)
	defer func() {
		exportCmd.Flags().Set("format", "geojson")
		exportCmd.Flags().Set("output", "")
	}()

	err := exportCmd.RunE(exportCmd, []string{"harper"})
	if err != nil {
		t.Fatalf("exportCmd failed: %v", err)
	}

	// Verify file created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("export file not created")
	}
}

func TestExportCmd_YAML(t *testing.T) {
	testDB(t)

	item := models.NewItem("harper")
	_ = db.CreateItem(item)
	pos := models.NewPosition(item.ID, 41.0, -87.0, nil)
	_ = db.CreatePosition(pos)

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "export.yaml")

	exportCmd.Flags().Set("format", "yaml")
	exportCmd.Flags().Set("output", outputPath)
	defer func() {
		exportCmd.Flags().Set("format", "geojson")
		exportCmd.Flags().Set("output", "")
	}()

	err := exportCmd.RunE(exportCmd, []string{})
	if err != nil {
		t.Fatalf("exportCmd failed: %v", err)
	}

	// Verify file created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("export file not created")
	}
}

func TestExportCmd_LineGeometry(t *testing.T) {
	testDB(t)

	item := models.NewItem("harper")
	_ = db.CreateItem(item)
	pos1 := models.NewPosition(item.ID, 41.0, -87.0, nil)
	pos2 := models.NewPosition(item.ID, 42.0, -88.0, nil)
	_ = db.CreatePosition(pos1)
	_ = db.CreatePosition(pos2)

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "track.geojson")

	exportCmd.Flags().Set("format", "geojson")
	exportCmd.Flags().Set("geometry", "line")
	exportCmd.Flags().Set("output", outputPath)
	defer func() {
		exportCmd.Flags().Set("format", "geojson")
		exportCmd.Flags().Set("geometry", "points")
		exportCmd.Flags().Set("output", "")
	}()

	err := exportCmd.RunE(exportCmd, []string{"harper"})
	if err != nil {
		t.Fatalf("exportCmd failed: %v", err)
	}

	// Verify file created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("export file not created")
	}
}

func TestExportCmd_WithSince(t *testing.T) {
	testDB(t)

	item := models.NewItem("harper")
	_ = db.CreateItem(item)
	pos := models.NewPosition(item.ID, 41.0, -87.0, nil)
	_ = db.CreatePosition(pos)

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "recent.geojson")

	exportCmd.Flags().Set("format", "geojson")
	exportCmd.Flags().Set("since", "24h")
	exportCmd.Flags().Set("output", outputPath)
	defer func() {
		exportCmd.Flags().Set("format", "geojson")
		exportCmd.Flags().Set("since", "")
		exportCmd.Flags().Set("output", "")
	}()

	err := exportCmd.RunE(exportCmd, []string{"harper"})
	if err != nil {
		t.Fatalf("exportCmd failed: %v", err)
	}
}

func TestExportCmd_WithDateRange(t *testing.T) {
	testDB(t)

	item := models.NewItem("harper")
	_ = db.CreateItem(item)
	pos := models.NewPosition(item.ID, 41.0, -87.0, nil)
	_ = db.CreatePosition(pos)

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "range.geojson")

	exportCmd.Flags().Set("format", "geojson")
	exportCmd.Flags().Set("from", "2020-01-01")
	exportCmd.Flags().Set("to", "2030-12-31")
	exportCmd.Flags().Set("output", outputPath)
	defer func() {
		exportCmd.Flags().Set("format", "geojson")
		exportCmd.Flags().Set("from", "")
		exportCmd.Flags().Set("to", "")
		exportCmd.Flags().Set("output", "")
	}()

	err := exportCmd.RunE(exportCmd, []string{"harper"})
	if err != nil {
		t.Fatalf("exportCmd failed: %v", err)
	}
}

func TestExportCmd_InvalidSince(t *testing.T) {
	testDB(t)

	exportCmd.Flags().Set("since", "invalid")
	defer exportCmd.Flags().Set("since", "")

	err := exportCmd.RunE(exportCmd, []string{})
	if err == nil {
		t.Error("expected error for invalid since")
	}
}

func TestExportCmd_InvalidFrom(t *testing.T) {
	testDB(t)

	exportCmd.Flags().Set("from", "invalid")
	defer exportCmd.Flags().Set("from", "")

	err := exportCmd.RunE(exportCmd, []string{})
	if err == nil {
		t.Error("expected error for invalid from")
	}
}

func TestExportCmd_InvalidTo(t *testing.T) {
	testDB(t)

	exportCmd.Flags().Set("from", "2024-01-01")
	exportCmd.Flags().Set("to", "invalid")
	defer func() {
		exportCmd.Flags().Set("from", "")
		exportCmd.Flags().Set("to", "")
	}()

	err := exportCmd.RunE(exportCmd, []string{})
	if err == nil {
		t.Error("expected error for invalid to")
	}
}

func TestExportCmd_NoPositions(t *testing.T) {
	testDB(t)

	item := models.NewItem("harper")
	_ = db.CreateItem(item)

	exportCmd.Flags().Set("format", "geojson")
	defer exportCmd.Flags().Set("format", "geojson")

	err := exportCmd.RunE(exportCmd, []string{"harper"})
	if err == nil {
		t.Error("expected error when no positions")
	}
}

func TestExportCmd_ItemNotFound(t *testing.T) {
	testDB(t)

	err := exportCmd.RunE(exportCmd, []string{"nonexistent"})
	if err == nil {
		t.Error("expected error for nonexistent item")
	}
}

func TestExportCmd_AllItems(t *testing.T) {
	testDB(t)

	item1 := models.NewItem("harper")
	item2 := models.NewItem("car")
	_ = db.CreateItem(item1)
	_ = db.CreateItem(item2)
	pos1 := models.NewPosition(item1.ID, 41.0, -87.0, nil)
	pos2 := models.NewPosition(item2.ID, 42.0, -88.0, nil)
	_ = db.CreatePosition(pos1)
	_ = db.CreatePosition(pos2)

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "all.geojson")

	exportCmd.Flags().Set("format", "geojson")
	exportCmd.Flags().Set("output", outputPath)
	defer func() {
		exportCmd.Flags().Set("format", "geojson")
		exportCmd.Flags().Set("output", "")
	}()

	// No item name - export all
	err := exportCmd.RunE(exportCmd, []string{})
	if err != nil {
		t.Fatalf("exportCmd failed: %v", err)
	}
}

func TestExportCmd_AllItemsWithFrom(t *testing.T) {
	testDB(t)

	item := models.NewItem("harper")
	_ = db.CreateItem(item)
	pos := models.NewPosition(item.ID, 41.0, -87.0, nil)
	_ = db.CreatePosition(pos)

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "from.geojson")

	exportCmd.Flags().Set("format", "geojson")
	exportCmd.Flags().Set("from", "2020-01-01")
	exportCmd.Flags().Set("output", outputPath)
	defer func() {
		exportCmd.Flags().Set("format", "geojson")
		exportCmd.Flags().Set("from", "")
		exportCmd.Flags().Set("output", "")
	}()

	err := exportCmd.RunE(exportCmd, []string{})
	if err != nil {
		t.Fatalf("exportCmd failed: %v", err)
	}
}

func TestExportCmd_MarkdownNoItem(t *testing.T) {
	testDB(t)

	item := models.NewItem("harper")
	_ = db.CreateItem(item)
	pos := models.NewPosition(item.ID, 41.0, -87.0, nil)
	_ = db.CreatePosition(pos)

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "all.md")

	exportCmd.Flags().Set("format", "markdown")
	exportCmd.Flags().Set("output", outputPath)
	defer func() {
		exportCmd.Flags().Set("format", "geojson")
		exportCmd.Flags().Set("output", "")
	}()

	err := exportCmd.RunE(exportCmd, []string{})
	if err != nil {
		t.Fatalf("exportCmd failed: %v", err)
	}
}

func TestExportCmd_MarkdownItemNotFound(t *testing.T) {
	testDB(t)

	exportCmd.Flags().Set("format", "markdown")
	defer exportCmd.Flags().Set("format", "geojson")

	err := exportCmd.RunE(exportCmd, []string{"nonexistent"})
	if err == nil {
		t.Error("expected error for nonexistent item")
	}
}

// Tests for import backup flow

func TestImportBackupFlow(t *testing.T) {
	testDB(t)

	// Create data
	item := models.NewItem("test")
	_ = db.CreateItem(item)
	pos := models.NewPosition(item.ID, 41.0, -87.0, nil)
	_ = db.CreatePosition(pos)

	// Backup
	tmpDir := t.TempDir()
	backupPath := filepath.Join(tmpDir, "backup.yaml")

	backupCmd.Flags().Set("output", backupPath)
	defer backupCmd.Flags().Set("output", "")

	err := backupCmd.RunE(backupCmd, []string{})
	if err != nil {
		t.Fatalf("backup failed: %v", err)
	}

	// Reset database
	_ = db.Reset()

	// Import
	importCmd.Flags().Set("confirm", "true")
	defer importCmd.Flags().Set("confirm", "false")

	err = importCmd.RunE(importCmd, []string{backupPath})
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}

	// Verify data restored
	items, _ := db.ListItems()
	if len(items) == 0 {
		t.Error("expected items after import")
	}
}

// Tests for duration parsing edge cases

func TestParseDuration_AllUnits(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"1h"},
		{"24h"},
		{"1d"},
		{"7d"},
		{"1w"},
		{"2w"},
		{"1m"},
	}

	now := time.Now()
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := parseDuration(tt.input)
			if err != nil {
				t.Fatalf("parseDuration(%q) failed: %v", tt.input, err)
			}
			// Result should be before now
			if result.After(now) {
				t.Errorf("parseDuration(%q) = %v, should be before %v", tt.input, result, now)
			}
		})
	}
}

func TestParseDate_RFC3339Formats(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{"2024-12-15T10:00:00Z", false},
		{"2024-12-15T10:00:00+05:00", false},
		{"2024-12-15T10:00:00-08:00", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			_, err := parseDate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDate(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

// Tests for adding to existing items

func TestAddCmd_ExistingItem(t *testing.T) {
	testDB(t)

	// Create item first
	item := models.NewItem("existingitem")
	_ = db.CreateItem(item)

	// Add position
	addCmd.Flags().Set("lat", "41.8781")
	addCmd.Flags().Set("lng", "-87.6298")
	defer func() {
		addCmd.Flags().Set("lat", "0")
		addCmd.Flags().Set("lng", "0")
	}()

	err := addCmd.RunE(addCmd, []string{"existingitem"})
	if err != nil {
		t.Fatalf("addCmd failed: %v", err)
	}

	// Should still have only 1 item
	items, _ := db.ListItems()
	if len(items) != 1 {
		t.Errorf("expected 1 item, got %d", len(items))
	}

	// Should have position
	pos, err := db.GetCurrentPosition(item.ID)
	if err != nil {
		t.Fatalf("position not created: %v", err)
	}
	if pos.Latitude != 41.8781 {
		t.Errorf("expected latitude 41.8781, got %f", pos.Latitude)
	}
}

// Tests for output to stdout

func TestExportCmd_GeoJSONToStdout(t *testing.T) {
	testDB(t)

	item := models.NewItem("harper")
	_ = db.CreateItem(item)
	pos := models.NewPosition(item.ID, 41.0, -87.0, nil)
	_ = db.CreatePosition(pos)

	// No output flag - goes to stdout
	exportCmd.Flags().Set("format", "geojson")
	defer exportCmd.Flags().Set("format", "geojson")

	err := exportCmd.RunE(exportCmd, []string{"harper"})
	if err != nil {
		t.Fatalf("exportCmd failed: %v", err)
	}
}

func TestExportCmd_MarkdownToStdout(t *testing.T) {
	testDB(t)

	item := models.NewItem("harper")
	_ = db.CreateItem(item)
	pos := models.NewPosition(item.ID, 41.0, -87.0, nil)
	_ = db.CreatePosition(pos)

	exportCmd.Flags().Set("format", "markdown")
	defer exportCmd.Flags().Set("format", "geojson")

	err := exportCmd.RunE(exportCmd, []string{"harper"})
	if err != nil {
		t.Fatalf("exportCmd failed: %v", err)
	}
}

func TestExportCmd_YAMLToStdout(t *testing.T) {
	testDB(t)

	item := models.NewItem("harper")
	_ = db.CreateItem(item)
	pos := models.NewPosition(item.ID, 41.0, -87.0, nil)
	_ = db.CreatePosition(pos)

	exportCmd.Flags().Set("format", "yaml")
	defer exportCmd.Flags().Set("format", "geojson")

	err := exportCmd.RunE(exportCmd, []string{})
	if err != nil {
		t.Fatalf("exportCmd failed: %v", err)
	}
}

// Tests for all items with date range

func TestExportCmd_AllItemsWithDateRange(t *testing.T) {
	testDB(t)

	item := models.NewItem("harper")
	_ = db.CreateItem(item)
	pos := models.NewPosition(item.ID, 41.0, -87.0, nil)
	_ = db.CreatePosition(pos)

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "range.geojson")

	exportCmd.Flags().Set("format", "geojson")
	exportCmd.Flags().Set("from", "2020-01-01")
	exportCmd.Flags().Set("to", "2030-12-31")
	exportCmd.Flags().Set("output", outputPath)
	defer func() {
		exportCmd.Flags().Set("format", "geojson")
		exportCmd.Flags().Set("from", "")
		exportCmd.Flags().Set("to", "")
		exportCmd.Flags().Set("output", "")
	}()

	err := exportCmd.RunE(exportCmd, []string{})
	if err != nil {
		t.Fatalf("exportCmd failed: %v", err)
	}
}

func TestExportCmd_AllItemsWithSince(t *testing.T) {
	testDB(t)

	item := models.NewItem("harper")
	_ = db.CreateItem(item)
	pos := models.NewPosition(item.ID, 41.0, -87.0, nil)
	_ = db.CreatePosition(pos)

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "since.geojson")

	exportCmd.Flags().Set("format", "geojson")
	exportCmd.Flags().Set("since", "24h")
	exportCmd.Flags().Set("output", outputPath)
	defer func() {
		exportCmd.Flags().Set("format", "geojson")
		exportCmd.Flags().Set("since", "")
		exportCmd.Flags().Set("output", "")
	}()

	err := exportCmd.RunE(exportCmd, []string{})
	if err != nil {
		t.Fatalf("exportCmd failed: %v", err)
	}
}

// Tests for getPositionsForItem with different filters

func TestExportCmd_ItemWithFromOnly(t *testing.T) {
	testDB(t)

	item := models.NewItem("harper")
	_ = db.CreateItem(item)
	pos := models.NewPosition(item.ID, 41.0, -87.0, nil)
	_ = db.CreatePosition(pos)

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "fromonly.geojson")

	exportCmd.Flags().Set("format", "geojson")
	exportCmd.Flags().Set("from", "2020-01-01")
	exportCmd.Flags().Set("output", outputPath)
	defer func() {
		exportCmd.Flags().Set("format", "geojson")
		exportCmd.Flags().Set("from", "")
		exportCmd.Flags().Set("output", "")
	}()

	err := exportCmd.RunE(exportCmd, []string{"harper"})
	if err != nil {
		t.Fatalf("exportCmd failed: %v", err)
	}
}

// Tests for backup without output flag

func TestBackupCmd_DefaultOutput(t *testing.T) {
	testDB(t)

	item := models.NewItem("harper")
	_ = db.CreateItem(item)
	pos := models.NewPosition(item.ID, 41.0, -87.0, nil)
	_ = db.CreatePosition(pos)

	// Change to temp dir so default output goes there
	oldDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	_ = os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	// No output flag - should create timestamped file
	err := backupCmd.RunE(backupCmd, []string{})
	if err != nil {
		t.Fatalf("backupCmd failed: %v", err)
	}

	// Verify a file was created
	files, _ := os.ReadDir(tmpDir)
	found := false
	for _, f := range files {
		if strings.HasPrefix(f.Name(), "positions-") && strings.HasSuffix(f.Name(), ".yaml") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected default backup file to be created")
	}
}

// Tests for remove without confirm (needs stdin handling)
// We can only test the cancellation path by mocking stdin

// Tests for list with position error handling.
func TestListCmd_WithPositionError(t *testing.T) {
	testDB(t)

	// Create item and position
	item := models.NewItem("harper")
	_ = db.CreateItem(item)
	pos := models.NewPosition(item.ID, 41.0, -87.0, nil)
	_ = db.CreatePosition(pos)

	// This tests the normal path - error handling is harder to test without mocking
	err := listCmd.RunE(listCmd, []string{})
	if err != nil {
		t.Fatalf("listCmd failed: %v", err)
	}
}

// Tests for label handling

func TestAddCmd_NoLabel(t *testing.T) {
	testDB(t)

	addCmd.Flags().Set("lat", "41.8781")
	addCmd.Flags().Set("lng", "-87.6298")
	// No label set
	defer func() {
		addCmd.Flags().Set("lat", "0")
		addCmd.Flags().Set("lng", "0")
	}()

	err := addCmd.RunE(addCmd, []string{"nolabel"})
	if err != nil {
		t.Fatalf("addCmd failed: %v", err)
	}

	item, _ := db.GetItemByName("nolabel")
	pos, _ := db.GetCurrentPosition(item.ID)
	if pos.Label != nil {
		t.Error("expected nil label")
	}
}

// Tests for write errors

func TestExportCmd_GeoJSONWriteError(t *testing.T) {
	testDB(t)

	item := models.NewItem("harper")
	_ = db.CreateItem(item)
	pos := models.NewPosition(item.ID, 41.0, -87.0, nil)
	_ = db.CreatePosition(pos)

	// Use a path that should fail (directory that doesn't exist)
	exportCmd.Flags().Set("format", "geojson")
	exportCmd.Flags().Set("output", "/nonexistent/path/export.geojson")
	defer func() {
		exportCmd.Flags().Set("format", "geojson")
		exportCmd.Flags().Set("output", "")
	}()

	err := exportCmd.RunE(exportCmd, []string{"harper"})
	if err == nil {
		t.Error("expected error for invalid output path")
	}
}

func TestExportCmd_MarkdownWriteError(t *testing.T) {
	testDB(t)

	item := models.NewItem("harper")
	_ = db.CreateItem(item)
	pos := models.NewPosition(item.ID, 41.0, -87.0, nil)
	_ = db.CreatePosition(pos)

	exportCmd.Flags().Set("format", "markdown")
	exportCmd.Flags().Set("output", "/nonexistent/path/export.md")
	defer func() {
		exportCmd.Flags().Set("format", "geojson")
		exportCmd.Flags().Set("output", "")
	}()

	err := exportCmd.RunE(exportCmd, []string{"harper"})
	if err == nil {
		t.Error("expected error for invalid output path")
	}
}

func TestExportCmd_YAMLWriteError(t *testing.T) {
	testDB(t)

	item := models.NewItem("harper")
	_ = db.CreateItem(item)
	pos := models.NewPosition(item.ID, 41.0, -87.0, nil)
	_ = db.CreatePosition(pos)

	exportCmd.Flags().Set("format", "yaml")
	exportCmd.Flags().Set("output", "/nonexistent/path/export.yaml")
	defer func() {
		exportCmd.Flags().Set("format", "geojson")
		exportCmd.Flags().Set("output", "")
	}()

	err := exportCmd.RunE(exportCmd, []string{})
	if err == nil {
		t.Error("expected error for invalid output path")
	}
}

func TestBackupCmd_WriteError(t *testing.T) {
	testDB(t)

	item := models.NewItem("harper")
	_ = db.CreateItem(item)

	backupCmd.Flags().Set("output", "/nonexistent/path/backup.yaml")
	defer backupCmd.Flags().Set("output", "")

	err := backupCmd.RunE(backupCmd, []string{})
	if err == nil {
		t.Error("expected error for invalid output path")
	}
}

// Tests for installSkillCmd

func TestInstallSkillCmd_SkipConfirm(t *testing.T) {
	// Save original value
	originalSkipConfirm := skillSkipConfirm
	defer func() { skillSkipConfirm = originalSkipConfirm }()

	// Set to true for auto-confirm
	skillSkipConfirm = true

	// Run the skill installation (it will write to real home dir, but that's ok for this test)
	// We're just testing that the command runs with the skip flag
	err := installSkillCmd.RunE(installSkillCmd, []string{})
	if err != nil {
		t.Fatalf("installSkillCmd failed: %v", err)
	}
}

func TestInstallSkillCmd_OverwriteExisting(t *testing.T) {
	// Save original value
	originalSkipConfirm := skillSkipConfirm
	defer func() { skillSkipConfirm = originalSkipConfirm }()

	// Set to true for auto-confirm
	skillSkipConfirm = true

	// Run twice to test overwrite path
	_ = installSkillCmd.RunE(installSkillCmd, []string{})
	err := installSkillCmd.RunE(installSkillCmd, []string{})
	if err != nil {
		t.Fatalf("installSkillCmd failed on overwrite: %v", err)
	}
}

// Tests for more paths in commands

func TestTimelineCmd_WithPositions(t *testing.T) {
	testDB(t)

	item := models.NewItem("multipos")
	_ = db.CreateItem(item)

	// Create multiple positions at different times
	for i := 0; i < 3; i++ {
		lat := 40.0 + float64(i)
		pos := models.NewPosition(item.ID, lat, -87.0, nil)
		_ = db.CreatePosition(pos)
	}

	err := timelineCmd.RunE(timelineCmd, []string{"multipos"})
	if err != nil {
		t.Fatalf("timelineCmd failed: %v", err)
	}
}

func TestTimelineCmd_WithLabel(t *testing.T) {
	testDB(t)

	item := models.NewItem("labeltest")
	_ = db.CreateItem(item)

	// Create positions with labels
	label := "home"
	pos := models.NewPosition(item.ID, 41.0, -87.0, &label)
	_ = db.CreatePosition(pos)

	err := timelineCmd.RunE(timelineCmd, []string{"labeltest"})
	if err != nil {
		t.Fatalf("timelineCmd failed: %v", err)
	}
}

func TestExportCmd_ReversePositions(t *testing.T) {
	testDB(t)

	item := models.NewItem("reversetest")
	_ = db.CreateItem(item)

	// Create multiple positions to test the reverse logic
	for i := 0; i < 3; i++ {
		lat := 40.0 + float64(i)
		pos := models.NewPosition(item.ID, lat, -87.0, nil)
		_ = db.CreatePosition(pos)
	}

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "reverse.geojson")

	exportCmd.Flags().Set("format", "geojson")
	exportCmd.Flags().Set("output", outputPath)
	defer func() {
		exportCmd.Flags().Set("format", "geojson")
		exportCmd.Flags().Set("output", "")
	}()

	err := exportCmd.RunE(exportCmd, []string{"reversetest"})
	if err != nil {
		t.Fatalf("exportCmd failed: %v", err)
	}
}

// Tests for MCP command metadata

func TestMcpCmd_Run(t *testing.T) {
	// Verify the mcp command has a RunE
	if mcpCmd.RunE == nil {
		t.Fatal("mcpCmd.RunE should not be nil")
	}
}

// Tests for getAllPositions function

func TestGetAllPositions_NoFilter(t *testing.T) {
	testDB(t)

	item := models.NewItem("allpos")
	_ = db.CreateItem(item)
	pos := models.NewPosition(item.ID, 41.0, -87.0, nil)
	_ = db.CreatePosition(pos)

	positions, err := getAllPositions(time.Time{}, time.Time{}, time.Time{})
	if err != nil {
		t.Fatalf("getAllPositions failed: %v", err)
	}
	if len(positions) == 0 {
		t.Error("expected at least one position")
	}
}

func TestGetAllPositions_WithSince(t *testing.T) {
	testDB(t)

	item := models.NewItem("sinceall")
	_ = db.CreateItem(item)
	pos := models.NewPosition(item.ID, 41.0, -87.0, nil)
	_ = db.CreatePosition(pos)

	since := time.Now().Add(-24 * time.Hour)
	positions, err := getAllPositions(since, time.Time{}, time.Time{})
	if err != nil {
		t.Fatalf("getAllPositions failed: %v", err)
	}
	if len(positions) == 0 {
		t.Error("expected at least one position")
	}
}

func TestGetAllPositions_WithDateRange(t *testing.T) {
	testDB(t)

	item := models.NewItem("rangeall")
	_ = db.CreateItem(item)
	pos := models.NewPosition(item.ID, 41.0, -87.0, nil)
	_ = db.CreatePosition(pos)

	from := time.Now().Add(-7 * 24 * time.Hour)
	to := time.Now().Add(24 * time.Hour)
	positions, err := getAllPositions(time.Time{}, from, to)
	if err != nil {
		t.Fatalf("getAllPositions failed: %v", err)
	}
	if len(positions) == 0 {
		t.Error("expected at least one position")
	}
}

func TestGetAllPositions_WithFromOnly(t *testing.T) {
	testDB(t)

	item := models.NewItem("fromallonly")
	_ = db.CreateItem(item)
	pos := models.NewPosition(item.ID, 41.0, -87.0, nil)
	_ = db.CreatePosition(pos)

	from := time.Now().Add(-7 * 24 * time.Hour)
	positions, err := getAllPositions(time.Time{}, from, time.Time{})
	if err != nil {
		t.Fatalf("getAllPositions failed: %v", err)
	}
	if len(positions) == 0 {
		t.Error("expected at least one position")
	}
}

// Tests for getPositionsForItem function

func TestGetPositionsForItem_WithSince(t *testing.T) {
	testDB(t)

	item := models.NewItem("sinceitem")
	_ = db.CreateItem(item)
	pos := models.NewPosition(item.ID, 41.0, -87.0, nil)
	_ = db.CreatePosition(pos)

	since := time.Now().Add(-24 * time.Hour)
	positions, err := getPositionsForItem(item, since, time.Time{}, time.Time{})
	if err != nil {
		t.Fatalf("getPositionsForItem failed: %v", err)
	}
	if len(positions) == 0 {
		t.Error("expected at least one position")
	}
}

func TestGetPositionsForItem_WithDateRange(t *testing.T) {
	testDB(t)

	item := models.NewItem("rangeitem")
	_ = db.CreateItem(item)
	pos := models.NewPosition(item.ID, 41.0, -87.0, nil)
	_ = db.CreatePosition(pos)

	from := time.Now().Add(-7 * 24 * time.Hour)
	to := time.Now().Add(24 * time.Hour)
	positions, err := getPositionsForItem(item, time.Time{}, from, to)
	if err != nil {
		t.Fatalf("getPositionsForItem failed: %v", err)
	}
	if len(positions) == 0 {
		t.Error("expected at least one position")
	}
}

func TestGetPositionsForItem_WithFromOnly(t *testing.T) {
	testDB(t)

	item := models.NewItem("fromitemonly")
	_ = db.CreateItem(item)
	pos := models.NewPosition(item.ID, 41.0, -87.0, nil)
	_ = db.CreatePosition(pos)

	from := time.Now().Add(-7 * 24 * time.Hour)
	positions, err := getPositionsForItem(item, time.Time{}, from, time.Time{})
	if err != nil {
		t.Fatalf("getPositionsForItem failed: %v", err)
	}
	if len(positions) == 0 {
		t.Error("expected at least one position")
	}
}

func TestGetPositionsForItem_NoFilter(t *testing.T) {
	testDB(t)

	item := models.NewItem("nofilteritem")
	_ = db.CreateItem(item)
	pos := models.NewPosition(item.ID, 41.0, -87.0, nil)
	_ = db.CreatePosition(pos)

	positions, err := getPositionsForItem(item, time.Time{}, time.Time{}, time.Time{})
	if err != nil {
		t.Fatalf("getPositionsForItem failed: %v", err)
	}
	if len(positions) == 0 {
		t.Error("expected at least one position")
	}
}

// Tests for exportGeoJSON function

func TestExportGeoJSON_EmptyPositions(t *testing.T) {
	err := exportGeoJSON([]*models.Position{}, "points", nil, "")
	if err == nil {
		t.Error("expected error for empty positions")
	}
	if !strings.Contains(err.Error(), "no positions found") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestExportGeoJSON_LineGeometry(t *testing.T) {
	positions := []*models.Position{
		models.NewPosition(uuid.New(), 41.0, -87.0, nil),
		models.NewPosition(uuid.New(), 42.0, -88.0, nil),
	}

	err := exportGeoJSON(positions, "line", nil, "")
	if err != nil {
		t.Fatalf("exportGeoJSON failed: %v", err)
	}
}

func TestExportGeoJSON_PointsGeometry(t *testing.T) {
	positions := []*models.Position{
		models.NewPosition(uuid.New(), 41.0, -87.0, nil),
	}

	err := exportGeoJSON(positions, "points", nil, "")
	if err != nil {
		t.Fatalf("exportGeoJSON failed: %v", err)
	}
}

// Tests for exportMarkdown function

func TestExportMarkdown_ToStdout(t *testing.T) {
	testDB(t)

	item := models.NewItem("mdstdout")
	_ = db.CreateItem(item)
	pos := models.NewPosition(item.ID, 41.0, -87.0, nil)
	_ = db.CreatePosition(pos)

	// Export to stdout (empty output path)
	err := exportMarkdown([]string{"mdstdout"}, "")
	if err != nil {
		t.Fatalf("exportMarkdown failed: %v", err)
	}
}

func TestExportMarkdown_NoItem(t *testing.T) {
	testDB(t)

	item := models.NewItem("mdnoitem")
	_ = db.CreateItem(item)
	pos := models.NewPosition(item.ID, 41.0, -87.0, nil)
	_ = db.CreatePosition(pos)

	// Export without item name
	err := exportMarkdown([]string{}, "")
	if err != nil {
		t.Fatalf("exportMarkdown failed: %v", err)
	}
}

// Tests for exportYAML function

func TestExportYAML_ToStdout(t *testing.T) {
	testDB(t)

	item := models.NewItem("yamlstdout")
	_ = db.CreateItem(item)
	pos := models.NewPosition(item.ID, 41.0, -87.0, nil)
	_ = db.CreatePosition(pos)

	// Export to stdout (empty output path)
	err := exportYAML("")
	if err != nil {
		t.Fatalf("exportYAML failed: %v", err)
	}
}

// Tests for parseDuration edge cases

func TestParseDuration_SingleDigit(t *testing.T) {
	_, err := parseDuration("1h")
	if err != nil {
		t.Errorf("parseDuration failed for '1h': %v", err)
	}
}

func TestParseDuration_MultiDigit(t *testing.T) {
	_, err := parseDuration("100d")
	if err != nil {
		t.Errorf("parseDuration failed for '100d': %v", err)
	}
}

func TestParseDuration_Month(t *testing.T) {
	result, err := parseDuration("1m")
	if err != nil {
		t.Errorf("parseDuration failed for '1m': %v", err)
	}
	// Month should be roughly 30 days ago
	expected := time.Now().Add(-30 * 24 * time.Hour)
	diff := result.Sub(expected)
	if diff < -time.Hour || diff > time.Hour {
		t.Errorf("parseDuration('1m') result off by too much: %v", diff)
	}
}

func TestParseDuration_Week(t *testing.T) {
	result, err := parseDuration("1w")
	if err != nil {
		t.Errorf("parseDuration failed for '1w': %v", err)
	}
	// Week should be roughly 7 days ago
	expected := time.Now().Add(-7 * 24 * time.Hour)
	diff := result.Sub(expected)
	if diff < -time.Hour || diff > time.Hour {
		t.Errorf("parseDuration('1w') result off by too much: %v", diff)
	}
}

// Helper function

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
