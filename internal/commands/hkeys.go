package commands

import (
	"errors"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

// HKeysHandler implements the HKEYS command.
type HKeysHandler struct {
	store *store.MemoryStore
}

// NewHKeysHandler returns a new HKeysHandler.
func NewHKeysHandler(store *store.MemoryStore) *HKeysHandler {
	return &HKeysHandler{store: store}
}

// Execute handles the HKEYS command.
func (h *HKeysHandler) Execute(cmd types.Command) types.Response {
	if len(cmd.Args) != 1 {
		return types.Response{
			Status:  types.StatusError,
			Message: "wrong number of arguments for 'hkeys' command",
		}
	}

	key := cmd.Args[0]

	keys, err := h.store.HKeys(key)
	if err != nil {
		if errors.Is(err, store.ErrWrongType) {
			return types.Response{
				Status:  types.StatusError,
				Message: store.ErrWrongType.Error(),
			}
		}
		return types.Response{
			Status:  types.StatusError,
			Message: err.Error(),
		}
	}

	// Read-only command; no persistence required.

	return types.Response{
		Status: types.StatusArray,
		Data:   keys,
	}
}
