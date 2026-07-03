package types

import "time"

// Value represents a single database record stored within IgnisKV.
// It encapsulates the type metadata, the actual payload, and the expiration
// lifecycle of the record.
type Value struct {
	// Type represents the internal IgnisKV data type of the record.
	Type DataType

	// Data holds the underlying database payload. In v0.1 this holds a Go
	// string, but it is typed as 'any' to support nested structures in future versions.
	Data any

	// ExpiresAt specifies the exact timestamp when this record should expire.
	// A zero time value (ExpiresAt.IsZero() returning true) indicates that the
	// record never expires.
	ExpiresAt time.Time
}
