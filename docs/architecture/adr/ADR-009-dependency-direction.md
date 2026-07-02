# ADR-009: Dependency Direction

## Status
Accepted

## Date
2026-07-02

## Context
Circular dependencies are a major issue in Go. If Package A imports Package B, and Package B imports Package A, the Go compiler will fail to compile. As a project grows, dependencies can easily become tangled, leading developers to merge unrelated packages or introduce interface hacks to bypass compiler errors.

## Problem Statement
How should package import directions be structured to guarantee compilation safety, maintain clear architectural boundaries, and eliminate circular dependencies?

## Decision
We enforce a strict downward-only dependency hierarchy. Package imports must flow sequentially in one direction. To maintain this architecture:
1. **Shared domain types package**: Shared structures, values, enums, and common objects are placed in a leaf package named `internal/types`. This provides common representations used across multiple components while avoiding circular dependencies.
2. **Package Graph**: We establish and enforce the following package dependency hierarchy:
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
   * `internal/store` references `internal/types` (leaf).
   * `internal/commands` references `internal/store` and `internal/types`.
   * `internal/protocol` references `internal/commands` and `internal/types`.
   * `internal/server` references `internal/protocol` and `internal/commands`.
   * `cmd/server` references `internal/server`.
3. **No upward imports**: A package located lower in the hierarchy (e.g., `store`) is strictly prohibited from importing packages located higher up (e.g., `commands` or `protocol`).

## Rationale
* Isolating shared data models (like `Value` and `DataType`) in `internal/types` allows `store` and `commands` to share representations without creating cyclic relationships.
* Enforcing a clear dependency flow makes the codebase easier to understand and prevents boundary leaks.

## Alternatives Considered
* **Alternative A: Allow packages to import each other at will**: Rejected because the Go compiler prohibits cyclic imports, leading to compilation failures.
* **Alternative B: Create a single massive `internal/` package with no sub-packages**: Rejected because it violates our tenet "One responsibility per package" and makes code organization impossible.

## Consequences
* **Positive**: Guarantees compilation, enforces clean boundary contracts, and prevents spaghetti imports.
* **Negative**: Structs that start off local to a package must be relocated to `internal/types` if they need to be referenced by other components later.

## Future Evolution
This dependency rule will be audited as part of code reviews. Any imports that violate the hierarchy or create package loops will be rejected immediately.
