// ABOUTME: Unit tests for position database operations
// ABOUTME: Tests CRUD and query operations for positions table

package db

import (
	"testing"
	"time"

	"github.com/harper/position/internal/models"
)

func TestCreatePosition(t *testing.T) {
	db := setupTestDB(t)
	item := models.NewItem("harper")
	_ = CreateItem(db, item)

	label := "chicago"
	pos := models.NewPosition(item.ID, 41.8781, -87.6298, &label)

	err := CreatePosition(db, pos)
	if err != nil {
		t.Fatalf("failed to create position: %v", err)
	}
}

func TestGetCurrentPosition(t *testing.T) {
	db := setupTestDB(t)
	item := models.NewItem("harper")
	_ = CreateItem(db, item)

	// Add older position
	label1 := "boston"
	pos1 := models.NewPositionWithRecordedAt(item.ID, 42.3601, -71.0589, &label1,
		time.Now().Add(-1*time.Hour))
	_ = CreatePosition(db, pos1)

	// Add newer position
	label2 := "chicago"
	pos2 := models.NewPosition(item.ID, 41.8781, -87.6298, &label2)
	_ = CreatePosition(db, pos2)

	current, err := GetCurrentPosition(db, item.ID)
	if err != nil {
		t.Fatalf("failed to get current position: %v", err)
	}
	if current.Label == nil || *current.Label != "chicago" {
		t.Error("expected most recent position (chicago)")
	}
}

func TestGetTimeline(t *testing.T) {
	db := setupTestDB(t)
	item := models.NewItem("harper")
	_ = CreateItem(db, item)

	// Add positions at different times
	for i := 0; i < 3; i++ {
		pos := models.NewPositionWithRecordedAt(item.ID, float64(i), float64(i), nil,
			time.Now().Add(time.Duration(-i)*time.Hour))
		_ = CreatePosition(db, pos)
	}

	timeline, err := GetTimeline(db, item.ID)
	if err != nil {
		t.Fatalf("failed to get timeline: %v", err)
	}
	if len(timeline) != 3 {
		t.Errorf("expected 3 positions, got %d", len(timeline))
	}
	// Should be sorted newest first
	if timeline[0].Latitude != 0 {
		t.Error("expected newest position first")
	}
}

func TestDeletePositionsForItem(t *testing.T) {
	db := setupTestDB(t)
	item := models.NewItem("harper")
	_ = CreateItem(db, item)

	pos := models.NewPosition(item.ID, 41.8781, -87.6298, nil)
	_ = CreatePosition(db, pos)

	// Delete item should cascade delete positions
	_ = DeleteItem(db, item.ID)

	// Positions should be gone (tested via item cascade)
	timeline, _ := GetTimeline(db, item.ID)
	if len(timeline) != 0 {
		t.Error("positions should be deleted with item")
	}
}
