package commands

import (
	"errors"
	"strconv"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

// HSetHandler implements the HSET command.
type HSetHandler struct {
	store *store.MemoryStore
}

// NewHSetHandler returns a new HSetHandler.
func NewHSetHandler(store *store.MemoryStore) *HSetHandler {
	return &HSetHandler{store: store}
}

// Execute handles the HSET command.
func (h *HSetHandler) Execute(cmd types.Command) types.Response {
	if len(cmd.Args) < 3 || len(cmd.Args)%2 == 0 {
		return types.Response{
			Status:  types.StatusError,
			Message: "wrong number of arguments for 'hset' command",
		}
	}

	key := cmd.Args[0]
	pairs := cmd.Args[1:]

	added, err := h.store.HSet(key, pairs)
	if err != nil {
		if errors.Is(err, store.ErrWrongType) {
			return types.Response{
				Status:  types.StatusError,
				Message: store.ErrWrongType.Error(),
			}
		}
		if errors.Is(err, store.ErrInvalidArguments) {
			return types.Response{
				Status:  types.StatusError,
				Message: store.ErrInvalidArguments.Error(),
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
		Status: types.StatusInteger,
		Data:   strconv.Itoa(added),
	}
}
