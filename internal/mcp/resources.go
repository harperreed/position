// ABOUTME: MCP resource definitions
// ABOUTME: Provides read-only views for AI agents

package mcp

import (
	"context"
	"encoding/json"
	"fmt"

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
	items, err := s.client.ListItems()
	if err != nil {
		return nil, fmt.Errorf("failed to list items: %w", err)
	}

	itemOutputs := make([]ItemOutput, len(items))
	for i, item := range items {
		itemOutputs[i] = ItemOutput{Name: item.Name}

		pos, err := s.client.GetCurrentPosition(item.ID)
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
