// ABOUTME: Unit tests for data models
// ABOUTME: Tests constructors and model methods

package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewItem(t *testing.T) {
	item := NewItem("harper")

	if item.Name != "harper" {
		t.Errorf("expected name 'harper', got '%s'", item.Name)
	}
	if item.ID == uuid.Nil {
		t.Error("expected non-nil UUID")
	}
	if item.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
}

func TestNewPosition(t *testing.T) {
	itemID := uuid.New()
	lat := 41.8781
	lng := -87.6298
	label := "chicago"

	pos := NewPosition(itemID, lat, lng, &label)

	if pos.ItemID != itemID {
		t.Error("item ID mismatch")
	}
	if pos.Latitude != lat {
		t.Errorf("expected lat %f, got %f", lat, pos.Latitude)
	}
	if pos.Longitude != lng {
		t.Errorf("expected lng %f, got %f", lng, pos.Longitude)
	}
	if pos.Label == nil || *pos.Label != label {
		t.Error("expected label 'chicago'")
	}
	if pos.ID == uuid.Nil {
		t.Error("expected non-nil UUID")
	}
}

func TestNewPositionWithRecordedAt(t *testing.T) {
	itemID := uuid.New()
	recordedAt := time.Date(2024, 12, 14, 15, 0, 0, 0, time.UTC)

	pos := NewPositionWithRecordedAt(itemID, 41.8781, -87.6298, nil, recordedAt)

	if !pos.RecordedAt.Equal(recordedAt) {
		t.Errorf("expected recordedAt %v, got %v", recordedAt, pos.RecordedAt)
	}
}
