// ABOUTME: MCP server initialization and configuration
// ABOUTME: Sets up server with tools and resources for AI agents

package mcp

import (
	"context"
	"fmt"

	"github.com/harper/position/internal/storage"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Server wraps MCP server with SQLite repository.
type Server struct {
	mcp  *mcp.Server
	repo storage.Repository
}

// NewServer creates MCP server with all capabilities.
func NewServer(repo storage.Repository) (*Server, error) {
	if repo == nil {
		return nil, fmt.Errorf("repository is required")
	}

	mcpServer := mcp.NewServer(
		&mcp.Implementation{
			Name:    "position",
			Version: "1.0.0",
		},
		nil,
	)

	s := &Server{
		mcp:  mcpServer,
		repo: repo,
	}

	s.registerTools()
	s.registerResources()

	return s, nil
}

// Serve starts the MCP server in stdio mode.
func (s *Server) Serve(ctx context.Context) error {
	return s.mcp.Run(ctx, &mcp.StdioTransport{})
}
