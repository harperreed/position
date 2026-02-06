// ABOUTME: Markdown file-based storage backend for position data
// ABOUTME: Stores items in _items.yaml and positions as markdown files in per-item directories

package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/harper/position/internal/models"
	"github.com/harper/suite/mdstore"
	"gopkg.in/yaml.v3"
)

// MarkdownStore provides file-based storage for position data using markdown files and YAML.
type MarkdownStore struct {
	dataDir string
}

// Compile-time check that MarkdownStore implements Repository.
var _ Repository = (*MarkdownStore)(nil)

// NewMarkdownStore creates a new markdown-backed store rooted at dataDir.
func NewMarkdownStore(dataDir string) (*MarkdownStore, error) {
	if err := mdstore.EnsureDir(dataDir); err != nil {
		return nil, fmt.Errorf("create data directory: %w", err)
	}
	return &MarkdownStore{dataDir: dataDir}, nil
}

// Close releases resources. For MarkdownStore this is a no-op.
func (s *MarkdownStore) Close() error {
	return nil
}

// Sync is a no-op for local markdown storage.
func (s *MarkdownStore) Sync() error {
	return nil
}

// Reset clears all data from the store.
func (s *MarkdownStore) Reset() error {
	return mdstore.WithLock(s.dataDir, func() error {
		entries, err := os.ReadDir(s.dataDir)
		if err != nil {
			return fmt.Errorf("read data directory: %w", err)
		}
		for _, entry := range entries {
			name := entry.Name()
			// Skip the lock file
			if name == ".lock" {
				continue
			}
			path := filepath.Join(s.dataDir, name)
			if err := os.RemoveAll(path); err != nil {
				return fmt.Errorf("remove %s: %w", name, err)
			}
		}
		return nil
	})
}

// --- Item YAML types ---

// itemEntry represents a single item in the _items.yaml file.
type itemEntry struct {
	ID        string `yaml:"id"`
	Name      string `yaml:"name"`
	CreatedAt string `yaml:"created_at"`
}

// toModel converts an itemEntry to a models.Item.
func (e *itemEntry) toModel() (*models.Item, error) {
	id, err := uuid.Parse(e.ID)
	if err != nil {
		return nil, fmt.Errorf("parse item ID %q: %w", e.ID, err)
	}
	createdAt, err := mdstore.ParseTime(e.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("parse item created_at %q: %w", e.CreatedAt, err)
	}
	return &models.Item{
		ID:        id,
		Name:      e.Name,
		CreatedAt: createdAt,
	}, nil
}

// fromItemModel converts a models.Item to an itemEntry.
func fromItemModel(item *models.Item) itemEntry {
	return itemEntry{
		ID:        item.ID.String(),
		Name:      item.Name,
		CreatedAt: mdstore.FormatTime(item.CreatedAt.UTC()),
	}
}

// --- Item file paths ---

// itemsFilePath returns the path to the _items.yaml file.
func (s *MarkdownStore) itemsFilePath() string {
	return filepath.Join(s.dataDir, "_items.yaml")
}

// itemDirPath returns the directory path for an item's positions.
func (s *MarkdownStore) itemDirPath(itemName string) string {
	return filepath.Join(s.dataDir, mdstore.Slugify(itemName))
}

// readItems reads the _items.yaml file.
func (s *MarkdownStore) readItems() ([]itemEntry, error) {
	var entries []itemEntry
	if err := mdstore.ReadYAML(s.itemsFilePath(), &entries); err != nil {
		return nil, fmt.Errorf("read items file: %w", err)
	}
	return entries, nil
}

// writeItems writes the _items.yaml file atomically.
func (s *MarkdownStore) writeItems(entries []itemEntry) error {
	return mdstore.WriteYAML(s.itemsFilePath(), entries)
}

// --- Item operations ---

// CreateItem creates a new item.
func (s *MarkdownStore) CreateItem(item *models.Item) error {
	return mdstore.WithLock(s.dataDir, func() error {
		entries, err := s.readItems()
		if err != nil {
			return err
		}

		// Check for duplicate name
		for _, e := range entries {
			if e.Name == item.Name {
				return fmt.Errorf("item with name %q already exists", item.Name)
			}
		}

		entries = append(entries, fromItemModel(item))
		if err := s.writeItems(entries); err != nil {
			return err
		}

		// Create item directory
		return mdstore.EnsureDir(s.itemDirPath(item.Name))
	})
}

// GetItemByID retrieves an item by its UUID.
func (s *MarkdownStore) GetItemByID(id uuid.UUID) (*models.Item, error) {
	entries, err := s.readItems()
	if err != nil {
		return nil, err
	}

	for _, e := range entries {
		if e.ID == id.String() {
			return e.toModel()
		}
	}
	return nil, ErrNotFound
}

// GetItemByName retrieves an item by its name.
func (s *MarkdownStore) GetItemByName(name string) (*models.Item, error) {
	entries, err := s.readItems()
	if err != nil {
		return nil, err
	}

	for _, e := range entries {
		if e.Name == name {
			return e.toModel()
		}
	}
	return nil, ErrNotFound
}

// ListItems returns all items sorted by name.
func (s *MarkdownStore) ListItems() ([]*models.Item, error) {
	entries, err := s.readItems()
	if err != nil {
		return nil, err
	}

	var items []*models.Item
	for _, e := range entries {
		item, err := e.toModel()
		if err != nil {
			// Skip malformed entries
			continue
		}
		items = append(items, item)
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})

	return items, nil
}

// DeleteItem removes an item and all its positions.
func (s *MarkdownStore) DeleteItem(id uuid.UUID) error {
	return mdstore.WithLock(s.dataDir, func() error {
		entries, err := s.readItems()
		if err != nil {
			return err
		}

		var remaining []itemEntry
		var itemName string
		for _, e := range entries {
			if e.ID == id.String() {
				itemName = e.Name
			} else {
				remaining = append(remaining, e)
			}
		}

		if itemName == "" {
			// Item not found, but match SQLite behavior (no error)
			return nil
		}

		if err := s.writeItems(remaining); err != nil {
			return err
		}

		// Remove item directory with all positions
		itemDir := s.itemDirPath(itemName)
		if _, err := os.Stat(itemDir); err == nil {
			return os.RemoveAll(itemDir)
		}
		return nil
	})
}

// --- Position frontmatter ---

// positionFrontmatter holds the YAML frontmatter of a position markdown file.
type positionFrontmatter struct {
	ID         string  `yaml:"id"`
	ItemID     string  `yaml:"item_id"`
	Latitude   float64 `yaml:"latitude"`
	Longitude  float64 `yaml:"longitude"`
	Label      string  `yaml:"label,omitempty"`
	RecordedAt string  `yaml:"recorded_at"`
	CreatedAt  string  `yaml:"created_at"`
}

// toModel converts a positionFrontmatter to a models.Position.
func (fm *positionFrontmatter) toModel() (*models.Position, error) {
	id, err := uuid.Parse(fm.ID)
	if err != nil {
		return nil, fmt.Errorf("parse position ID %q: %w", fm.ID, err)
	}
	itemID, err := uuid.Parse(fm.ItemID)
	if err != nil {
		return nil, fmt.Errorf("parse item ID %q: %w", fm.ItemID, err)
	}
	recordedAt, err := mdstore.ParseTime(fm.RecordedAt)
	if err != nil {
		return nil, fmt.Errorf("parse recorded_at %q: %w", fm.RecordedAt, err)
	}
	createdAt, err := mdstore.ParseTime(fm.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("parse created_at %q: %w", fm.CreatedAt, err)
	}

	var label *string
	if fm.Label != "" {
		label = &fm.Label
	}

	return &models.Position{
		ID:         id,
		ItemID:     itemID,
		Latitude:   fm.Latitude,
		Longitude:  fm.Longitude,
		Label:      label,
		RecordedAt: recordedAt,
		CreatedAt:  createdAt,
	}, nil
}

// fromPositionModel converts a models.Position to a positionFrontmatter.
func fromPositionModel(pos *models.Position) positionFrontmatter {
	fm := positionFrontmatter{
		ID:         pos.ID.String(),
		ItemID:     pos.ItemID.String(),
		Latitude:   pos.Latitude,
		Longitude:  pos.Longitude,
		RecordedAt: mdstore.FormatTime(pos.RecordedAt.UTC()),
		CreatedAt:  mdstore.FormatTime(pos.CreatedAt.UTC()),
	}
	if pos.Label != nil {
		fm.Label = *pos.Label
	}
	return fm
}

// positionFileName generates a filename for a position record.
// Format: <timestamp>-<id-prefix>.md.
func positionFileName(pos *models.Position) string {
	ts := pos.RecordedAt.UTC().Format("2006-01-02T15-04-05")
	idPrefix := pos.ID.String()[:8]
	return fmt.Sprintf("%s-%s.md", ts, idPrefix)
}

// --- Position file helpers ---

// readPositionFile reads a position from a markdown file.
func readPositionFile(path string) (*models.Position, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	yamlStr, _ := mdstore.ParseFrontmatter(string(data))
	if yamlStr == "" {
		return nil, fmt.Errorf("no frontmatter found in %s", path)
	}

	var fm positionFrontmatter
	if err := yaml.Unmarshal([]byte(yamlStr), &fm); err != nil {
		return nil, fmt.Errorf("parse frontmatter in %s: %w", path, err)
	}

	return fm.toModel()
}

// writePositionFile writes a position as a markdown file.
func writePositionFile(path string, pos *models.Position) error {
	fm := fromPositionModel(pos)

	var body string
	if pos.Label != nil {
		body = fmt.Sprintf("\n%s\n", *pos.Label)
	}

	content, err := mdstore.RenderFrontmatter(&fm, body)
	if err != nil {
		return fmt.Errorf("render position frontmatter: %w", err)
	}

	return mdstore.AtomicWrite(path, []byte(content))
}

// resolveItemDir finds the item directory path for a given item ID.
// Reads _items.yaml to find the item name, then returns the slugified directory path.
func (s *MarkdownStore) resolveItemDir(itemID uuid.UUID) (string, error) {
	entries, err := s.readItems()
	if err != nil {
		return "", err
	}

	for _, e := range entries {
		if e.ID == itemID.String() {
			return s.itemDirPath(e.Name), nil
		}
	}
	return "", ErrNotFound
}

// readAllPositionsInDir reads all position files from a directory.
func readAllPositionsInDir(dir string) ([]*models.Position, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read directory %s: %w", dir, err)
	}

	var positions []*models.Position
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		pos, err := readPositionFile(path)
		if err != nil {
			// Skip malformed files
			continue
		}
		positions = append(positions, pos)
	}

	return positions, nil
}

// --- Position operations ---

// CreatePosition creates a new position with deduplication.
// If the new position matches the current position for the item, it's silently skipped.
func (s *MarkdownStore) CreatePosition(pos *models.Position) error {
	// Check for duplicate against current position
	current, err := s.GetCurrentPosition(pos.ItemID)
	if err == nil && coordsEqual(current.Latitude, current.Longitude, pos.Latitude, pos.Longitude) {
		// Same location as current position - skip
		return nil
	}

	itemDir, err := s.resolveItemDir(pos.ItemID)
	if err != nil {
		return fmt.Errorf("resolve item directory: %w", err)
	}

	if err := mdstore.EnsureDir(itemDir); err != nil {
		return fmt.Errorf("create item directory: %w", err)
	}

	filename := positionFileName(pos)
	path := filepath.Join(itemDir, filename)

	return writePositionFile(path, pos)
}

// GetPosition retrieves a position by its UUID.
func (s *MarkdownStore) GetPosition(id uuid.UUID) (*models.Position, error) {
	items, err := s.readItems()
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		dir := s.itemDirPath(item.Name)
		positions, err := readAllPositionsInDir(dir)
		if err != nil {
			continue
		}
		for _, pos := range positions {
			if pos.ID == id {
				return pos, nil
			}
		}
	}

	return nil, ErrNotFound
}

// GetCurrentPosition returns the most recent position for an item.
func (s *MarkdownStore) GetCurrentPosition(itemID uuid.UUID) (*models.Position, error) {
	itemDir, err := s.resolveItemDir(itemID)
	if err != nil {
		return nil, err
	}

	positions, err := readAllPositionsInDir(itemDir)
	if err != nil {
		return nil, err
	}

	if len(positions) == 0 {
		return nil, ErrNotFound
	}

	// Sort by recorded_at descending, return newest
	sort.Slice(positions, func(i, j int) bool {
		return positions[i].RecordedAt.After(positions[j].RecordedAt)
	})

	return positions[0], nil
}

// GetTimeline returns all positions for an item, sorted by recorded_at descending (newest first).
func (s *MarkdownStore) GetTimeline(itemID uuid.UUID) ([]*models.Position, error) {
	itemDir, err := s.resolveItemDir(itemID)
	if err != nil {
		return nil, err
	}

	positions, err := readAllPositionsInDir(itemDir)
	if err != nil {
		return nil, err
	}

	// Sort by recorded_at descending
	sort.Slice(positions, func(i, j int) bool {
		return positions[i].RecordedAt.After(positions[j].RecordedAt)
	})

	return positions, nil
}

// GetPositionsSince returns positions for an item recorded after the given time.
func (s *MarkdownStore) GetPositionsSince(itemID uuid.UUID, since time.Time) ([]*models.Position, error) {
	itemDir, err := s.resolveItemDir(itemID)
	if err != nil {
		return nil, err
	}

	all, err := readAllPositionsInDir(itemDir)
	if err != nil {
		return nil, err
	}

	var filtered []*models.Position
	for _, pos := range all {
		if pos.RecordedAt.After(since) {
			filtered = append(filtered, pos)
		}
	}

	// Sort by recorded_at descending
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].RecordedAt.After(filtered[j].RecordedAt)
	})

	return filtered, nil
}

// GetPositionsInRange returns positions for an item within a time range.
func (s *MarkdownStore) GetPositionsInRange(itemID uuid.UUID, from, to time.Time) ([]*models.Position, error) {
	itemDir, err := s.resolveItemDir(itemID)
	if err != nil {
		return nil, err
	}

	all, err := readAllPositionsInDir(itemDir)
	if err != nil {
		return nil, err
	}

	var filtered []*models.Position
	for _, pos := range all {
		if !pos.RecordedAt.Before(from) && !pos.RecordedAt.After(to) {
			filtered = append(filtered, pos)
		}
	}

	// Sort by recorded_at descending
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].RecordedAt.After(filtered[j].RecordedAt)
	})

	return filtered, nil
}

// GetAllPositions returns all positions across all items.
func (s *MarkdownStore) GetAllPositions() ([]*models.Position, error) {
	items, err := s.readItems()
	if err != nil {
		return nil, err
	}

	var allPositions []*models.Position
	for _, item := range items {
		dir := s.itemDirPath(item.Name)
		positions, err := readAllPositionsInDir(dir)
		if err != nil {
			continue
		}
		allPositions = append(allPositions, positions...)
	}

	// Sort by recorded_at descending
	sort.Slice(allPositions, func(i, j int) bool {
		return allPositions[i].RecordedAt.After(allPositions[j].RecordedAt)
	})

	return allPositions, nil
}

// GetAllPositionsSince returns all positions across all items after the given time.
func (s *MarkdownStore) GetAllPositionsSince(since time.Time) ([]*models.Position, error) {
	all, err := s.GetAllPositions()
	if err != nil {
		return nil, err
	}

	var filtered []*models.Position
	for _, pos := range all {
		if pos.RecordedAt.After(since) {
			filtered = append(filtered, pos)
		}
	}

	return filtered, nil
}

// GetAllPositionsInRange returns all positions across all items within a time range.
func (s *MarkdownStore) GetAllPositionsInRange(from, to time.Time) ([]*models.Position, error) {
	all, err := s.GetAllPositions()
	if err != nil {
		return nil, err
	}

	var filtered []*models.Position
	for _, pos := range all {
		if !pos.RecordedAt.Before(from) && !pos.RecordedAt.After(to) {
			filtered = append(filtered, pos)
		}
	}

	return filtered, nil
}

// DeletePosition removes a single position.
func (s *MarkdownStore) DeletePosition(id uuid.UUID) error {
	items, err := s.readItems()
	if err != nil {
		return err
	}

	for _, item := range items {
		dir := s.itemDirPath(item.Name)
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
				continue
			}
			path := filepath.Join(dir, entry.Name())
			pos, err := readPositionFile(path)
			if err != nil {
				continue
			}
			if pos.ID == id {
				return os.Remove(path)
			}
		}
	}

	return nil
}
