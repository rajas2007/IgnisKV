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
//   - Support for multiple Redis-compatible data types.
//   - String values.
//   - List values.
//   - List operations implemented within the Store rather than handlers.
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
//   - Absolute expiration observation.
//   - EXPIRETIME command support.
//   - Unix timestamp expiration queries.
//   - Second-precision expiration timestamp retrieval.
//   - Absolute expiration observation in milliseconds.
//   - PEXPIRETIME command support.
//   - Unix millisecond expiration timestamp queries.
//   - Complete expiration observation subsystem.
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
//   - EXPIRETIME
//   - PEXPIRETIME
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
//   - EXPIRETIME
//   - PEXPIRETIME
//
// Higher-level packages decide *when* commands execute, *when* to ask for TTL
// information, *when* expiration should change, and *when* expiration should be
// removed (command policy). Handlers continue deciding WHEN commands execute.
// Handlers perform only argument validation and response translation.
//
// The Store determines *how* data is stored, *how* data expires, *how* expired
// data is removed, *how* TTL is calculated, *how* expiration changes, and *how*
// expiration is cleared (mechanism). The Store owns all type validation.
//
// The Store decides:
//   - whether a key exists
//   - whether a key has expired
//   - whether the stored type matches the requested operation
//
// Every command operates against exactly one DataType.
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
// Every stored object consists of:
//   - Type
//   - Data
//   - ExpiresAt
//
// String values store:
//
//	Data -> string
//
// List values store:
//
//	Data -> []string
//
// Lists use Type: ListType to distinguish them from String values while
// continuing to reuse the same Value abstraction.
//
// This architecture was intentionally designed to allow additional data types
// without changing the Store architecture. Future collection types (Hashes and
// Sets) will reuse the same Value abstraction.
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
//	        SET / GET / TTL / PTTL / EXPIRETIME / PEXPIRETIME / EXPIRE / PERSIST / EXPIREAT / PEXPIRE / PEXPIREAT
//	                        ↓
//	               Background Cleanup
//
// Background cleanup, TTL reporting, PTTL reporting, EXPIRETIME reporting,
// PEXPIRETIME reporting, EXPIRE modifications, and PERSIST removals are
// implemented entirely inside the Store and do not alter the overall application
// architecture. No higher-level package is aware of expiration mechanics, performs
// expiration calculations, or manipulates timestamps.
//
// GET, TTL, PTTL, EXPIRETIME, PEXPIRETIME, EXPIRE, PEXPIRE, PERSIST, EXPIREAT, PEXPIREAT, and background cleanup all reuse:
//   - ExpiresAt
//   - isExpired()
//   - lazyExpire()
//
// No additional expiration metadata exists. ExpiresAt remains the single
// storage location for expiration metadata.
//
// TTL, PTTL, EXPIRETIME, and PEXPIRETIME are observation commands. They never modify state.
// All four commands reuse ExpiresAt, isExpired(), and lazyExpire().
//
// TTL and PTTL compute the remaining lifetime from the ExpiresAt field. TTL
// reports whole seconds. PTTL reports whole milliseconds.
// EXPIRETIME and PEXPIRETIME expose the stored expiration timestamp.
// EXPIRETIME returns the stored absolute expiration timestamp in Unix seconds.
// PEXPIRETIME returns the stored absolute expiration timestamp in Unix milliseconds.
//
// Observation comparison:
//
//	Command         Returns               Unit
//	TTL             remaining/-1/-2       seconds
//	PTTL            remaining/-1/-2       milliseconds
//	EXPIRETIME      absolute/-1/-2        Unix seconds
//	PEXPIRETIME     absolute/-1/-2        Unix milliseconds
//
// TTL and PTTL derive the remaining lifetime.
// EXPIRETIME and PEXPIRETIME expose the stored expiration timestamp.
// Within each pair, only the returned unit differs.
//
// PEXPIRETIME requires no clamping.
//
// Unlike TTL/PTTL, it returns a fixed stored timestamp rather than a computed duration.
// Since ExpiresAt.UnixMilli() reads immutable metadata, the returned value never
// becomes negative due to elapsed execution time.
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
// # Type Safety
//
// Every collection command verifies the stored DataType before operating.
// If the stored type does not match the requested command, the Store returns
// ErrWrongType. This provides a single shared type-checking mechanism for
// every collection command.
//
// # Expiration
//
// Expiration behavior is identical for all supported data types. Strings and
// Lists both reuse ExpiresAt, isExpired(), and lazyExpire(). Collection
// commands never implement separate expiration logic.
//
// # Current Scope (Sprint 24)
//
// The Collections subsystem now supports:
//   - LPush
//   - RPush
//   - LLen
//   - LRange
//
// LRANGE is the first range-based collection command.
// It introduces index normalization.
// It supports positive and negative indices.
// It performs lazy expiration.
// It never mutates collection contents.
// It never performs persistence.
// It returns a copy of the requested range.
//
// # Collection Command Categories
//
// The architectural distinction for collection commands is as follows:
//
// Mutating commands
//   - LPUSH
//   - RPUSH
//
// Characteristics:
//   - Acquire Lock immediately
//   - Modify state
//   - Trigger persistence
//   - Return updated state
//
// Read-only commands
//   - LLEN
//   - LRANGE
//
// Characteristics:
//   - Begin with RLock
//   - May lazily expire keys
//   - Never modify collection contents
//   - Never trigger persistence
//
// This distinction becomes the architectural guideline for all future collection commands.
//
// # Range-Based Collection Operations
//
// Collection readers are now divided into:
//
// Point readers
//   - LLEN
//
// Range readers
//   - LRANGE
//
// Point readers observe a single property of a collection.
// Range readers return a normalized subset of collection elements.
//
// Both:
//   - begin with RLock
//   - may lazily expire keys
//   - never mutate data
//   - never perform persistence
//
// This distinction becomes the architectural guideline for future range-based commands.
//
// # Collection Invariants
//
// Collections never exist in an empty state.
//
// When the final element of a collection is removed,
// the collection key itself is deleted from the store.
//
// Consequently:
//
// LLEN returning 0 always means:
//   - the key does not exist
//   - the key was lazily expired
//
// It never means:
//   - an existing empty list
//
// This invariant applies to all future collection types unless
// explicitly documented otherwise.
//
// # Range Normalization
//
// The canonical normalization algorithm:
//
// normalizeRange(start, stop, length)
//
// Rules:
//
// 1. Negative indices count from the end.
// Examples:
// -1 = last element
// -2 = second last
//
// 2. Indices are clamped to valid bounds.
//
// 3. If start exceeds the normalized stop:
// Return an empty range.
//
// 4. The stop index is inclusive.
//
// 5. The returned ordering always matches the underlying list.
//
// This algorithm becomes the reusable specification for:
//   - LRANGE
//   - LTRIM
//   - LINDEX
//   - Future sorted-set range commands
//
// # Slice Ownership
//
// LRANGE never returns a reference to the internal slice.
// Instead it returns a copy.
//
// Reason:
// The returned slice survives after RLock has been released.
// Returning internal storage would expose shared mutable memory
// to concurrent writers such as LPUSH and RPUSH.
//
// Returning a copy:
//   - preserves thread safety
//   - avoids data races
//   - protects internal storage
//   - provides stable ownership semantics
//
// This rule applies to every future command returning collections.
//
// # Read-only Collection Rule
//
// LLEN, LRANGE, and future collection readers:
//   - never mutate state
//   - never trigger persistence
//   - may lazily expire keys
//   - always preserve Redis-compatible semantics
//
// Persistence is exclusively the responsibility of mutating commands.
//
// # Concurrency Model
//
// The following concurrency model has been established for collection write operations:
//
// Always-write operations (SET, DEL, LPUSH, RPUSH) execute as:
//
//	Lock
//	  ↓
//	Read current value
//	  ↓
//	Validate
//	  ↓
//	Modify
//	  ↓
//	Write
//	  ↓
//	Unlock
//
// These operations acquire the write lock immediately because:
//   - They always mutate state.
//   - Reading and writing must be one atomic operation.
//   - sync.RWMutex does not support lock upgrading.
//   - Using RLock followed by Lock would introduce a race window.
//
// Contrast this with read-then-maybe-write operations (for example:
// GET, TTL, PTTL, EXPIRE, and PEXPIRE), which may begin with RLock
// because they often complete without modifying state.
// This distinction becomes the concurrency guideline for all future collection commands.
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
//   - EXPIRETIME command.
//   - Absolute expiration timestamp queries.
//   - Unix second timestamp observation.
//   - PEXPIRETIME command.
//   - Absolute expiration timestamp queries in milliseconds.
//
// Expiration observation now supports:
//   - remaining lifetime
//   - absolute expiration time
//   - seconds
//   - milliseconds
//
// through one shared expiration implementation.
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
//
// Absolute expiration timestamps can be assigned in both seconds and milliseconds.
// Absolute expiration observation is available in both Unix seconds and milliseconds.
// No additional expiration commands are planned before v1.0.
//
// Absolute timestamps must be in the future.
//
// Only LPUSH is implemented during Sprint 21. Other collection commands remain
// future work.
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
// The original Value abstraction now demonstrates its intended purpose. The
// combination of DataType, Data (interface/any), and ExpiresAt allows multiple
// Redis-compatible data structures without modifying the Store architecture.
// This validates the architectural decisions established during the early
// foundation sprints.
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
// EXPIRETIME reuses:
//   - ExpiresAt
//   - isExpired()
//   - lazyExpire()
//   - time.Time
//
// PEXPIRETIME reuses:
//   - ExpiresAt
//   - isExpired()
//   - lazyExpire()
//   - time.Time
//   - UnixMilli()
//
// TTL, PTTL, EXPIRETIME, and PEXPIRETIME introduce no new expiration infrastructure.
// All four reuse the same ExpiresAt field, isExpired(), and lazyExpire().
// They differ only in what they report and the unit returned to the client.
// PEXPIRETIME simply exposes the existing ExpiresAt value as a Unix timestamp in milliseconds.
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
//   - EXPIRETIME
//   - PEXPIRETIME
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
// # Completion Note
//
// Sprint 20 completes the expiration subsystem.
//
// Setting:
//   - EXPIRE
//   - PEXPIRE
//   - EXPIREAT
//   - PEXPIREAT
//
// Observing:
//   - TTL
//   - PTTL
//   - EXPIRETIME
//   - PEXPIRETIME
//
// Removing:
//   - PERSIST
//
// All nine commands share:
//   - one ExpiresAt field
//   - one isExpired() implementation
//   - one lazyExpire() implementation
//   - one background cleanup goroutine
//
// This is the intended end state of the expiration subsystem before v1.0.
//
// # Future Scope
//
// Future milestones may extend the Store with:
//
// Additional collection types:
//
//	Hashes
//	  ↓
//	Sets
//	  ↓
//	Sorted Sets
//
// The Lists subsystem is currently under active development.
//
// All future collection types will reuse the existing Value abstraction and
// type validation infrastructure.
//
// Expiration:
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
