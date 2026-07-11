// Package store implements the core storage engine of IgnisKV.
//
// It is the authoritative owner of the database's storage layer. It manages both
// the volatile (memory) and durable (disk) representations of the data, as well
// as the lifetime of every key stored in the database.
//
// # Responsibilities
//
// The package is responsible for:
//   - In-memory key-value storage.
//   - CRUD operations.
//   - Thread-safe access using sync.RWMutex.
//   - Snapshot persistence.
//   - Manual SAVE triggered by explicit client command.
//   - Automatic persistence after successful write operations.
//   - Automatic startup recovery from persisted snapshots.
//   - Expiration metadata management.
//   - Lazy expiration during key access.
//   - Persistence of expiration timestamps.
//   - Expiration-aware snapshot loading and restoration.
//
// Persistence and expiration are both considered native features of the storage
// layer rather than separate subsystems. All interactions with the database
// keyspace, memory, or disk must go through this package. It is responsible for
// serializing the current database state to disk and restoring previously
// persisted state during startup.
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
// The Store package is responsible for both data storage and data lifetime.
// Higher-level packages decide *when* data should be stored, persisted, or
// expired (command policy), while the Store determines *how* expiration is
// represented, checked, stored, restored, and enforced (mechanism).
//
// This separation keeps persistence and expiration independent of command
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
// Expiration is implemented inside the Store without introducing new packages
// or changing the overall architecture. No higher-level package is aware of
// expiration mechanics.
//
// # Current Scope (Sprint 10)
//
// The Store currently supports:
//   - Thread-safe in-memory storage.
//   - Manual SAVE.
//   - Automatic persistence after successful writes.
//   - Automatic startup recovery.
//   - JSON snapshots.
//   - Atomic snapshot writes.
//   - Expiration metadata using ExpiresAt.
//   - Lazy expiration during GET.
//   - Persistence of expiration timestamps.
//
// Sprint 10 intentionally implements lazy expiration only. Expired keys are
// removed when discovered during access. Background expiration is intentionally
// deferred to Sprint 11.
//
// # Current Limitations
//
// Persistence is intentionally non-transactional. If Save() fails after a
// successful write, the in-memory state reflects the latest write while the
// snapshot reflects the previous state. The server continues operating normally.
//
// Expiration is intentionally lazy. Expired keys may remain in memory until
// accessed. No background cleanup exists yet. Active expiration will be
// introduced in Sprint 11.
//
// # Future Scope
//
// Future milestones may extend the Store with:
//   - Background expiration.
//   - TTL command.
//   - EXPIRE command.
//   - PERSIST command.
//   - Append Only File (AOF).
//   - Write-Ahead Logging (WAL).
//   - Snapshot optimization.
//   - Additional persistence strategies.
//
// # Engineering Notes
//
// Expiration mechanics are owned entirely by the Store. Expiration checks are
// centralized through a private helper, providing a single definition of key
// expiration that is reused by Get, Save, and Load.
//
// The helper performs no synchronization and has no side effects. Callers
// remain responsible for lock management.
//
// Synchronization remains owned entirely by MemoryStore using sync.RWMutex.
// All expiration implementations must preserve the existing concurrency
// guarantees.
package store
