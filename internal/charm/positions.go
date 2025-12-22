// ABOUTME: Position CRUD operations using Charm KV
// ABOUTME: Handles creation, retrieval, timeline queries, and deletion of positions

package charm

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/charmbracelet/charm/kv"
	"github.com/google/uuid"
	"github.com/harper/position/internal/models"
)

// coordEpsilon defines the threshold for considering coordinates equal.
// 0.0000001 degrees ≈ 1.1cm at the equator, sufficient for GPS deduplication.
const coordEpsilon = 0.0000001

// coordsEqual compares two coordinate pairs using epsilon for floating-point safety.
func coordsEqual(lat1, lng1, lat2, lng2 float64) bool {
	return math.Abs(lat1-lat2) < coordEpsilon && math.Abs(lng1-lng2) < coordEpsilon
}

// CreatePosition creates a new position in the KV store.
// Deduplicates: if the new position matches the current position for the item,
// it's silently swallowed (returns nil without storing).
func (c *Client) CreatePosition(pos *models.Position) error {
	if c.kv.IsReadOnly() {
		return fmt.Errorf("cannot write: database is locked by another process (MCP server?)")
	}

	// Check if this is a duplicate of the current position
	current, err := c.GetCurrentPosition(pos.ItemID)
	if err == nil && coordsEqual(current.Latitude, current.Longitude, pos.Latitude, pos.Longitude) {
		// Same location as current position - swallow it
		return nil
	}

	key := fmt.Sprintf("%s%s", PositionPrefix, pos.ID.String())
	data, err := json.Marshal(pos)
	if err != nil {
		return fmt.Errorf("marshal position: %w", err)
	}

	if err := c.kv.Set([]byte(key), data); err != nil {
		return fmt.Errorf("set position: %w", err)
	}

	c.syncIfEnabled()
	return nil
}

// GetPosition retrieves a position by its UUID.
func (c *Client) GetPosition(id uuid.UUID) (*models.Position, error) {
	key := fmt.Sprintf("%s%s", PositionPrefix, id.String())
	data, err := c.kv.Get([]byte(key))
	if err != nil {
		if errors.Is(err, kv.ErrMissingKey) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get position: %w", err)
	}

	var pos models.Position
	if err := json.Unmarshal(data, &pos); err != nil {
		return nil, fmt.Errorf("unmarshal position: %w", err)
	}

	return &pos, nil
}

// GetCurrentPosition returns the most recent position for an item.
func (c *Client) GetCurrentPosition(itemID uuid.UUID) (*models.Position, error) {
	positions, err := c.GetTimeline(itemID)
	if err != nil {
		return nil, err
	}

	if len(positions) == 0 {
		return nil, ErrNotFound
	}

	// Timeline is sorted newest first
	return positions[0], nil
}

// GetTimeline returns all positions for an item, sorted by recorded_at descending (newest first).
func (c *Client) GetTimeline(itemID uuid.UUID) ([]*models.Position, error) {
	positions := []*models.Position{}
	prefix := []byte(PositionPrefix)

	keys, err := c.kv.Keys()
	if err != nil {
		return nil, fmt.Errorf("list keys: %w", err)
	}

	for _, key := range keys {
		if !bytes.HasPrefix(key, prefix) {
			continue
		}

		data, err := c.kv.Get(key)
		if err != nil {
			return nil, fmt.Errorf("get position %s: %w", key, err)
		}

		var pos models.Position
		if err := json.Unmarshal(data, &pos); err != nil {
			return nil, fmt.Errorf("unmarshal position: %w", err)
		}

		// Filter by item ID
		if pos.ItemID == itemID {
			positions = append(positions, &pos)
		}
	}

	// Sort by recorded_at descending (newest first)
	sort.Slice(positions, func(i, j int) bool {
		return positions[i].RecordedAt.After(positions[j].RecordedAt)
	})

	return positions, nil
}

// GetPositionsSince returns positions for an item recorded after the given time.
func (c *Client) GetPositionsSince(itemID uuid.UUID, since time.Time) ([]*models.Position, error) {
	allPositions, err := c.GetTimeline(itemID)
	if err != nil {
		return nil, err
	}

	filtered := []*models.Position{}
	for _, pos := range allPositions {
		if pos.RecordedAt.After(since) {
			filtered = append(filtered, pos)
		}
	}

	return filtered, nil
}

// GetPositionsInRange returns positions for an item within a time range.
func (c *Client) GetPositionsInRange(itemID uuid.UUID, from, to time.Time) ([]*models.Position, error) {
	allPositions, err := c.GetTimeline(itemID)
	if err != nil {
		return nil, err
	}

	filtered := []*models.Position{}
	for _, pos := range allPositions {
		if (pos.RecordedAt.Equal(from) || pos.RecordedAt.After(from)) &&
			(pos.RecordedAt.Equal(to) || pos.RecordedAt.Before(to)) {
			filtered = append(filtered, pos)
		}
	}

	return filtered, nil
}

// GetAllPositions returns all positions across all items.
func (c *Client) GetAllPositions() ([]*models.Position, error) {
	positions := []*models.Position{}
	prefix := []byte(PositionPrefix)

	keys, err := c.kv.Keys()
	if err != nil {
		return nil, fmt.Errorf("list keys: %w", err)
	}

	for _, key := range keys {
		if !bytes.HasPrefix(key, prefix) {
			continue
		}

		data, err := c.kv.Get(key)
		if err != nil {
			return nil, fmt.Errorf("get position %s: %w", key, err)
		}

		var pos models.Position
		if err := json.Unmarshal(data, &pos); err != nil {
			return nil, fmt.Errorf("unmarshal position: %w", err)
		}
		positions = append(positions, &pos)
	}

	// Sort by recorded_at descending
	sort.Slice(positions, func(i, j int) bool {
		return positions[i].RecordedAt.After(positions[j].RecordedAt)
	})

	return positions, nil
}

// GetAllPositionsSince returns all positions across all items after the given time.
func (c *Client) GetAllPositionsSince(since time.Time) ([]*models.Position, error) {
	allPositions, err := c.GetAllPositions()
	if err != nil {
		return nil, err
	}

	filtered := []*models.Position{}
	for _, pos := range allPositions {
		if pos.RecordedAt.After(since) {
			filtered = append(filtered, pos)
		}
	}

	return filtered, nil
}

// GetAllPositionsInRange returns all positions across all items within a time range.
func (c *Client) GetAllPositionsInRange(from, to time.Time) ([]*models.Position, error) {
	allPositions, err := c.GetAllPositions()
	if err != nil {
		return nil, err
	}

	filtered := []*models.Position{}
	for _, pos := range allPositions {
		if (pos.RecordedAt.Equal(from) || pos.RecordedAt.After(from)) &&
			(pos.RecordedAt.Equal(to) || pos.RecordedAt.Before(to)) {
			filtered = append(filtered, pos)
		}
	}

	return filtered, nil
}

// DeletePosition removes a single position.
func (c *Client) DeletePosition(id uuid.UUID) error {
	if c.kv.IsReadOnly() {
		return fmt.Errorf("cannot write: database is locked by another process (MCP server?)")
	}

	key := fmt.Sprintf("%s%s", PositionPrefix, id.String())
	if err := c.kv.Delete([]byte(key)); err != nil {
		return fmt.Errorf("delete position: %w", err)
	}

	c.syncIfEnabled()
	return nil
}
