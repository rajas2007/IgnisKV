package commands

import (
	"errors"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

// DelHandler implements the DEL command, which removes a key from the database.
// It translates storage-layer outcomes into protocol-independent Response values,
// distinguishing between a successful deletion, an absent key, and an unexpected
// error.
type DelHandler struct {
	// store is the storage engine this handler removes keys from.
	store *store.MemoryStore
}

// NewDelHandler returns a new DelHandler with the provided MemoryStore injected
// as its storage dependency.
func NewDelHandler(store *store.MemoryStore) *DelHandler {
	return &DelHandler{store: store}
}

// Execute handles the DEL command.
//
// It expects exactly one argument: the key to delete. If the argument count is
// incorrect it returns a StatusError response. When the key exists and is
// removed successfully, it returns a StatusOK response. When the key does not
// exist, it returns a StatusNil response to distinguish absence from an error
// condition. Any other storage error is returned as a StatusError response.
func (h *DelHandler) Execute(cmd types.Command) types.Response {
	if len(cmd.Args) != 1 {
		return types.Response{
			Status:  types.StatusError,
			Message: "wrong number of arguments",
		}
	}

	err := h.store.Delete(cmd.Args[0])
	if err != nil {
		if errors.Is(err, store.ErrKeyNotFound) {
			return types.Response{
				Status: types.StatusNil,
			}
		}
		return types.Response{
			Status:  types.StatusError,
			Message: err.Error(),
		}
	}

	if err := h.store.Save(store.DefaultSnapshotFile); err != nil {
		return types.Response{
			Status:  types.StatusError,
			Message: err.Error(),
		}
	}

	return types.Response{
		Status:  types.StatusOK,
		Message: "OK",
	}
}
