// ABOUTME: Position add command
// ABOUTME: Creates new positions for items with optional label and timestamp

package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/fatih/color"
	"github.com/harper/position/internal/db"
	"github.com/harper/position/internal/models"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:     "add <name> <latitude> <longitude>",
	Aliases: []string{"a"},
	Short:   "Add a position for an item",
	Long: `Add a new position for an item. Creates the item if it doesn't exist.

Examples:
  position add harper 41.8781 -87.6298
  position add harper 41.8781 -87.6298 --label chicago
  position add harper 41.8781 -87.6298 --label chicago --at 2024-12-14T15:00:00Z`,
	Args: cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		lat, err := strconv.ParseFloat(args[1], 64)
		if err != nil {
			return fmt.Errorf("invalid latitude: %w", err)
		}
		if lat < -90 || lat > 90 {
			return fmt.Errorf("latitude must be between -90 and 90")
		}

		lng, err := strconv.ParseFloat(args[2], 64)
		if err != nil {
			return fmt.Errorf("invalid longitude: %w", err)
		}
		if lng < -180 || lng > 180 {
			return fmt.Errorf("longitude must be between -180 and 180")
		}

		// Get or create item
		item, err := db.GetItemByName(dbConn, name)
		if err != nil {
			item = models.NewItem(name)
			if err := db.CreateItem(dbConn, item); err != nil {
				return fmt.Errorf("failed to create item: %w", err)
			}
		}

		// Parse optional flags
		var label *string
		if labelStr, _ := cmd.Flags().GetString("label"); labelStr != "" {
			label = &labelStr
		}

		var pos *models.Position
		if atStr, _ := cmd.Flags().GetString("at"); atStr != "" {
			recordedAt, err := time.Parse(time.RFC3339, atStr)
			if err != nil {
				return fmt.Errorf("invalid timestamp format (use RFC3339, e.g., 2024-12-14T15:00:00Z): %w", err)
			}
			pos = models.NewPositionWithRecordedAt(item.ID, lat, lng, label, recordedAt)
		} else {
			pos = models.NewPosition(item.ID, lat, lng, label)
		}

		if err := db.CreatePosition(dbConn, pos); err != nil {
			return fmt.Errorf("failed to create position: %w", err)
		}

		color.Green("✓ Added position for %s", name)
		if label != nil {
			fmt.Printf("  %s @ %s (%.4f, %.4f)\n",
				color.New(color.Faint).Sprint(pos.ID.String()[:6]),
				*label, lat, lng)
		} else {
			fmt.Printf("  %s @ (%.4f, %.4f)\n",
				color.New(color.Faint).Sprint(pos.ID.String()[:6]),
				lat, lng)
		}

		return nil
	},
}

func init() {
	addCmd.Flags().StringP("label", "l", "", "location label (e.g., 'chicago')")
	addCmd.Flags().String("at", "", "recorded time (RFC3339, e.g., 2024-12-14T15:00:00Z)")
	addCmd.Flags().SetInterspersed(false)

	rootCmd.AddCommand(addCmd)
}
