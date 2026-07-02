# IgnisKV Architecture Index

Welcome to the **IgnisKV** architectural documentation. This index serves as the central hub for design principles, architecture diagrams, decision records, package layouts, and system guidelines.

---

## 1. Architecture Overview

IgnisKV is built using a **Component-Based Architecture**. Subsystems are modeled as decoupled, self-contained packages (components) that represent core domains rather than generic model-view-controller (MVC) layers. Key principles governing this design include:
* **Simplicity and Directness**: High cohesion inside packages, low coupling between packages.
* **Natural Evolution**: Packages are created incrementally only when actual functional implementations exist. We strictly avoid creating placeholder or boilerplate directories.
* **Unidirectional Request Pipeline**: Control flows sequentially in one direction, preventing cross-cutting concerns from bleeding into database core logic.

---

## 2. Architecture Decision Records (ADRs)

### What is an ADR?
An Architectural Decision Record (ADR) is a document that captures an important architectural decision, the context in which it was made, its rationale, alternatives considered, and its consequences.

### Why do they exist?
We use ADRs to:
* Ensure design transparency for developers and open-source contributors.
* Maintain consistency by documenting the constraints and trade-offs of decisions.
* Speed up onboarding by explaining *why* the codebase is structured a certain way.

### Naming & Numbering Convention
All ADRs are stored under `docs/architecture/adr/` and follow the format:
`ADR-[number]-[slug].md` (e.g., `ADR-001-project-vision.md`).
The numbers are sequentially incremented starting at `001`.

### How to Add a Future ADR
1. **Draft**: Create a copy of the ADR template and populate the context, problem statement, decisions, and options.
2. **Review**: Open a Pull Request (PR) for review by the maintainers and the technical architect.
3. **Approve**: Once consensus is reached, change the status to `Accepted` and merge it.

### Current ADR Registry

| ID | Title | Status | Date |
|:---|:---|:---|:---|
| [ADR-001](adr/ADR-001-project-vision.md) | Project Vision | Accepted | 2026-07-02 |
| [ADR-002](adr/ADR-002-engineering-philosophy.md) | Engineering Philosophy | Accepted | 2026-07-02 |
| [ADR-003](adr/ADR-003-component-architecture.md) | Component Architecture | Accepted | 2026-07-02 |
| [ADR-004](adr/ADR-004-request-pipeline.md) | Request Pipeline | Accepted | 2026-07-02 |
| [ADR-005](adr/ADR-005-store-design.md) | Store Design | Accepted | 2026-07-02 |
| [ADR-006](adr/ADR-006-command-architecture.md) | Command Architecture | Accepted | 2026-07-02 |
| [ADR-007](adr/ADR-007-value-model.md) | Value Model | Accepted | 2026-07-02 |
| [ADR-008](adr/ADR-008-response-model.md) | Response Model | Accepted | 2026-07-02 |
| [ADR-009](adr/ADR-009-dependency-direction.md) | Dependency Direction | Accepted | 2026-07-02 |
| [ADR-010](adr/ADR-010-testing-philosophy.md) | Testing Philosophy | Accepted | 2026-07-02 |

---

## 3. Architecture Diagrams

To visualize the system's operational flows and boundaries, review the following diagrams:
* **[Request Flow Diagram](diagrams/request-flow.md)**: Details the lifecycle of a client command passing through the database engine.
* **[Package Dependency Graph](diagrams/dependency-graph.md)**: Visualizes import rules and boundary conditions preventing circular imports.
* **[Package Layout map](diagrams/package-layout.md)**: Maps the physical directory tree structure and responsibilities of each folder.

---

## 4. Development Rules & Guidelines

When contributing to IgnisKV, you must follow these rules:
1. **Zero Global Variables**: Do not declare package-level mutable variables. State must be instantiated, encapsulated, and passed explicitly.
2. **Downward-Only Imports**: Check imports before committing. Circular dependencies are a sign of architectural failure.
3. **No Premature Abstraction**: Do not design interface wrappers (e.g., `Store` interface) until there are at least two distinct implementations requiring polymorphism.
4. **Test Before Commit**: Every new component or command handler must be accompanied by comprehensive unit tests validating correct and boundary conditions.
