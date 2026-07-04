// Package store implements the core in-memory storage engine of IgnisKV.
//
// It is the authoritative owner of the database keyspace, mapping string keys
// to types.Value records. All reads and writes to the keyspace must go through
// this package's methods, which coordinate safe concurrent access using a
// read-write mutex.
//
// This package is intentionally isolated from networking, protocol serialisation,
// command routing, and parsing. It knows only how to store, retrieve, and delete
// values. This separation ensures that the storage engine remains independently
// testable and free from the concerns of higher-level subsystems.
package store
