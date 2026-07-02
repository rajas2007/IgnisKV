# IgnisKV Project Roadmap

This document outlines the milestones, planned capabilities, and completion criteria for the evolution of **IgnisKV** from a basic local key-value store to a production-ready, concurrent, in-memory database.

---

## Milestone v0.1: Foundation

* **Goal**: Build the core, synchronous, in-memory database engine and interact with it locally using a Command-Line Interface (CLI).
* **Features**:
  * Concrete `MemoryStore` mapping string keys to `Value` structs.
  * Extensible `Value` model (supporting strings initially).
  * Decoupled Command Parser extracting commands and parameters.
  * Registry-based Dispatcher routing commands to their respective handlers.
  * Core command set: `PING`, `SET`, `GET`, `DEL`, `QUIT`.
  * Interactive CLI REPL for manual database interaction.
* **Deliverables**:
  * Root executable testing pipeline.
  * Internal packages: `store`, `commands`, `types`, `parser`, `dispatcher`.
  * Fully interactive console loop (CLI REPL).
* **Completion Criteria**:
  * All unit tests pass with zero errors.
  * The CLI correctly executes all core commands and handles syntax errors gracefully.
  * Exiting the CLI using `QUIT` terminates the program cleanly.

---

## Milestone v0.2: Networking

* **Goal**: Open the database engine to external clients over the network using TCP and the Redis Serialization Protocol (RESP).
* **Features**:
  * TCP Server establishing and managing socket listener events.
  * Concurrent client connection handling (basic multiplexing or connection lifecycle hooks).
  * RESP parser and serializer (handling simple strings, errors, integers, bulk strings, and arrays).
* **Deliverables**:
  * `internal/server` package managing socket connections.
  * `internal/protocol` package implementing RESP specifications.
* **Completion Criteria**:
  * A client (e.g., standard `netcat` or a Redis client) can connect over a local TCP socket.
  * Commands sent in RESP format are successfully parsed, executed by the engine, and responded to in RESP wire format.

---

## Milestone v0.3: Durability & Persistence

* **Goal**: Implement write-ahead logging to guarantee database recovery and durability on server restarts.
* **Features**:
  * Append-Only File (AOF) logger tracking write transactions (`SET`, `DEL`).
  * Startup initialization sequence that replays log files to reconstruct state.
  * Configurable fsync options (always, every second, or handled by the OS).
* **Deliverables**:
  * `internal/persistence` package.
  * Durability test suite containing state recovery scenarios.
* **Completion Criteria**:
  * Writing keys, shutting down the server, and restarting it restores the exact state prior to termination.
  * Transactions are written cleanly without corrupting the active database.

---

## Milestone v0.4: Concurrency & Scale

* **Goal**: Transition from a single-threaded server model to a high-concurrency architecture capable of handling multiple client connections safely.
* **Features**:
  * Safe concurrent state mutations using internal reader-writer locks (`sync.RWMutex`).
  * Connection pool limits and resource exhaustion controls.
  * Prevention of race conditions during parser execution and write-back.
* **Deliverables**:
  * Concurrent client load test scripts.
  * Go race detector validation pipeline.
* **Completion Criteria**:
  * Clean verification from Go's race detector under active write/read concurrency stress tests.
  * The server handles multiple concurrent TCP connections simultaneously without deadlock.

---

## Milestone v0.5: Key Expiration (TTL)

* **Goal**: Add time-to-live policies to allow transient keys to expire and release memory resources.
* **Features**:
  * Option to assign expiration durations during key creation (e.g., via `SET` parameters).
  * Passive key eviction (keys are verified for expiration on access).
  * Active key eviction (background routine scanning the database and evicting expired keys).
* **Deliverables**:
  * Extensible timestamp checks inside `Value`.
  * Background eviction thread loop.
* **Completion Criteria**:
  * Expired keys become inaccessible to `GET` commands immediately upon expiration.
  * Expired keys are deleted from memory in the background, freeing up allocated resources.

---

## Milestone v0.6: Publish / Subscribe (Pub/Sub)

* **Goal**: Implement real-time publish/subscribe channel messaging to act as a message broker.
* **Features**:
  * Commands: `SUBSCRIBE`, `UNSUBSCRIBE`, `PUBLISH`.
  * Connection state shifting (clients in subscribe mode can only execute Pub/Sub actions).
  * In-memory message bus routing messages to active subscribers.
* **Deliverables**:
  * `internal/pubsub` messaging broker package.
* **Completion Criteria**:
  * Messages published to a channel are instantaneously broadcasted to all active client sockets subscribed to that channel.
  * Clean socket cleanup occurs when a subscriber disconnects abruptly.

---

## Milestone v0.7: Verification & Performance

* **Goal**: Audit, profile, and optimize the engine's throughput and resource utilization.
* **Features**:
  * Formal microbenchmark suite for the Parser, Store, and Dispatcher.
  * Memory profiling (pprof) and allocation optimization (reducing heap allocations in parser/wire formats).
  * End-to-end load testing with concurrent simulated workloads.
* **Deliverables**:
  * Performance regression reports and benchmarking suites.
  * Optimized parser with zero-allocation slicing where possible.
* **Completion Criteria**:
  * Operations-per-second target is met on local hardware under heavy read/write ratios.
  * CPU/Memory profile shows no obvious leaks or allocation hotspots.

---

## Milestone v1.0: Production Ready

* **Goal**: Prepare the database for actual integration as a reliable, stable dependency in user software architectures.
* **Features**:
  * Production logging subsystem (structured log levels).
  * Connection management limits and client timeouts.
  * Comprehensive error-handling fallback systems.
  * Security evaluation (command authorization or simple passwords).
* **Deliverables**:
  * User reference documentation.
  * Formal production configuration schema (YAML/JSON configuration files).
* **Completion Criteria**:
  * Safe operation runs continuously over long-term stability testing periods.
  * 100% architectural conformance to specifications and coding principles.
