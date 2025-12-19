// ABOUTME: MCP serve command
// ABOUTME: Starts the MCP server for AI agent integration

package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/harper/position/internal/mcp"
	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start MCP server for AI agents",
	RunE: func(cmd *cobra.Command, args []string) error {
		server, err := mcp.NewServer(charmClient)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sigCh
			cancel()
		}()

		return server.Serve(ctx)
	},
}

func init() {
	rootCmd.AddCommand(mcpCmd)
}
