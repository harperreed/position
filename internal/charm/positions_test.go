// ABOUTME: Tests for position CRUD operations
// ABOUTME: Verifies deduplication behavior for repeated positions

package charm

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/harper/position/internal/models"
)

func TestCreatePosition_DeduplicatesSameLocation(t *testing.T) {
	// Setup test environment
	tmpDir := t.TempDir()
	t.Setenv("CHARM_DATA_DIR", tmpDir)

	client, err := NewTestClient("test-dedup")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	// Create an item
	item := models.NewItem("test-item")
	if err := client.CreateItem(item); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	// Add first position
	pos1 := models.NewPosition(item.ID, 41.8781, -87.6298, nil)
	if err := client.CreatePosition(pos1); err != nil {
		t.Fatalf("failed to create first position: %v", err)
	}

	// Try to add same location again
	pos2 := models.NewPosition(item.ID, 41.8781, -87.6298, nil)
	if err := client.CreatePosition(pos2); err != nil {
		t.Fatalf("failed to create second position: %v", err)
	}

	// Get timeline - should only have ONE entry (deduplicated)
	timeline, err := client.GetTimeline(item.ID)
	if err != nil {
		t.Fatalf("failed to get timeline: %v", err)
	}

	if len(timeline) != 1 {
		t.Errorf("expected 1 position after deduplication, got %d", len(timeline))
	}
}

func TestCreatePosition_StoresDifferentLocation(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("CHARM_DATA_DIR", tmpDir)

	client, err := NewTestClient("test-positions")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	item := models.NewItem("test-item")
	if err := client.CreateItem(item); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	// Add first position (Chicago)
	pos1 := models.NewPosition(item.ID, 41.8781, -87.6298, nil)
	if err := client.CreatePosition(pos1); err != nil {
		t.Fatalf("failed to create first position: %v", err)
	}

	// Add second position (New York - different location)
	pos2 := models.NewPosition(item.ID, 40.7128, -74.0060, nil)
	if err := client.CreatePosition(pos2); err != nil {
		t.Fatalf("failed to create second position: %v", err)
	}

	// Get timeline - should have TWO entries
	timeline, err := client.GetTimeline(item.ID)
	if err != nil {
		t.Fatalf("failed to get timeline: %v", err)
	}

	if len(timeline) != 2 {
		t.Errorf("expected 2 positions for different locations, got %d", len(timeline))
	}
}

func TestCreatePosition_StoresFirstPosition(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("CHARM_DATA_DIR", tmpDir)

	client, err := NewTestClient("test-positions")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	item := models.NewItem("test-item")
	if err := client.CreateItem(item); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	// Add first position - should always work
	pos := models.NewPosition(item.ID, 41.8781, -87.6298, nil)
	if err := client.CreatePosition(pos); err != nil {
		t.Fatalf("failed to create position: %v", err)
	}

	timeline, err := client.GetTimeline(item.ID)
	if err != nil {
		t.Fatalf("failed to get timeline: %v", err)
	}

	if len(timeline) != 1 {
		t.Errorf("expected 1 position, got %d", len(timeline))
	}
}

func TestCreatePosition_IgnoresLabelForDeduplication(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("CHARM_DATA_DIR", tmpDir)

	client, err := NewTestClient("test-positions")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	item := models.NewItem("test-item")
	if err := client.CreateItem(item); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	// Add position with label "home"
	label1 := "home"
	pos1 := models.NewPosition(item.ID, 41.8781, -87.6298, &label1)
	if err := client.CreatePosition(pos1); err != nil {
		t.Fatalf("failed to create first position: %v", err)
	}

	// Try to add same location with different label "apartment"
	label2 := "apartment"
	pos2 := models.NewPosition(item.ID, 41.8781, -87.6298, &label2)
	if err := client.CreatePosition(pos2); err != nil {
		t.Fatalf("failed to create second position: %v", err)
	}

	// Should deduplicate - label difference doesn't matter
	timeline, err := client.GetTimeline(item.ID)
	if err != nil {
		t.Fatalf("failed to get timeline: %v", err)
	}

	if len(timeline) != 1 {
		t.Errorf("expected 1 position (labels ignored for dedup), got %d", len(timeline))
	}
}

func TestCreatePosition_DifferentItemsSameLocation(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("CHARM_DATA_DIR", tmpDir)

	client, err := NewTestClient("test-positions")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	// Create two different items
	item1 := models.NewItem("harper")
	item2 := models.NewItem("car")
	if err := client.CreateItem(item1); err != nil {
		t.Fatalf("failed to create item1: %v", err)
	}
	if err := client.CreateItem(item2); err != nil {
		t.Fatalf("failed to create item2: %v", err)
	}

	// Add same location for both items
	pos1 := models.NewPosition(item1.ID, 41.8781, -87.6298, nil)
	pos2 := models.NewPosition(item2.ID, 41.8781, -87.6298, nil)

	if err := client.CreatePosition(pos1); err != nil {
		t.Fatalf("failed to create pos1: %v", err)
	}
	if err := client.CreatePosition(pos2); err != nil {
		t.Fatalf("failed to create pos2: %v", err)
	}

	// Each item should have its own position (not deduplicated across items)
	timeline1, _ := client.GetTimeline(item1.ID)
	timeline2, _ := client.GetTimeline(item2.ID)

	if len(timeline1) != 1 || len(timeline2) != 1 {
		t.Errorf("expected 1 position per item, got %d and %d", len(timeline1), len(timeline2))
	}
}

func TestCreatePosition_SmallLocationChange(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("CHARM_DATA_DIR", tmpDir)

	client, err := NewTestClient("test-positions")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	item := models.NewItem("test-item")
	if err := client.CreateItem(item); err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	// Add first position
	pos1 := models.NewPosition(item.ID, 41.8781, -87.6298, nil)
	if err := client.CreatePosition(pos1); err != nil {
		t.Fatalf("failed to create first position: %v", err)
	}

	// Add position with very small difference (sub-meter, ~0.00001 = ~1m)
	// This SHOULD be stored as a different position
	pos2 := models.NewPosition(item.ID, 41.8782, -87.6298, nil)
	if err := client.CreatePosition(pos2); err != nil {
		t.Fatalf("failed to create second position: %v", err)
	}

	timeline, err := client.GetTimeline(item.ID)
	if err != nil {
		t.Fatalf("failed to get timeline: %v", err)
	}

	// Small but non-zero difference should still be stored
	if len(timeline) != 2 {
		t.Errorf("expected 2 positions for small location change, got %d", len(timeline))
	}
}

func TestCreatePosition_NewItemHasNoCurrentPosition(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("CHARM_DATA_DIR", tmpDir)

	client, err := NewTestClient("test-positions")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	// Generate a random UUID for an item that doesn't exist
	fakeItemID := uuid.New()

	// Verify no current position exists
	_, err = client.GetCurrentPosition(fakeItemID)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound for non-existent item, got %v", err)
	}
}
