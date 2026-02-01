// ABOUTME: Unit tests for data models
// ABOUTME: Tests constructors, validators, and model methods

package models

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewItem(t *testing.T) {
	item := NewItem("harper")

	if item.Name != "harper" {
		t.Errorf("expected name 'harper', got '%s'", item.Name)
	}
	if item.ID == uuid.Nil {
		t.Error("expected non-nil UUID")
	}
	if item.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
}

func TestNewItem_UniqueIDs(t *testing.T) {
	item1 := NewItem("item1")
	item2 := NewItem("item2")

	if item1.ID == item2.ID {
		t.Error("expected unique IDs for different items")
	}
}

func TestNewPosition(t *testing.T) {
	itemID := uuid.New()
	lat := 41.8781
	lng := -87.6298
	label := "chicago"

	pos := NewPosition(itemID, lat, lng, &label)

	if pos.ItemID != itemID {
		t.Error("item ID mismatch")
	}
	if pos.Latitude != lat {
		t.Errorf("expected lat %f, got %f", lat, pos.Latitude)
	}
	if pos.Longitude != lng {
		t.Errorf("expected lng %f, got %f", lng, pos.Longitude)
	}
	if pos.Label == nil || *pos.Label != label {
		t.Error("expected label 'chicago'")
	}
	if pos.ID == uuid.Nil {
		t.Error("expected non-nil UUID")
	}
}

func TestNewPosition_NilLabel(t *testing.T) {
	pos := NewPosition(uuid.New(), 0, 0, nil)
	if pos.Label != nil {
		t.Error("expected nil label")
	}
}

func TestNewPosition_SetsTimestamps(t *testing.T) {
	before := time.Now()
	pos := NewPosition(uuid.New(), 0, 0, nil)
	after := time.Now()

	if pos.RecordedAt.Before(before) || pos.RecordedAt.After(after) {
		t.Error("RecordedAt should be between before and after test times")
	}
	if pos.CreatedAt.Before(before) || pos.CreatedAt.After(after) {
		t.Error("CreatedAt should be between before and after test times")
	}
}

func TestNewPositionWithRecordedAt(t *testing.T) {
	itemID := uuid.New()
	recordedAt := time.Date(2024, 12, 14, 15, 0, 0, 0, time.UTC)

	pos := NewPositionWithRecordedAt(itemID, 41.8781, -87.6298, nil, recordedAt)

	if !pos.RecordedAt.Equal(recordedAt) {
		t.Errorf("expected recordedAt %v, got %v", recordedAt, pos.RecordedAt)
	}
}

func TestNewPositionWithRecordedAt_CreatedAtStillNow(t *testing.T) {
	recordedAt := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	before := time.Now()
	pos := NewPositionWithRecordedAt(uuid.New(), 0, 0, nil, recordedAt)
	after := time.Now()

	// RecordedAt should be the provided time
	if !pos.RecordedAt.Equal(recordedAt) {
		t.Errorf("expected recordedAt %v, got %v", recordedAt, pos.RecordedAt)
	}
	// CreatedAt should be now
	if pos.CreatedAt.Before(before) || pos.CreatedAt.After(after) {
		t.Error("CreatedAt should be between before and after test times")
	}
}

func TestValidateCoordinates(t *testing.T) {
	tests := []struct {
		name    string
		lat     float64
		lng     float64
		wantErr bool
	}{
		{"valid_chicago", 41.8781, -87.6298, false},
		{"valid_origin", 0, 0, false},
		{"valid_north_pole", 90, 0, false},
		{"valid_south_pole", -90, 0, false},
		{"valid_antimeridian_east", 0, 180, false},
		{"valid_antimeridian_west", 0, -180, false},
		{"invalid_lat_too_high", 91, 0, true},
		{"invalid_lat_too_low", -91, 0, true},
		{"invalid_lng_too_high", 0, 181, true},
		{"invalid_lng_too_low", 0, -181, true},
		{"invalid_lat_nan", math.NaN(), 0, true},
		{"invalid_lng_nan", 0, math.NaN(), true},
		{"invalid_lat_inf", math.Inf(1), 0, true},
		{"invalid_lat_neg_inf", math.Inf(-1), 0, true},
		{"invalid_lng_inf", 0, math.Inf(1), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCoordinates(tt.lat, tt.lng)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCoordinates(%f, %f) error = %v, wantErr %v", tt.lat, tt.lng, err, tt.wantErr)
			}
		})
	}
}

func TestValidateName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid_simple", "harper", false},
		{"valid_with_spaces", "my car", false},
		{"valid_single_char", "a", false},
		{"invalid_empty", "", true},
		{"invalid_whitespace_only", "   ", true},
		{"invalid_tabs_only", "\t\t", true},
		{"invalid_newlines_only", "\n\n", true},
		{"valid_max_length", strings.Repeat("a", 255), false},
		{"invalid_too_long", strings.Repeat("a", 256), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateName_ErrorMessages(t *testing.T) {
	err := ValidateName("")
	if err == nil || !strings.Contains(err.Error(), "empty") {
		t.Errorf("expected error about empty name, got %v", err)
	}

	err = ValidateName(strings.Repeat("a", 300))
	if err == nil || !strings.Contains(err.Error(), "too long") {
		t.Errorf("expected error about length, got %v", err)
	}
}

func TestValidateCoordinates_ErrorMessages(t *testing.T) {
	err := ValidateCoordinates(math.NaN(), 0)
	if err == nil || !strings.Contains(err.Error(), "NaN") {
		t.Errorf("expected error about NaN, got %v", err)
	}

	err = ValidateCoordinates(math.Inf(1), 0)
	if err == nil || !strings.Contains(err.Error(), "infinite") {
		t.Errorf("expected error about infinite, got %v", err)
	}

	err = ValidateCoordinates(100, 0)
	if err == nil || !strings.Contains(err.Error(), "latitude") {
		t.Errorf("expected error about latitude, got %v", err)
	}

	err = ValidateCoordinates(0, 200)
	if err == nil || !strings.Contains(err.Error(), "longitude") {
		t.Errorf("expected error about longitude, got %v", err)
	}
}

func TestPosition_UniqueIDs(t *testing.T) {
	itemID := uuid.New()
	pos1 := NewPosition(itemID, 0, 0, nil)
	pos2 := NewPosition(itemID, 0, 0, nil)

	if pos1.ID == pos2.ID {
		t.Error("expected unique IDs for different positions")
	}
}

func TestPosition_WithLabel(t *testing.T) {
	label := "home"
	pos := NewPosition(uuid.New(), 41.0, -87.0, &label)

	if pos.Label == nil {
		t.Fatal("label should not be nil")
	}
	if *pos.Label != "home" {
		t.Errorf("expected label 'home', got '%s'", *pos.Label)
	}
}

func TestPosition_EmptyLabel(t *testing.T) {
	label := ""
	pos := NewPosition(uuid.New(), 41.0, -87.0, &label)

	if pos.Label == nil {
		t.Fatal("label should not be nil even when empty")
	}
	if *pos.Label != "" {
		t.Errorf("expected empty label, got '%s'", *pos.Label)
	}
}
