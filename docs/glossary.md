# IgnisKV Glossary

This glossary defines the core terminology, components, and concepts used throughout the **IgnisKV** codebase and system documentation.

---

### Store
The state management engine of the database. In IgnisKV, this is typically a concrete component (`MemoryStore`) that isolates raw state (hash maps) and provides synchronized read/write methods to query or update keys.

### Value
The internal data representation associated with a database key. A Value encapsulates both the raw payload (data) and metadata indicating its core type (e.g., `StringType`).

### Command
An instruction parsed from client input representing a specific database operation (e.g., `SET`, `GET`, `DEL`). It contains the operation identity and any associated arguments or options.

### Handler (Command Handler)
An interface-driven component representing the execution logic of a specific command. A handler accepts the command parameters, runs the appropriate core database operations on the Store, and returns a unified Response structure.

### Dispatcher
A central registry component that matches incoming command names with their corresponding command handler implementation and manages routing command execution.

### Parser
The pipeline component responsible for taking raw input bytes or strings from the client connection (or CLI), verifying syntactical correctness, and transforming them into structured Command objects.

### Protocol
The agreed-upon format and set of rules governing how a client and server communicate data, parse parameters, and format values over a socket connection or interface.

### RESP (Redis Serialization Protocol)
A simple, binary-safe serialization protocol designed for Redis. It supports data structures like simple strings, bulk strings, integers, arrays, and errors. IgnisKV plans to implement RESP for its networking protocol.

### AOF (Append-Only File)
A persistence technique where every write transaction is logged sequentially to a file on disk. This log file is replayed at startup to reconstruct the database's in-memory state.

### TTL (Time-To-Live)
A mechanism that limits the lifespan of a key-value mapping. The key is automatically marked as expired and scheduled for eviction once its duration is exceeded.

### Pub/Sub (Publish/Subscribe)
A messaging model where senders (publishers) do not programmatically address messages to specific receivers (subscribers). Instead, publishers release messages to specific channels, which are automatically distributed to all active subscribers.

### Pipeline (Request Pipeline)
The linear sequence of processing stages that a single request flows through from arrival at the server to returning a response. Each stage has a single, decoupled responsibility.

### Component
A self-contained, logically isolated package or subsystem within the database codebase (e.g., the parser package, the store package) designed with low coupling and high cohesion.

### Context
A standard object or structure passed through execution pipelines to carry configuration data, request deadlines, tracing tokens, or cancellation signals.

### Response
A structured object returned by command handlers representing the results of a command execution. It abstracts away protocol-specific details, separating logical execution output from final serialization formatting.
