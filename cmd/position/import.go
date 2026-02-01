// ABOUTME: Import command for restoring data from YAML backup
// ABOUTME: Supports importing backup files created by the backup command

package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/harper/position/internal/storage"
	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
	Use:   "import <file>",
	Short: "Import data from a YAML backup",
	Long: `Import items and positions from a YAML backup file.

This restores data from a backup created with 'position backup'.

WARNING: This will add to existing data, not replace it.
Use 'position reset' first if you want a clean import.

Examples:
  position import positions.yaml
  position import ~/backups/positions-20241214.yaml`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filename := args[0]

		data, err := os.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}

		confirm, _ := cmd.Flags().GetBool("confirm")
		if !confirm {
			fmt.Printf("Import data from '%s'? [y/N] ", filename)
			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))
			if response != "y" && response != "yes" {
				fmt.Println("Canceled.")
				return nil
			}
		}

		if err := storage.ImportBackup(db, data); err != nil {
			return fmt.Errorf("failed to import: %w", err)
		}

		items, _ := db.ListItems()
		positions, _ := db.GetAllPositions()

		color.Green("Import complete")
		fmt.Printf("  %d items, %d positions in database\n", len(items), len(positions))

		return nil
	},
}

func init() {
	importCmd.Flags().Bool("confirm", false, "skip confirmation prompt")

	rootCmd.AddCommand(importCmd)
}
