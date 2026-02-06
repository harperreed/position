// ABOUTME: Data migration between position storage backends
// ABOUTME: Copies items and positions from source to destination repository

package storage

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/harper/position/internal/models"
	"github.com/harper/suite/mdstore"
)

// MigrateSummary holds counts of migrated entities.
type MigrateSummary struct {
	Items     int
	Positions int
}

// MigrateData copies all data from src to dst storage.
// It iterates through items and their positions in order,
// creating each entity in the destination. The destination should be empty
// before calling this function.
func MigrateData(src, dst Repository) (*MigrateSummary, error) {
	summary := &MigrateSummary{}

	// List all items
	items, err := src.ListItems()
	if err != nil {
		return nil, fmt.Errorf("list source items: %w", err)
	}

	for _, item := range items {
		if err := dst.CreateItem(item); err != nil {
			return nil, fmt.Errorf("create item %q: %w", item.Name, err)
		}
		summary.Items++

		// Get all positions for this item (timeline returns newest first)
		positions, err := src.GetTimeline(item.ID)
		if err != nil {
			return nil, fmt.Errorf("get timeline for item %q: %w", item.Name, err)
		}

		// Write positions oldest first so deduplication doesn't interfere
		// We need to bypass the deduplication logic during migration,
		// so we write positions directly for markdown and use direct insert for sqlite.
		for i := len(positions) - 1; i >= 0; i-- {
			pos := positions[i]
			if err := createPositionDirect(dst, pos); err != nil {
				return nil, fmt.Errorf("create position for item %q: %w", item.Name, err)
			}
			summary.Positions++
		}
	}

	return summary, nil
}

// createPositionDirect creates a position without deduplication.
// For SQLiteDB, it uses direct SQL insert.
// For MarkdownStore, it writes the file directly.
func createPositionDirect(dst Repository, pos *models.Position) error {
	switch d := dst.(type) {
	case *SQLiteDB:
		return importPositionDirect(d, pos)
	case *MarkdownStore:
		itemDir, err := d.resolveItemDir(pos.ItemID)
		if err != nil {
			return fmt.Errorf("resolve item directory: %w", err)
		}
		if err := mdstore.EnsureDir(itemDir); err != nil {
			return fmt.Errorf("create item directory: %w", err)
		}
		filename := positionFileName(pos)
		path := filepath.Join(itemDir, filename)
		return writePositionFile(path, pos)
	default:
		// Fallback: use the regular CreatePosition (may deduplicate)
		return dst.CreatePosition(pos)
	}
}

// IsDirNonEmpty checks whether a directory exists and contains any files or subdirectories.
// Returns false if the directory does not exist or is empty.
func IsDirNonEmpty(path string) (bool, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("read directory %q: %w", path, err)
	}
	return len(entries) > 0, nil
}
