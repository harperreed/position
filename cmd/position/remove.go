// ABOUTME: Position remove command
// ABOUTME: Removes an item and all its position history

package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/harperreed/sweet/vault"

	"github.com/fatih/color"
	"github.com/harper/position/internal/db"
	"github.com/harper/position/internal/models"
	"github.com/harper/position/internal/sync"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:     "remove <name>",
	Aliases: []string{"rm"},
	Short:   "Remove an item and all its positions",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		item, err := db.GetItemByName(dbConn, name)
		if err != nil {
			return fmt.Errorf("item '%s' not found", name)
		}

		confirm, _ := cmd.Flags().GetBool("confirm")
		if !confirm {
			fmt.Printf("Remove '%s' and all position history? [y/N] ", name)
			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))
			if response != "y" && response != "yes" {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		// Get positions before deleting (for sync)
		positions, _ := db.GetTimeline(dbConn, item.ID)

		// Queue sync deletes before local delete
		if err := queueRemoveToSync(cmd.Context(), item, positions); err != nil {
			color.Yellow("⚠ Sync queue failed: %v", err)
		}

		if err := db.DeleteItem(dbConn, item.ID); err != nil {
			return fmt.Errorf("failed to remove item: %w", err)
		}

		color.Green("✓ Removed %s", name)
		return nil
	},
}

func init() {
	removeCmd.Flags().Bool("confirm", false, "skip confirmation prompt")

	rootCmd.AddCommand(removeCmd)
}

// queueRemoveToSync queues delete operations to vault if sync is configured.
func queueRemoveToSync(ctx context.Context, item *models.Item, positions []*models.Position) error {
	cfg, err := sync.LoadConfig()
	if err != nil {
		return nil // No config, skip silently
	}

	if !cfg.IsConfigured() {
		return nil // Not configured, skip silently
	}

	syncer, err := sync.NewSyncer(cfg, dbConn)
	if err != nil {
		return fmt.Errorf("create syncer: %w", err)
	}
	defer func() { _ = syncer.Close() }()

	// Queue position deletes first
	for _, pos := range positions {
		if err := syncer.QueuePositionChange(ctx, pos.ID, item.Name, pos.Latitude, pos.Longitude, pos.Label, pos.RecordedAt, vault.OpDelete); err != nil {
			return fmt.Errorf("queue position delete: %w", err)
		}
	}

	// Queue item delete
	if err := syncer.QueueItemChange(ctx, item.ID, item.Name, vault.OpDelete); err != nil {
		return fmt.Errorf("queue item delete: %w", err)
	}

	return nil
}
