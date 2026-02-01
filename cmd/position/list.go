// ABOUTME: Position list command
// ABOUTME: Lists all tracked items with their current positions

package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/harper/position/internal/storage"
	"github.com/harper/position/internal/ui"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all tracked items",
	RunE: func(cmd *cobra.Command, args []string) error {
		items, err := db.ListItems()
		if err != nil {
			return fmt.Errorf("failed to list items: %w", err)
		}

		if len(items) == 0 {
			fmt.Println("No items tracked yet. Use 'position add' to add one.")
			return nil
		}

		for _, item := range items {
			pos, err := db.GetCurrentPosition(item.ID)
			if err != nil {
				// ErrNotFound is expected for items without positions
				if !errors.Is(err, storage.ErrNotFound) {
					// Unexpected error - log but continue with other items
					fmt.Fprintf(os.Stderr, "warning: failed to get position for %s: %v\n", item.Name, err)
				}
				// pos will be nil, which FormatItemWithPosition handles
			}
			fmt.Println(ui.FormatItemWithPosition(item, pos))
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
