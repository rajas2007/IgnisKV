package commands

import "github.com/rajas2007/IgnisKV/internal/types"

// CommandHandler is the contract that every database command handler must satisfy.
// It allows the Dispatcher to execute any registered command without knowing its
// concrete type, decoupling command routing from command implementation.
//
// Every supported command — PING, SET, GET, DEL, QUIT, and any future additions —
// is implemented as a separate struct that satisfies this interface. The Dispatcher
// stores handlers using this interface type, which means new commands can be
// registered without modifying the Dispatcher itself.
//
// Execute receives a types.Command containing the command name and its parsed
// string arguments. The handler is responsible for validating those arguments,
// calling the appropriate store methods, and returning a types.Response that
// describes the outcome. Execute never reads raw input or writes to a socket; those
// concerns belong to the layers above it.
//
// Dependencies such as a *store.MemoryStore are injected through each handler's
// constructor rather than through Execute. This keeps the interface signature
// uniform across all handlers regardless of what resources they require, and makes
// each handler independently testable by allowing its dependencies to be supplied
// at construction time.
type CommandHandler interface {
	Execute(cmd types.Command) types.Response
}
