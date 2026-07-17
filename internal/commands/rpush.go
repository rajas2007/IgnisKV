package commands

import (
	"errors"
	"strconv"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

// RPushHandler implements the RPUSH command.
type RPushHandler struct {
	store *store.MemoryStore
}

// NewRPushHandler returns a new RPushHandler.
func NewRPushHandler(store *store.MemoryStore) *RPushHandler {
	return &RPushHandler{store: store}
}

// Execute handles the RPUSH command.
func (h *RPushHandler) Execute(cmd types.Command) types.Response {
	if len(cmd.Args) < 2 {
		return types.Response{
			Status:  types.StatusError,
			Message: "wrong number of arguments",
		}
	}

	key := cmd.Args[0]
	values := cmd.Args[1:]

	length, err := h.store.RPush(key, values...)
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
		Data:   strconv.FormatInt(length, 10),
	}
}
