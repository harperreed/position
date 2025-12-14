// ABOUTME: Database schema and migrations
// ABOUTME: Defines tables for items and positions

package db

import (
	"database/sql"
	"fmt"
)

const schema = `
CREATE TABLE IF NOT EXISTS items (
	id TEXT PRIMARY KEY,
	name TEXT UNIQUE NOT NULL,
	created_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS positions (
	id TEXT PRIMARY KEY,
	item_id TEXT NOT NULL,
	latitude REAL NOT NULL,
	longitude REAL NOT NULL,
	label TEXT,
	recorded_at DATETIME NOT NULL,
	created_at DATETIME NOT NULL,
	FOREIGN KEY (item_id) REFERENCES items(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_positions_item_id ON positions(item_id);
CREATE INDEX IF NOT EXISTS idx_positions_recorded_at ON positions(recorded_at);
`

func runMigrations(db *sql.DB) error {
	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}
	return nil
}
