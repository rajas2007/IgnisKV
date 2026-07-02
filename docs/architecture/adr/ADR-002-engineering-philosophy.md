# ADR-002: Engineering Philosophy

## Status
Accepted

## Date
2026-07-02

## Context
In new infrastructure projects, there is a strong temptation to over-engineer from day one. Developers often design deep inheritance chains, write generic interfaces for single implementations, and add complex concurrency utilities before establishing basic correctness. This can lead to unreadable code, maintenance blockages, and high refactoring overhead.

## Problem Statement
What core software engineering guidelines and philosophies will guide the day-to-day coding, package structure, and abstraction choices of IgnisKV?

## Decision
We adopt the following engineering tenets for all IgnisKV development:
1. **Build like a product**: Even if this is an educational project, treat it as a library or application that will run in production (e.g., proper error handling, logging, testing, and zero tolerance for data races).
2. **Simplicity first**: Prefer simple, clear code over clever or highly optimized constructs that are difficult to read and verify.
3. **Grow architecture naturally**: Add structures and layers only when concrete requirements demand them, rather than guessing future requirements.
4. **Avoid premature abstraction**: Do not introduce Go interfaces until there are multiple concrete implementations.
5. **One responsibility per package**: Every package must have a single, highly focused purpose.
6. **Refactor only when design improves**: Code refactoring must be driven by making the design cleaner, more decoupled, or more readable.

## Rationale
Adopting these guidelines ensures that the codebase remains accessible to learners while displaying the highest standards of professional software craftsmanship. It keeps the development velocity high by avoiding the cognitive load of unnecessary layers of abstraction.

## Alternatives Considered
* **Alternative A: Design a highly generic framework first**: Write interfaces for every package, abstract the transport layer immediately, and prepare for distributed clustering on day one. Rejected because it violates "Simplicity First" and results in bloated, hard-to-follow boilerplate.
* **Alternative B: Write quick, unstructured code to get features working**: Rejected because it violates "Build like a product" and creates technical debt that makes future features (like concurrency and persistence) impossible to implement cleanly.

## Consequences
* **Positive**: The codebase remains small, readable, highly testable, and naturally modular.
* **Negative**: Refactoring is expected as we move between milestones (e.g., adding interfaces when we introduce multiple stores or persistence formats later).

## Future Evolution
These guidelines will serve as code review standards. Any PR that introduces abstractions without immediate, concrete use cases will be rejected.
