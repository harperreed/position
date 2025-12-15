// ABOUTME: Vault sync integration for position
// ABOUTME: Handles change queuing, syncing, and applying remote changes

package sync

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"suitesync/vault"

	"github.com/google/uuid"
)

const (
	EntityItem     = "item"
	EntityPosition = "position"
)

// Syncer manages vault sync for position data.
type Syncer struct {
	config *Config
	store  *vault.Store
	keys   vault.Keys
	client *vault.Client
	appDB  *sql.DB
}

// NewSyncer creates a new syncer from config.
func NewSyncer(cfg *Config, appDB *sql.DB) (*Syncer, error) {
	if cfg.DerivedKey == "" {
		return nil, errors.New("derived key not configured - run 'position sync login' first")
	}

	// DerivedKey is stored as hex-encoded seed
	seed, err := vault.ParseSeedPhrase(cfg.DerivedKey)
	if err != nil {
		return nil, fmt.Errorf("invalid derived key: %w", err)
	}

	keys, err := vault.DeriveKeys(seed, "", vault.DefaultKDFParams())
	if err != nil {
		return nil, fmt.Errorf("derive keys: %w", err)
	}

	store, err := vault.OpenStore(cfg.VaultDB)
	if err != nil {
		return nil, fmt.Errorf("open vault store: %w", err)
	}

	client := vault.NewClient(vault.SyncConfig{
		BaseURL:   cfg.Server,
		DeviceID:  cfg.DeviceID,
		AuthToken: cfg.Token,
	})

	return &Syncer{
		config: cfg,
		store:  store,
		keys:   keys,
		client: client,
		appDB:  appDB,
	}, nil
}

// Close releases syncer resources.
func (s *Syncer) Close() error {
	if s.store != nil {
		return s.store.Close()
	}
	return nil
}

// QueueItemChange queues a change for an item.
func (s *Syncer) QueueItemChange(ctx context.Context, itemID uuid.UUID, name string, op vault.Op) error {
	var payload map[string]any
	if op != vault.OpDelete {
		payload = map[string]any{
			"name":       name,
			"created_at": time.Now().UTC().Unix(),
		}
	}

	return s.queueChange(ctx, EntityItem, itemID.String(), op, payload)
}

// QueuePositionChange queues a change for a position.
func (s *Syncer) QueuePositionChange(ctx context.Context, posID uuid.UUID, itemName string, lat, lng float64, label *string, recordedAt time.Time, op vault.Op) error {
	var payload map[string]any
	if op != vault.OpDelete {
		payload = map[string]any{
			"item_name":   itemName,
			"latitude":    lat,
			"longitude":   lng,
			"recorded_at": recordedAt.UTC().Unix(),
		}
		if label != nil {
			payload["label"] = *label
		}
	}

	return s.queueChange(ctx, EntityPosition, posID.String(), op, payload)
}

func (s *Syncer) queueChange(ctx context.Context, entity, entityID string, op vault.Op, payload map[string]any) error {
	change, err := vault.NewChange(entity, entityID, op, payload)
	if err != nil {
		return fmt.Errorf("create change: %w", err)
	}
	if op == vault.OpDelete {
		change.Deleted = true
	}

	plain, err := json.Marshal(change)
	if err != nil {
		return fmt.Errorf("marshal change: %w", err)
	}

	aad := change.AAD(s.keys.UserID(), s.config.DeviceID)
	env, err := vault.Encrypt(s.keys.EncKey, plain, aad)
	if err != nil {
		return fmt.Errorf("encrypt change: %w", err)
	}

	if err := s.store.EnqueueEncryptedChange(ctx, change, s.keys.UserID(), s.config.DeviceID, env); err != nil {
		return fmt.Errorf("enqueue change: %w", err)
	}

	// Auto-sync if enabled
	if s.config.AutoSync && s.canSync() {
		return s.Sync(ctx)
	}

	return nil
}

func (s *Syncer) canSync() bool {
	return s.config.Server != "" && s.config.Token != "" && s.config.UserID != ""
}

// Sync pushes local changes and pulls remote changes.
func (s *Syncer) Sync(ctx context.Context) error {
	if !s.canSync() {
		return errors.New("sync not configured - run 'position sync login' first")
	}

	return vault.Sync(ctx, s.store, s.client, s.keys, s.config.UserID, s.applyChange)
}

// applyChange applies a remote change to the local database.
func (s *Syncer) applyChange(ctx context.Context, c vault.Change) error {
	switch c.Entity {
	case EntityItem:
		return s.applyItemChange(ctx, c)
	case EntityPosition:
		return s.applyPositionChange(ctx, c)
	default:
		// Ignore unknown entities
		return nil
	}
}

func (s *Syncer) applyItemChange(ctx context.Context, c vault.Change) error {
	if c.Op == vault.OpDelete || c.Deleted {
		_, err := s.appDB.ExecContext(ctx,
			`DELETE FROM items WHERE id = ?`, c.EntityID)
		return err
	}

	var payload struct {
		Name      string `json:"name"`
		CreatedAt int64  `json:"created_at"`
	}
	if err := json.Unmarshal(c.Payload, &payload); err != nil {
		return fmt.Errorf("unmarshal item payload: %w", err)
	}

	createdAt := time.Unix(payload.CreatedAt, 0)
	_, err := s.appDB.ExecContext(ctx, `
		INSERT INTO items (id, name, created_at) VALUES (?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET name = excluded.name
	`, c.EntityID, payload.Name, createdAt)

	return err
}

func (s *Syncer) applyPositionChange(ctx context.Context, c vault.Change) error {
	if c.Op == vault.OpDelete || c.Deleted {
		_, err := s.appDB.ExecContext(ctx,
			`DELETE FROM positions WHERE id = ?`, c.EntityID)
		return err
	}

	var payload struct {
		ItemName   string  `json:"item_name"`
		Latitude   float64 `json:"latitude"`
		Longitude  float64 `json:"longitude"`
		Label      *string `json:"label,omitempty"`
		RecordedAt int64   `json:"recorded_at"`
	}
	if err := json.Unmarshal(c.Payload, &payload); err != nil {
		return fmt.Errorf("unmarshal position payload: %w", err)
	}

	// Look up item by name to get ID
	var itemID string
	err := s.appDB.QueryRowContext(ctx,
		`SELECT id FROM items WHERE name = ?`, payload.ItemName).Scan(&itemID)
	if errors.Is(err, sql.ErrNoRows) {
		// Create item if it doesn't exist
		itemID = uuid.New().String()
		_, err = s.appDB.ExecContext(ctx,
			`INSERT INTO items (id, name, created_at) VALUES (?, ?, ?)`,
			itemID, payload.ItemName, time.Now())
		if err != nil {
			return fmt.Errorf("create item for position: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("lookup item: %w", err)
	}

	recordedAt := time.Unix(payload.RecordedAt, 0)
	_, err = s.appDB.ExecContext(ctx, `
		INSERT INTO positions (id, item_id, latitude, longitude, label, recorded_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			latitude = excluded.latitude,
			longitude = excluded.longitude,
			label = excluded.label,
			recorded_at = excluded.recorded_at
	`, c.EntityID, itemID, payload.Latitude, payload.Longitude, payload.Label, recordedAt, time.Now())

	return err
}

// PendingCount returns the number of changes waiting to be synced.
func (s *Syncer) PendingCount(ctx context.Context) (int, error) {
	batch, err := s.store.DequeueBatch(ctx, 1000)
	if err != nil {
		return 0, err
	}
	return len(batch), nil
}
