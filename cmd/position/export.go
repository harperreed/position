// ABOUTME: Export command for generating GeoJSON output
// ABOUTME: Supports time filtering and multiple geometry types

package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/harper/position/internal/geojson"
	"github.com/harper/position/internal/models"
	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:     "export [name]",
	Aliases: []string{"e"},
	Short:   "Export positions as GeoJSON",
	Long: `Export positions as GeoJSON for mapping applications.

Examples:
  # Export all positions for an item
  position export harper --format geojson

  # Export with time filter (relative)
  position export harper --format geojson --since 24h
  position export harper --format geojson --since 7d

  # Export with time filter (absolute)
  position export harper --format geojson --from 2024-12-01 --to 2024-12-14

  # Export all items
  position export --format geojson --since 7d

  # Export as LineString (path/track)
  position export harper --format geojson --geometry line

  # Save to file
  position export harper --format geojson --output map.geojson`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		format, _ := cmd.Flags().GetString("format")
		if format != "geojson" {
			return fmt.Errorf("unsupported format: %s (only 'geojson' is supported)", format)
		}

		geometry, _ := cmd.Flags().GetString("geometry")
		if geometry != "points" && geometry != "line" {
			return fmt.Errorf("unsupported geometry: %s (use 'points' or 'line')", geometry)
		}

		// Parse time filters
		since, _ := cmd.Flags().GetString("since")
		from, _ := cmd.Flags().GetString("from")
		to, _ := cmd.Flags().GetString("to")

		var sinceTime, fromTime, toTime time.Time
		var err error

		if since != "" {
			sinceTime, err = parseDuration(since)
			if err != nil {
				return fmt.Errorf("invalid --since value: %w", err)
			}
		}
		if from != "" {
			fromTime, err = parseDate(from)
			if err != nil {
				return fmt.Errorf("invalid --from value: %w", err)
			}
		}
		if to != "" {
			toTime, err = parseDate(to)
			if err != nil {
				return fmt.Errorf("invalid --to value: %w", err)
			}
			// Set to end of day
			toTime = toTime.Add(24*time.Hour - time.Second)
		}

		// Build item name cache for resolving IDs to names
		items, err := charmClient.ListItems()
		if err != nil {
			return fmt.Errorf("failed to list items: %w", err)
		}
		itemNames := make(map[string]string)
		for _, item := range items {
			itemNames[item.ID.String()] = item.Name
		}
		nameResolver := func(itemID string) string {
			return itemNames[itemID]
		}

		var positions []*models.Position

		if len(args) == 1 {
			// Export single item
			name := args[0]
			item, err := charmClient.GetItemByName(name)
			if err != nil {
				return fmt.Errorf("item '%s' not found", name)
			}

			positions, err = getPositionsForItem(item, sinceTime, fromTime, toTime)
			if err != nil {
				return err
			}
		} else {
			// Export all items
			positions, err = getAllPositions(sinceTime, fromTime, toTime)
			if err != nil {
				return err
			}
		}

		if len(positions) == 0 {
			return fmt.Errorf("no positions found")
		}

		// Generate GeoJSON
		var fc *geojson.FeatureCollection
		if geometry == "line" {
			fc = geojson.ToLineFeatureCollection(positions, nameResolver)
		} else {
			fc = geojson.ToPointsFeatureCollection(positions, nameResolver)
		}

		jsonBytes, err := fc.ToJSONIndent()
		if err != nil {
			return fmt.Errorf("failed to generate GeoJSON: %w", err)
		}

		// Output
		output, _ := cmd.Flags().GetString("output")
		if output != "" {
			if err := os.WriteFile(output, jsonBytes, 0600); err != nil {
				return fmt.Errorf("failed to write file: %w", err)
			}
			fmt.Fprintf(os.Stderr, "Wrote %d positions to %s\n", len(positions), output)
		} else {
			fmt.Println(string(jsonBytes))
		}

		return nil
	},
}

func getPositionsForItem(item *models.Item, since, from, to time.Time) ([]*models.Position, error) {
	if !since.IsZero() {
		return charmClient.GetPositionsSince(item.ID, since)
	}
	if !from.IsZero() && !to.IsZero() {
		return charmClient.GetPositionsInRange(item.ID, from, to)
	}
	if !from.IsZero() {
		return charmClient.GetPositionsSince(item.ID, from)
	}
	// No time filter - get all (use timeline which is DESC, but we want ASC)
	positions, err := charmClient.GetTimeline(item.ID)
	if err != nil {
		return nil, err
	}
	// Reverse to get chronological order
	for i, j := 0, len(positions)-1; i < j; i, j = i+1, j-1 {
		positions[i], positions[j] = positions[j], positions[i]
	}
	return positions, nil
}

func getAllPositions(since, from, to time.Time) ([]*models.Position, error) {
	if !since.IsZero() {
		return charmClient.GetAllPositionsSince(since)
	}
	if !from.IsZero() && !to.IsZero() {
		return charmClient.GetAllPositionsInRange(from, to)
	}
	if !from.IsZero() {
		return charmClient.GetAllPositionsSince(from)
	}
	return charmClient.GetAllPositions()
}

// parseDuration parses relative duration strings like "24h", "7d", "1w".
func parseDuration(s string) (time.Time, error) {
	re := regexp.MustCompile(`^(\d+)([hdwm])$`)
	matches := re.FindStringSubmatch(s)
	if matches == nil {
		return time.Time{}, fmt.Errorf("invalid duration format (use e.g., 24h, 7d, 1w)")
	}

	num, _ := strconv.Atoi(matches[1])
	unit := matches[2]

	var duration time.Duration
	switch unit {
	case "h":
		duration = time.Duration(num) * time.Hour
	case "d":
		duration = time.Duration(num) * 24 * time.Hour
	case "w":
		duration = time.Duration(num) * 7 * 24 * time.Hour
	case "m":
		duration = time.Duration(num) * 30 * 24 * time.Hour
	}

	return time.Now().Add(-duration), nil
}

// parseDate parses date strings in RFC3339 or YYYY-MM-DD format.
func parseDate(s string) (time.Time, error) {
	// Try RFC3339 first
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	// Try YYYY-MM-DD
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("invalid date format (use YYYY-MM-DD or RFC3339)")
}

func init() {
	exportCmd.Flags().StringP("format", "f", "geojson", "output format (geojson)")
	exportCmd.Flags().StringP("geometry", "g", "points", "geometry type (points, line)")
	exportCmd.Flags().String("since", "", "relative time filter (e.g., 24h, 7d, 1w)")
	exportCmd.Flags().String("from", "", "start date (YYYY-MM-DD or RFC3339)")
	exportCmd.Flags().String("to", "", "end date (YYYY-MM-DD or RFC3339)")
	exportCmd.Flags().StringP("output", "o", "", "output file (default: stdout)")

	rootCmd.AddCommand(exportCmd)
}
