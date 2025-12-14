// ABOUTME: Unit tests for terminal UI formatting
// ABOUTME: Tests human-readable output for items and positions

package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/harper/position/internal/models"
)

func TestFormatPosition(t *testing.T) {
	label := "chicago"
	pos := &models.Position{
		ID:         uuid.New(),
		Latitude:   41.8781,
		Longitude:  -87.6298,
		Label:      &label,
		RecordedAt: time.Now(),
	}

	output := FormatPosition(pos)
	if !strings.Contains(output, "chicago") {
		t.Error("expected output to contain label")
	}
	if !strings.Contains(output, "41.8781") {
		t.Error("expected output to contain latitude")
	}
}

func TestFormatPosition_NoLabel(t *testing.T) {
	pos := &models.Position{
		ID:         uuid.New(),
		Latitude:   41.8781,
		Longitude:  -87.6298,
		Label:      nil,
		RecordedAt: time.Now(),
	}

	output := FormatPosition(pos)
	if !strings.Contains(output, "41.8781") {
		t.Error("expected output to contain latitude")
	}
}

func TestFormatRelativeTime(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{30 * time.Second, "just now"},
		{5 * time.Minute, "5 minutes ago"},
		{2 * time.Hour, "2 hours ago"},
		{25 * time.Hour, "1 day ago"},
	}

	for _, tc := range tests {
		tm := time.Now().Add(-tc.duration)
		result := FormatRelativeTime(tm)
		if !strings.Contains(result, tc.expected[:4]) {
			t.Errorf("for %v: expected %q in result, got %q", tc.duration, tc.expected, result)
		}
	}
}
