# ADR-005: Store Design

## Status
Accepted

## Date
2026-07-02

## Context
In many applications, data storage is treated as a global, package-level variable. While this makes it easy to read/write from anywhere in the codebase, it creates a massive issue for concurrency safety, testing isolation (tests running in parallel can corrupt each other's data), and clean software engineering. Furthermore, defining abstract interfaces too early can lead to unnecessary interface boilerplate.

## Problem Statement
How should the core database storage layer be represented, accessed, and secured to ensure testing isolation, thread safety, and natural extensibility?

## Decision
We implement a concrete `MemoryStore` struct under `internal/store`.
1. **Encapsulation**: The store maintains a private, unexported map: `map[string]Value`. All access to this keyspace must go through explicit public methods (e.g., `Get()`, `Set()`, `Delete()`).
2. **Zero Globals**: The store must be instantiated dynamically via a constructor function (e.g., `NewMemoryStore()`). No global store instances are permitted.
3. **Thread Safety**: An internal `sync.RWMutex` is encapsulated within the `MemoryStore` struct to coordinate safe concurrent reads and exclusive writes.
4. **Concrete Dependency**: The Store itself remains a concrete implementation rather than an interface. We will not define a `Store` interface until multiple storage backends (e.g., disk-only, clustered, proxy) actually exist. However, command handlers depend on abstractions where appropriate, and the concrete store design must be clean enough that an interface can be introduced later without major refactoring.

## Rationale
* Encapsulation prevents other packages from bypassing the read-write locks or mutating the keyspace directly.
* Eliminating global state allows unit tests to instantiate fresh, isolated database instances, ensuring test independence.
* Holding off on a `Store` interface complies with our core tenet "Avoid premature abstraction," keeping codebase navigation simple.

## Alternatives Considered
* **Alternative A: Package-level variables and global map**: Make the storage map public and access it directly across packages. Rejected because it makes thread-safety coordination impossible and breaks test isolation.
* **Alternative B: Create a generic `Store` interface immediately**: Write an interface representing all key-value operations. Rejected because we only have a single in-memory store implementation at this stage. Creating an interface now adds code clutter without immediate benefit.

## Consequences
* **Positive**: Thread-safe operations, high test isolation, clean encapsulation boundaries.
* **Negative**: Introducing a new store type later will require updating command handler signatures or defining a new interface.

## Future Evolution
When persistence (AOF) or new storage engines are introduced in later milestones, we can easily wrap `MemoryStore` with a transaction logger or extract a common storage interface.
