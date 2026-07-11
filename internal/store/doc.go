// Package store implements the core storage engine of IgnisKV.
//
// It is the authoritative owner of the database's storage layer. It manages
// the physical storage of data, the durable persistence of state, and the
// lifetime of every key stored in the database.
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
//   - Active (background) expiration.
//   - Background cleanup of expired keys.
//   - Lifetime management of keys.
//   - Automatic expiration enforcement.
//   - Coordination of lazy and active expiration.
//   - Persistence of expiration timestamps.
//   - Expiration-aware snapshot loading and restoration.
//   - Remaining lifetime calculation.
//   - TTL metadata inspection.
//   - Read-only expiration queries.
//   - Expiration state reporting.
//   - Expiration metadata modification.
//   - EXPIRE command support.
//   - Updating key lifetime.
//   - Expiration policy enforcement.
//   - Expiration metadata persistence.
//
// Persistence and expiration are both considered native features of the storage
// layer rather than separate subsystems. All interactions with the database
// keyspace, memory, or disk must go through this package. It is responsible for
// serializing the current database state to disk, restoring previously
// persisted state during startup, and ensuring that expired keys are removed
// both on access and proactively in the background.
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
// The Store package is responsible for data storage, data lifetime, the removal
// of expired data, reporting expiration state, and modifying expiration.
// Higher-level packages decide *when* commands execute, *when* to ask for TTL
// information, and *when* expiration should change (command policy).
// The Store determines *how* data is stored, *how* data expires, *how* expired
// data is removed, *how* TTL is calculated, and *how* expiration changes (mechanism).
//
// Handlers never manipulate timestamps. The Store remains the single owner
// of expiration semantics.
//
// Neither the Server nor the Dispatcher knows anything about expiration
// mechanics. Expiration remains an implementation detail of the Store. This
// separation keeps persistence and expiration independent of command execution,
// networking, and protocol concerns while allowing the storage implementation
// to evolve without affecting the rest of the system.
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
//	   ↙    ↓      ↘
//	Memory Persistence Expiration
//	               ↙        ↘
//	             TTL      EXPIRE
//
// Background cleanup, TTL reporting, and EXPIRE modifications are implemented
// entirely inside the Store and do not alter the overall application architecture.
// No higher-level package is aware of expiration mechanics, performs expiration
// calculations, or manipulates timestamps.
//
// # Current Scope (Sprint 13)
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
//   - Active background expiration.
//   - Expiration-aware persistence.
//   - Concurrent-safe expiration cleanup.
//   - TTL command support.
//   - Remaining lifetime calculation.
//   - Read-only expiration queries.
//   - Centralized expiration observation.
//   - EXPIRE infrastructure.
//   - Expiration updates.
//   - Expiration metadata modification.
//   - Updating existing TTL.
//   - Converting persistent keys into expiring keys.
//   - Automatic persistence after expiration updates.
//   - Shared lazy expiration infrastructure.
//   - Shared expiration lifecycle.
//
// Both TTL and EXPIRE reuse the existing expiration subsystem without
// introducing a second implementation or new metadata.
//
// # Current Limitations
//
// Persistence is intentionally non-transactional. If Save() fails after a
// successful write, the in-memory state reflects the latest write while the
// snapshot reflects the previous state. The server continues operating normally.
//
// Cleanup scans the entire keyspace while holding a write lock. Sprint 11
// intentionally favors correctness and simplicity over scalability. All client
// operations pause briefly during a cleanup cycle. Future milestones may
// introduce sampling, sharding, or finer-grained locking to reduce contention.
//
// TTL reports whole seconds only. Millisecond precision is intentionally deferred.
// TTL observes expiration but never modifies it.
//
// EXPIRE currently supports only relative durations measured in seconds.
// Millisecond precision remains deferred. Absolute expiration commands
// remain future work.
//
// # Background Cleanup
//
// Each MemoryStore instance owns exactly one background cleanup goroutine.
// The goroutine is started automatically by NewMemoryStore() and runs for the
// lifetime of the application. It uses a fixed time.Ticker to periodically
// wake, scan the keyspace, and remove expired keys.
//
// No shutdown mechanism currently exists.
// Graceful shutdown using context.Context is intentionally deferred to a future milestone.
//
// Tests should create only the MemoryStore instances they actually require
// because each Store owns one background cleanup goroutine.
//
// # Engineering Notes
//
// Lazy expiration and active expiration coexist. Either mechanism may delete
// the same key. This is safe because delete(map, key) on a missing key is a
// no-op in Go. No additional coordination is required between the two paths.
//
// Expiration logic remains centralized through a single private helper,
// providing one authoritative definition of key expiration that is reused by
// Get, Save, Load, and the background cleanup goroutine.
//
// The helper performs no synchronization and has no side effects. Callers
// remain responsible for lock management.
//
// Synchronization remains owned entirely by MemoryStore using sync.RWMutex.
// All expiration implementations must preserve the existing concurrency
// guarantees.
//
// TTL is a read-only operation. TTL never performs persistence. TTL never changes
// expiration policy.
//
// EXPIRE reuses:
//   - isExpired()
//   - lazyExpire()
//   - ExpiresAt
//   - background cleanup
//   - Save()
//   - Load()
//   - TTL()
//
// The Store remains the single authoritative owner of expiration metadata.
//
// Engineering Rule:
//
// Any command that reads or modifies a key must verify expiration before
// operating on it. If the key has already expired, it is lazily removed and
// treated as absent.
//
// Persistence Rule: Any command that modifies logical database state (data
// or metadata) automatically triggers persistence.
//
// The current write commands are:
//   - SET
//   - DEL
//   - EXPIRE
//
// Future write commands are expected to follow the same persistence rule.
//
// TTL follows the same check-then-act lazy expiration pattern used by Get():
//
//	Acquire RLock
//	↓
//	Read value
//	↓
//	Release RLock
//	↓
//	Check isExpired()
//	↓
//	Expired?
//	↓
//	Acquire Lock
//	↓
//	Verify key still exists
//	↓
//	Verify still expired
//	↓
//	Delete
//	↓
//	Release Lock
//
// This verification step prevents races where another goroutine modifies or
// removes the key between the read and write locks. This check-then-act pattern
// is now the standard concurrency pattern for every Store operation that
// performs lazy expiration.
//
// The Store-layer contract for TTL() returns:
//   - n >= 0, nil → key exists with n seconds remaining.
//   - -1, nil → key exists without expiration.
//   - 0, ErrKeyExpired → key expired and requires lazy deletion.
//   - 0, ErrKeyNotFound → key does not exist.
//
// Higher-level handlers translate both ErrKeyExpired and ErrKeyNotFound into the
// same client-visible response while preserving richer information inside the Store.
//
// # Future Scope
//
// Future milestones may extend the Store with:
//
// Expiration:
//   - Absolute expiration commands.
//   - PERSIST command.
//   - Millisecond precision expiration.
//   - Configurable cleanup interval.
//   - Expiration statistics.
//   - Redis-style sampling improvements.
//
// Lifecycle:
//   - Graceful goroutine shutdown.
//   - Context cancellation.
//
// Persistence:
//   - Append Only File (AOF).
//   - Write-Ahead Logging (WAL).
//   - Snapshot optimization.
//   - Additional persistence strategies.
package store
