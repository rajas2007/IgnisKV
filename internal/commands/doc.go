// Package commands implements the command execution layer of IgnisKV.
//
// It occupies the third tier of the request pipeline, sitting between the
// parser that produces Command values and the storage engine that manages
// the keyspace. Its responsibility is to receive parsed commands, validate
// their arguments, call the appropriate MemoryStore methods, and return
// structured Response values to the layer above it.
//
// # Structure
//
// The package is organised around two collaborating components. The Dispatcher
// maintains a registry that maps command names to their handlers. Each command
// is implemented as a separate handler that satisfies the CommandHandler
// interface. When a Command arrives, the Dispatcher looks up the handler by
// name and delegates execution to it.
//
// # What belongs here
//
// Command handler implementations, the CommandHandler interface, the Dispatcher
// and its handler registry, and argument validation logic. Each command that
// IgnisKV supports corresponds to one handler file in this package.
//
// # What does not belong here
//
// Raw input reading, RESP parsing, socket I/O, and output formatting. Handlers
// receive types.Command values, not byte slices or protocol frames. They return
// types.Response values, not serialised wire data. The conversion between wire
// format and domain types belongs in internal/protocol; the conversion between
// domain responses and client output belongs in the output layer.
//
// Storage implementation also does not belong here. Handlers call exported
// methods on a *store.MemoryStore. They do not access the underlying map, own
// synchronisation primitives, or duplicate storage logic.
//
// # Interactions
//
// This package imports internal/types for the Command and Response models and
// imports internal/store to receive a *store.MemoryStore through handler
// constructors. It is imported by internal/protocol (v0.2) and cmd/server,
// which wire the dispatcher into the request loop. It never imports either of
// those packages; the dependency flows strictly downward.
//
// # Why this separation exists
//
// Isolating command execution from the storage engine means that the store can
// be tested and reasoned about with no knowledge of what commands exist.
// Isolating it from the protocol layer means that command handlers do not
// change when the wire format changes. Each component is independently
// testable and replaceable within its own boundary.
package commands
