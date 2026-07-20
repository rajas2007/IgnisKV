package commands

import (
	"errors"
	"strconv"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

// HDelHandler implements the HDEL command.
type HDelHandler struct {
	store *store.MemoryStore
}

// NewHDelHandler returns a new HDelHandler.
func NewHDelHandler(store *store.MemoryStore) *HDelHandler {
	return &HDelHandler{store: store}
}

// Execute handles the HDEL command.
func (h *HDelHandler) Execute(cmd types.Command) types.Response {
	if len(cmd.Args) < 2 {
		return types.Response{
			Status:  types.StatusError,
			Message: "wrong number of arguments for 'hdel' command",
		}
	}

	key := cmd.Args[0]
	fields := cmd.Args[1:]

	deleted, err := h.store.HDel(key, fields)
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

	if err := h.store.Save(store.DefaultSnapshotFile); err != nil {
		return types.Response{
			Status:  types.StatusError,
			Message: err.Error(),
		}
	}

	return types.Response{
		Status: types.StatusInteger,
		Data:   strconv.Itoa(deleted),
	}
}
