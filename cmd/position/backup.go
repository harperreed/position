// ABOUTME: Backup command for exporting data to YAML
// ABOUTME: Creates portable backup files for data migration

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/harper/position/internal/storage"
	"github.com/spf13/cobra"
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Create a YAML backup of all data",
	Long: `Create a YAML backup file containing all items and positions.

The backup file can be used to:
- Migrate data between machines
- Restore after data loss
- Import into a fresh database

Examples:
  position backup --output positions.yaml
  position backup -o ~/backups/positions-$(date +%Y%m%d).yaml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		output, _ := cmd.Flags().GetString("output")

		data, err := storage.ExportBackup(db)
		if err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}

		if output == "" {
			// Default filename with timestamp
			output = fmt.Sprintf("positions-%s.yaml", time.Now().Format("20060102-150405"))
		}

		if err := os.WriteFile(output, data, 0644); err != nil { //nolint:gosec // 0644 is intentional for backup files
			return fmt.Errorf("failed to write backup: %w", err)
		}

		items, _ := db.ListItems()
		positions, _ := db.GetAllPositions()

		color.Green("Backup created: %s", output)
		fmt.Printf("  %d items, %d positions\n", len(items), len(positions))

		return nil
	},
}

func init() {
	backupCmd.Flags().StringP("output", "o", "", "output file (default: positions-YYYYMMDD-HHMMSS.yaml)")

	rootCmd.AddCommand(backupCmd)
}
