// ABOUTME: Database operations for positions
// ABOUTME: Provides CRUD and query functions for the positions table

package db

import (
	"database/sql"
	"fmt"
	"time"

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
	defer func() { _ = rows.Close() }()

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

// GetPositionsSince retrieves positions for an item since a given time, oldest first.
func GetPositionsSince(db *sql.DB, itemID uuid.UUID, since time.Time) ([]*models.Position, error) {
	rows, err := db.Query(
		`SELECT id, item_id, latitude, longitude, label, recorded_at, created_at
		 FROM positions WHERE item_id = ? AND recorded_at >= ? ORDER BY recorded_at ASC`,
		itemID.String(), since,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query positions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return scanPositions(rows)
}

// GetPositionsInRange retrieves positions for an item within a time range, oldest first.
func GetPositionsInRange(db *sql.DB, itemID uuid.UUID, from, to time.Time) ([]*models.Position, error) {
	rows, err := db.Query(
		`SELECT id, item_id, latitude, longitude, label, recorded_at, created_at
		 FROM positions WHERE item_id = ? AND recorded_at >= ? AND recorded_at <= ? ORDER BY recorded_at ASC`,
		itemID.String(), from, to,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query positions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return scanPositions(rows)
}

// GetAllPositions retrieves all positions for all items, oldest first.
func GetAllPositions(db *sql.DB) ([]*models.Position, error) {
	rows, err := db.Query(
		`SELECT id, item_id, latitude, longitude, label, recorded_at, created_at
		 FROM positions ORDER BY recorded_at ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query positions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return scanPositions(rows)
}

// GetAllPositionsSince retrieves all positions since a given time, oldest first.
func GetAllPositionsSince(db *sql.DB, since time.Time) ([]*models.Position, error) {
	rows, err := db.Query(
		`SELECT id, item_id, latitude, longitude, label, recorded_at, created_at
		 FROM positions WHERE recorded_at >= ? ORDER BY recorded_at ASC`,
		since,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query positions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return scanPositions(rows)
}

// GetAllPositionsInRange retrieves all positions within a time range, oldest first.
func GetAllPositionsInRange(db *sql.DB, from, to time.Time) ([]*models.Position, error) {
	rows, err := db.Query(
		`SELECT id, item_id, latitude, longitude, label, recorded_at, created_at
		 FROM positions WHERE recorded_at >= ? AND recorded_at <= ? ORDER BY recorded_at ASC`,
		from, to,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query positions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return scanPositions(rows)
}

// scanPositions scans multiple position rows into a slice.
func scanPositions(rows *sql.Rows) ([]*models.Position, error) {
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
	pos.ID, err = uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse position ID: %w", err)
	}
	pos.ItemID, err = uuid.Parse(itemIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse item ID: %w", err)
	}
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
	pos.ID, err = uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse position ID: %w", err)
	}
	pos.ItemID, err = uuid.Parse(itemIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse item ID: %w", err)
	}
	return &pos, nil
}
