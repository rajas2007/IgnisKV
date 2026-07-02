# IgnisKV Specification — Version 0.1

This document specifies the behavior, interface, limits, and architecture of **IgnisKV** version 0.1. Version 0.1 is the foundational release of the database engine, focusing purely on in-memory storage correctness and local CLI command execution.

---

## 1. Project Scope

The scope of Version 0.1 is strictly limited to a local-only key-value storage engine. The application executes as a single-process, single-threaded interactive Command-Line Interface (CLI) REPL (Read-Eval-Print Loop). 

All interactions occur via standard input (`stdin`) and standard output (`stdout`/`stderr`). There are no external networking mechanisms, socket bindings, or serialization protocols implemented in this release.

---

## 2. Supported Commands

The CLI parser processes incoming lines of text as commands. The commands supported in Version 0.1 are:

### 2.1 PING
Returns a pong verification to check that the database is active.
* **Syntax**: `PING [message]`
* **Behavior**: If an optional message is provided, returns that message as-is. If no message is provided, returns `"PONG"`.
* **Output Example**: `PONG` or `hello`

### 2.2 SET
Associates a key with a value in the memory store. If the key already exists, overwrites its value.
* **Syntax**: `SET <key> <value>`
* **Behavior**: Maps the literal string `<key>` to `<value>`.
* **Output Example**: `OK`

### 2.3 GET
Retrieves the value associated with the specified key.
* **Syntax**: `GET <key>`
* **Behavior**: Looks up the `<key>`. Returns the string value if found, or an error if the key does not exist.
* **Output Example**: `"some_value"` or `(nil)` (representing absence of key)

### 2.4 DEL
Deletes a key-value mapping from the memory store.
* **Syntax**: `DEL <key>`
* **Behavior**: Removes the `<key>` mapping from the store. Returns the number of keys successfully deleted (either `1` or `0`).
* **Output Example**: `(integer) 1` or `(integer) 0`

### 2.5 QUIT
Terminates the interactive session cleanly.
* **Syntax**: `QUIT`
* **Behavior**: Flushes state if necessary, exits the main execution loop, and closes the process with an exit status code of `0`.
* **Output Example**: `Bye!` (followed by terminal exit)

---

## 3. Supported Data Types

Version 0.1 supports only one primary primitive data representation:
* **StringType**: Raw, binary-safe strings. All values provided to the database are parsed, stored, and retrieved as Go strings. 

No complex types (such as lists, sets, hashes, or sorted sets) are supported in this version.

---

## 4. Storage Model

Storage is entirely volatile and resides in memory. 
* **Data Structure**: A core Go `map[string]Value`, where `Value` is a struct containing:
  * `DataType` (an enum representing the type of data, currently only `StringType`).
  * `Data` (the raw string representation).
* **Concurrency Protection**: The store encapsulates its map behind a `sync.RWMutex`. In Version 0.1, although the loop runs synchronously in a single thread, the mutex is introduced to establish the correct architectural boundaries for multi-threading in future milestones.
* **Global Access**: Global variables are strictly prohibited. The database instantiation must return an isolated instance pointer to a concrete `MemoryStore` struct.

---

## 5. Architecture Overview

The system is organized into a clean, unidirectional processing pipeline:

```
[Stdin / CLI Reader] ────► [Parser] ────► [Dispatcher] ────► [Command Handlers] ────► [MemoryStore] ────► [Output Writer]
```

1. **CLI Reader**: Listens on `stdin` for a line terminated by a newline.
2. **Parser**: Tokens are parsed out of the raw line. The first token determines the command; subsequent tokens represent the command arguments.
3. **Dispatcher**: The parsed command is matched against a registry of handlers.
4. **Command Handlers**: Execute specific business logic (e.g., calling `Store.Get()` or `Store.Set()`) and package the outcome into a structured response.
5. **MemoryStore**: The underlying database engine containing the hash map.
6. **Output Writer**: Converts the handler response into standard output format and prints it to the console.

---

## 6. Project Limitations

* **Single-Process**: Must be run in the foreground of the current shell.
* **No Persistence**: Restarting the process resets the database state to empty.
* **No TCP Listener**: Cannot receive connections from external client tools (e.g., `redis-cli`).
* **No TTL**: Keys remain in memory indefinitely until deleted via `DEL` or the process terminates.

---

## 7. Out of Scope

The following items are explicitly excluded from Version 0.1:
* TCP socket listening, socket connections, or multiplexing.
* Redis Serialization Protocol (RESP) format parsing and formatting.
* Persistence journaling (AOF) or snapshots.
* Key eviction algorithms or TTL metadata.
* Multi-key complex transactions.

---

## 8. Success Criteria

Version 0.1 is considered successfully completed when:
1. **Compilation**: The source code compiles cleanly using the default Go compiler (`go build`) without compiler warnings.
2. **Unit Tests**: A comprehensive suite of unit tests validates the behavior of the `Store`, `Parser`, `Dispatcher`, and each individual command handler independently. All tests must pass.
3. **Correct Behavior**:
   * Executing `SET key value` followed by `GET key` prints `value`.
   * Executing `DEL key` prints `(integer) 1` and subsequent `GET key` yields `(nil)`.
   * Entering invalid commands returns clear, readable error messages (e.g., `ERR unknown command`).
