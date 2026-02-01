// ABOUTME: Tests for MCP server, tools, and resources
// ABOUTME: Verifies MCP integration with repository interface

package mcp

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/harper/position/internal/models"
	"github.com/harper/position/internal/storage"
)

// mockRepo implements storage.Repository for testing.
type mockRepo struct {
	items     map[uuid.UUID]*models.Item
	positions map[uuid.UUID]*models.Position

	createItemErr     error
	getItemByNameErr  error
	getItemByIDErr    error
	listItemsErr      error
	deleteItemErr     error
	createPositionErr error
	getPositionErr    error
	getCurrentPosErr  error
	getTimelineErr    error
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		items:     make(map[uuid.UUID]*models.Item),
		positions: make(map[uuid.UUID]*models.Position),
	}
}

func (m *mockRepo) CreateItem(item *models.Item) error {
	if m.createItemErr != nil {
		return m.createItemErr
	}
	// Check for duplicate names
	for _, existing := range m.items {
		if existing.Name == item.Name {
			return errors.New("duplicate name")
		}
	}
	m.items[item.ID] = item
	return nil
}

func (m *mockRepo) GetItemByID(id uuid.UUID) (*models.Item, error) {
	if m.getItemByIDErr != nil {
		return nil, m.getItemByIDErr
	}
	item, ok := m.items[id]
	if !ok {
		return nil, storage.ErrNotFound
	}
	return item, nil
}

func (m *mockRepo) GetItemByName(name string) (*models.Item, error) {
	if m.getItemByNameErr != nil {
		return nil, m.getItemByNameErr
	}
	for _, item := range m.items {
		if item.Name == name {
			return item, nil
		}
	}
	return nil, storage.ErrNotFound
}

func (m *mockRepo) ListItems() ([]*models.Item, error) {
	if m.listItemsErr != nil {
		return nil, m.listItemsErr
	}
	var items []*models.Item
	for _, item := range m.items {
		items = append(items, item)
	}
	return items, nil
}

func (m *mockRepo) DeleteItem(id uuid.UUID) error {
	if m.deleteItemErr != nil {
		return m.deleteItemErr
	}
	delete(m.items, id)
	// Also delete positions for item
	for pid, pos := range m.positions {
		if pos.ItemID == id {
			delete(m.positions, pid)
		}
	}
	return nil
}

func (m *mockRepo) CreatePosition(pos *models.Position) error {
	if m.createPositionErr != nil {
		return m.createPositionErr
	}
	m.positions[pos.ID] = pos
	return nil
}

func (m *mockRepo) GetPosition(id uuid.UUID) (*models.Position, error) {
	if m.getPositionErr != nil {
		return nil, m.getPositionErr
	}
	pos, ok := m.positions[id]
	if !ok {
		return nil, storage.ErrNotFound
	}
	return pos, nil
}

func (m *mockRepo) GetCurrentPosition(itemID uuid.UUID) (*models.Position, error) {
	if m.getCurrentPosErr != nil {
		return nil, m.getCurrentPosErr
	}
	var latest *models.Position
	for _, pos := range m.positions {
		if pos.ItemID == itemID {
			if latest == nil || pos.RecordedAt.After(latest.RecordedAt) {
				latest = pos
			}
		}
	}
	if latest == nil {
		return nil, storage.ErrNotFound
	}
	return latest, nil
}

func (m *mockRepo) GetTimeline(itemID uuid.UUID) ([]*models.Position, error) {
	if m.getTimelineErr != nil {
		return nil, m.getTimelineErr
	}
	var positions []*models.Position
	for _, pos := range m.positions {
		if pos.ItemID == itemID {
			positions = append(positions, pos)
		}
	}
	return positions, nil
}

func (m *mockRepo) GetPositionsSince(itemID uuid.UUID, since time.Time) ([]*models.Position, error) {
	var positions []*models.Position
	for _, pos := range m.positions {
		if pos.ItemID == itemID && pos.RecordedAt.After(since) {
			positions = append(positions, pos)
		}
	}
	return positions, nil
}

func (m *mockRepo) GetPositionsInRange(itemID uuid.UUID, from, to time.Time) ([]*models.Position, error) {
	var positions []*models.Position
	for _, pos := range m.positions {
		if pos.ItemID == itemID && pos.RecordedAt.After(from) && pos.RecordedAt.Before(to) {
			positions = append(positions, pos)
		}
	}
	return positions, nil
}

func (m *mockRepo) GetAllPositions() ([]*models.Position, error) {
	positions := make([]*models.Position, 0, len(m.positions))
	for _, pos := range m.positions {
		positions = append(positions, pos)
	}
	return positions, nil
}

func (m *mockRepo) GetAllPositionsSince(since time.Time) ([]*models.Position, error) {
	var positions []*models.Position
	for _, pos := range m.positions {
		if pos.RecordedAt.After(since) {
			positions = append(positions, pos)
		}
	}
	return positions, nil
}

func (m *mockRepo) GetAllPositionsInRange(from, to time.Time) ([]*models.Position, error) {
	var positions []*models.Position
	for _, pos := range m.positions {
		if pos.RecordedAt.After(from) && pos.RecordedAt.Before(to) {
			positions = append(positions, pos)
		}
	}
	return positions, nil
}

func (m *mockRepo) DeletePosition(id uuid.UUID) error {
	delete(m.positions, id)
	return nil
}

func (m *mockRepo) Sync() error {
	return nil
}

func (m *mockRepo) Reset() error {
	m.items = make(map[uuid.UUID]*models.Item)
	m.positions = make(map[uuid.UUID]*models.Position)
	return nil
}

func (m *mockRepo) Close() error {
	return nil
}

// Tests

func TestNewServer(t *testing.T) {
	repo := newMockRepo()
	server, err := NewServer(repo)
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}
	if server == nil {
		t.Fatal("expected non-nil server")
	}
	if server.repo == nil {
		t.Error("expected non-nil repo")
	}
	if server.mcp == nil {
		t.Error("expected non-nil mcp server")
	}
}

func TestNewServer_NilRepo(t *testing.T) {
	_, err := NewServer(nil)
	if err == nil {
		t.Error("expected error for nil repo")
	}
}

func TestHandleAddPosition_NewItem(t *testing.T) {
	repo := newMockRepo()
	server, _ := NewServer(repo)

	input := AddPositionInput{
		Name:      "harper",
		Latitude:  41.8781,
		Longitude: -87.6298,
	}

	result, output, err := server.handleAddPosition(context.Background(), nil, input)
	if err != nil {
		t.Fatalf("handleAddPosition failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if output.ItemName != "harper" {
		t.Errorf("expected item name 'harper', got %q", output.ItemName)
	}
	if output.Latitude != 41.8781 {
		t.Errorf("expected latitude 41.8781, got %f", output.Latitude)
	}

	// Verify item was created
	items, _ := repo.ListItems()
	if len(items) != 1 {
		t.Errorf("expected 1 item, got %d", len(items))
	}
}

func TestHandleAddPosition_ExistingItem(t *testing.T) {
	repo := newMockRepo()
	item := models.NewItem("harper")
	_ = repo.CreateItem(item)

	server, _ := NewServer(repo)

	input := AddPositionInput{
		Name:      "harper",
		Latitude:  41.8781,
		Longitude: -87.6298,
	}

	_, _, err := server.handleAddPosition(context.Background(), nil, input)
	if err != nil {
		t.Fatalf("handleAddPosition failed: %v", err)
	}

	// Should still have just 1 item
	items, _ := repo.ListItems()
	if len(items) != 1 {
		t.Errorf("expected 1 item, got %d", len(items))
	}
}

func TestHandleAddPosition_WithLabel(t *testing.T) {
	repo := newMockRepo()
	server, _ := NewServer(repo)

	label := "chicago"
	input := AddPositionInput{
		Name:      "harper",
		Latitude:  41.8781,
		Longitude: -87.6298,
		Label:     &label,
	}

	_, output, err := server.handleAddPosition(context.Background(), nil, input)
	if err != nil {
		t.Fatalf("handleAddPosition failed: %v", err)
	}
	if output.Label == nil || *output.Label != "chicago" {
		t.Error("expected label 'chicago'")
	}
}

func TestHandleAddPosition_WithTimestamp(t *testing.T) {
	repo := newMockRepo()
	server, _ := NewServer(repo)

	at := "2024-12-15T10:00:00Z"
	input := AddPositionInput{
		Name:      "harper",
		Latitude:  41.8781,
		Longitude: -87.6298,
		At:        &at,
	}

	_, output, err := server.handleAddPosition(context.Background(), nil, input)
	if err != nil {
		t.Fatalf("handleAddPosition failed: %v", err)
	}

	expected, _ := time.Parse(time.RFC3339, at)
	if !output.RecordedAt.Equal(expected) {
		t.Errorf("expected recorded_at %v, got %v", expected, output.RecordedAt)
	}
}

func TestHandleAddPosition_InvalidTimestamp(t *testing.T) {
	repo := newMockRepo()
	server, _ := NewServer(repo)

	at := "not-a-timestamp"
	input := AddPositionInput{
		Name:      "harper",
		Latitude:  41.8781,
		Longitude: -87.6298,
		At:        &at,
	}

	_, _, err := server.handleAddPosition(context.Background(), nil, input)
	if err == nil {
		t.Error("expected error for invalid timestamp")
	}
}

func TestHandleAddPosition_InvalidCoordinates(t *testing.T) {
	repo := newMockRepo()
	server, _ := NewServer(repo)

	input := AddPositionInput{
		Name:      "harper",
		Latitude:  100, // Invalid
		Longitude: -87.6298,
	}

	_, _, err := server.handleAddPosition(context.Background(), nil, input)
	if err == nil {
		t.Error("expected error for invalid coordinates")
	}
}

func TestHandleAddPosition_InvalidName(t *testing.T) {
	repo := newMockRepo()
	server, _ := NewServer(repo)

	input := AddPositionInput{
		Name:      "", // Invalid
		Latitude:  41.8781,
		Longitude: -87.6298,
	}

	_, _, err := server.handleAddPosition(context.Background(), nil, input)
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestHandleAddPosition_CreateItemError(t *testing.T) {
	repo := newMockRepo()
	repo.createItemErr = errors.New("database error")
	server, _ := NewServer(repo)

	input := AddPositionInput{
		Name:      "harper",
		Latitude:  41.8781,
		Longitude: -87.6298,
	}

	_, _, err := server.handleAddPosition(context.Background(), nil, input)
	if err == nil {
		t.Error("expected error when create item fails")
	}
}

func TestHandleAddPosition_GetItemError(t *testing.T) {
	repo := newMockRepo()
	repo.getItemByNameErr = errors.New("database error")
	server, _ := NewServer(repo)

	input := AddPositionInput{
		Name:      "harper",
		Latitude:  41.8781,
		Longitude: -87.6298,
	}

	_, _, err := server.handleAddPosition(context.Background(), nil, input)
	if err == nil {
		t.Error("expected error when get item fails")
	}
}

func TestHandleAddPosition_CreatePositionError(t *testing.T) {
	repo := newMockRepo()
	repo.createPositionErr = errors.New("database error")
	server, _ := NewServer(repo)

	input := AddPositionInput{
		Name:      "harper",
		Latitude:  41.8781,
		Longitude: -87.6298,
	}

	_, _, err := server.handleAddPosition(context.Background(), nil, input)
	if err == nil {
		t.Error("expected error when create position fails")
	}
}

func TestHandleGetCurrent(t *testing.T) {
	repo := newMockRepo()
	item := models.NewItem("harper")
	_ = repo.CreateItem(item)
	pos := models.NewPosition(item.ID, 41.8781, -87.6298, nil)
	_ = repo.CreatePosition(pos)

	server, _ := NewServer(repo)

	input := GetCurrentInput{Name: "harper"}
	result, output, err := server.handleGetCurrent(context.Background(), nil, input)
	if err != nil {
		t.Fatalf("handleGetCurrent failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if output.Latitude != 41.8781 {
		t.Errorf("expected latitude 41.8781, got %f", output.Latitude)
	}
}

func TestHandleGetCurrent_ItemNotFound(t *testing.T) {
	repo := newMockRepo()
	server, _ := NewServer(repo)

	input := GetCurrentInput{Name: "nonexistent"}
	_, _, err := server.handleGetCurrent(context.Background(), nil, input)
	if err == nil {
		t.Error("expected error for nonexistent item")
	}
}

func TestHandleGetCurrent_NoPosition(t *testing.T) {
	repo := newMockRepo()
	item := models.NewItem("harper")
	_ = repo.CreateItem(item)

	server, _ := NewServer(repo)

	input := GetCurrentInput{Name: "harper"}
	_, _, err := server.handleGetCurrent(context.Background(), nil, input)
	if err == nil {
		t.Error("expected error when no position exists")
	}
}

func TestHandleGetCurrent_InvalidName(t *testing.T) {
	repo := newMockRepo()
	server, _ := NewServer(repo)

	input := GetCurrentInput{Name: ""}
	_, _, err := server.handleGetCurrent(context.Background(), nil, input)
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestHandleGetTimeline(t *testing.T) {
	repo := newMockRepo()
	item := models.NewItem("harper")
	_ = repo.CreateItem(item)
	pos1 := models.NewPosition(item.ID, 41.0, -87.0, nil)
	pos2 := models.NewPosition(item.ID, 42.0, -88.0, nil)
	_ = repo.CreatePosition(pos1)
	_ = repo.CreatePosition(pos2)

	server, _ := NewServer(repo)

	input := GetCurrentInput{Name: "harper"}
	result, output, err := server.handleGetTimeline(context.Background(), nil, input)
	if err != nil {
		t.Fatalf("handleGetTimeline failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if output.Count != 2 {
		t.Errorf("expected count 2, got %d", output.Count)
	}
	if len(output.Positions) != 2 {
		t.Errorf("expected 2 positions, got %d", len(output.Positions))
	}
}

func TestHandleGetTimeline_ItemNotFound(t *testing.T) {
	repo := newMockRepo()
	server, _ := NewServer(repo)

	input := GetCurrentInput{Name: "nonexistent"}
	_, _, err := server.handleGetTimeline(context.Background(), nil, input)
	if err == nil {
		t.Error("expected error for nonexistent item")
	}
}

func TestHandleGetTimeline_InvalidName(t *testing.T) {
	repo := newMockRepo()
	server, _ := NewServer(repo)

	input := GetCurrentInput{Name: "   "}
	_, _, err := server.handleGetTimeline(context.Background(), nil, input)
	if err == nil {
		t.Error("expected error for whitespace-only name")
	}
}

func TestHandleGetTimeline_GetTimelineError(t *testing.T) {
	repo := newMockRepo()
	item := models.NewItem("harper")
	_ = repo.CreateItem(item)
	repo.getTimelineErr = errors.New("database error")

	server, _ := NewServer(repo)

	input := GetCurrentInput{Name: "harper"}
	_, _, err := server.handleGetTimeline(context.Background(), nil, input)
	if err == nil {
		t.Error("expected error when timeline query fails")
	}
}

func TestHandleListItems(t *testing.T) {
	repo := newMockRepo()
	item1 := models.NewItem("harper")
	item2 := models.NewItem("car")
	_ = repo.CreateItem(item1)
	_ = repo.CreateItem(item2)
	pos := models.NewPosition(item1.ID, 41.0, -87.0, nil)
	_ = repo.CreatePosition(pos)

	server, _ := NewServer(repo)

	result, output, err := server.handleListItems(context.Background(), nil, ListItemsInput{})
	if err != nil {
		t.Fatalf("handleListItems failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if output.Count != 2 {
		t.Errorf("expected count 2, got %d", output.Count)
	}

	// Find harper's item and verify it has position
	for _, item := range output.Items {
		if item.Name == "harper" {
			if item.CurrentPosition == nil {
				t.Error("expected harper to have current position")
			}
		}
		if item.Name == "car" {
			if item.CurrentPosition != nil {
				t.Error("expected car to have no position")
			}
		}
	}
}

func TestHandleListItems_Empty(t *testing.T) {
	repo := newMockRepo()
	server, _ := NewServer(repo)

	_, output, err := server.handleListItems(context.Background(), nil, ListItemsInput{})
	if err != nil {
		t.Fatalf("handleListItems failed: %v", err)
	}
	if output.Count != 0 {
		t.Errorf("expected count 0, got %d", output.Count)
	}
}

func TestHandleListItems_Error(t *testing.T) {
	repo := newMockRepo()
	repo.listItemsErr = errors.New("database error")
	server, _ := NewServer(repo)

	_, _, err := server.handleListItems(context.Background(), nil, ListItemsInput{})
	if err == nil {
		t.Error("expected error when list items fails")
	}
}

func TestHandleRemoveItem(t *testing.T) {
	repo := newMockRepo()
	item := models.NewItem("harper")
	_ = repo.CreateItem(item)
	pos := models.NewPosition(item.ID, 41.0, -87.0, nil)
	_ = repo.CreatePosition(pos)

	server, _ := NewServer(repo)

	input := RemoveItemInput{Name: "harper"}
	result, output, err := server.handleRemoveItem(context.Background(), nil, input)
	if err != nil {
		t.Fatalf("handleRemoveItem failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if !output.Success {
		t.Error("expected success to be true")
	}

	// Verify item was deleted
	items, _ := repo.ListItems()
	if len(items) != 0 {
		t.Errorf("expected 0 items, got %d", len(items))
	}
}

func TestHandleRemoveItem_NotFound(t *testing.T) {
	repo := newMockRepo()
	server, _ := NewServer(repo)

	input := RemoveItemInput{Name: "nonexistent"}
	_, _, err := server.handleRemoveItem(context.Background(), nil, input)
	if err == nil {
		t.Error("expected error for nonexistent item")
	}
}

func TestHandleRemoveItem_InvalidName(t *testing.T) {
	repo := newMockRepo()
	server, _ := NewServer(repo)

	input := RemoveItemInput{Name: ""}
	_, _, err := server.handleRemoveItem(context.Background(), nil, input)
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestHandleRemoveItem_DeleteError(t *testing.T) {
	repo := newMockRepo()
	item := models.NewItem("harper")
	_ = repo.CreateItem(item)
	repo.deleteItemErr = errors.New("database error")

	server, _ := NewServer(repo)

	input := RemoveItemInput{Name: "harper"}
	_, _, err := server.handleRemoveItem(context.Background(), nil, input)
	if err == nil {
		t.Error("expected error when delete fails")
	}
}

func TestHandleItemsResource(t *testing.T) {
	repo := newMockRepo()
	item := models.NewItem("harper")
	_ = repo.CreateItem(item)
	pos := models.NewPosition(item.ID, 41.0, -87.0, nil)
	_ = repo.CreatePosition(pos)

	server, _ := NewServer(repo)

	result, err := server.handleItemsResource(context.Background(), nil)
	if err != nil {
		t.Fatalf("handleItemsResource failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Contents) != 1 {
		t.Fatalf("expected 1 content, got %d", len(result.Contents))
	}
	if result.Contents[0].URI != "position://items" {
		t.Errorf("expected URI 'position://items', got %q", result.Contents[0].URI)
	}
	if result.Contents[0].MIMEType != "application/json" {
		t.Errorf("expected MIME type 'application/json', got %q", result.Contents[0].MIMEType)
	}
}

func TestHandleItemsResource_Error(t *testing.T) {
	repo := newMockRepo()
	repo.listItemsErr = errors.New("database error")
	server, _ := NewServer(repo)

	_, err := server.handleItemsResource(context.Background(), nil)
	if err == nil {
		t.Error("expected error when list items fails")
	}
}
