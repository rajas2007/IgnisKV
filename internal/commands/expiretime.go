package commands

import (
	"errors"
	"strconv"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

// ExpireTimeHandler implements the EXPIRETIME command.
type ExpireTimeHandler struct {
	store *store.MemoryStore
}

// NewExpireTimeHandler returns a new ExpireTimeHandler.
func NewExpireTimeHandler(store *store.MemoryStore) *ExpireTimeHandler {
	return &ExpireTimeHandler{store: store}
}

// Execute handles the EXPIRETIME command.
func (h *ExpireTimeHandler) Execute(cmd types.Command) types.Response {
	if len(cmd.Args) != 1 {
		return types.Response{
			Status:  types.StatusError,
			Message: "wrong number of arguments",
		}
	}

	ts, err := h.store.ExpireTime(cmd.Args[0])
	if err != nil {
		if errors.Is(err, store.ErrKeyNotFound) || errors.Is(err, store.ErrKeyExpired) {
			return types.Response{
				Status: types.StatusInteger,
				Data:   "-2",
			}
		}
		return types.Response{
			Status:  types.StatusError,
			Message: err.Error(),
		}
	}

	return types.Response{
		Status: types.StatusInteger,
		Data:   strconv.FormatInt(ts, 10),
	}
}
