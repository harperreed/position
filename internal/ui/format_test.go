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
	if !strings.Contains(output, "-87.6298") {
		t.Error("expected output to contain longitude")
	}
}

func TestFormatPosition_NilPosition(t *testing.T) {
	output := FormatPosition(nil)
	if !strings.Contains(output, "no position") {
		t.Errorf("expected nil position message, got %q", output)
	}
}

func TestFormatPosition_EmptyLabel(t *testing.T) {
	empty := ""
	pos := &models.Position{
		ID:         uuid.New(),
		Latitude:   41.8781,
		Longitude:  -87.6298,
		Label:      &empty,
		RecordedAt: time.Now(),
	}

	output := FormatPosition(pos)
	// Should show coordinates when label is empty
	if !strings.Contains(output, "41.8781") {
		t.Error("expected output to contain latitude for empty label")
	}
}

func TestFormatPositionForTimeline(t *testing.T) {
	label := "office"
	pos := &models.Position{
		ID:         uuid.New(),
		Latitude:   40.7128,
		Longitude:  -74.0060,
		Label:      &label,
		RecordedAt: time.Date(2024, 12, 15, 14, 30, 0, 0, time.Local),
	}

	output := FormatPositionForTimeline(pos)
	if !strings.Contains(output, "office") {
		t.Error("expected output to contain label")
	}
	if !strings.Contains(output, "40.7128") {
		t.Error("expected output to contain latitude")
	}
	// Should have date formatting
	if !strings.Contains(output, "Dec") || !strings.Contains(output, "15") {
		t.Errorf("expected date in output, got %q", output)
	}
}

func TestFormatPositionForTimeline_NoLabel(t *testing.T) {
	pos := &models.Position{
		ID:         uuid.New(),
		Latitude:   40.7128,
		Longitude:  -74.0060,
		Label:      nil,
		RecordedAt: time.Date(2024, 12, 15, 14, 30, 0, 0, time.Local),
	}

	output := FormatPositionForTimeline(pos)
	if !strings.Contains(output, "40.7128") {
		t.Error("expected output to contain latitude")
	}
}

func TestFormatPositionForTimeline_NilPosition(t *testing.T) {
	output := FormatPositionForTimeline(nil)
	if !strings.Contains(output, "no position") {
		t.Errorf("expected nil position message, got %q", output)
	}
}

func TestFormatItemWithPosition(t *testing.T) {
	item := &models.Item{
		ID:        uuid.New(),
		Name:      "harper",
		CreatedAt: time.Now(),
	}
	label := "home"
	pos := &models.Position{
		ID:         uuid.New(),
		Latitude:   41.8781,
		Longitude:  -87.6298,
		Label:      &label,
		RecordedAt: time.Now(),
	}

	output := FormatItemWithPosition(item, pos)
	if !strings.Contains(output, "harper") {
		t.Error("expected output to contain item name")
	}
	if !strings.Contains(output, "home") {
		t.Error("expected output to contain position label")
	}
}

func TestFormatItemWithPosition_NoPosition(t *testing.T) {
	item := &models.Item{
		ID:        uuid.New(),
		Name:      "car",
		CreatedAt: time.Now(),
	}

	output := FormatItemWithPosition(item, nil)
	if !strings.Contains(output, "car") {
		t.Error("expected output to contain item name")
	}
	if !strings.Contains(output, "no position") {
		t.Error("expected 'no position' message")
	}
}

func TestFormatItemWithPosition_NilItem(t *testing.T) {
	output := FormatItemWithPosition(nil, nil)
	if !strings.Contains(output, "invalid item") {
		t.Errorf("expected invalid item message, got %q", output)
	}
}

func TestFormatItemWithPosition_NoLabel(t *testing.T) {
	item := &models.Item{
		ID:        uuid.New(),
		Name:      "bike",
		CreatedAt: time.Now(),
	}
	pos := &models.Position{
		ID:         uuid.New(),
		Latitude:   41.8781,
		Longitude:  -87.6298,
		Label:      nil,
		RecordedAt: time.Now(),
	}

	output := FormatItemWithPosition(item, pos)
	if !strings.Contains(output, "bike") {
		t.Error("expected output to contain item name")
	}
	if !strings.Contains(output, "41.8781") {
		t.Error("expected output to contain coordinates")
	}
}

func TestFormatRelativeTime(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		contains string
	}{
		{"just_now", 30 * time.Second, "just now"},
		{"one_minute", 1 * time.Minute, "1 minute ago"},
		{"five_minutes", 5 * time.Minute, "5 minutes ago"},
		{"one_hour", 1 * time.Hour, "1 hour ago"},
		{"two_hours", 2 * time.Hour, "2 hours ago"},
		{"one_day", 25 * time.Hour, "1 day ago"},
		{"multiple_days", 72 * time.Hour, "3 days ago"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tm := time.Now().Add(-tc.duration)
			result := FormatRelativeTime(tm)
			if !strings.Contains(result, tc.contains) {
				t.Errorf("FormatRelativeTime for %v: expected to contain %q, got %q", tc.duration, tc.contains, result)
			}
		})
	}
}

func TestFormatRelativeTime_FutureTime(t *testing.T) {
	futureTime := time.Now().Add(1 * time.Hour)
	result := FormatRelativeTime(futureTime)
	if !strings.Contains(result, "future") {
		t.Errorf("expected future time message, got %q", result)
	}
}

func TestFormatRelativeTime_EdgeCases(t *testing.T) {
	// Test just under one minute
	tm := time.Now().Add(-59 * time.Second)
	result := FormatRelativeTime(tm)
	if !strings.Contains(result, "just now") {
		t.Errorf("59 seconds ago should be 'just now', got %q", result)
	}

	// Test exactly one minute
	tm = time.Now().Add(-60 * time.Second)
	result = FormatRelativeTime(tm)
	if !strings.Contains(result, "minute") {
		t.Errorf("60 seconds ago should contain 'minute', got %q", result)
	}

	// Test 59 minutes
	tm = time.Now().Add(-59 * time.Minute)
	result = FormatRelativeTime(tm)
	if !strings.Contains(result, "59 minutes") {
		t.Errorf("59 minutes ago should be '59 minutes ago', got %q", result)
	}

	// Test 23 hours
	tm = time.Now().Add(-23 * time.Hour)
	result = FormatRelativeTime(tm)
	if !strings.Contains(result, "23 hours") {
		t.Errorf("23 hours ago should be '23 hours ago', got %q", result)
	}
}

func TestFormatPositionForTimeline_EmptyLabel(t *testing.T) {
	empty := ""
	pos := &models.Position{
		ID:         uuid.New(),
		Latitude:   40.7128,
		Longitude:  -74.0060,
		Label:      &empty,
		RecordedAt: time.Date(2024, 12, 15, 14, 30, 0, 0, time.Local),
	}

	output := FormatPositionForTimeline(pos)
	// Should show coordinates when label is empty
	if !strings.Contains(output, "40.7128") {
		t.Error("expected output to contain latitude for empty label")
	}
}

func TestFormatItemWithPosition_EmptyLabel(t *testing.T) {
	item := &models.Item{
		ID:        uuid.New(),
		Name:      "test",
		CreatedAt: time.Now(),
	}
	empty := ""
	pos := &models.Position{
		ID:         uuid.New(),
		Latitude:   41.8781,
		Longitude:  -87.6298,
		Label:      &empty,
		RecordedAt: time.Now(),
	}

	output := FormatItemWithPosition(item, pos)
	if !strings.Contains(output, "test") {
		t.Error("expected output to contain item name")
	}
	// Should show coordinates when label is empty
	if !strings.Contains(output, "41.8781") {
		t.Error("expected output to contain coordinates for empty label")
	}
}
