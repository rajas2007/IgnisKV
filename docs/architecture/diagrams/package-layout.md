# Package Layout Map

This document outlines the physical directory layout of the IgnisKV repository and maps each directory to its architectural role.

---

## 1. ASCII Package Tree Layout

```
IgnisKV/
├── cmd/
│   └── server/
│       └── main.go           # Application bootstrapper & configuration loader
│
├── docs/                     # Comprehensive system documentation
│   ├── architecture/
│   │   ├── INDEX.md          # Architecture homepage
│   │   ├── adr/              # Architecture Decision Records
│   │   └── diagrams/         # System ASCII diagrams
│   ├── decisions.md          # Live registry of open design decisions
│   ├── glossary.md           # Terminology and concept definition table
│   ├── roadmap.md            # Project milestone roadmap
│   └── specification.md      # Detailed v0.1 engine specifications
│
├── internal/                 # Private application and database engine code
│   ├── commands/             # Dispatcher registry and specific command handlers
│   ├── protocol/             # Serialization/deserialization code (planned v0.2)
│   ├── server/               # TCP listener and network socket loop (planned v0.2)
│   ├── store/                # MemoryStore in-memory keyspace mapping
│   └── types/                # Shared leaf package for domain models and enums
│
├── .gitignore                # OS, IDE, and build-output filters
├── go.mod                    # Go module dependencies and toolchain setup
├── LICENSE                   # Open-source license declarations
└── README.md                 # Project landing page and quickstart instructions
```

---

## 2. Directory Separation Rules

1. **`cmd/` vs. `internal/`**: The `cmd/` directory contains small main entrypoints. No reusable database logic should live here. All implementation details reside in package subdirectories under `internal/`.
2. **`internal/` Enforced Privacy**: Go prohibits external projects from importing packages located under `/internal/`. This guarantees that the core implementation details of IgnisKV remain private to the module.
3. **Flat and Focused Packages**: Under `internal/`, we maintain flat directories (e.g., `internal/store` contains store files, not sub-packages). We do not nest packages unnecessarily.
