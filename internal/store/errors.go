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

// ErrFieldNotFound is returned by hash operations when the requested field does
// not exist in the hash.
var ErrFieldNotFound = errors.New("field not found")

// ErrNotInteger is returned when a mathematical operation is attempted on a
// value that cannot be parsed as an integer.
var ErrNotInteger = errors.New("ERR hash value is not an integer")

// ErrNotFloat is returned when a mathematical operation is attempted on a
// value that cannot be parsed as a float.
var ErrNotFloat = errors.New("ERR hash value is not a float")

// ErrOverflow is returned when an increment operation would overflow the bounds
// of a 64-bit signed integer.
var ErrOverflow = errors.New("ERR increment or decrement would overflow")

// ErrNaNOrInfinity is returned when a float operation would produce NaN or Infinity.
var ErrNaNOrInfinity = errors.New("ERR increment would produce NaN or Infinity")
