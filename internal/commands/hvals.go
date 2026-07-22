package commands

import (
	"errors"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

// HValsHandler implements the HVALS command.
type HValsHandler struct {
	store *store.MemoryStore
}

// NewHValsHandler returns a new HValsHandler.
func NewHValsHandler(store *store.MemoryStore) *HValsHandler {
	return &HValsHandler{store: store}
}

// Execute handles the HVALS command.
func (h *HValsHandler) Execute(cmd types.Command) types.Response {
	if len(cmd.Args) != 1 {
		return types.Response{
			Status:  types.StatusError,
			Message: "wrong number of arguments for 'hvals' command",
		}
	}

	key := cmd.Args[0]

	vals, err := h.store.HVals(key)
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
		Data:   vals,
	}
}
