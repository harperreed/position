// ABOUTME: Position remove command
// ABOUTME: Removes an item and all its position history

package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/harper/position/internal/db"
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
