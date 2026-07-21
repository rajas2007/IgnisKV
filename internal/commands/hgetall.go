package commands

import (
	"errors"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

// HGetAllHandler implements the HGETALL command.
type HGetAllHandler struct {
	store *store.MemoryStore
}

// NewHGetAllHandler returns a new HGetAllHandler.
func NewHGetAllHandler(store *store.MemoryStore) *HGetAllHandler {
	return &HGetAllHandler{store: store}
}

// Execute handles the HGETALL command.
func (h *HGetAllHandler) Execute(cmd types.Command) types.Response {
	if len(cmd.Args) != 1 {
		return types.Response{
			Status:  types.StatusError,
			Message: "wrong number of arguments for 'hgetall' command",
		}
	}

	key := cmd.Args[0]

	result, err := h.store.HGetAll(key)
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
