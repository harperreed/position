// ABOUTME: Item CRUD operations using Charm KV
// ABOUTME: Handles creation, retrieval, listing, and deletion of tracked items

package charm

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/charmbracelet/charm/kv"
	"github.com/google/uuid"
	"github.com/harper/position/internal/models"
	"github.com/harper/position/internal/storage"
)

// Compile-time check that Client implements storage.Repository.
var _ storage.Repository = (*Client)(nil)

// ErrNotFound is returned when an item or position is not found.
var ErrNotFound = errors.New("not found")

// CreateItem creates a new item in the KV store.
func (c *Client) CreateItem(item *models.Item) error {
	if c.kv.IsReadOnly() {
		return fmt.Errorf("cannot write: database is locked by another process (MCP server?)")
	}

	key := fmt.Sprintf("%s%s", ItemPrefix, item.ID.String())
	data, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("marshal item: %w", err)
	}

	if err := c.kv.Set([]byte(key), data); err != nil {
		return fmt.Errorf("set item: %w", err)
	}

	c.syncIfEnabled()
	return nil
}

// GetItemByID retrieves an item by its UUID.
func (c *Client) GetItemByID(id uuid.UUID) (*models.Item, error) {
	key := fmt.Sprintf("%s%s", ItemPrefix, id.String())
	data, err := c.kv.Get([]byte(key))
	if err != nil {
		if errors.Is(err, kv.ErrMissingKey) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get item: %w", err)
	}

	var item models.Item
	if err := json.Unmarshal(data, &item); err != nil {
		return nil, fmt.Errorf("unmarshal item: %w", err)
	}

	return &item, nil
}

// GetItemByName retrieves an item by its name.
// This requires a full scan since we're filtering by name.
func (c *Client) GetItemByName(name string) (*models.Item, error) {
	items, err := c.ListItems()
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		if item.Name == name {
			return item, nil
		}
	}

	return nil, ErrNotFound
}

// ListItems returns all items sorted by name.
func (c *Client) ListItems() ([]*models.Item, error) {
	items := []*models.Item{}
	prefix := []byte(ItemPrefix)

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
			return nil, fmt.Errorf("get item %s: %w", key, err)
		}

		var item models.Item
		if err := json.Unmarshal(data, &item); err != nil {
			return nil, fmt.Errorf("unmarshal item: %w", err)
		}
		items = append(items, &item)
	}

	// Sort by name
	sort.Slice(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})

	return items, nil
}

// DeleteItem removes an item and all its positions (cascade delete).
func (c *Client) DeleteItem(id uuid.UUID) error {
	if c.kv.IsReadOnly() {
		return fmt.Errorf("cannot write: database is locked by another process (MCP server?)")
	}

	// First, delete all positions for this item
	if err := c.deletePositionsForItem(id); err != nil {
		return fmt.Errorf("delete positions: %w", err)
	}

	// Then delete the item itself
	key := fmt.Sprintf("%s%s", ItemPrefix, id.String())
	if err := c.kv.Delete([]byte(key)); err != nil {
		return fmt.Errorf("delete item: %w", err)
	}

	c.syncIfEnabled()
	return nil
}

// deletePositionsForItem removes all positions associated with an item.
func (c *Client) deletePositionsForItem(itemID uuid.UUID) error {
	positions, err := c.GetTimeline(itemID)
	if err != nil {
		return err
	}

	for _, pos := range positions {
		key := fmt.Sprintf("%s%s", PositionPrefix, pos.ID.String())
		if err := c.kv.Delete([]byte(key)); err != nil {
			return fmt.Errorf("delete position %s: %w", pos.ID, err)
		}
	}

	return nil
}
