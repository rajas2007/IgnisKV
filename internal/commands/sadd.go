package commands

import (
	"errors"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

// SAddHandler implements the SADD command.
type SAddHandler struct {
	store *store.MemoryStore
}

// NewSAddHandler returns a new SAddHandler.
func NewSAddHandler(store *store.MemoryStore) *SAddHandler {
	return &SAddHandler{store: store}
}

// Execute handles the SADD command.
func (h *SAddHandler) Execute(cmd types.Command) types.Response {
	if len(cmd.Args) < 2 {
		return types.Response{
			Status:  types.StatusError,
			Message: "wrong number of arguments for 'sadd' command",
		}
	}

	key := cmd.Args[0]
	members := cmd.Args[1:]

	added, err := h.store.SAdd(key, members)
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

	if added > 0 {
		// Mutated, so we persist
		h.store.Save(store.DefaultSnapshotFile)
	}

	return types.Response{
		Status: types.StatusInteger,
		Data:   added,
	}
}
