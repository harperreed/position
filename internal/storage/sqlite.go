// ABOUTME: SQLite storage implementation for position data
// ABOUTME: Provides local-only persistence using pure Go SQLite driver

package storage

import (
	"database/sql"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/harper/position/internal/models"
	_ "modernc.org/sqlite"
)

// coordEpsilon defines the threshold for considering coordinates equal.
// 0.0000001 degrees is roughly 1.1cm at the equator, sufficient for GPS deduplication.
const coordEpsilon = 0.0000001

// SQLiteDB implements Repository with a local SQLite database.
type SQLiteDB struct {
	db   *sql.DB
	path string
}

// Compile-time check that SQLiteDB implements Repository.
var _ Repository = (*SQLiteDB)(nil)

// DefaultDBPath returns the default database path.
func DefaultDBPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".local", "share", "position", "position.db")
}

// NewSQLiteDB creates a new SQLite database at the given path.
// Creates the directory and database file if they don't exist.
func NewSQLiteDB(path string) (*SQLiteDB, error) {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0750); err != nil { //nolint:gosec // 0750 is appropriate for user data directory
		return nil, fmt.Errorf("create directory: %w", err)
	}

	db, err := sql.Open("sqlite", path+"?_pragma=foreign_keys(1)")
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	s := &SQLiteDB{db: db, path: path}

	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return s, nil
}

// migrate creates or updates the database schema.
func (s *SQLiteDB) migrate() error {
	schema := `
		CREATE TABLE IF NOT EXISTS items (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS positions (
			id TEXT PRIMARY KEY,
			item_id TEXT NOT NULL REFERENCES items(id) ON DELETE CASCADE,
			latitude REAL NOT NULL,
			longitude REAL NOT NULL,
			label TEXT,
			recorded_at DATETIME NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS idx_positions_item_id ON positions(item_id);
		CREATE INDEX IF NOT EXISTS idx_positions_recorded_at ON positions(recorded_at);
	`
	_, err := s.db.Exec(schema)
	return err
}

// Close closes the database connection.
func (s *SQLiteDB) Close() error {
	return s.db.Close()
}

// Sync is a no-op for local SQLite (no cloud sync).
func (s *SQLiteDB) Sync() error {
	return nil
}

// Reset clears all data from the database.
func (s *SQLiteDB) Reset() error {
	_, err := s.db.Exec("DELETE FROM positions; DELETE FROM items;")
	return err
}

// CreateItem creates a new item.
func (s *SQLiteDB) CreateItem(item *models.Item) error {
	_, err := s.db.Exec(
		"INSERT INTO items (id, name, created_at) VALUES (?, ?, ?)",
		item.ID.String(), item.Name, item.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert item: %w", err)
	}
	return nil
}

// GetItemByID retrieves an item by its UUID.
func (s *SQLiteDB) GetItemByID(id uuid.UUID) (*models.Item, error) {
	row := s.db.QueryRow(
		"SELECT id, name, created_at FROM items WHERE id = ?",
		id.String(),
	)
	return s.scanItem(row)
}

// GetItemByName retrieves an item by its name.
func (s *SQLiteDB) GetItemByName(name string) (*models.Item, error) {
	row := s.db.QueryRow(
		"SELECT id, name, created_at FROM items WHERE name = ?",
		name,
	)
	return s.scanItem(row)
}

// ListItems returns all items sorted by name.
func (s *SQLiteDB) ListItems() ([]*models.Item, error) {
	rows, err := s.db.Query("SELECT id, name, created_at FROM items ORDER BY name")
	if err != nil {
		return nil, fmt.Errorf("query items: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var items []*models.Item
	for rows.Next() {
		item, err := s.scanItemFromRows(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// DeleteItem removes an item (positions cascade delete automatically).
func (s *SQLiteDB) DeleteItem(id uuid.UUID) error {
	_, err := s.db.Exec("DELETE FROM items WHERE id = ?", id.String())
	return err
}

func (s *SQLiteDB) scanItem(row *sql.Row) (*models.Item, error) {
	var idStr string
	var item models.Item
	err := row.Scan(&idStr, &item.Name, &item.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan item: %w", err)
	}
	item.ID, _ = uuid.Parse(idStr)
	return &item, nil
}

func (s *SQLiteDB) scanItemFromRows(rows *sql.Rows) (*models.Item, error) {
	var idStr string
	var item models.Item
	err := rows.Scan(&idStr, &item.Name, &item.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("scan item: %w", err)
	}
	item.ID, _ = uuid.Parse(idStr)
	return &item, nil
}

// CreatePosition creates a new position with deduplication.
// If the new position matches the current position for the item, it's silently skipped.
func (s *SQLiteDB) CreatePosition(pos *models.Position) error {
	// Check for duplicate against current position
	current, err := s.GetCurrentPosition(pos.ItemID)
	if err == nil && coordsEqual(current.Latitude, current.Longitude, pos.Latitude, pos.Longitude) {
		// Same location as current position - skip
		return nil
	}

	_, err = s.db.Exec(
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

// coordsEqual compares two coordinate pairs using epsilon for floating-point safety.
func coordsEqual(lat1, lng1, lat2, lng2 float64) bool {
	return math.Abs(lat1-lat2) < coordEpsilon && math.Abs(lng1-lng2) < coordEpsilon
}

// GetPosition retrieves a position by its UUID.
func (s *SQLiteDB) GetPosition(id uuid.UUID) (*models.Position, error) {
	row := s.db.QueryRow(
		`SELECT id, item_id, latitude, longitude, label, recorded_at, created_at
		 FROM positions WHERE id = ?`,
		id.String(),
	)
	return s.scanPosition(row)
}

// GetCurrentPosition returns the most recent position for an item.
func (s *SQLiteDB) GetCurrentPosition(itemID uuid.UUID) (*models.Position, error) {
	row := s.db.QueryRow(
		`SELECT id, item_id, latitude, longitude, label, recorded_at, created_at
		 FROM positions WHERE item_id = ? ORDER BY recorded_at DESC LIMIT 1`,
		itemID.String(),
	)
	return s.scanPosition(row)
}

// GetTimeline returns all positions for an item, sorted by recorded_at descending (newest first).
func (s *SQLiteDB) GetTimeline(itemID uuid.UUID) ([]*models.Position, error) {
	rows, err := s.db.Query(
		`SELECT id, item_id, latitude, longitude, label, recorded_at, created_at
		 FROM positions WHERE item_id = ? ORDER BY recorded_at DESC`,
		itemID.String(),
	)
	if err != nil {
		return nil, fmt.Errorf("query positions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return s.scanPositions(rows)
}

// GetPositionsSince returns positions for an item recorded after the given time.
func (s *SQLiteDB) GetPositionsSince(itemID uuid.UUID, since time.Time) ([]*models.Position, error) {
	rows, err := s.db.Query(
		`SELECT id, item_id, latitude, longitude, label, recorded_at, created_at
		 FROM positions WHERE item_id = ? AND recorded_at > ? ORDER BY recorded_at DESC`,
		itemID.String(), since,
	)
	if err != nil {
		return nil, fmt.Errorf("query positions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return s.scanPositions(rows)
}

// GetPositionsInRange returns positions for an item within a time range.
func (s *SQLiteDB) GetPositionsInRange(itemID uuid.UUID, from, to time.Time) ([]*models.Position, error) {
	rows, err := s.db.Query(
		`SELECT id, item_id, latitude, longitude, label, recorded_at, created_at
		 FROM positions WHERE item_id = ? AND recorded_at >= ? AND recorded_at <= ?
		 ORDER BY recorded_at DESC`,
		itemID.String(), from, to,
	)
	if err != nil {
		return nil, fmt.Errorf("query positions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return s.scanPositions(rows)
}

// GetAllPositions returns all positions across all items.
func (s *SQLiteDB) GetAllPositions() ([]*models.Position, error) {
	rows, err := s.db.Query(
		`SELECT id, item_id, latitude, longitude, label, recorded_at, created_at
		 FROM positions ORDER BY recorded_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("query positions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return s.scanPositions(rows)
}

// GetAllPositionsSince returns all positions across all items after the given time.
func (s *SQLiteDB) GetAllPositionsSince(since time.Time) ([]*models.Position, error) {
	rows, err := s.db.Query(
		`SELECT id, item_id, latitude, longitude, label, recorded_at, created_at
		 FROM positions WHERE recorded_at > ? ORDER BY recorded_at DESC`,
		since,
	)
	if err != nil {
		return nil, fmt.Errorf("query positions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return s.scanPositions(rows)
}

// GetAllPositionsInRange returns all positions across all items within a time range.
func (s *SQLiteDB) GetAllPositionsInRange(from, to time.Time) ([]*models.Position, error) {
	rows, err := s.db.Query(
		`SELECT id, item_id, latitude, longitude, label, recorded_at, created_at
		 FROM positions WHERE recorded_at >= ? AND recorded_at <= ? ORDER BY recorded_at DESC`,
		from, to,
	)
	if err != nil {
		return nil, fmt.Errorf("query positions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return s.scanPositions(rows)
}

// DeletePosition removes a single position.
func (s *SQLiteDB) DeletePosition(id uuid.UUID) error {
	_, err := s.db.Exec("DELETE FROM positions WHERE id = ?", id.String())
	return err
}

func (s *SQLiteDB) scanPosition(row *sql.Row) (*models.Position, error) {
	var idStr, itemIDStr string
	var pos models.Position
	err := row.Scan(&idStr, &itemIDStr, &pos.Latitude, &pos.Longitude,
		&pos.Label, &pos.RecordedAt, &pos.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan position: %w", err)
	}
	pos.ID, _ = uuid.Parse(idStr)
	pos.ItemID, _ = uuid.Parse(itemIDStr)
	return &pos, nil
}

func (s *SQLiteDB) scanPositions(rows *sql.Rows) ([]*models.Position, error) {
	var positions []*models.Position
	for rows.Next() {
		var idStr, itemIDStr string
		var pos models.Position
		err := rows.Scan(&idStr, &itemIDStr, &pos.Latitude, &pos.Longitude,
			&pos.Label, &pos.RecordedAt, &pos.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan position: %w", err)
		}
		pos.ID, _ = uuid.Parse(idStr)
		pos.ItemID, _ = uuid.Parse(itemIDStr)
		positions = append(positions, &pos)
	}
	return positions, rows.Err()
}
