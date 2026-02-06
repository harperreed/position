// ABOUTME: Tests for storage migration between position backends
// ABOUTME: Covers sqlite-to-markdown, markdown-to-sqlite, data integrity, and roundtrips

package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/harper/position/internal/models"
)

// seedPositionData populates a storage backend with a representative data set
// and returns the items and positions for verification.
func seedPositionData(t *testing.T, src Repository) (items []*models.Item, positions []*models.Position) {
	t.Helper()

	// Create items
	item1 := models.NewItem("harper")
	item2 := models.NewItem("car")
	mustNoError(t, src.CreateItem(item1))
	mustNoError(t, src.CreateItem(item2))
	items = append(items, item1, item2)

	// Create positions for item1 at different times and locations
	label1 := "chicago"
	label2 := "new york"
	pos1 := models.NewPositionWithRecordedAt(item1.ID, 41.8781, -87.6298, &label1,
		time.Date(2026, 2, 1, 10, 0, 0, 0, time.UTC))
	pos2 := models.NewPositionWithRecordedAt(item1.ID, 40.7128, -74.0060, &label2,
		time.Date(2026, 2, 2, 10, 0, 0, 0, time.UTC))
	// Position without label
	pos3 := models.NewPositionWithRecordedAt(item1.ID, 34.0522, -118.2437, nil,
		time.Date(2026, 2, 3, 10, 0, 0, 0, time.UTC))

	// Position for item2
	label4 := "garage"
	pos4 := models.NewPositionWithRecordedAt(item2.ID, 42.0, -88.0, &label4,
		time.Date(2026, 2, 1, 8, 0, 0, 0, time.UTC))

	// Use direct insert to bypass deduplication for SQLite
	for _, pos := range []*models.Position{pos1, pos2, pos3, pos4} {
		mustNoError(t, createPositionDirect(src, pos))
	}
	positions = append(positions, pos1, pos2, pos3, pos4)

	return
}

func mustNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// verifyMigratedPositionData checks that the destination contains all expected data.
func verifyMigratedPositionData(t *testing.T, dst Repository, items []*models.Item, positions []*models.Position) {
	t.Helper()

	// Verify items
	for _, orig := range items {
		got, err := dst.GetItemByID(orig.ID)
		if err != nil {
			t.Errorf("item %s (%s) not found in destination: %v", orig.Name, orig.ID, err)
			continue
		}
		if got.Name != orig.Name {
			t.Errorf("item name mismatch: want %q, got %q", orig.Name, got.Name)
		}
	}

	// Verify positions
	for _, orig := range positions {
		got, err := dst.GetPosition(orig.ID)
		if err != nil {
			t.Errorf("position %s not found in destination: %v", orig.ID, err)
			continue
		}
		if got.Latitude != orig.Latitude {
			t.Errorf("position latitude mismatch: want %f, got %f", orig.Latitude, got.Latitude)
		}
		if got.Longitude != orig.Longitude {
			t.Errorf("position longitude mismatch: want %f, got %f", orig.Longitude, got.Longitude)
		}
		if got.ItemID != orig.ItemID {
			t.Errorf("position itemID mismatch: want %s, got %s", orig.ItemID, got.ItemID)
		}
		// Check label
		if (orig.Label == nil) != (got.Label == nil) {
			t.Errorf("position label nil mismatch: orig=%v, got=%v", orig.Label, got.Label)
		} else if orig.Label != nil && *orig.Label != *got.Label {
			t.Errorf("position label mismatch: want %q, got %q", *orig.Label, *got.Label)
		}
	}
}

func TestMigrateData_SqliteToMarkdown(t *testing.T) {
	// Set up source (sqlite)
	srcDir := t.TempDir()
	src, err := NewSQLiteDB(filepath.Join(srcDir, "position.db"))
	if err != nil {
		t.Fatalf("create source store: %v", err)
	}
	defer src.Close()

	items, positions := seedPositionData(t, src)

	// Set up destination (markdown)
	dstDir := t.TempDir()
	dst, err := NewMarkdownStore(dstDir)
	if err != nil {
		t.Fatalf("create destination store: %v", err)
	}
	defer dst.Close()

	// Run migration
	summary, err := MigrateData(src, dst)
	if err != nil {
		t.Fatalf("MigrateData failed: %v", err)
	}

	// Verify summary counts
	if summary.Items != len(items) {
		t.Errorf("summary items: want %d, got %d", len(items), summary.Items)
	}
	if summary.Positions != len(positions) {
		t.Errorf("summary positions: want %d, got %d", len(positions), summary.Positions)
	}

	// Verify all data was migrated correctly
	verifyMigratedPositionData(t, dst, items, positions)
}

func TestMigrateData_MarkdownToSqlite(t *testing.T) {
	// Set up source (markdown)
	srcDir := t.TempDir()
	src, err := NewMarkdownStore(srcDir)
	if err != nil {
		t.Fatalf("create source store: %v", err)
	}
	defer src.Close()

	items, positions := seedPositionData(t, src)

	// Set up destination (sqlite)
	dstDir := t.TempDir()
	dst, err := NewSQLiteDB(filepath.Join(dstDir, "position.db"))
	if err != nil {
		t.Fatalf("create destination store: %v", err)
	}
	defer dst.Close()

	// Run migration
	summary, err := MigrateData(src, dst)
	if err != nil {
		t.Fatalf("MigrateData failed: %v", err)
	}

	// Verify summary counts
	if summary.Items != len(items) {
		t.Errorf("summary items: want %d, got %d", len(items), summary.Items)
	}
	if summary.Positions != len(positions) {
		t.Errorf("summary positions: want %d, got %d", len(positions), summary.Positions)
	}

	// Verify all data was migrated correctly
	verifyMigratedPositionData(t, dst, items, positions)
}

func TestMigrateData_EmptySource(t *testing.T) {
	// Set up empty source (sqlite)
	srcDir := t.TempDir()
	src, err := NewSQLiteDB(filepath.Join(srcDir, "position.db"))
	if err != nil {
		t.Fatalf("create source store: %v", err)
	}
	defer src.Close()

	// Set up destination (markdown)
	dstDir := t.TempDir()
	dst, err := NewMarkdownStore(dstDir)
	if err != nil {
		t.Fatalf("create destination store: %v", err)
	}
	defer dst.Close()

	summary, err := MigrateData(src, dst)
	if err != nil {
		t.Fatalf("MigrateData failed: %v", err)
	}

	if summary.Items != 0 || summary.Positions != 0 {
		t.Errorf("expected all zero counts for empty source, got items=%d positions=%d",
			summary.Items, summary.Positions)
	}
}

func TestMigrateRoundTrip_SqliteToMarkdownToSqlite(t *testing.T) {
	// Phase 1: Create data in SQLite
	srcDir := t.TempDir()
	original, err := NewSQLiteDB(filepath.Join(srcDir, "original.db"))
	if err != nil {
		t.Fatalf("create original store: %v", err)
	}
	defer original.Close()

	items, positions := seedPositionData(t, original)

	// Phase 2: Migrate SQLite -> Markdown
	mdDir := t.TempDir()
	mdStore, err := NewMarkdownStore(mdDir)
	if err != nil {
		t.Fatalf("create markdown store: %v", err)
	}
	defer mdStore.Close()

	summary1, err := MigrateData(original, mdStore)
	if err != nil {
		t.Fatalf("MigrateData (sqlite->markdown) failed: %v", err)
	}
	if summary1.Items != len(items) || summary1.Positions != len(positions) {
		t.Errorf("phase 1 summary mismatch: items=%d/%d positions=%d/%d",
			summary1.Items, len(items), summary1.Positions, len(positions))
	}

	// Phase 3: Migrate Markdown -> new SQLite
	dstDir := t.TempDir()
	final, err := NewSQLiteDB(filepath.Join(dstDir, "final.db"))
	if err != nil {
		t.Fatalf("create final store: %v", err)
	}
	defer final.Close()

	summary2, err := MigrateData(mdStore, final)
	if err != nil {
		t.Fatalf("MigrateData (markdown->sqlite) failed: %v", err)
	}
	if summary2.Items != len(items) || summary2.Positions != len(positions) {
		t.Errorf("phase 2 summary mismatch: items=%d/%d positions=%d/%d",
			summary2.Items, len(items), summary2.Positions, len(positions))
	}

	// Phase 4: Field-by-field verification against original data
	verifyMigratedPositionData(t, final, items, positions)
}

func TestMigrateRoundTrip_MarkdownToSqliteToMarkdown(t *testing.T) {
	// Phase 1: Create data in Markdown
	srcDir := t.TempDir()
	original, err := NewMarkdownStore(srcDir)
	if err != nil {
		t.Fatalf("create original store: %v", err)
	}
	defer original.Close()

	items, positions := seedPositionData(t, original)

	// Phase 2: Migrate Markdown -> SQLite
	sqlDir := t.TempDir()
	sqlStore, err := NewSQLiteDB(filepath.Join(sqlDir, "mid.db"))
	if err != nil {
		t.Fatalf("create sqlite store: %v", err)
	}
	defer sqlStore.Close()

	_, err = MigrateData(original, sqlStore)
	if err != nil {
		t.Fatalf("MigrateData (markdown->sqlite) failed: %v", err)
	}

	// Phase 3: Migrate SQLite -> new Markdown
	dstDir := t.TempDir()
	final, err := NewMarkdownStore(dstDir)
	if err != nil {
		t.Fatalf("create final store: %v", err)
	}
	defer final.Close()

	_, err = MigrateData(sqlStore, final)
	if err != nil {
		t.Fatalf("MigrateData (sqlite->markdown) failed: %v", err)
	}

	// Phase 4: Verify all data
	verifyMigratedPositionData(t, final, items, positions)
}

func TestIsDirNonEmpty_Empty(t *testing.T) {
	emptyDir := t.TempDir()
	nonEmpty, err := IsDirNonEmpty(emptyDir)
	if err != nil {
		t.Fatalf("IsDirNonEmpty on empty dir: %v", err)
	}
	if nonEmpty {
		t.Error("expected empty dir to be reported as empty")
	}
}

func TestIsDirNonEmpty_NonEmpty(t *testing.T) {
	nonEmptyDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(nonEmptyDir, "file.txt"), []byte("data"), 0644); err != nil {
		t.Fatalf("create file: %v", err)
	}
	nonEmpty, err := IsDirNonEmpty(nonEmptyDir)
	if err != nil {
		t.Fatalf("IsDirNonEmpty on non-empty dir: %v", err)
	}
	if !nonEmpty {
		t.Error("expected non-empty dir to be reported as non-empty")
	}
}

func TestIsDirNonEmpty_NonExistent(t *testing.T) {
	nonEmpty, err := IsDirNonEmpty(filepath.Join(t.TempDir(), "does-not-exist"))
	if err != nil {
		t.Fatalf("IsDirNonEmpty on non-existent dir: %v", err)
	}
	if nonEmpty {
		t.Error("expected non-existent dir to be reported as empty")
	}
}

func TestMigrateData_PreservesPositionOrdering(t *testing.T) {
	// Verify positions arrive in correct order after migration
	srcDir := t.TempDir()
	src, err := NewSQLiteDB(filepath.Join(srcDir, "position.db"))
	if err != nil {
		t.Fatalf("create source store: %v", err)
	}
	defer src.Close()

	item := models.NewItem("harper")
	mustNoError(t, src.CreateItem(item))

	// Create positions with defined order
	for i := 0; i < 5; i++ {
		ts := time.Date(2026, 2, 1+i, 10, 0, 0, 0, time.UTC)
		lat := 40.0 + float64(i)
		pos := models.NewPositionWithRecordedAt(item.ID, lat, -80.0, nil, ts)
		mustNoError(t, createPositionDirect(src, pos))
	}

	// Migrate to markdown
	dstDir := t.TempDir()
	dst, err := NewMarkdownStore(dstDir)
	if err != nil {
		t.Fatalf("create destination store: %v", err)
	}
	defer dst.Close()

	_, err = MigrateData(src, dst)
	if err != nil {
		t.Fatalf("MigrateData failed: %v", err)
	}

	// Verify ordering: timeline returns newest first
	timeline, err := dst.GetTimeline(item.ID)
	if err != nil {
		t.Fatalf("GetTimeline failed: %v", err)
	}
	if len(timeline) != 5 {
		t.Fatalf("expected 5 positions, got %d", len(timeline))
	}
	for i := 0; i < len(timeline)-1; i++ {
		if !timeline[i].RecordedAt.After(timeline[i+1].RecordedAt) {
			t.Errorf("positions not in order: pos[%d].RecordedAt=%v <= pos[%d].RecordedAt=%v",
				i, timeline[i].RecordedAt, i+1, timeline[i+1].RecordedAt)
		}
	}
}

func TestMigrateData_PreservesLabels(t *testing.T) {
	srcDir := t.TempDir()
	src, err := NewSQLiteDB(filepath.Join(srcDir, "position.db"))
	if err != nil {
		t.Fatalf("create source store: %v", err)
	}
	defer src.Close()

	item := models.NewItem("harper")
	mustNoError(t, src.CreateItem(item))

	// Position with label
	label := "downtown"
	pos1 := models.NewPositionWithRecordedAt(item.ID, 41.0, -87.0, &label,
		time.Date(2026, 2, 1, 10, 0, 0, 0, time.UTC))
	mustNoError(t, createPositionDirect(src, pos1))

	// Position without label
	pos2 := models.NewPositionWithRecordedAt(item.ID, 42.0, -88.0, nil,
		time.Date(2026, 2, 2, 10, 0, 0, 0, time.UTC))
	mustNoError(t, createPositionDirect(src, pos2))

	// Migrate to markdown
	dstDir := t.TempDir()
	dst, err := NewMarkdownStore(dstDir)
	if err != nil {
		t.Fatalf("create destination store: %v", err)
	}
	defer dst.Close()

	_, err = MigrateData(src, dst)
	if err != nil {
		t.Fatalf("MigrateData failed: %v", err)
	}

	// Verify labels preserved
	got1, err := dst.GetPosition(pos1.ID)
	if err != nil {
		t.Fatalf("GetPosition (with label) failed: %v", err)
	}
	if got1.Label == nil || *got1.Label != "downtown" {
		t.Errorf("expected label 'downtown', got %v", got1.Label)
	}

	got2, err := dst.GetPosition(pos2.ID)
	if err != nil {
		t.Fatalf("GetPosition (without label) failed: %v", err)
	}
	if got2.Label != nil {
		t.Errorf("expected nil label, got %v", got2.Label)
	}
}
