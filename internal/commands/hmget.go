package commands

import (
	"errors"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

// HMGetHandler implements the HMGET command.
type HMGetHandler struct {
	store *store.MemoryStore
}

// NewHMGetHandler returns a new HMGetHandler.
func NewHMGetHandler(store *store.MemoryStore) *HMGetHandler {
	return &HMGetHandler{store: store}
}

// Execute handles the HMGET command.
func (h *HMGetHandler) Execute(cmd types.Command) types.Response {
	if len(cmd.Args) < 2 {
		return types.Response{
			Status:  types.StatusError,
			Message: "wrong number of arguments for 'hmget' command",
		}
	}

	key := cmd.Args[0]
	fields := cmd.Args[1:]

	result, err := h.store.HMGet(key, fields)
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
		Data:   result,
	}
}
