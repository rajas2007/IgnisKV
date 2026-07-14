package commands

import (
	"errors"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

// PersistHandler implements the PERSIST command, which removes the expiration
// from a key, converting it back into a persistent key without modifying
// its stored value.
type PersistHandler struct {
	// store is the storage engine this handler interacts with.
	store *store.MemoryStore
}

// NewPersistHandler returns a new PersistHandler with the provided MemoryStore
// injected as its storage dependency.
func NewPersistHandler(store *store.MemoryStore) *PersistHandler {
	return &PersistHandler{store: store}
}

// Execute handles the PERSIST command.
//
// It expects exactly one argument: the key whose expiration should be removed.
// If the argument count is incorrect, it returns a StatusError response. On
// success (the key existed and had an expiration), it triggers automatic
// persistence via store.Save(store.DefaultSnapshotFile) and returns a
// StatusInteger response of "1". If the key does not exist, has already
// expired, or has no expiration to remove, it returns a StatusInteger response
// of "0".
func (h *PersistHandler) Execute(cmd types.Command) types.Response {
	if len(cmd.Args) != 1 {
		return types.Response{
			Status:  types.StatusError,
			Message: "wrong number of arguments",
		}
	}

	result, err := h.store.Persist(cmd.Args[0])
	if err != nil {
		if errors.Is(err, store.ErrKeyNotFound) || errors.Is(err, store.ErrKeyExpired) {
			return types.Response{
				Status: types.StatusInteger,
				Data:   "0",
			}
		}
		return types.Response{
			Status:  types.StatusError,
			Message: err.Error(),
		}
	}

	if result != 1 {
		return types.Response{
			Status: types.StatusInteger,
			Data:   "0",
		}
	}

	// Trigger persistence after successful expiration removal
	if err := h.store.Save(store.DefaultSnapshotFile); err != nil {
		return types.Response{
			Status:  types.StatusError,
			Message: err.Error(),
		}
	}

	return types.Response{
		Status: types.StatusInteger,
		Data:   "1",
	}
}
