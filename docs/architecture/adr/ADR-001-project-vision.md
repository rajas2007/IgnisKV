# ADR-001: Project Vision

## Status
Accepted

## Date
2026-07-02

## Context
When embarking on learning systems programming and database design, developers often either build toy projects that lack architectural rigor or attempt to rewrite complex production tools (like Redis) line-for-line, getting bogged down in extensive feature parity. We need a clear guiding vision for the IgnisKV project that establishes its purpose and boundaries.

## Problem Statement
What is the primary objective of IgnisKV, and what are the core boundaries that guide its design and feature prioritization?

## Decision
We define IgnisKV as a Redis-inspired, in-memory, key-value database written in Go. The goal is to explore systems programming, networking, concurrency, durability, and persistence using production-quality software engineering practices. We explicitly prioritize software craftsmanship, design patterns, testing, and clean architecture over replicating Redis feature-for-feature.

## Rationale
By establishing that IgnisKV is a teaching and learning vehicle built to production standards rather than a drop-in Redis clone:
1. We can focus on the architectural concepts of database design (parsing, dispatching, thread-safe memory storage, persistence formats).
2. We prevent scope creep.
3. We encourage writing highly readable, well-documented, and thoroughly tested code instead of optimized but opaque hacks.

## Alternatives Considered
* **Alternative A: Rebuild Redis exactly**: Rejected because it would require implementing hundreds of commands and protocol edge cases, adding massive scope without adding core educational value.
* **Alternative B: Create a simple, un-architected toy memory store**: Rejected because it would fail to teach the production-grade engineering principles (concurrency safety, clean package separation, logging, error recovery) that are key to the project.

## Consequences
* **Positive**: Clean boundaries, clear educational focus, and high standards of implementation.
* **Negative**: The project is not intended to be a drop-in replacement for Redis in existing highly scaled production environments.

## Future Evolution
As the project matures from a single-process engine to a distributed database or a multi-model store, this decision will continue to govern new features. We will only add features that provide architectural learning opportunities.
