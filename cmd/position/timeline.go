// ABOUTME: Position timeline command
// ABOUTME: Shows location history for an item

package main

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/harper/position/internal/ui"
	"github.com/spf13/cobra"
)

var timelineCmd = &cobra.Command{
	Use:     "timeline <name>",
	Aliases: []string{"t"},
	Short:   "Get position history for an item",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		item, err := charmClient.GetItemByName(name)
		if err != nil {
			return fmt.Errorf("item '%s' not found", name)
		}

		positions, err := charmClient.GetTimeline(item.ID)
		if err != nil {
			return fmt.Errorf("failed to get timeline: %w", err)
		}

		if len(positions) == 0 {
			fmt.Printf("%s has no position history\n", color.GreenString(name))
			return nil
		}

		fmt.Printf("%s timeline:\n", color.GreenString(name))
		for _, pos := range positions {
			fmt.Println(ui.FormatPositionForTimeline(pos))
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(timelineCmd)
}
