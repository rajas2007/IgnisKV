# ADR-008: Response Model

## Status
Accepted

## Date
2026-07-02

## Context
If database command executors or storage components return raw protocol-serialized payloads (such as raw RESP bytes like `+OK\r\n`), then core components become directly dependent on the networking protocol. If we later decide to support a HTTP JSON API or a custom binary protocol, we would be forced to rewrite or duplicate the core store and command execution logic.

## Problem Statement
How should operation outputs be structured, returned, and translated to ensure that core execution modules are isolated from client serialization formats?

## Decision
We decouple database output representations into distinct layers:
1. **Store Primitives**: Store methods return standard, idiomatic Go values (e.g., `string`, `bool`, `error`). The store does not know about command structures or wire formats.
2. **Structured Command Response**: Command handlers process commands and package their outcome into a structured `Response` domain object defined under `internal/types`. This object contains:
   * **Status**: An execution status code (e.g., Success, Error, ConnectionClose).
   * **Payload**: The logical output content (such as a string slice, an integer, or nil).
3. **Protocol Serialization**: The output serialization component (such as CLI output formatter in `v0.1` or RESP serializer in `v0.2`) receives the generic `Response` object and converts it into the appropriate wire format (e.g., raw text console printout or standard RESP string like `*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n`).

## Rationale
This design respects the boundaries of the request pipeline. By separating logical execution outputs from wire formatting, we ensure that:
* Command handlers and database maps remain clean, reusable, and easy to unit test.
* Swapping or adding new serialization protocols (such as migrating from a text terminal shell to a binary TCP protocol) requires zero changes to the dispatcher, the store, or command handler files.

## Alternatives Considered
* **Alternative A: Handlers return RESP byte streams directly**: Have command handlers format their own responses directly into wire formats. Rejected because it binds command logic to a specific networking protocol format, making local testing or alternative protocol integrations highly complex.
* **Alternative B: Store returns serialized formats**: Have the store methods output standard response formats. Rejected because the store should remain a clean, protocol-agnostic, low-level data structure manager.

## Consequences
* **Positive**: High protocol flexibility, clean separation of concerns, and simple mock validation in unit tests.
* **Negative**: Command execution outcomes must be mapped twice (first from Store to Response, then from Response to wire formats), adding small performance and allocation overheads.

## Future Evolution
When introducing networking in `v0.2`, the `internal/protocol` package will consume these structured `Response` objects and serialize them into standard RESP formatting, while the core database code remains unchanged.
