package commands

import (
	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

// Dispatcher routes parsed commands to their registered CommandHandler implementations.
// It is the central coordinator of the command execution layer. It does not execute
// commands itself, validate arguments, or interact with the storage engine directly.
// Its sole responsibility is looking up the correct handler and delegating execution.
type Dispatcher struct {
	// handlers maps each command name to the CommandHandler responsible for executing it.
	// The map is populated at construction time and never modified afterwards.
	handlers map[string]CommandHandler
}

// NewDispatcher allocates a Dispatcher and registers all built-in command handlers.
// Handlers receive their dependencies at this point through their own constructors.
// No handler registration is required by the caller after this function returns.
func NewDispatcher(s *store.MemoryStore) *Dispatcher {
	d := &Dispatcher{
		handlers: make(map[string]CommandHandler),
	}

	d.handlers["PING"] = NewPingHandler()
	d.handlers["SET"] = NewSetHandler(s)
	d.handlers["LPUSH"] = NewLPushHandler(s)
	d.handlers["RPUSH"] = NewRPushHandler(s)
	d.handlers["LLEN"] = NewLLenHandler(s)
	d.handlers["GET"] = NewGetHandler(s)
	d.handlers["DEL"] = NewDelHandler(s)
	d.handlers["TTL"] = NewTTLHandler(s)
	d.handlers["PTTL"] = NewPTTLHandler(s)
	d.handlers["EXPIRETIME"] = NewExpireTimeHandler(s)
	d.handlers["PEXPIRETIME"] = NewPExpireTimeHandler(s)
	d.handlers["EXPIRE"] = NewExpireHandler(s)
	d.handlers["PEXPIRE"] = NewPExpireHandler(s)
	d.handlers["PERSIST"] = NewPersistHandler(s)
	d.handlers["EXPIREAT"] = NewExpireAtHandler(s)
	d.handlers["PEXPIREAT"] = NewPExpireAtHandler(s)
	d.handlers["SAVE"] = NewSaveHandler(s)
	d.handlers["HELP"] = NewHelpHandler()
	d.handlers["QUIT"] = NewQuitHandler()

	return d
}

// Dispatch looks up the handler registered for cmd.Name and delegates execution to it.
// If no handler is registered for the given command name, it returns a StatusError
// response with the message "unknown command". It never panics or returns a nil Response.
func (d *Dispatcher) Dispatch(cmd types.Command) types.Response {
	handler, ok := d.handlers[cmd.Name]
	if !ok {
		return types.Response{
			Status:  types.StatusError,
			Message: "unknown command",
		}
	}

	return handler.Execute(cmd)
}
