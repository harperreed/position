// ABOUTME: Charm KV client wrapper for position tracking
// ABOUTME: Provides thread-safe client initialization and sync configuration

package charm

import (
	"fmt"
	"os"
	"sync"

	"github.com/charmbracelet/charm/kv"
)

const (
	// DBName is the name of the Charm KV database for position data.
	DBName = "position"

	// DefaultCharmHost is the default Charm server to use.
	DefaultCharmHost = "charm.2389.dev"

	// Key prefixes for type-based organization.
	ItemPrefix     = "item:"
	PositionPrefix = "position:"
)

var (
	globalClient *Client
	clientOnce   sync.Once
	clientErr    error
)

// Client wraps a Charm KV store with position-specific operations.
type Client struct {
	kv       *kv.KV
	autoSync bool
}

// Config holds client configuration options.
type Config struct {
	// CharmHost is the Charm server to use (default: charm.2389.dev).
	CharmHost string
	// AutoSync enables automatic sync after writes.
	AutoSync bool
}

// DefaultConfig returns the default client configuration.
func DefaultConfig() *Config {
	host := os.Getenv("CHARM_HOST")
	if host == "" {
		host = DefaultCharmHost
	}
	return &Config{
		CharmHost: host,
		AutoSync:  true,
	}
}

// InitClient initializes the global Charm client.
// Safe to call multiple times - uses sync.Once internally.
func InitClient(cfg *Config) error {
	clientOnce.Do(func() {
		if cfg == nil {
			cfg = DefaultConfig()
		}

		// Set CHARM_HOST before opening KV
		if err := os.Setenv("CHARM_HOST", cfg.CharmHost); err != nil {
			clientErr = err
			return
		}

		db, err := kv.OpenWithDefaultsFallback(DBName)
		if err != nil {
			clientErr = err
			return
		}

		globalClient = &Client{
			kv:       db,
			autoSync: cfg.AutoSync,
		}

		// Pull remote data on startup (skip in read-only mode)
		if cfg.AutoSync && !db.IsReadOnly() {
			_ = db.Sync()
		}
	})
	return clientErr
}

// GetClient returns the global client.
// Returns nil if InitClient wasn't called or failed.
func GetClient() *Client {
	return globalClient
}

// NewClient creates a new client with the given config.
// For CLI usage, prefer InitClient + GetClient for global access.
func NewClient(cfg *Config) (*Client, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	// Set CHARM_HOST before opening KV
	if err := os.Setenv("CHARM_HOST", cfg.CharmHost); err != nil {
		return nil, err
	}

	db, err := kv.OpenWithDefaultsFallback(DBName)
	if err != nil {
		return nil, err
	}

	client := &Client{
		kv:       db,
		autoSync: cfg.AutoSync,
	}

	// Pull remote data on startup (skip in read-only mode)
	if cfg.AutoSync && !db.IsReadOnly() {
		_ = db.Sync()
	}

	return client, nil
}

// Close releases client resources.
func (c *Client) Close() error {
	if c.kv != nil {
		return c.kv.Close()
	}
	return nil
}

// syncIfEnabled syncs to remote if auto-sync is enabled and not in read-only mode.
func (c *Client) syncIfEnabled() {
	if c.autoSync && !c.kv.IsReadOnly() {
		_ = c.kv.Sync()
	}
}

// IsReadOnly returns true if the database is in read-only mode.
// This happens when another process (like an MCP server) holds the lock.
func (c *Client) IsReadOnly() bool {
	return c.kv.IsReadOnly()
}

// Sync forces a sync with the remote server.
func (c *Client) Sync() error {
	if c.kv.IsReadOnly() {
		return fmt.Errorf("cannot sync: database is locked by another process (MCP server?)")
	}
	return c.kv.Sync()
}

// Reset clears all local data and resets the KV store.
func (c *Client) Reset() error {
	if c.kv.IsReadOnly() {
		return fmt.Errorf("cannot reset: database is locked by another process (MCP server?)")
	}
	return c.kv.Reset()
}

// NewTestClient creates a client for testing without network access.
// Uses kv.OpenWithDefaults to avoid Charm Cloud authentication.
func NewTestClient(dbName string) (*Client, error) {
	db, err := kv.OpenWithDefaults(dbName)
	if err != nil {
		return nil, err
	}
	return &Client{
		kv:       db,
		autoSync: false,
	}, nil
}
