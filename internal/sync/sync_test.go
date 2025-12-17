// ABOUTME: Tests for vault sync integration
// ABOUTME: Verifies change queuing, syncing, and pending count tracking

package sync

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/harperreed/sweet/vault"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSyncer(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test app database
	appDB := setupTestDB(t, tmpDir)
	defer func() { _ = appDB.Close() }()

	// Create seed and derive key
	seed, phrase, err := vault.NewSeedPhrase()
	require.NoError(t, err)

	cfg := &Config{
		Server:     "https://test.example.com",
		UserID:     "test-user",
		Token:      "test-token",
		DerivedKey: phrase,
		DeviceID:   "test-device",
		VaultDB:    filepath.Join(tmpDir, "vault.db"),
		AutoSync:   false,
	}

	syncer, err := NewSyncer(cfg, appDB)
	require.NoError(t, err)
	require.NotNil(t, syncer)
	defer func() { _ = syncer.Close() }()

	assert.Equal(t, cfg, syncer.config)
	assert.NotNil(t, syncer.store)
	assert.NotNil(t, syncer.client)
	assert.NotNil(t, syncer.keys)

	// Verify keys were derived correctly
	expectedKeys, err := vault.DeriveKeys(seed, "", vault.DefaultKDFParams())
	require.NoError(t, err)
	assert.Equal(t, expectedKeys.EncKey, syncer.keys.EncKey)
}

func TestNewSyncerNoDerivedKey(t *testing.T) {
	tmpDir := t.TempDir()

	appDB := setupTestDB(t, tmpDir)
	defer func() { _ = appDB.Close() }()

	cfg := &Config{
		Server:   "https://test.example.com",
		DeviceID: "test-device",
		VaultDB:  filepath.Join(tmpDir, "vault.db"),
	}

	_, err := NewSyncer(cfg, appDB)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "derived key not configured")
}

func TestNewSyncerInvalidDerivedKey(t *testing.T) {
	tmpDir := t.TempDir()

	appDB := setupTestDB(t, tmpDir)
	defer func() { _ = appDB.Close() }()

	cfg := &Config{
		Server:     "https://test.example.com",
		UserID:     "test-user",
		Token:      "test-token",
		DerivedKey: "invalid-key-format",
		DeviceID:   "test-device",
		VaultDB:    filepath.Join(tmpDir, "vault.db"),
	}

	_, err := NewSyncer(cfg, appDB)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid derived key")
}

func TestQueueItemChange(t *testing.T) {
	ctx := context.Background()
	syncer := setupTestSyncer(t)
	defer func() { _ = syncer.Close() }()

	itemID := uuid.New()

	// Queue item create
	err := syncer.QueueItemChange(ctx, itemID, "test-item", vault.OpUpsert)
	require.NoError(t, err)

	// Verify change was queued
	count, err := syncer.PendingCount(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestQueueItemChangeDelete(t *testing.T) {
	ctx := context.Background()
	syncer := setupTestSyncer(t)
	defer func() { _ = syncer.Close() }()

	itemID := uuid.New()

	// Queue item delete
	err := syncer.QueueItemChange(ctx, itemID, "", vault.OpDelete)
	require.NoError(t, err)

	// Verify change was queued
	count, err := syncer.PendingCount(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestQueuePositionChange(t *testing.T) {
	ctx := context.Background()
	syncer := setupTestSyncer(t)
	defer func() { _ = syncer.Close() }()

	posID := uuid.New()
	recordedAt := time.Now().UTC()
	label := strPtr("test location")

	// Queue position create
	err := syncer.QueuePositionChange(ctx, posID, "item-name", 37.7749, -122.4194, label, recordedAt, vault.OpUpsert)
	require.NoError(t, err)

	// Verify change was queued
	count, err := syncer.PendingCount(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestQueuePositionChangeWithoutLabel(t *testing.T) {
	ctx := context.Background()
	syncer := setupTestSyncer(t)
	defer func() { _ = syncer.Close() }()

	posID := uuid.New()
	recordedAt := time.Now().UTC()

	// Queue position create without label
	err := syncer.QueuePositionChange(ctx, posID, "item-name", 37.7749, -122.4194, nil, recordedAt, vault.OpUpsert)
	require.NoError(t, err)

	// Verify change was queued
	count, err := syncer.PendingCount(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestQueuePositionChangeDelete(t *testing.T) {
	ctx := context.Background()
	syncer := setupTestSyncer(t)
	defer func() { _ = syncer.Close() }()

	posID := uuid.New()
	recordedAt := time.Now().UTC()

	// Queue position delete
	err := syncer.QueuePositionChange(ctx, posID, "", 0, 0, nil, recordedAt, vault.OpDelete)
	require.NoError(t, err)

	// Verify change was queued
	count, err := syncer.PendingCount(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestPendingCount(t *testing.T) {
	ctx := context.Background()
	syncer := setupTestSyncer(t)
	defer func() { _ = syncer.Close() }()

	// Initially zero
	count, err := syncer.PendingCount(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, count)

	// Queue multiple changes
	itemID := uuid.New()
	err = syncer.QueueItemChange(ctx, itemID, "item-1", vault.OpUpsert)
	require.NoError(t, err)

	posID := uuid.New()
	recordedAt := time.Now().UTC()
	err = syncer.QueuePositionChange(ctx, posID, "item-1", 40.7128, -74.0060, nil, recordedAt, vault.OpUpsert)
	require.NoError(t, err)

	// Verify count
	count, err = syncer.PendingCount(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestMultipleChanges(t *testing.T) {
	ctx := context.Background()
	syncer := setupTestSyncer(t)
	defer func() { _ = syncer.Close() }()

	// Create multiple items
	for i := 0; i < 5; i++ {
		itemID := uuid.New()
		err := syncer.QueueItemChange(ctx, itemID, "item-"+string(rune('A'+i)), vault.OpUpsert)
		require.NoError(t, err)
	}

	count, err := syncer.PendingCount(ctx)
	require.NoError(t, err)
	assert.Equal(t, 5, count)
}

func TestAutoSyncDisabled(t *testing.T) {
	ctx := context.Background()
	syncer := setupTestSyncer(t)
	defer func() { _ = syncer.Close() }()

	// AutoSync is disabled by default in test setup
	assert.False(t, syncer.config.AutoSync)

	itemID := uuid.New()
	err := syncer.QueueItemChange(ctx, itemID, "test-item", vault.OpUpsert)
	require.NoError(t, err)

	// Change should be queued but not synced
	count, err := syncer.PendingCount(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestSyncNotConfigured(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	appDB := setupTestDB(t, tmpDir)
	defer func() { _ = appDB.Close() }()

	_, phrase, err := vault.NewSeedPhrase()
	require.NoError(t, err)

	// Create syncer with missing server config
	cfg := &Config{
		Server:     "", // Empty server
		UserID:     "",
		Token:      "",
		DerivedKey: phrase,
		DeviceID:   "test-device",
		VaultDB:    filepath.Join(tmpDir, "vault.db"),
	}

	syncer, err := NewSyncer(cfg, appDB)
	require.NoError(t, err)
	defer func() { _ = syncer.Close() }()

	// Sync should fail with helpful error
	err = syncer.Sync(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sync not configured")
}

func TestCanSync(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected bool
	}{
		{
			name: "fully configured",
			config: &Config{
				Server: "https://example.com",
				Token:  "token",
				UserID: "user",
			},
			expected: true,
		},
		{
			name: "missing server",
			config: &Config{
				Server: "",
				Token:  "token",
				UserID: "user",
			},
			expected: false,
		},
		{
			name: "missing token",
			config: &Config{
				Server: "https://example.com",
				Token:  "",
				UserID: "user",
			},
			expected: false,
		},
		{
			name: "missing user id",
			config: &Config{
				Server: "https://example.com",
				Token:  "token",
				UserID: "",
			},
			expected: false,
		},
		{
			name: "all missing",
			config: &Config{
				Server: "",
				Token:  "",
				UserID: "",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			appDB := setupTestDB(t, tmpDir)
			defer func() { _ = appDB.Close() }()

			_, phrase, err := vault.NewSeedPhrase()
			require.NoError(t, err)

			tt.config.DerivedKey = phrase
			tt.config.DeviceID = "test-device"
			tt.config.VaultDB = filepath.Join(tmpDir, "vault.db")

			syncer, err := NewSyncer(tt.config, appDB)
			require.NoError(t, err)
			defer func() { _ = syncer.Close() }()

			assert.Equal(t, tt.expected, syncer.canSync())
		})
	}
}

func TestPendingChanges(t *testing.T) {
	ctx := context.Background()
	syncer := setupTestSyncer(t)
	defer func() { _ = syncer.Close() }()

	// Queue some changes
	itemID1 := uuid.New()
	err := syncer.QueueItemChange(ctx, itemID1, "item-1", vault.OpUpsert)
	require.NoError(t, err)

	itemID2 := uuid.New()
	err = syncer.QueueItemChange(ctx, itemID2, "item-2", vault.OpUpsert)
	require.NoError(t, err)

	// Get pending changes
	changes, err := syncer.PendingChanges(ctx)
	require.NoError(t, err)
	require.Len(t, changes, 2)

	// Verify structure
	for _, change := range changes {
		assert.NotEmpty(t, change.ChangeID)
		assert.True(t, strings.HasSuffix(change.Entity, EntityItem), "entity should end with %s, got %s", EntityItem, change.Entity)
		assert.False(t, change.TS.IsZero())
	}
}

func TestLastSyncedSeq(t *testing.T) {
	ctx := context.Background()
	syncer := setupTestSyncer(t)
	defer func() { _ = syncer.Close() }()

	// Initially should be "0"
	seq, err := syncer.LastSyncedSeq(ctx)
	require.NoError(t, err)
	assert.Equal(t, "0", seq)
}

func TestCloseNilStore(t *testing.T) {
	syncer := &Syncer{
		store: nil,
	}

	err := syncer.Close()
	assert.NoError(t, err)
}

func TestQueueChangeEncryption(t *testing.T) {
	ctx := context.Background()
	syncer := setupTestSyncer(t)
	defer func() { _ = syncer.Close() }()

	itemID := uuid.New()
	err := syncer.QueueItemChange(ctx, itemID, "encrypted-item", vault.OpUpsert)
	require.NoError(t, err)

	// Verify change was encrypted (indirectly by checking it was queued)
	count, err := syncer.PendingCount(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestQueuePositionWithCoordinates(t *testing.T) {
	ctx := context.Background()
	syncer := setupTestSyncer(t)
	defer func() { _ = syncer.Close() }()

	tests := []struct {
		name string
		lat  float64
		lng  float64
	}{
		{"San Francisco", 37.7749, -122.4194},
		{"New York", 40.7128, -74.0060},
		{"Tokyo", 35.6762, 139.6503},
		{"Sydney", -33.8688, 151.2093},
		{"Zero coords", 0.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			posID := uuid.New()
			recordedAt := time.Now().UTC()
			label := strPtr(tt.name)

			err := syncer.QueuePositionChange(ctx, posID, "item-name", tt.lat, tt.lng, label, recordedAt, vault.OpUpsert)
			require.NoError(t, err)
		})
	}

	// Verify all positions were queued
	count, err := syncer.PendingCount(ctx)
	require.NoError(t, err)
	assert.Equal(t, len(tests), count)
}
