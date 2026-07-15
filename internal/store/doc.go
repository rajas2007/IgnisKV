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
//   - Removing expiration metadata.
//   - PERSIST command support.
//   - Converting expiring keys back into persistent keys.
//   - Expiration metadata lifecycle management.
//   - Shared expiration modification infrastructure.
//   - Absolute expiration management.
//   - EXPIREAT command support.
//   - Absolute expiration metadata updates.
//   - Unix timestamp expiration handling.
//   - Shared expiration timestamp management.
//   - Millisecond expiration management.
//   - PEXPIRE command support.
//   - Relative millisecond expiration updates.
//   - High-precision expiration scheduling.
//   - Millisecond duration handling.
//   - Absolute millisecond expiration management.
//   - PEXPIREAT command support.
//   - Unix millisecond timestamp handling.
//   - High-precision absolute expiration assignment.
//   - Millisecond TTL observation.
//   - PTTL command support.
//   - High-precision expiration queries.
//   - Millisecond lifetime calculation.
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
// The Store package owns every supported expiration operation:
//   - SET EX
//   - TTL
//   - PTTL
//   - EXPIRE
//   - PEXPIRE
//   - PERSIST
//   - EXPIREAT
//   - PEXPIREAT
//   - Lazy expiration
//   - Active expiration
//
// The Store also owns every supported expiration observation operation:
//   - TTL
//   - PTTL
//
// Higher-level packages decide *when* commands execute, *when* to ask for TTL
// information, *when* expiration should change, and *when* expiration should be
// removed (command policy). Handlers continue deciding WHEN commands execute.
// Handlers perform only argument validation.
//
// The Store determines *how* data is stored, *how* data expires, *how* expired
// data is removed, *how* TTL is calculated, *how* expiration changes, and *how*
// expiration is cleared (mechanism).
//
// Handlers may convert client input into Go types. For example, Unix timestamp
// → time.Time, and for PEXPIRE this means converting milliseconds into
// time.Duration. For PEXPIREAT this means converting Unix milliseconds into
// time.Time using time.UnixMilli(). The Store remains the sole owner of
// expiration semantics and ExpiresAt. Handlers never manipulate timestamps.
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
//	   ↙        ↓          ↘
//	Memory  Persistence  Expiration
//	                        ↓
//	        SET / GET / TTL / PTTL / EXPIRE / PERSIST / EXPIREAT / PEXPIRE / PEXPIREAT
//	                        ↓
//	               Background Cleanup
//
// Background cleanup, TTL reporting, PTTL reporting, EXPIRE modifications, and
// PERSIST removals are implemented entirely inside the Store and do not alter
// the overall application architecture. No higher-level package is aware of
// expiration mechanics, performs expiration calculations, or manipulates
// timestamps.
//
// GET, TTL, PTTL, EXPIRE, PEXPIRE, PERSIST, EXPIREAT, PEXPIREAT, and background cleanup all reuse:
//   - ExpiresAt
//   - isExpired()
//   - lazyExpire()
//
// No additional expiration metadata exists. ExpiresAt remains the single
// storage location for expiration metadata.
//
// TTL and PTTL are observation commands. They never modify state. They both
// reuse ExpiresAt, isExpired(), and lazyExpire(). TTL reports whole seconds.
// PTTL reports whole milliseconds. Both compute remaining lifetime from the
// same ExpiresAt field.
//
// Observation comparison:
//
//	Command       Returns              Unit
//	TTL           remaining/-1/-2      seconds
//	PTTL          remaining/-1/-2      milliseconds
//
// Everything except the returned unit is identical.
//
// EXPIRE and PEXPIRE both compute relative expiration. EXPIREAT and PEXPIREAT
// both assign absolute expiration. All four ultimately update the same
// ExpiresAt field. No duplicate expiration implementation exists.
//
// Completed command pairs:
//
//	EXPIRE    ↔ EXPIREAT
//	(relative seconds ↔ absolute seconds)
//
//	PEXPIRE   ↔ PEXPIREAT
//	(relative milliseconds ↔ absolute milliseconds)
//
// Each pair shares identical Store semantics.
// Only the handler input conversion differs.
//
// # Current Scope (Sprint 18)
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
//   - PERSIST command support.
//   - Removing expiration metadata.
//   - Persistent key restoration.
//   - Shared expiration modification.
//   - Complete expiration lifecycle management.
//   - EXPIREAT command support.
//   - Absolute expiration assignment.
//   - Unix timestamp expiration.
//   - Shared expiration metadata updates.
//   - Absolute expiration persistence.
//   - PEXPIRE command.
//   - Millisecond relative expiration.
//   - High-precision expiration scheduling.
//   - time.Duration based expiration updates.
//   - PEXPIREAT command.
//   - Absolute millisecond expiration.
//   - Unix millisecond timestamps.
//   - High-precision absolute expiration assignment.
//   - time.UnixMilli based expiration updates.
//   - PTTL command.
//   - Millisecond TTL observation.
//   - High-precision remaining lifetime queries.
//   - Millisecond lifetime calculation.
//
// Expiration observation now supports both seconds and milliseconds through
// one shared expiration implementation.
//
// Expiration now supports observation, creation, modification, removal,
// relative assignment, absolute assignment, second precision input,
// millisecond precision input, and automatic enforcement through one
// shared implementation.
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
// TTL continues to report whole seconds.
// PTTL reports whole milliseconds.
// EXPIRETIME and PEXPIRETIME remain future work.
//
// Absolute timestamps must be in the future.
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
// TTL and PTTL are read-only operations. Neither performs persistence. Neither
// changes expiration policy.
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
// PERSIST reuses:
//   - ExpiresAt
//   - isExpired()
//   - lazyExpire()
//   - Save()
//   - background cleanup
//   - integer RESP replies
//
// EXPIREAT reuses:
//   - time.Time
//   - ExpiresAt
//   - isExpired()
//   - lazyExpire()
//   - Save()
//   - background cleanup
//   - integer RESP replies
//
// PEXPIRE reuses:
//   - ExpiresAt
//   - isExpired()
//   - lazyExpire()
//   - Save()
//   - background cleanup
//   - time.Duration
//
// PEXPIREAT reuses:
//   - ExpiresAt
//   - isExpired()
//   - lazyExpire()
//   - Save()
//   - background cleanup
//   - time.Time
//   - time.UnixMilli() (via handler conversion)
//
// PTTL reuses:
//   - ExpiresAt
//   - isExpired()
//   - lazyExpire()
//   - time.Time
//
// PTTL introduces only a different observation unit. Internally both TTL and
// PTTL compute the remaining lifetime from the same ExpiresAt field.
//
// Millisecond clamping:
//
// Between reading ExpiresAt and computing the remaining lifetime, a key may
// expire. If the computed millisecond lifetime is negative, clamp it to zero
// before returning. This prevents returning a nonsensical negative lifetime
// for a key that is in the process of expiring.
//
// The Store remains the single authoritative owner of expiration metadata.
//
// Engineering Rule:
//
// Every command that reads or modifies a key verifies expiration before
// operating on it. If the key has already expired, it is lazily removed and
// treated as absent.
//
// Commands following this rule:
//   - GET
//   - TTL
//   - PTTL
//   - EXPIRE
//   - PEXPIRE
//   - PERSIST
//   - EXPIREAT
//   - PEXPIREAT
//
// Persistence Rule:
//
// Every command that modifies logical database state (data or metadata)
// must trigger automatic persistence.
//
// Current write commands:
//   - SET
//   - DEL
//   - EXPIRE
//   - PEXPIRE
//   - PERSIST
//   - EXPIREAT
//   - PEXPIREAT
//
// PEXPIREAT introduces only a new client input format. PTTL introduces only
// a new observation unit. Internally the expiration subsystem remains
// unchanged because all expiration metadata continues to be stored exclusively
// in ExpiresAt.
//
//   - EXPIRE and PEXPIRE share one implementation pattern.
//   - EXPIREAT and PEXPIREAT share one implementation pattern.
//
// The only difference inside each pair is the handler's input conversion.
//
// Future write commands are expected to follow the same persistence rule.
//
// Idempotency:
//
// PERSIST is idempotent. Calling PERSIST on an already persistent key performs
// no state modification and returns the existing state unchanged.
//
// Clearing ExpiresAt requires no changes to the background cleanup subsystem
// because isExpired() already treats the zero time value as persistent.
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
//   - Advanced timeline observation (EXPIRETIME, PEXPIRETIME).
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
