// ABOUTME: Database operations for items
// ABOUTME: Provides CRUD functions for the items table

package db

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/harper/position/internal/models"
)

// CreateItem inserts a new item into the database.
func CreateItem(db *sql.DB, item *models.Item) error {
	_, err := db.Exec(
		"INSERT INTO items (id, name, created_at) VALUES (?, ?, ?)",
		item.ID.String(), item.Name, item.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert item: %w", err)
	}
	return nil
}

// GetItemByID retrieves an item by its UUID.
func GetItemByID(db *sql.DB, id uuid.UUID) (*models.Item, error) {
	row := db.QueryRow("SELECT id, name, created_at FROM items WHERE id = ?", id.String())
	return scanItem(row)
}

// GetItemByName retrieves an item by its name.
func GetItemByName(db *sql.DB, name string) (*models.Item, error) {
	row := db.QueryRow("SELECT id, name, created_at FROM items WHERE name = ?", name)
	return scanItem(row)
}

// ListItems retrieves all items sorted by name.
func ListItems(db *sql.DB) ([]*models.Item, error) {
	rows, err := db.Query("SELECT id, name, created_at FROM items ORDER BY name")
	if err != nil {
		return nil, fmt.Errorf("failed to query items: %w", err)
	}
	defer rows.Close()

	var items []*models.Item
	for rows.Next() {
		item, err := scanItemFromRows(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// DeleteItem removes an item and all its positions (via CASCADE).
func DeleteItem(db *sql.DB, id uuid.UUID) error {
	result, err := db.Exec("DELETE FROM items WHERE id = ?", id.String())
	if err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("item not found")
	}
	return nil
}

func scanItem(row *sql.Row) (*models.Item, error) {
	var item models.Item
	var idStr string
	err := row.Scan(&idStr, &item.Name, &item.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to scan item: %w", err)
	}
	item.ID, _ = uuid.Parse(idStr)
	return &item, nil
}

func scanItemFromRows(rows *sql.Rows) (*models.Item, error) {
	var item models.Item
	var idStr string
	err := rows.Scan(&idStr, &item.Name, &item.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to scan item: %w", err)
	}
	item.ID, _ = uuid.Parse(idStr)
	return &item, nil
}
