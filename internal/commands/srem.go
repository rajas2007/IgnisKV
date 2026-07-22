package commands

import (
	"errors"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

// SRemHandler implements the SREM command.
type SRemHandler struct {
	store *store.MemoryStore
}

// NewSRemHandler returns a new SRemHandler.
func NewSRemHandler(store *store.MemoryStore) *SRemHandler {
	return &SRemHandler{store: store}
}

// Execute handles the SREM command.
func (h *SRemHandler) Execute(cmd types.Command) types.Response {
	if len(cmd.Args) < 2 {
		return types.Response{
			Status:  types.StatusError,
			Message: "wrong number of arguments for 'srem' command",
		}
	}

	key := cmd.Args[0]
	members := cmd.Args[1:]

	removed, err := h.store.SRem(key, members)
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

	if removed > 0 {
		// Mutated, so we persist
		h.store.Save(store.DefaultSnapshotFile)
	}

	return types.Response{
		Status: types.StatusInteger,
		Data:   removed,
	}
}
