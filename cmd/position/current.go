// ABOUTME: Position current command
// ABOUTME: Shows the current (most recent) position for an item

package main

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/harper/position/internal/db"
	"github.com/harper/position/internal/ui"
	"github.com/spf13/cobra"
)

var currentCmd = &cobra.Command{
	Use:     "current <name>",
	Aliases: []string{"c"},
	Short:   "Get current position of an item",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		item, err := db.GetItemByName(dbConn, name)
		if err != nil {
			return fmt.Errorf("item '%s' not found", name)
		}

		pos, err := db.GetCurrentPosition(dbConn, item.ID)
		if err != nil {
			return fmt.Errorf("no position found for '%s'", name)
		}

		fmt.Printf("%s @ %s\n",
			color.GreenString(name),
			ui.FormatPosition(pos))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(currentCmd)
}
