// ABOUTME: Common storage errors
// ABOUTME: Enables consistent error handling across storage implementations

package storage

import "errors"

// ErrNotFound is returned when a requested entity does not exist.
var ErrNotFound = errors.New("not found")

// ErrReadOnly is returned when attempting to write to a read-only store.
var ErrReadOnly = errors.New("storage is read-only")
