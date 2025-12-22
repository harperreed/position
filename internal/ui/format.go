// ABOUTME: Terminal UI formatting utilities
// ABOUTME: Provides human-readable output for items and positions

package ui

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/harper/position/internal/models"
)

// FormatPosition formats a position for terminal display.
func FormatPosition(pos *models.Position) string {
	if pos == nil {
		return color.New(color.Faint).Sprint("(no position)")
	}
	coords := fmt.Sprintf("(%.4f, %.4f)", pos.Latitude, pos.Longitude)
	relTime := FormatRelativeTime(pos.RecordedAt)

	if pos.Label != nil && *pos.Label != "" {
		return fmt.Sprintf("%s %s - %s",
			color.CyanString(*pos.Label),
			color.New(color.Faint).Sprint(coords),
			color.New(color.Faint).Sprint(relTime))
	}
	return fmt.Sprintf("%s - %s",
		color.CyanString(coords),
		color.New(color.Faint).Sprint(relTime))
}

// FormatPositionForTimeline formats a position for timeline display.
func FormatPositionForTimeline(pos *models.Position) string {
	if pos == nil {
		return color.New(color.Faint).Sprint("  (no position)")
	}
	coords := fmt.Sprintf("(%.4f, %.4f)", pos.Latitude, pos.Longitude)
	timeStr := pos.RecordedAt.Format("Jan 2, 3:04 PM")

	if pos.Label != nil && *pos.Label != "" {
		return fmt.Sprintf("  %s %s - %s",
			color.CyanString(*pos.Label),
			color.New(color.Faint).Sprint(coords),
			timeStr)
	}
	return fmt.Sprintf("  %s - %s",
		color.CyanString(coords),
		timeStr)
}

// FormatItemWithPosition formats an item with its current position.
func FormatItemWithPosition(item *models.Item, pos *models.Position) string {
	if item == nil {
		return color.New(color.Faint).Sprint("(invalid item)")
	}
	if pos == nil {
		return fmt.Sprintf("%s - %s",
			color.GreenString(item.Name),
			color.New(color.Faint).Sprint("no position"))
	}

	var posStr string
	if pos.Label != nil && *pos.Label != "" {
		posStr = *pos.Label
	} else {
		posStr = fmt.Sprintf("(%.4f, %.4f)", pos.Latitude, pos.Longitude)
	}

	relTime := FormatRelativeTime(pos.RecordedAt)
	return fmt.Sprintf("%s - %s (%s)",
		color.GreenString(item.Name),
		posStr,
		color.New(color.Faint).Sprint(relTime))
}

// FormatRelativeTime formats a time as relative to now.
func FormatRelativeTime(t time.Time) string {
	diff := time.Since(t)

	// Handle future times (clock skew, bad data)
	if diff < 0 {
		return color.YellowString("in the future")
	}

	if diff < time.Minute {
		return "just now"
	}
	if diff < time.Hour {
		mins := int(diff.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	}
	if diff < 24*time.Hour {
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	}
	days := int(diff.Hours() / 24)
	if days == 1 {
		return "1 day ago"
	}
	return fmt.Sprintf("%d days ago", days)
}
