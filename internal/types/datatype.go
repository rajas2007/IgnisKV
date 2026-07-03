// Package types defines shared domain types, models, and enums used across
// the IgnisKV database components. These common representations prevent
// circular package dependencies.
package types

// DataType represents the internal value data types supported by IgnisKV.
// This defines the database's type system rather than Go's native types.
type DataType int

const (
	// StringType represents a string value stored by IgnisKV.
	StringType DataType = iota

	// Planned future data types for later milestones:
	// - ListType
	// - HashType
	// - SetType
	// - SortedSetType
)
