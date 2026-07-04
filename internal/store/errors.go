package store

import "errors"

// ErrKeyNotFound is returned by store operations when the requested key does
// not exist in the keyspace.
var ErrKeyNotFound = errors.New("key not found")

// ErrKeyExpired is returned by store operations when the requested key exists
// in the keyspace but its expiration time has passed. It is defined now as it
// belongs to the storage domain, ahead of TTL implementation.
var ErrKeyExpired = errors.New("key expired")
