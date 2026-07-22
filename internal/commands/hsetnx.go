package commands

import (
	"errors"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

// HSetNXHandler implements the HSETNX command.
type HSetNXHandler struct {
	store *store.MemoryStore
}

// NewHSetNXHandler returns a new HSetNXHandler.
func NewHSetNXHandler(store *store.MemoryStore) *HSetNXHandler {
	return &HSetNXHandler{store: store}
}

// Execute handles the HSETNX command.
func (h *HSetNXHandler) Execute(cmd types.Command) types.Response {
	if len(cmd.Args) != 3 {
		return types.Response{
			Status:  types.StatusError,
			Message: "wrong number of arguments for 'hsetnx' command",
		}
	}

	key := cmd.Args[0]
	field := cmd.Args[1]
	value := cmd.Args[2]

	set, err := h.store.HSetNX(key, field, value)
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

	if set {
		// Mutated, so we persist
		h.store.Save(store.DefaultSnapshotFile)
		return types.Response{
			Status: types.StatusInteger,
			Data:   1,
		}
	}

	// No mutation, do not persist
	return types.Response{
		Status: types.StatusInteger,
		Data:   0,
	}
}
