# Package Dependency Graph

This document details the direction of package dependencies in IgnisKV. We enforce a strict downward-pointing DAG (Directed Acyclic Graph) to prevent circular imports and maintain clean boundary interfaces.

---

## 1. ASCII Dependency Graph

```
                 cmd/server
                      │
                      ▼
              internal/server
                      │
                      ▼
             internal/protocol
                      │
                      ▼
             internal/commands
                 │          │
                 ▼          ▼
         internal/types  internal/store
                 ▲          │
                 └──────────┘
```

---

## 2. Package Roles and Boundaries

### cmd/server
The server bootstrapper. It parses runtime flags, sets up config files, and instantiates components from the `internal/server` package.

### internal/server
Owns the connection loop, TCP listener, client session tracking, and socket lifecycle hooks. (Planned for `v0.2`).

### internal/protocol
Handles deserialization of wire data (such as RESP payloads) into raw tokens, and serialization of database structures back into wire formats. (Planned for `v0.2`).

### internal/commands
Contains individual command handlers (e.g., `GetHandler`, `SetHandler`) and the Dispatcher registry. It acts as the orchestration layer between incoming parsed inputs and core database storage.

### internal/store
The state storage engine (`MemoryStore`). It has no knowledge of network protocols, CLI interfaces, or specific commands. It only implements raw primitives like keyspace checks, retrievals, and updates.

### internal/types
A leaf package containing shared domain definitions, custom errors, value structures, and data types (such as `Value` and `DataType`). Placing these shared types here allows packages like `store`, `commands`, and `protocol` to use them without referencing each other directly, eliminating circular dependency loops.
