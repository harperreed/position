// ABOUTME: Core data models for items and positions
// ABOUTME: Provides constructor functions for creating new entities

package models

import (
	"time"

	"github.com/google/uuid"
)

// Item represents something being tracked (person, car, etc.).
type Item struct {
	ID        uuid.UUID
	Name      string
	CreatedAt time.Time
}

// Position represents a location entry for an item.
type Position struct {
	ID         uuid.UUID
	ItemID     uuid.UUID
	Latitude   float64
	Longitude  float64
	Label      *string
	RecordedAt time.Time
	CreatedAt  time.Time
}

// NewItem creates a new item with generated UUID and timestamp.
func NewItem(name string) *Item {
	return &Item{
		ID:        uuid.New(),
		Name:      name,
		CreatedAt: time.Now(),
	}
}

// NewPosition creates a new position with generated UUID and current timestamps.
func NewPosition(itemID uuid.UUID, lat, lng float64, label *string) *Position {
	now := time.Now()
	return &Position{
		ID:         uuid.New(),
		ItemID:     itemID,
		Latitude:   lat,
		Longitude:  lng,
		Label:      label,
		RecordedAt: now,
		CreatedAt:  now,
	}
}

// NewPositionWithRecordedAt creates a position with a specific recorded time.
func NewPositionWithRecordedAt(itemID uuid.UUID, lat, lng float64, label *string, recordedAt time.Time) *Position {
	return &Position{
		ID:         uuid.New(),
		ItemID:     itemID,
		Latitude:   lat,
		Longitude:  lng,
		Label:      label,
		RecordedAt: recordedAt,
		CreatedAt:  time.Now(),
	}
}
