// ABOUTME: Charm KV client wrapper using transactional Do API
// ABOUTME: Short-lived connections to avoid lock contention with other MCP servers

package charm

import (
	"os"

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

// Client holds configuration for KV operations.
// Unlike the previous implementation, it does NOT hold a persistent connection.
// Each operation opens the database, performs the operation, and closes it.
type Client struct {
	dbName   string
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

// NewClient creates a new client with the given config.
func NewClient(cfg *Config) (*Client, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	// Set CHARM_HOST before any KV operations
	if err := os.Setenv("CHARM_HOST", cfg.CharmHost); err != nil {
		return nil, err
	}

	return &Client{
		dbName:   DBName,
		autoSync: cfg.AutoSync,
	}, nil
}

// Get retrieves a value by key (read-only, no lock contention).
func (c *Client) Get(key []byte) ([]byte, error) {
	var val []byte
	err := kv.DoReadOnly(c.dbName, func(k *kv.KV) error {
		var err error
		val, err = k.Get(key)
		return err
	})
	return val, err
}

// Set stores a value with the given key.
func (c *Client) Set(key, value []byte) error {
	return kv.Do(c.dbName, func(k *kv.KV) error {
		if err := k.Set(key, value); err != nil {
			return err
		}
		if c.autoSync {
			return k.Sync()
		}
		return nil
	})
}

// Delete removes a key.
func (c *Client) Delete(key []byte) error {
	return kv.Do(c.dbName, func(k *kv.KV) error {
		if err := k.Delete(key); err != nil {
			return err
		}
		if c.autoSync {
			return k.Sync()
		}
		return nil
	})
}

// Keys returns all keys in the database.
func (c *Client) Keys() ([][]byte, error) {
	var keys [][]byte
	err := kv.DoReadOnly(c.dbName, func(k *kv.KV) error {
		var err error
		keys, err = k.Keys()
		return err
	})
	return keys, err
}

// DoReadOnly executes a function with read-only database access.
// Use this for batch read operations that need multiple Gets.
func (c *Client) DoReadOnly(fn func(k *kv.KV) error) error {
	return kv.DoReadOnly(c.dbName, fn)
}

// Do executes a function with write access to the database.
// Use this for batch write operations.
func (c *Client) Do(fn func(k *kv.KV) error) error {
	return kv.Do(c.dbName, func(k *kv.KV) error {
		if err := fn(k); err != nil {
			return err
		}
		if c.autoSync {
			return k.Sync()
		}
		return nil
	})
}

// Sync triggers a manual sync with the charm server.
func (c *Client) Sync() error {
	return kv.Do(c.dbName, func(k *kv.KV) error {
		return k.Sync()
	})
}

// Reset clears all data (nuclear option).
func (c *Client) Reset() error {
	return kv.Do(c.dbName, func(k *kv.KV) error {
		return k.Reset()
	})
}

// --- Legacy compatibility layer ---
// These functions maintain backwards compatibility with existing code.

var globalClient *Client

// InitClient initializes the global charm client.
func InitClient(cfg *Config) error {
	if globalClient != nil {
		return nil
	}
	var err error
	globalClient, err = NewClient(cfg)
	return err
}

// GetClient returns the global client.
func GetClient() *Client {
	return globalClient
}

// Close is a no-op for backwards compatibility.
// With Do API, connections are automatically closed after each operation.
func (c *Client) Close() error {
	return nil
}

// NewTestClient creates a client for testing without network access.
func NewTestClient(dbName string) (*Client, error) {
	return &Client{
		dbName:   dbName,
		autoSync: false,
	}, nil
}
