// ABOUTME: MCP tool definitions and handlers
// ABOUTME: Provides CRUD operations for AI agents

package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/harper/position/internal/db"
	"github.com/harper/position/internal/models"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (s *Server) registerTools() {
	s.registerAddPositionTool()
	s.registerGetCurrentTool()
	s.registerGetTimelineTool()
	s.registerListItemsTool()
	s.registerRemoveItemTool()
}

// AddPositionInput defines input for add_position tool.
type AddPositionInput struct {
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Label     *string `json:"label,omitempty"`
	At        *string `json:"at,omitempty"`
}

// PositionOutput defines output for position tools.
type PositionOutput struct {
	ItemName   string    `json:"item_name"`
	Latitude   float64   `json:"latitude"`
	Longitude  float64   `json:"longitude"`
	Label      *string   `json:"label,omitempty"`
	RecordedAt time.Time `json:"recorded_at"`
}

func (s *Server) registerAddPositionTool() {
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "add_position",
		Description: "Add a position for an item (creates item if needed). Use this to track where something or someone is located.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the item to track (e.g., 'harper', 'car')",
				},
				"latitude": map[string]interface{}{
					"type":        "number",
					"description": "Latitude coordinate (-90 to 90)",
				},
				"longitude": map[string]interface{}{
					"type":        "number",
					"description": "Longitude coordinate (-180 to 180)",
				},
				"label": map[string]interface{}{
					"type":        "string",
					"description": "Optional location label (e.g., 'chicago', '123 Main St')",
				},
				"at": map[string]interface{}{
					"type":        "string",
					"description": "Optional recorded time in RFC3339 format",
				},
			},
			"required": []string{"name", "latitude", "longitude"},
		},
	}, s.handleAddPosition)
}

func (s *Server) handleAddPosition(_ context.Context, req *mcp.CallToolRequest, input AddPositionInput) (*mcp.CallToolResult, PositionOutput, error) {
	if err := models.ValidateCoordinates(input.Latitude, input.Longitude); err != nil {
		return nil, PositionOutput{}, err
	}

	item, err := db.GetItemByName(s.db, input.Name)
	if err != nil {
		item = models.NewItem(input.Name)
		if err := db.CreateItem(s.db, item); err != nil {
			return nil, PositionOutput{}, fmt.Errorf("failed to create item: %w", err)
		}
	}

	var pos *models.Position
	if input.At != nil {
		recordedAt, err := time.Parse(time.RFC3339, *input.At)
		if err != nil {
			return nil, PositionOutput{}, fmt.Errorf("invalid timestamp: %w", err)
		}
		pos = models.NewPositionWithRecordedAt(item.ID, input.Latitude, input.Longitude, input.Label, recordedAt)
	} else {
		pos = models.NewPosition(item.ID, input.Latitude, input.Longitude, input.Label)
	}

	if err := db.CreatePosition(s.db, pos); err != nil {
		return nil, PositionOutput{}, fmt.Errorf("failed to create position: %w", err)
	}

	output := PositionOutput{
		ItemName:   input.Name,
		Latitude:   pos.Latitude,
		Longitude:  pos.Longitude,
		Label:      pos.Label,
		RecordedAt: pos.RecordedAt,
	}

	jsonBytes, _ := json.MarshalIndent(output, "", "  ") //nolint:errchkjson // output is always serializable
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(jsonBytes)}},
	}, output, nil
}

// GetCurrentInput defines input for get_current tool.
type GetCurrentInput struct {
	Name string `json:"name"`
}

func (s *Server) registerGetCurrentTool() {
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "get_current",
		Description: "Get the current (most recent) position of an item.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the item",
				},
			},
			"required": []string{"name"},
		},
	}, s.handleGetCurrent)
}

func (s *Server) handleGetCurrent(_ context.Context, req *mcp.CallToolRequest, input GetCurrentInput) (*mcp.CallToolResult, PositionOutput, error) {
	item, err := db.GetItemByName(s.db, input.Name)
	if err != nil {
		return nil, PositionOutput{}, fmt.Errorf("item '%s' not found", input.Name)
	}

	pos, err := db.GetCurrentPosition(s.db, item.ID)
	if err != nil {
		return nil, PositionOutput{}, fmt.Errorf("no position found for '%s'", input.Name)
	}

	output := PositionOutput{
		ItemName:   input.Name,
		Latitude:   pos.Latitude,
		Longitude:  pos.Longitude,
		Label:      pos.Label,
		RecordedAt: pos.RecordedAt,
	}

	jsonBytes, _ := json.MarshalIndent(output, "", "  ") //nolint:errchkjson // output is always serializable
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(jsonBytes)}},
	}, output, nil
}

// TimelineOutput defines output for timeline tool.
type TimelineOutput struct {
	ItemName  string           `json:"item_name"`
	Positions []PositionOutput `json:"positions"`
	Count     int              `json:"count"`
}

func (s *Server) registerGetTimelineTool() {
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "get_timeline",
		Description: "Get the position history for an item, newest first.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the item",
				},
			},
			"required": []string{"name"},
		},
	}, s.handleGetTimeline)
}

func (s *Server) handleGetTimeline(_ context.Context, req *mcp.CallToolRequest, input GetCurrentInput) (*mcp.CallToolResult, TimelineOutput, error) {
	item, err := db.GetItemByName(s.db, input.Name)
	if err != nil {
		return nil, TimelineOutput{}, fmt.Errorf("item '%s' not found", input.Name)
	}

	positions, err := db.GetTimeline(s.db, item.ID)
	if err != nil {
		return nil, TimelineOutput{}, fmt.Errorf("failed to get timeline: %w", err)
	}

	posOutputs := make([]PositionOutput, len(positions))
	for i, pos := range positions {
		posOutputs[i] = PositionOutput{
			ItemName:   input.Name,
			Latitude:   pos.Latitude,
			Longitude:  pos.Longitude,
			Label:      pos.Label,
			RecordedAt: pos.RecordedAt,
		}
	}

	output := TimelineOutput{
		ItemName:  input.Name,
		Positions: posOutputs,
		Count:     len(posOutputs),
	}

	jsonBytes, _ := json.MarshalIndent(output, "", "  ") //nolint:errchkjson // output is always serializable
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(jsonBytes)}},
	}, output, nil
}

// ItemOutput defines output for item tools.
type ItemOutput struct {
	Name            string          `json:"name"`
	CurrentPosition *PositionOutput `json:"current_position,omitempty"`
}

// ListItemsOutput defines output for list_items tool.
type ListItemsOutput struct {
	Items []ItemOutput `json:"items"`
	Count int          `json:"count"`
}

// ListItemsInput is empty but required for type.
type ListItemsInput struct{}

func (s *Server) registerListItemsTool() {
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "list_items",
		Description: "List all tracked items with their current positions.",
		InputSchema: map[string]interface{}{
			"type": "object",
		},
	}, s.handleListItems)
}

func (s *Server) handleListItems(_ context.Context, req *mcp.CallToolRequest, input ListItemsInput) (*mcp.CallToolResult, ListItemsOutput, error) {
	items, err := db.ListItems(s.db)
	if err != nil {
		return nil, ListItemsOutput{}, fmt.Errorf("failed to list items: %w", err)
	}

	itemOutputs := make([]ItemOutput, len(items))
	for i, item := range items {
		itemOutputs[i] = ItemOutput{Name: item.Name}

		pos, err := db.GetCurrentPosition(s.db, item.ID)
		if err == nil {
			itemOutputs[i].CurrentPosition = &PositionOutput{
				ItemName:   item.Name,
				Latitude:   pos.Latitude,
				Longitude:  pos.Longitude,
				Label:      pos.Label,
				RecordedAt: pos.RecordedAt,
			}
		}
	}

	output := ListItemsOutput{
		Items: itemOutputs,
		Count: len(itemOutputs),
	}

	jsonBytes, _ := json.MarshalIndent(output, "", "  ") //nolint:errchkjson // output is always serializable
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(jsonBytes)}},
	}, output, nil
}

// RemoveItemInput defines input for remove_item tool.
type RemoveItemInput struct {
	Name string `json:"name"`
}

// RemoveItemOutput defines output for remove_item tool.
type RemoveItemOutput struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func (s *Server) registerRemoveItemTool() {
	mcp.AddTool(s.mcp, &mcp.Tool{
		Name:        "remove_item",
		Description: "Remove an item and all its position history. This cannot be undone.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the item to remove",
				},
			},
			"required": []string{"name"},
		},
	}, s.handleRemoveItem)
}

func (s *Server) handleRemoveItem(_ context.Context, req *mcp.CallToolRequest, input RemoveItemInput) (*mcp.CallToolResult, RemoveItemOutput, error) {
	item, err := db.GetItemByName(s.db, input.Name)
	if err != nil {
		return nil, RemoveItemOutput{}, fmt.Errorf("item '%s' not found", input.Name)
	}

	if err := db.DeleteItem(s.db, item.ID); err != nil {
		return nil, RemoveItemOutput{}, fmt.Errorf("failed to remove item: %w", err)
	}

	output := RemoveItemOutput{
		Success: true,
		Message: fmt.Sprintf("Removed '%s' and all position history", input.Name),
	}

	jsonBytes, _ := json.MarshalIndent(output, "", "  ") //nolint:errchkjson // output is always serializable
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(jsonBytes)}},
	}, output, nil
}
