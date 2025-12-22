// ABOUTME: Core data models for items and positions
// ABOUTME: Provides constructor functions for creating new entities

package models

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ValidateCoordinates checks if latitude and longitude are within valid ranges.
func ValidateCoordinates(lat, lng float64) error {
	if math.IsNaN(lat) || math.IsNaN(lng) {
		return fmt.Errorf("coordinates cannot be NaN")
	}
	if math.IsInf(lat, 0) || math.IsInf(lng, 0) {
		return fmt.Errorf("coordinates cannot be infinite")
	}
	if lat < -90 || lat > 90 {
		return fmt.Errorf("latitude must be between -90 and 90")
	}
	if lng < -180 || lng > 180 {
		return fmt.Errorf("longitude must be between -180 and 180")
	}
	return nil
}

// ValidateName checks if a name is valid (non-empty, within length limits).
// Note: This validates the raw input - callers should trim whitespace themselves if needed.
func ValidateName(name string) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return fmt.Errorf("name cannot be empty or whitespace")
	}
	if len(name) > 255 {
		return fmt.Errorf("name too long (max 255 characters)")
	}
	return nil
}

// Item represents something being tracked (person, car, etc.).
type Item struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// Position represents a location entry for an item.
type Position struct {
	ID         uuid.UUID `json:"id"`
	ItemID     uuid.UUID `json:"item_id"`
	Latitude   float64   `json:"latitude"`
	Longitude  float64   `json:"longitude"`
	Label      *string   `json:"label,omitempty"`
	RecordedAt time.Time `json:"recorded_at"`
	CreatedAt  time.Time `json:"created_at"`
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
