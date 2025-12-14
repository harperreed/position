// ABOUTME: GeoJSON generation utilities
// ABOUTME: Converts positions to GeoJSON FeatureCollections

package geojson

import (
	"encoding/json"
	"time"

	"github.com/harper/position/internal/models"
)

// FeatureCollection represents a GeoJSON FeatureCollection.
type FeatureCollection struct {
	Type     string    `json:"type"`
	Features []Feature `json:"features"`
}

// Feature represents a GeoJSON Feature.
type Feature struct {
	Type       string                 `json:"type"`
	Geometry   Geometry               `json:"geometry"`
	Properties map[string]interface{} `json:"properties"`
}

// Geometry represents a GeoJSON Geometry.
type Geometry struct {
	Type        string      `json:"type"`
	Coordinates interface{} `json:"coordinates"`
}

// PointCoordinates represents [longitude, latitude] for a Point.
type PointCoordinates [2]float64

// LineCoordinates represents [[lng, lat], [lng, lat], ...] for a LineString.
type LineCoordinates []PointCoordinates

// ItemNameResolver is a function that resolves an item ID to its name.
type ItemNameResolver func(itemID string) string

// ToPointsFeatureCollection converts positions to a FeatureCollection of Points.
func ToPointsFeatureCollection(positions []*models.Position, nameResolver ItemNameResolver) *FeatureCollection {
	features := make([]Feature, 0, len(positions))

	for _, pos := range positions {
		name := ""
		if nameResolver != nil {
			name = nameResolver(pos.ItemID.String())
		}

		props := map[string]interface{}{
			"name":        name,
			"recorded_at": pos.RecordedAt.Format(time.RFC3339),
		}
		if pos.Label != nil {
			props["label"] = *pos.Label
		}

		features = append(features, Feature{
			Type: "Feature",
			Geometry: Geometry{
				Type:        "Point",
				Coordinates: PointCoordinates{pos.Longitude, pos.Latitude},
			},
			Properties: props,
		})
	}

	return &FeatureCollection{
		Type:     "FeatureCollection",
		Features: features,
	}
}

// ToLineFeatureCollection converts positions to a FeatureCollection of LineStrings.
// Positions are grouped by item and sorted chronologically.
func ToLineFeatureCollection(positions []*models.Position, nameResolver ItemNameResolver) *FeatureCollection {
	// Group positions by item ID
	byItem := make(map[string][]*models.Position)
	for _, pos := range positions {
		key := pos.ItemID.String()
		byItem[key] = append(byItem[key], pos)
	}

	features := make([]Feature, 0, len(byItem))

	for itemID, itemPositions := range byItem {
		if len(itemPositions) < 2 {
			// Need at least 2 points for a line
			continue
		}

		name := ""
		if nameResolver != nil {
			name = nameResolver(itemID)
		}

		coords := make(LineCoordinates, len(itemPositions))
		for i, pos := range itemPositions {
			coords[i] = PointCoordinates{pos.Longitude, pos.Latitude}
		}

		features = append(features, Feature{
			Type: "Feature",
			Geometry: Geometry{
				Type:        "LineString",
				Coordinates: coords,
			},
			Properties: map[string]interface{}{
				"name":        name,
				"point_count": len(itemPositions),
			},
		})
	}

	return &FeatureCollection{
		Type:     "FeatureCollection",
		Features: features,
	}
}

// ToJSON serializes a FeatureCollection to JSON.
func (fc *FeatureCollection) ToJSON() ([]byte, error) {
	return json.Marshal(fc)
}

// ToJSONIndent serializes a FeatureCollection to indented JSON.
func (fc *FeatureCollection) ToJSONIndent() ([]byte, error) {
	return json.MarshalIndent(fc, "", "  ")
}
