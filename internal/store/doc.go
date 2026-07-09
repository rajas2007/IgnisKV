// Package store implements the core storage engine of IgnisKV.
//
// It is the authoritative owner of the database's storage layer. It manages both
// the volatile (memory) and durable (disk) representations of the data.
//
// # Responsibilities
//
// The package is responsible for:
//   - In-memory key-value storage.
//   - CRUD operations.
//   - Thread-safe access using sync.RWMutex.
//   - Snapshot persistence.
//
// Persistence is considered a native feature of the storage layer rather than a
// separate subsystem. All interactions with the database keyspace, memory, or disk
// must go through this package. It is responsible for serializing the current database
// state to disk and restoring previously persisted state during startup.
//
// # What Does NOT Belong Here
//
// This package is strictly isolated from the concerns of higher-level components.
// It does NOT handle:
//   - Networking.
//   - RESP parsing.
//   - Command dispatching.
//   - Business logic.
//   - Client management.
//
// # Design Philosophy
//
// The Store package is responsible only for data storage. Higher-level packages
// decide *when* data should be stored or persisted, while the Store determines
// *how* it is stored. This separation keeps persistence independent of command
// execution, networking, and protocol concerns while allowing the storage
// implementation to evolve without affecting the rest of the system.
//
// # Architecture
//
// The Store sits at the base of the application architecture:
//
//	Client
//	    ↓
//	Server
//	    ↓
//	Protocol
//	    ↓
//	Dispatcher
//	    ↓
//	Store
//	   ↙     ↘
//	Memory   JSON Snapshot
//
// Current Scope (Sprint 8)
//
// Persistence is introduced incrementally. The current scope covers:
//   - Manual SAVE command.
//   - Automatic loading during startup.
//   - JSON snapshot persistence.
//   - No automatic background saving.
//   - No Append Only File.
//   - No snapshot compression.
//   - No incremental persistence.
//
// # Future Scope
//
// Future milestones may extend the Store with automatic persistence,
// Append Only Files (AOF), snapshot optimization, compression,
// and additional persistence strategies without changing its role
// as the owner of the storage layer.
package store
