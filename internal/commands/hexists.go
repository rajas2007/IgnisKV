package commands

import (
	"errors"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

// HExistsHandler implements the HEXISTS command.
type HExistsHandler struct {
	store *store.MemoryStore
}

// NewHExistsHandler returns a new HExistsHandler.
func NewHExistsHandler(store *store.MemoryStore) *HExistsHandler {
	return &HExistsHandler{store: store}
}

// Execute handles the HEXISTS command.
func (h *HExistsHandler) Execute(cmd types.Command) types.Response {
	if len(cmd.Args) != 2 {
		return types.Response{
			Status:  types.StatusError,
			Message: "wrong number of arguments for 'hexists' command",
		}
	}

	key := cmd.Args[0]
	field := cmd.Args[1]

	exists, err := h.store.HExists(key, field)
	if err != nil {
		if errors.Is(err, store.ErrKeyNotFound) || errors.Is(err, store.ErrKeyExpired) {
			return types.Response{
				Status: types.StatusInteger,
				Data:   "0",
			}
		}
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

	if exists {
		return types.Response{
			Status: types.StatusInteger,
			Data:   "1",
		}
	}

	return types.Response{
		Status: types.StatusInteger,
		Data:   "0",
	}
}
