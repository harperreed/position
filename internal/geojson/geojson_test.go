// ABOUTME: Unit tests for GeoJSON generation
// ABOUTME: Tests Point and LineString feature collection builders

package geojson

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/harper/position/internal/models"
)

func TestToPointsFeatureCollection(t *testing.T) {
	label := "chicago"
	itemID := uuid.New()
	positions := []*models.Position{
		{
			ID:         uuid.New(),
			ItemID:     itemID,
			Latitude:   41.8781,
			Longitude:  -87.6298,
			Label:      &label,
			RecordedAt: time.Now(),
		},
	}

	nameResolver := func(id string) string {
		if id == itemID.String() {
			return "harper"
		}
		return ""
	}

	fc := ToPointsFeatureCollection(positions, nameResolver)

	if fc.Type != "FeatureCollection" {
		t.Errorf("expected FeatureCollection type, got %s", fc.Type)
	}
	if len(fc.Features) != 1 {
		t.Fatalf("expected 1 feature, got %d", len(fc.Features))
	}

	feature := fc.Features[0]
	if feature.Type != "Feature" {
		t.Errorf("expected Feature type, got %s", feature.Type)
	}
	if feature.Geometry.Type != "Point" {
		t.Errorf("expected Point geometry, got %s", feature.Geometry.Type)
	}

	coords, ok := feature.Geometry.Coordinates.(PointCoordinates)
	if !ok {
		t.Fatal("expected PointCoordinates")
	}
	// GeoJSON uses [lng, lat] order
	if coords[0] != -87.6298 {
		t.Errorf("expected longitude -87.6298, got %f", coords[0])
	}
	if coords[1] != 41.8781 {
		t.Errorf("expected latitude 41.8781, got %f", coords[1])
	}

	if feature.Properties["name"] != "harper" {
		t.Errorf("expected name 'harper', got %v", feature.Properties["name"])
	}
	if feature.Properties["label"] != "chicago" {
		t.Errorf("expected label 'chicago', got %v", feature.Properties["label"])
	}
}

func TestToLineFeatureCollection(t *testing.T) {
	itemID := uuid.New()
	positions := []*models.Position{
		{
			ID:         uuid.New(),
			ItemID:     itemID,
			Latitude:   41.8781,
			Longitude:  -87.6298,
			RecordedAt: time.Now().Add(-time.Hour),
		},
		{
			ID:         uuid.New(),
			ItemID:     itemID,
			Latitude:   40.7128,
			Longitude:  -74.0060,
			RecordedAt: time.Now(),
		},
	}

	nameResolver := func(id string) string {
		if id == itemID.String() {
			return "harper"
		}
		return ""
	}

	fc := ToLineFeatureCollection(positions, nameResolver)

	if fc.Type != "FeatureCollection" {
		t.Errorf("expected FeatureCollection type, got %s", fc.Type)
	}
	if len(fc.Features) != 1 {
		t.Fatalf("expected 1 feature, got %d", len(fc.Features))
	}

	feature := fc.Features[0]
	if feature.Geometry.Type != "LineString" {
		t.Errorf("expected LineString geometry, got %s", feature.Geometry.Type)
	}

	coords, ok := feature.Geometry.Coordinates.(LineCoordinates)
	if !ok {
		t.Fatal("expected LineCoordinates")
	}
	if len(coords) != 2 {
		t.Errorf("expected 2 coordinates, got %d", len(coords))
	}

	if feature.Properties["point_count"] != 2 {
		t.Errorf("expected point_count 2, got %v", feature.Properties["point_count"])
	}
}

func TestToLineFeatureCollection_SinglePoint(t *testing.T) {
	// A single point should not create a LineString
	positions := []*models.Position{
		{
			ID:         uuid.New(),
			ItemID:     uuid.New(),
			Latitude:   41.8781,
			Longitude:  -87.6298,
			RecordedAt: time.Now(),
		},
	}

	fc := ToLineFeatureCollection(positions, nil)

	if len(fc.Features) != 0 {
		t.Errorf("expected 0 features for single point, got %d", len(fc.Features))
	}
}

func TestFeatureCollection_ToJSON(t *testing.T) {
	fc := &FeatureCollection{
		Type:     "FeatureCollection",
		Features: []Feature{},
	}

	jsonBytes, err := fc.ToJSON()
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if parsed["type"] != "FeatureCollection" {
		t.Error("expected type FeatureCollection in JSON")
	}
}
