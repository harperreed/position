// ABOUTME: Repository interfaces for position tracking storage
// ABOUTME: Enables testability and storage backend swapping

package storage

import (
	"time"

	"github.com/google/uuid"
	"github.com/harper/position/internal/models"
)

// ItemRepository defines operations for managing tracked items.
type ItemRepository interface {
	CreateItem(item *models.Item) error
	GetItemByID(id uuid.UUID) (*models.Item, error)
	GetItemByName(name string) (*models.Item, error)
	ListItems() ([]*models.Item, error)
	DeleteItem(id uuid.UUID) error
}

// PositionRepository defines operations for managing positions.
type PositionRepository interface {
	CreatePosition(pos *models.Position) error
	GetPosition(id uuid.UUID) (*models.Position, error)
	GetCurrentPosition(itemID uuid.UUID) (*models.Position, error)
	GetTimeline(itemID uuid.UUID) ([]*models.Position, error)
	GetPositionsSince(itemID uuid.UUID, since time.Time) ([]*models.Position, error)
	GetPositionsInRange(itemID uuid.UUID, from, to time.Time) ([]*models.Position, error)
	GetAllPositions() ([]*models.Position, error)
	GetAllPositionsSince(since time.Time) ([]*models.Position, error)
	GetAllPositionsInRange(from, to time.Time) ([]*models.Position, error)
	DeletePosition(id uuid.UUID) error
}

// Repository combines all repository operations with lifecycle management.
type Repository interface {
	ItemRepository
	PositionRepository
	Close() error
	Sync() error
	Reset() error
	IsReadOnly() bool
}

// Compile-time interface implementation check is in the charm package:
// var _ Repository = (*charm.Client)(nil)
// This ensures the charm.Client satisfies the Repository interface.
