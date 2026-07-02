# Pending Design Decisions

This document serves as a living registry for architectural and design considerations that are open for discussion, pending research, or scheduled for decision-making in future development phases. Unlike Architecture Decision Records (ADRs), which document finalized and accepted decisions, this registry tracks ideas and trade-offs that have not yet been set in stone.

---

## 1. TTL Representation (Key Expiration Mechanics)

* **Status**: Open
* **Target Milestone**: `v0.5`
* **Problem**: How should we represent and track key-level Time-To-Live (TTL) inside the core engine?

### Considered Options

#### Option A: Store Expiration Inline within the `Value` struct
Add an optional `ExpiresAt *time.Time` field to the shared `Value` domain model.
* **Pros**: Simple to check expiration on access (passive eviction); minimal extra storage layout changes.
* **Cons**: Checking expiration requires reading the value struct; makes memory sweeps (active eviction) slow as we have to scan the entire main map.

#### Option B: Maintain a Separate Expiration Index
Maintain a secondary data structure (such as a min-heap or a sorted list/btree) mapping `ExpirationTime -> Key`.
* **Pros**: Highly efficient active eviction sweeps. The background thread only checks keys that are guaranteed to be expired by looking at the top of the heap.
* **Cons**: Higher complexity; must keep the main map and the expiration index in sync under concurrent writes.

---

## 2. Persistence Format

* **Status**: Open
* **Target Milestone**: `v0.3`
* **Problem**: What format and strategy should be used to guarantee durability on restart?

### Considered Options

#### Option A: Append-Only File (AOF)
Log every write command to disk in a sequential journal. On restart, replay the journal to reconstruct the state.
* **Pros**: Excellent durability guarantees (minimal data loss depending on `fsync` policy); simple implementation.
* **Cons**: Log files can grow indefinitely; requires an AOF rewriting/compaction mechanism.

#### Option B: Point-in-Time Snapshotting (RDB style)
Periodically serialize the entire memory store to a binary file on disk.
* **Pros**: Extremely fast startup times for large datasets; compact file representation.
* **Cons**: High risk of data loss between snapshots; writing snapshots can block the main database thread or require complex background process forks (`fork` is not idiomatic/safe in multi-threaded Go processes).

#### Option C: Hybrid Persistence Model
Use periodic snapshotting as the baseline state and append write deltas to an active journal. On recovery, load the snapshot first, then replay the remaining journal entries.
* **Pros**: Best of both worlds (small journals, fast startup times).
* **Cons**: Most complex to implement and debug.
