// ABOUTME: Tests for applying remote changes to local database
// ABOUTME: Verifies item and position change application, including edge cases

package sync

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/harperreed/sweet/vault"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyItemChangeUpsert(t *testing.T) {
	ctx := context.Background()
	syncer, appDB, cleanup := setupTestSyncerWithDB(t)
	defer cleanup()

	itemID := uuid.New()
	createdAt := time.Now().UTC().Unix()

	payload := map[string]any{
		"name":       "test-item",
		"created_at": createdAt,
	}

	payloadBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	change := vault.Change{
		Entity:   EntityItem,
		EntityID: itemID.String(),
		Op:       vault.OpUpsert,
		Payload:  payloadBytes,
	}

	err = syncer.applyItemChange(ctx, change)
	require.NoError(t, err)

	// Verify item was created
	var name string
	var dbCreatedAt time.Time
	err = appDB.QueryRowContext(ctx,
		`SELECT name, created_at FROM items WHERE id = ?`,
		itemID.String()).Scan(&name, &dbCreatedAt)
	require.NoError(t, err)
	assert.Equal(t, "test-item", name)
	assert.Equal(t, time.Unix(createdAt, 0).Unix(), dbCreatedAt.Unix())
}

func TestApplyItemChangeUpdate(t *testing.T) {
	ctx := context.Background()
	syncer, appDB, cleanup := setupTestSyncerWithDB(t)
	defer cleanup()

	itemID := uuid.New()

	// Insert initial item
	_, err := appDB.ExecContext(ctx,
		`INSERT INTO items (id, name, created_at) VALUES (?, ?, ?)`,
		itemID.String(), "original-name", time.Now())
	require.NoError(t, err)

	// Apply update
	createdAt := time.Now().UTC().Unix()
	payload := map[string]any{
		"name":       "updated-name",
		"created_at": createdAt,
	}

	payloadBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	change := vault.Change{
		Entity:   EntityItem,
		EntityID: itemID.String(),
		Op:       vault.OpUpsert,
		Payload:  payloadBytes,
	}

	err = syncer.applyItemChange(ctx, change)
	require.NoError(t, err)

	// Verify item was updated
	var name string
	err = appDB.QueryRowContext(ctx,
		`SELECT name FROM items WHERE id = ?`,
		itemID.String()).Scan(&name)
	require.NoError(t, err)
	assert.Equal(t, "updated-name", name)
}

func TestApplyItemChangeDelete(t *testing.T) {
	ctx := context.Background()
	syncer, appDB, cleanup := setupTestSyncerWithDB(t)
	defer cleanup()

	itemID := uuid.New()

	// Insert item
	_, err := appDB.ExecContext(ctx,
		`INSERT INTO items (id, name, created_at) VALUES (?, ?, ?)`,
		itemID.String(), "test-item", time.Now())
	require.NoError(t, err)

	// Apply delete
	change := vault.Change{
		Entity:   EntityItem,
		EntityID: itemID.String(),
		Op:       vault.OpDelete,
		Deleted:  true,
	}

	err = syncer.applyItemChange(ctx, change)
	require.NoError(t, err)

	// Verify item was deleted
	var count int
	err = appDB.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM items WHERE id = ?`,
		itemID.String()).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestApplyItemChangeDeleteWithDeletedFlag(t *testing.T) {
	ctx := context.Background()
	syncer, appDB, cleanup := setupTestSyncerWithDB(t)
	defer cleanup()

	itemID := uuid.New()

	// Insert item
	_, err := appDB.ExecContext(ctx,
		`INSERT INTO items (id, name, created_at) VALUES (?, ?, ?)`,
		itemID.String(), "test-item", time.Now())
	require.NoError(t, err)

	// Apply delete with Deleted flag
	change := vault.Change{
		Entity:   EntityItem,
		EntityID: itemID.String(),
		Op:       vault.OpUpsert,
		Deleted:  true,
	}

	err = syncer.applyItemChange(ctx, change)
	require.NoError(t, err)

	// Verify item was deleted
	var count int
	err = appDB.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM items WHERE id = ?`,
		itemID.String()).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestApplyPositionChangeUpsert(t *testing.T) {
	ctx := context.Background()
	syncer, appDB, cleanup := setupTestSyncerWithDB(t)
	defer cleanup()

	// Create item first
	itemID := uuid.New()
	_, err := appDB.ExecContext(ctx,
		`INSERT INTO items (id, name, created_at) VALUES (?, ?, ?)`,
		itemID.String(), "test-item", time.Now())
	require.NoError(t, err)

	posID := uuid.New()
	recordedAt := time.Now().UTC().Unix()
	label := "test location"

	payload := map[string]any{
		"item_name":   "test-item",
		"latitude":    37.7749,
		"longitude":   -122.4194,
		"label":       label,
		"recorded_at": recordedAt,
	}

	payloadBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	change := vault.Change{
		Entity:   EntityPosition,
		EntityID: posID.String(),
		Op:       vault.OpUpsert,
		Payload:  payloadBytes,
	}

	err = syncer.applyPositionChange(ctx, change)
	require.NoError(t, err)

	// Verify position was created
	var lat, lng float64
	var dbLabel *string
	var dbRecordedAt time.Time
	err = appDB.QueryRowContext(ctx,
		`SELECT latitude, longitude, label, recorded_at FROM positions WHERE id = ?`,
		posID.String()).Scan(&lat, &lng, &dbLabel, &dbRecordedAt)
	require.NoError(t, err)
	assert.Equal(t, 37.7749, lat)
	assert.Equal(t, -122.4194, lng)
	require.NotNil(t, dbLabel)
	assert.Equal(t, label, *dbLabel)
	assert.Equal(t, time.Unix(recordedAt, 0).Unix(), dbRecordedAt.Unix())
}

func TestApplyPositionChangeWithoutLabel(t *testing.T) {
	ctx := context.Background()
	syncer, appDB, cleanup := setupTestSyncerWithDB(t)
	defer cleanup()

	// Create item first
	itemID := uuid.New()
	_, err := appDB.ExecContext(ctx,
		`INSERT INTO items (id, name, created_at) VALUES (?, ?, ?)`,
		itemID.String(), "test-item", time.Now())
	require.NoError(t, err)

	posID := uuid.New()
	recordedAt := time.Now().UTC().Unix()

	payload := map[string]any{
		"item_name":   "test-item",
		"latitude":    40.7128,
		"longitude":   -74.0060,
		"recorded_at": recordedAt,
	}

	payloadBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	change := vault.Change{
		Entity:   EntityPosition,
		EntityID: posID.String(),
		Op:       vault.OpUpsert,
		Payload:  payloadBytes,
	}

	err = syncer.applyPositionChange(ctx, change)
	require.NoError(t, err)

	// Verify position was created without label
	var lat, lng float64
	var dbLabel *string
	err = appDB.QueryRowContext(ctx,
		`SELECT latitude, longitude, label FROM positions WHERE id = ?`,
		posID.String()).Scan(&lat, &lng, &dbLabel)
	require.NoError(t, err)
	assert.Equal(t, 40.7128, lat)
	assert.Equal(t, -74.0060, lng)
	assert.Nil(t, dbLabel)
}

func TestApplyPositionChangeCreatesItem(t *testing.T) {
	ctx := context.Background()
	syncer, appDB, cleanup := setupTestSyncerWithDB(t)
	defer cleanup()

	posID := uuid.New()
	recordedAt := time.Now().UTC().Unix()

	payload := map[string]any{
		"item_name":   "new-item",
		"latitude":    35.6762,
		"longitude":   139.6503,
		"recorded_at": recordedAt,
	}

	payloadBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	change := vault.Change{
		Entity:   EntityPosition,
		EntityID: posID.String(),
		Op:       vault.OpUpsert,
		Payload:  payloadBytes,
	}

	err = syncer.applyPositionChange(ctx, change)
	require.NoError(t, err)

	// Verify item was created
	var itemName string
	err = appDB.QueryRowContext(ctx,
		`SELECT name FROM items WHERE name = ?`,
		"new-item").Scan(&itemName)
	require.NoError(t, err)
	assert.Equal(t, "new-item", itemName)

	// Verify position was created with correct item reference
	var posItemID string
	err = appDB.QueryRowContext(ctx,
		`SELECT item_id FROM positions WHERE id = ?`,
		posID.String()).Scan(&posItemID)
	require.NoError(t, err)
	assert.NotEmpty(t, posItemID)
}

func TestApplyPositionChangeUpdate(t *testing.T) {
	ctx := context.Background()
	syncer, appDB, cleanup := setupTestSyncerWithDB(t)
	defer cleanup()

	// Create item
	itemID := uuid.New()
	_, err := appDB.ExecContext(ctx,
		`INSERT INTO items (id, name, created_at) VALUES (?, ?, ?)`,
		itemID.String(), "test-item", time.Now())
	require.NoError(t, err)

	// Create initial position
	posID := uuid.New()
	_, err = appDB.ExecContext(ctx,
		`INSERT INTO positions (id, item_id, latitude, longitude, recorded_at, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		posID.String(), itemID.String(), 1.0, 2.0, time.Now(), time.Now())
	require.NoError(t, err)

	// Apply update
	recordedAt := time.Now().UTC().Unix()
	payload := map[string]any{
		"item_name":   "test-item",
		"latitude":    37.7749,
		"longitude":   -122.4194,
		"label":       "updated location",
		"recorded_at": recordedAt,
	}

	payloadBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	change := vault.Change{
		Entity:   EntityPosition,
		EntityID: posID.String(),
		Op:       vault.OpUpsert,
		Payload:  payloadBytes,
	}

	err = syncer.applyPositionChange(ctx, change)
	require.NoError(t, err)

	// Verify position was updated
	var lat, lng float64
	var label *string
	err = appDB.QueryRowContext(ctx,
		`SELECT latitude, longitude, label FROM positions WHERE id = ?`,
		posID.String()).Scan(&lat, &lng, &label)
	require.NoError(t, err)
	assert.Equal(t, 37.7749, lat)
	assert.Equal(t, -122.4194, lng)
	require.NotNil(t, label)
	assert.Equal(t, "updated location", *label)
}

func TestApplyPositionChangeDelete(t *testing.T) {
	ctx := context.Background()
	syncer, appDB, cleanup := setupTestSyncerWithDB(t)
	defer cleanup()

	// Create item
	itemID := uuid.New()
	_, err := appDB.ExecContext(ctx,
		`INSERT INTO items (id, name, created_at) VALUES (?, ?, ?)`,
		itemID.String(), "test-item", time.Now())
	require.NoError(t, err)

	// Create position
	posID := uuid.New()
	_, err = appDB.ExecContext(ctx,
		`INSERT INTO positions (id, item_id, latitude, longitude, recorded_at, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		posID.String(), itemID.String(), 1.0, 2.0, time.Now(), time.Now())
	require.NoError(t, err)

	// Apply delete
	change := vault.Change{
		Entity:   EntityPosition,
		EntityID: posID.String(),
		Op:       vault.OpDelete,
		Deleted:  true,
	}

	err = syncer.applyPositionChange(ctx, change)
	require.NoError(t, err)

	// Verify position was deleted
	var count int
	err = appDB.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM positions WHERE id = ?`,
		posID.String()).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestApplyChangeUnknownEntity(t *testing.T) {
	ctx := context.Background()
	syncer, _, cleanup := setupTestSyncerWithDB(t)
	defer cleanup()

	change := vault.Change{
		Entity:   "unknown-entity",
		EntityID: uuid.New().String(),
		Op:       vault.OpUpsert,
		Payload:  []byte("{}"),
	}

	// Should not error, just ignore
	err := syncer.applyChange(ctx, change)
	assert.NoError(t, err)
}

func TestApplyChangeItem(t *testing.T) {
	ctx := context.Background()
	syncer, appDB, cleanup := setupTestSyncerWithDB(t)
	defer cleanup()

	itemID := uuid.New()
	createdAt := time.Now().UTC().Unix()

	payload := map[string]any{
		"name":       "routed-item",
		"created_at": createdAt,
	}

	payloadBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	change := vault.Change{
		Entity:   EntityItem,
		EntityID: itemID.String(),
		Op:       vault.OpUpsert,
		Payload:  payloadBytes,
	}

	err = syncer.applyChange(ctx, change)
	require.NoError(t, err)

	// Verify item was created
	var name string
	err = appDB.QueryRowContext(ctx,
		`SELECT name FROM items WHERE id = ?`,
		itemID.String()).Scan(&name)
	require.NoError(t, err)
	assert.Equal(t, "routed-item", name)
}

func TestApplyChangePosition(t *testing.T) {
	ctx := context.Background()
	syncer, appDB, cleanup := setupTestSyncerWithDB(t)
	defer cleanup()

	// Create item
	itemID := uuid.New()
	_, err := appDB.ExecContext(ctx,
		`INSERT INTO items (id, name, created_at) VALUES (?, ?, ?)`,
		itemID.String(), "test-item", time.Now())
	require.NoError(t, err)

	posID := uuid.New()
	recordedAt := time.Now().UTC().Unix()

	payload := map[string]any{
		"item_name":   "test-item",
		"latitude":    -33.8688,
		"longitude":   151.2093,
		"recorded_at": recordedAt,
	}

	payloadBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	change := vault.Change{
		Entity:   EntityPosition,
		EntityID: posID.String(),
		Op:       vault.OpUpsert,
		Payload:  payloadBytes,
	}

	err = syncer.applyChange(ctx, change)
	require.NoError(t, err)

	// Verify position was created
	var lat, lng float64
	err = appDB.QueryRowContext(ctx,
		`SELECT latitude, longitude FROM positions WHERE id = ?`,
		posID.String()).Scan(&lat, &lng)
	require.NoError(t, err)
	assert.Equal(t, -33.8688, lat)
	assert.Equal(t, 151.2093, lng)
}

func TestApplyItemChangeInvalidPayload(t *testing.T) {
	ctx := context.Background()
	syncer, _, cleanup := setupTestSyncerWithDB(t)
	defer cleanup()

	change := vault.Change{
		Entity:   EntityItem,
		EntityID: uuid.New().String(),
		Op:       vault.OpUpsert,
		Payload:  []byte("invalid json"),
	}

	err := syncer.applyItemChange(ctx, change)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal")
}

func TestApplyPositionChangeInvalidPayload(t *testing.T) {
	ctx := context.Background()
	syncer, _, cleanup := setupTestSyncerWithDB(t)
	defer cleanup()

	change := vault.Change{
		Entity:   EntityPosition,
		EntityID: uuid.New().String(),
		Op:       vault.OpUpsert,
		Payload:  []byte("invalid json"),
	}

	err := syncer.applyPositionChange(ctx, change)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal")
}

func TestApplyMultiplePositionsForSameItem(t *testing.T) {
	ctx := context.Background()
	syncer, appDB, cleanup := setupTestSyncerWithDB(t)
	defer cleanup()

	// Create item
	itemID := uuid.New()
	_, err := appDB.ExecContext(ctx,
		`INSERT INTO items (id, name, created_at) VALUES (?, ?, ?)`,
		itemID.String(), "tracked-item", time.Now())
	require.NoError(t, err)

	// Add multiple positions for the same item
	positions := []struct {
		lat   float64
		lng   float64
		label string
	}{
		{37.7749, -122.4194, "San Francisco"},
		{40.7128, -74.0060, "New York"},
		{35.6762, 139.6503, "Tokyo"},
	}

	for _, pos := range positions {
		posID := uuid.New()
		recordedAt := time.Now().UTC().Unix()

		payload := map[string]any{
			"item_name":   "tracked-item",
			"latitude":    pos.lat,
			"longitude":   pos.lng,
			"label":       pos.label,
			"recorded_at": recordedAt,
		}

		payloadBytes, err := json.Marshal(payload)
		require.NoError(t, err)

		change := vault.Change{
			Entity:   EntityPosition,
			EntityID: posID.String(),
			Op:       vault.OpUpsert,
			Payload:  payloadBytes,
		}

		err = syncer.applyPositionChange(ctx, change)
		require.NoError(t, err)
	}

	// Verify all positions were created
	var count int
	err = appDB.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM positions WHERE item_id = ?`,
		itemID.String()).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, len(positions), count)
}
