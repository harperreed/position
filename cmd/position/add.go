// ABOUTME: Position add command
// ABOUTME: Creates new positions for items with optional label and timestamp

package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/harper/position/internal/charm"
	"github.com/harper/position/internal/models"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:     "add <name> --lat <latitude> --lng <longitude>",
	Aliases: []string{"a"},
	Short:   "Add a position for an item",
	Long: `Add a new position for an item. Creates the item if it doesn't exist.

Examples:
  position add harper --lat 41.8781 --lng -87.6298
  position add harper --lat 41.8781 --lng -87.6298 --label chicago
  position add harper --lat 41.8781 --lng -87.6298 -l chicago --at 2024-12-14T15:00:00Z`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		lat, _ := cmd.Flags().GetFloat64("lat")
		lng, _ := cmd.Flags().GetFloat64("lng")

		if err := models.ValidateCoordinates(lat, lng); err != nil {
			return err
		}

		// Get or create item
		item, err := charmClient.GetItemByName(name)
		if err != nil {
			if errors.Is(err, charm.ErrNotFound) {
				item = models.NewItem(name)
				if err := charmClient.CreateItem(item); err != nil {
					return fmt.Errorf("failed to create item: %w", err)
				}
			} else {
				return fmt.Errorf("failed to get item: %w", err)
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

		if err := charmClient.CreatePosition(pos); err != nil {
			return fmt.Errorf("failed to create position: %w", err)
		}

		color.Green("✓ Position set for %s", name)
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
	addCmd.Flags().Float64("lat", 0, "latitude coordinate (-90 to 90)")
	addCmd.Flags().Float64("lng", 0, "longitude coordinate (-180 to 180)")
	addCmd.Flags().StringP("label", "l", "", "location label (e.g., 'chicago')")
	addCmd.Flags().String("at", "", "recorded time (RFC3339, e.g., 2024-12-14T15:00:00Z)")

	_ = addCmd.MarkFlagRequired("lat")
	_ = addCmd.MarkFlagRequired("lng")

	rootCmd.AddCommand(addCmd)
}
