package commands

import (
	"errors"
	"strconv"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

// PExpireTimeHandler implements the PEXPIRETIME command.
type PExpireTimeHandler struct {
	store *store.MemoryStore
}

// NewPExpireTimeHandler returns a new PExpireTimeHandler.
func NewPExpireTimeHandler(store *store.MemoryStore) *PExpireTimeHandler {
	return &PExpireTimeHandler{store: store}
}

// Execute handles the PEXPIRETIME command.
func (h *PExpireTimeHandler) Execute(cmd types.Command) types.Response {
	if len(cmd.Args) != 1 {
		return types.Response{
			Status:  types.StatusError,
			Message: "wrong number of arguments",
		}
	}

	ts, err := h.store.PExpireTime(cmd.Args[0])
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
