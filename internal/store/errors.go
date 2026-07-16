package store

import "errors"

// ErrKeyNotFound is returned by store operations when the requested key does
// not exist in the keyspace.
var ErrKeyNotFound = errors.New("key not found")

// ErrKeyExpired is returned by store operations when the requested key exists
// in the keyspace but its expiration time has passed. It is defined now as it
// belongs to the storage domain, ahead of TTL implementation.
var ErrKeyExpired = errors.New("key expired")

// ErrInvalidDuration is returned when an expiration command is provided with
// a zero or negative duration.
var ErrInvalidDuration = errors.New("invalid duration")

// ErrInvalidTimestamp is returned when an absolute expiration command is provided with
// a timestamp that is not in the future.
var ErrInvalidTimestamp = errors.New("invalid timestamp")

// ErrWrongType is returned when a command is executed against a key holding
// a value of a different DataType than the command expects.
var ErrWrongType = errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")
