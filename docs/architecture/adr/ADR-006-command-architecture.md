# ADR-006: Command Architecture

## Status
Accepted

## Date
2026-07-02

## Context
If a database engine handles command processing via a giant `switch` statement inside the core storage package (e.g., `switch cmdType { case SET: ... }`), adding new commands becomes highly error-prone and violates the Open-Closed Principle. The core store should remain focused entirely on low-level data structure updates and should be completely decoupled from command-specific execution rules, argument validation, and return behaviors.

## Problem Statement
How should client commands be represented, routed, and executed to allow adding new commands with minimal modification to other components?

## Decision
We implement a decoupled Command and Dispatcher architecture:
1. **Generic Command representation**: The `Parser` outputs a generic `Command` struct containing the command name (as a string) and a slice of string arguments.
2. **CommandHandler Interface**: Every supported command is implemented as a separate struct that implements a shared interface under `internal/commands`:
   ```go
   type CommandHandler interface {
       Execute(ctx context.Context, store *store.MemoryStore, args []string) Response
   }
   ```
3. **Dispatcher Registry**: A centralized `Dispatcher` struct manages a thread-safe registry (`map[string]CommandHandler`). The dispatcher matches the parsed command name with its handler and executes it.
4. **Store Isolation**: The core `MemoryStore` is completely unaware of the concept of commands. It only exposes primitive storage methods.

## Rationale
* Decoupling commands into distinct handler files (e.g., `set.go`, `get.go`) ensures that adding a new command requires only creating a new file and registering it, leaving existing command handlers and the store untouched.
* The store does not need to maintain logic for argument parsing or user response formatting, keeping it clean and simple.

## Alternatives Considered
* **Alternative A: Monolithic switch block in the store**: Run all command logic directly inside a method on `MemoryStore` based on the command string. Rejected because it directly couples store logic to argument formatting and makes the storage class grow indefinitely.
* **Alternative B: Commands directly call network routines**: Have command handlers write back to client sockets. Rejected because it violates the request pipeline direction and prevents reusing command handlers in non-network environments (such as tests or CLI utilities).

## Consequences
* **Positive**: High extensibility, clean compliance with the Single Responsibility Principle, and high testability for individual command behaviors.
* **Negative**: Introduces boilerplate for each command (each requires its own file, struct definition, registry mapping, and tests).

## Future Evolution
If we need to support dynamic commands or modules in the future, the Dispatcher registry can be easily modified to support runtime registration interfaces.
