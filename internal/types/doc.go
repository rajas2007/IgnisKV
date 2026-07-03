// Package types defines the shared domain models used throughout IgnisKV.
//
// It forms the common language of the database, providing the core data
// structures that every subsystem communicates with. This package is
// intentionally kept free of business logic, containing only plain data
// definitions.
//
// It is imported by internal/store, internal/commands, internal/protocol,
// and internal/server. Centralising shared representations here ensures
// that those packages can reference common types without importing each
// other, which prevents circular dependencies.
//
// This package must never import any other internal IgnisKV package.
package types
