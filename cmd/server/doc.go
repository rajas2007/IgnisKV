// Package server is the executable entry point for IgnisKV.
//
// It provides an interactive command-line interface — a read-eval-print loop
// (REPL) — that allows users to interact with the database locally. It
// bootstraps all components, wires them together, and starts the application.
// It contains no database logic of its own.
//
// # Responsibilities
//
// This package owns the application boundary. It creates the MemoryStore,
// constructs the Dispatcher, and starts the REPL loop. Within the REPL loop
// it owns three additional concerns: parsing user input into structured
// commands, dispatching those commands through the execution layer, and
// printing the resulting responses to the terminal.
//
// # What does not belong here
//
// Storage logic belongs in internal/store. Command execution and validation
// belong in internal/commands. Wire-protocol serialisation belongs in
// internal/protocol. This package coordinates those components; it does not
// reimplement them. Keeping business logic out of the application entry point
// ensures that every component remains independently testable, and that
// replacing the CLI with a TCP server in v0.2 requires no changes to the
// components themselves.
//
// # Request pipeline
//
// User input arrives as a line of text. The CLI parser tokenises it and
// produces a types.Command value. The Dispatcher routes that value to the
// registered handler, which calls the MemoryStore and returns a types.Response.
// The printer formats the response for the terminal. The REPL then inspects
// Response.Status to decide whether to continue or exit. The loop terminates
// when a handler returns types.StatusExit; it never inspects command names
// directly.
//
// # Interactions
//
// This package imports internal/commands for the Dispatcher and handler
// registry, internal/store for the MemoryStore constructor, and internal/types
// for the Command and Response models. It is not imported by any other package;
// it is a binary entry point, not a library.
package main
