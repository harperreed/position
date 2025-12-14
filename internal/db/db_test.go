// ABOUTME: Unit tests for database initialization
// ABOUTME: Tests connection, migrations, and XDG path handling

package db

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitDB(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Verify tables exist
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='items'").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query: %v", err)
	}
	if count != 1 {
		t.Error("items table not created")
	}

	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='positions'").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query: %v", err)
	}
	if count != 1 {
		t.Error("positions table not created")
	}
}

func TestGetDefaultDBPath(t *testing.T) {
	path := GetDefaultDBPath()
	if !filepath.IsAbs(path) {
		t.Error("expected absolute path")
	}
	if !contains(path, "position") {
		t.Error("expected path to contain 'position'")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestInitDB_CreatesDirIfNotExists(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "subdir", "nested", "test.db")

	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	defer func() { _ = db.Close() }()

	if _, err := os.Stat(filepath.Dir(dbPath)); os.IsNotExist(err) {
		t.Error("directory was not created")
	}
}
