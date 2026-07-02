# ADR-007: Value Model

## Status
Accepted

## Date
2026-07-02

## Context
A key-value store can support multiple data representations. If the store maps string keys directly to string values (e.g., `map[string]string`), expanding the store to support list, set, or hash models later becomes extremely difficult and requires a complete rewrite of the storage signatures and handler methods. We need a flexible in-memory data representation that supports types polymorphically.

## Problem Statement
How should database values be modeled internally to support multiple data structures cleanly while keeping initial implementation simple?

## Decision
We implement a shared domain struct called `Value` under `internal/types`:
1. **Value Structure**:
   ```go
   type Value struct {
       Type DataType
       Data any
   }
   ```
2. **Data Types**: `DataType` is implemented as an enum type representing the layout.
3. **v0.1 Limits**: In Version 0.1, we only implement and support `StringType`. The `Data` field is cast and treated exclusively as a string.
4. **Future Support**: This architecture explicitly provides the foundation for future milestones to support advanced Redis-like data structures including:
   * `ListType` (linked lists or array lists)
   * `HashType` (nested maps)
   * `SetType` (hashsets)
   * `SortedSetType` (skip-lists / search trees)
5. **No Expiration Fields**: We explicitly do NOT add any TTL (time-to-live) or expiration metadata fields (such as `ExpiresAt`) to the `Value` struct at this stage, as the design decision around TTL representation is still pending evaluation (see `docs/decisions.md`).

## Rationale
* Encapsulating the payload in a `Value` struct containing `DataType` allows the dispatcher and storage methods to verify correctness, handle type mismatches cleanly, and serialise data polymorphically.
* Using `any` for the internal payload type keeps the database engine adaptable, ready to accept complex structures without major engine refactoring.
* Excluding TTL fields maintains compliance with our core tenet "Grow architecture naturally" and keeps us focused strictly on what has been finalized.

## Alternatives Considered
* **Alternative A: Raw interface maps (`map[string]any`)**: Store raw interfaces directly in the keyspace map. Rejected because it lacks explicit type metadata, forcing runtime reflection to verify data types, which is slow and error-prone.
* **Alternative B: Type-specific storage maps**: Maintain distinct maps for each type (e.g., `stringMap`, `listMap`, `hashMap`). Rejected because it creates massive code duplication and makes key name collisions difficult to coordinate safely.

## Consequences
* **Positive**: Clean type polymorphism, simple type enforcement, and easy onboarding for new data structures.
* **Negative**: Using `any` requires type assertions when reading data, which adds minor performance overhead.

## Future Evolution
When expanding the database to include lists or sets in future milestones, we will add new enum entries to `DataType` and write new type-asserting methods inside the respective command handlers.
