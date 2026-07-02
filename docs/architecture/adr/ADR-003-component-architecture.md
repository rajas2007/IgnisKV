# ADR-003: Component Architecture

## Status
Accepted

## Date
2026-07-02

## Context
Standard enterprise applications often adopt horizontal layered architectures (such as Model-View-Controller or Clean/Hexagonal Architecture structures) which group code by technical roles: `models/`, `views/`, `controllers/`, or `repositories/`. In a database system, this structure can lead to fragmentation, forcing developers to look across multiple layers to modify a single logical feature (e.g., adding a command).

## Problem Statement
How should code and packages be organized to optimize developer navigation, minimize coupling, and keep relevant domain files grouped together?

## Decision
We adopt a **Component-Based Architecture** layout. Under this model:
1. Packages correspond to cohesive, self-contained subsystems or domains (such as `store`, `commands`, `parser`, `protocol`, `server`) rather than generic MVC technical layers.
2. Package internal files remain highly cohesive, and package interfaces expose only necessary primitives.
3. We do not pre-allocate or stub package directories. Packages are only created when their implementation code actually exists.

## Rationale
Component-based package structures:
* Keep all files relating to a single domain (e.g., command registration and command handlers) in one location (`internal/commands`).
* Enforce clear boundaries: package imports correspond to actual domain dependency boundaries.
* Avoid the clutter of empty folders or "layer boilerplate" files that contain no real logic.

## Alternatives Considered
* **Alternative A: Layered (MVC-like) Package Layout**: Organizing code by `internal/handlers`, `internal/models`, `internal/storage`. Rejected because database systems do not fit MVC models and separating the dispatcher registry from command handlers adds unnecessary import hopping.
* **Alternative B: Single flat root package**: Storing all files directly in the root directory or a single `internal/` directory. Rejected because as we add concurrency, persistence, and Pub/Sub, a single folder would become too cluttered, making it difficult to maintain strict boundaries.

## Consequences
* **Positive**: High cohesion, localized changes, clear boundaries, and zero placeholder package clutter.
* **Negative**: Developers must carefully plan where new structs reside (domain types must go to `internal/types` if referenced across multiple components to avoid cycle imports).

## Future Evolution
If a component grows too large, it may be subdivided locally (e.g., `internal/commands` could contain subfolders for groups of commands) but must maintain downward-only imports.
