// ABOUTME: Unit tests for item database operations
// ABOUTME: Tests CRUD operations for items table

package db

import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/harper/position/internal/models"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestCreateItem(t *testing.T) {
	db := setupTestDB(t)
	item := models.NewItem("harper")

	err := CreateItem(db, item)
	if err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	// Verify it exists
	found, err := GetItemByName(db, "harper")
	if err != nil {
		t.Fatalf("failed to get item: %v", err)
	}
	if found.ID != item.ID {
		t.Error("ID mismatch")
	}
}

func TestGetItemByName_NotFound(t *testing.T) {
	db := setupTestDB(t)

	_, err := GetItemByName(db, "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent item")
	}
}

func TestListItems(t *testing.T) {
	db := setupTestDB(t)

	_ = CreateItem(db, models.NewItem("harper"))
	_ = CreateItem(db, models.NewItem("hiromi"))

	items, err := ListItems(db)
	if err != nil {
		t.Fatalf("failed to list items: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
}

func TestDeleteItem(t *testing.T) {
	db := setupTestDB(t)
	item := models.NewItem("harper")
	_ = CreateItem(db, item)

	err := DeleteItem(db, item.ID)
	if err != nil {
		t.Fatalf("failed to delete item: %v", err)
	}

	_, err = GetItemByName(db, "harper")
	if err == nil {
		t.Error("item should be deleted")
	}
}
