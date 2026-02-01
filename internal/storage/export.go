// ABOUTME: Export and import functionality for position data
// ABOUTME: Supports YAML backup format and markdown export

package storage

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/harper/position/internal/models"
	"gopkg.in/yaml.v3"
)

// BackupVersion is the current backup format version.
const BackupVersion = "1.0"

// Backup represents the YAML backup format.
type Backup struct {
	Version    string           `yaml:"version"`
	ExportedAt time.Time        `yaml:"exported_at"`
	Tool       string           `yaml:"tool"`
	Items      []ItemBackup     `yaml:"items"`
	Positions  []PositionBackup `yaml:"positions"`
}

// ItemBackup represents an item in the backup format.
type ItemBackup struct {
	ID        string    `yaml:"id"`
	Name      string    `yaml:"name"`
	CreatedAt time.Time `yaml:"created_at"`
}

// PositionBackup represents a position in the backup format.
type PositionBackup struct {
	ID         string    `yaml:"id"`
	ItemID     string    `yaml:"item_id"`
	Latitude   float64   `yaml:"latitude"`
	Longitude  float64   `yaml:"longitude"`
	Label      string    `yaml:"label,omitempty"`
	RecordedAt time.Time `yaml:"recorded_at"`
	CreatedAt  time.Time `yaml:"created_at"`
}

// ItemWithPositions groups an item with its positions.
type ItemWithPositions struct {
	Item      *models.Item
	Positions []*models.Position
}

// ExportToYAML exports all data to YAML format.
func ExportToYAML(repo Repository) ([]byte, error) {
	items, err := repo.ListItems()
	if err != nil {
		return nil, fmt.Errorf("list items: %w", err)
	}

	positions, err := repo.GetAllPositions()
	if err != nil {
		return nil, fmt.Errorf("list positions: %w", err)
	}

	backup := Backup{
		Version:    BackupVersion,
		ExportedAt: time.Now().UTC(),
		Tool:       "position",
		Items:      make([]ItemBackup, len(items)),
		Positions:  make([]PositionBackup, len(positions)),
	}

	for i, item := range items {
		backup.Items[i] = ItemBackup{
			ID:        item.ID.String(),
			Name:      item.Name,
			CreatedAt: item.CreatedAt,
		}
	}

	for i, pos := range positions {
		backup.Positions[i] = PositionBackup{
			ID:         pos.ID.String(),
			ItemID:     pos.ItemID.String(),
			Latitude:   pos.Latitude,
			Longitude:  pos.Longitude,
			RecordedAt: pos.RecordedAt,
			CreatedAt:  pos.CreatedAt,
		}
		if pos.Label != nil {
			backup.Positions[i].Label = *pos.Label
		}
	}

	return yaml.Marshal(backup)
}

// ImportFromYAML imports data from YAML format.
// This is a restore operation and does NOT deduplicate positions.
func ImportFromYAML(repo Repository, data []byte) error {
	var backup Backup
	if err := yaml.Unmarshal(data, &backup); err != nil {
		return fmt.Errorf("parse yaml: %w", err)
	}

	if backup.Version != BackupVersion {
		return fmt.Errorf("unsupported backup version: %s (expected %s)", backup.Version, BackupVersion)
	}

	if backup.Tool != "position" {
		return fmt.Errorf("wrong tool: %s (expected position)", backup.Tool)
	}

	// Get direct database access for imports to bypass deduplication
	sqliteDB, ok := repo.(*SQLiteDB)
	if !ok {
		return fmt.Errorf("import requires SQLiteDB")
	}

	// Import items
	for _, itemBackup := range backup.Items {
		id, err := uuid.Parse(itemBackup.ID)
		if err != nil {
			return fmt.Errorf("invalid item ID %s: %w", itemBackup.ID, err)
		}

		item := &models.Item{
			ID:        id,
			Name:      itemBackup.Name,
			CreatedAt: itemBackup.CreatedAt,
		}

		if err := repo.CreateItem(item); err != nil {
			return fmt.Errorf("create item %s: %w", itemBackup.Name, err)
		}
	}

	// Import positions (directly, bypassing deduplication)
	for _, posBackup := range backup.Positions {
		id, err := uuid.Parse(posBackup.ID)
		if err != nil {
			return fmt.Errorf("invalid position ID %s: %w", posBackup.ID, err)
		}

		itemID, err := uuid.Parse(posBackup.ItemID)
		if err != nil {
			return fmt.Errorf("invalid item ID %s: %w", posBackup.ItemID, err)
		}

		var label *string
		if posBackup.Label != "" {
			label = &posBackup.Label
		}

		pos := &models.Position{
			ID:         id,
			ItemID:     itemID,
			Latitude:   posBackup.Latitude,
			Longitude:  posBackup.Longitude,
			Label:      label,
			RecordedAt: posBackup.RecordedAt,
			CreatedAt:  posBackup.CreatedAt,
		}

		// Direct insert to bypass deduplication
		if err := importPositionDirect(sqliteDB, pos); err != nil {
			return fmt.Errorf("create position: %w", err)
		}
	}

	return nil
}

// importPositionDirect inserts a position directly without deduplication.
func importPositionDirect(db *SQLiteDB, pos *models.Position) error {
	_, err := db.db.Exec(
		`INSERT INTO positions (id, item_id, latitude, longitude, label, recorded_at, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		pos.ID.String(), pos.ItemID.String(), pos.Latitude, pos.Longitude,
		pos.Label, pos.RecordedAt, pos.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert position: %w", err)
	}
	return nil
}

// ExportToMarkdown exports data to markdown format.
// If itemID is nil, exports all items.
func ExportToMarkdown(repo Repository, itemID *uuid.UUID) ([]byte, error) {
	data, err := GetItemsWithPositions(repo, itemID)
	if err != nil {
		return nil, err
	}

	var sb strings.Builder

	// Header per spec
	now := time.Now().UTC()
	sb.WriteString(fmt.Sprintf("# Position Export - %s\n\n", now.Format("2006-01-02")))
	sb.WriteString(fmt.Sprintf("Generated: %s\n\n", now.Format(time.RFC3339)))

	if len(data) == 0 {
		sb.WriteString("No items tracked.\n")
		return []byte(sb.String()), nil
	}

	for _, iwp := range data {
		sb.WriteString(fmt.Sprintf("## %s\n\n", iwp.Item.Name))

		if len(iwp.Positions) == 0 {
			sb.WriteString("No positions recorded.\n\n")
			continue
		}

		sb.WriteString("| Date | Location | Coordinates |\n")
		sb.WriteString("|------|----------|-------------|\n")

		for _, pos := range iwp.Positions {
			date := pos.RecordedAt.Format("2006-01-02 15:04")
			location := "-"
			if pos.Label != nil {
				location = *pos.Label
			}
			coords := fmt.Sprintf("(%.4f, %.4f)", pos.Latitude, pos.Longitude)
			sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n", date, location, coords))
		}

		sb.WriteString("\n")
	}

	return []byte(sb.String()), nil
}

// GetItemsWithPositions retrieves items with their positions.
// If itemID is nil, returns all items.
func GetItemsWithPositions(repo Repository, itemID *uuid.UUID) ([]ItemWithPositions, error) {
	var items []*models.Item

	if itemID != nil {
		item, err := repo.GetItemByID(*itemID)
		if err != nil {
			return nil, err
		}
		items = []*models.Item{item}
	} else {
		var err error
		items, err = repo.ListItems()
		if err != nil {
			return nil, err
		}
	}

	result := make([]ItemWithPositions, len(items))
	for i, item := range items {
		positions, err := repo.GetTimeline(item.ID)
		if err != nil {
			return nil, fmt.Errorf("get timeline for %s: %w", item.Name, err)
		}

		result[i] = ItemWithPositions{
			Item:      item,
			Positions: positions,
		}
	}

	return result, nil
}

// ExportBackup creates a YAML backup (alias for ExportToYAML).
func ExportBackup(repo Repository) ([]byte, error) {
	return ExportToYAML(repo)
}

// ImportBackup restores from a YAML backup (alias for ImportFromYAML).
func ImportBackup(repo Repository, data []byte) error {
	return ImportFromYAML(repo, data)
}
