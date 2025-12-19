// ABOUTME: Position list command
// ABOUTME: Lists all tracked items with their current positions

package main

import (
	"fmt"

	"github.com/harper/position/internal/ui"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all tracked items",
	RunE: func(cmd *cobra.Command, args []string) error {
		items, err := charmClient.ListItems()
		if err != nil {
			return fmt.Errorf("failed to list items: %w", err)
		}

		if len(items) == 0 {
			fmt.Println("No items tracked yet. Use 'position add' to add one.")
			return nil
		}

		for _, item := range items {
			pos, _ := charmClient.GetCurrentPosition(item.ID)
			fmt.Println(ui.FormatItemWithPosition(item, pos))
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
