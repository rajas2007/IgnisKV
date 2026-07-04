package commands

import (
	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

// SetHandler implements the SET command, which stores a string value under a
// given key in the database. It holds a reference to the MemoryStore through
// which all write operations are performed.
type SetHandler struct {
	// store is the storage engine this handler writes values to.
	store *store.MemoryStore
}

// NewSetHandler returns a new SetHandler with the provided MemoryStore injected
// as its storage dependency.
func NewSetHandler(store *store.MemoryStore) *SetHandler {
	return &SetHandler{store: store}
}

// Execute handles the SET command.
//
// It expects exactly two arguments: a key and a value. If the argument count
// is incorrect it returns a StatusError response. When valid, it constructs a
// StringType Value and stores it via MemoryStore.Set, then returns a StatusOK
// response with the message "OK".
func (h *SetHandler) Execute(cmd types.Command) types.Response {
	if len(cmd.Args) != 2 {
		return types.Response{
			Status:  types.StatusError,
			Message: "wrong number of arguments",
		}
	}

	value := types.Value{
		Type: types.StringType,
		Data: cmd.Args[1],
	}

	h.store.Set(cmd.Args[0], value)

	return types.Response{
		Status:  types.StatusOK,
		Message: "OK",
	}
}
