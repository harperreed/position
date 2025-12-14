// ABOUTME: MCP resource definitions
// ABOUTME: Provides read-only views for AI agents

package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/harper/position/internal/db"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func (s *Server) registerResources() {
	s.mcp.AddResource(&mcp.Resource{
		Name:        "position://items",
		Description: "All tracked items with their current positions",
		URI:         "position://items",
		MIMEType:    "application/json",
	}, s.handleItemsResource)
}

func (s *Server) handleItemsResource(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	items, err := db.ListItems(s.db)
	if err != nil {
		return nil, fmt.Errorf("failed to list items: %w", err)
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

	jsonBytes, _ := json.MarshalIndent(output, "", "  ")

	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:      "position://items",
				MIMEType: "application/json",
				Text:     string(jsonBytes),
			},
		},
	}, nil
}
