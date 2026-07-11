package commands

import (
	"errors"
	"strconv"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

// TTLHandler implements the TTL command.
type TTLHandler struct {
	store *store.MemoryStore
}

// NewTTLHandler returns a new TTLHandler.
func NewTTLHandler(store *store.MemoryStore) *TTLHandler {
	return &TTLHandler{store: store}
}

// Execute handles the TTL command.
func (h *TTLHandler) Execute(cmd types.Command) types.Response {
	if len(cmd.Args) != 1 {
		return types.Response{
			Status:  types.StatusError,
			Message: "wrong number of arguments",
		}
	}

	ttl, err := h.store.TTL(cmd.Args[0])
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
		Data:   strconv.FormatInt(ttl, 10),
	}
}
