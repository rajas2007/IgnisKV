package commands

import (
	"errors"
	"strconv"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

// PTTLHandler implements the PTTL command.
type PTTLHandler struct {
	store *store.MemoryStore
}

// NewPTTLHandler returns a new PTTLHandler.
func NewPTTLHandler(store *store.MemoryStore) *PTTLHandler {
	return &PTTLHandler{store: store}
}

// Execute handles the PTTL command.
func (h *PTTLHandler) Execute(cmd types.Command) types.Response {
	if len(cmd.Args) != 1 {
		return types.Response{
			Status:  types.StatusError,
			Message: "wrong number of arguments",
		}
	}

	pttl, err := h.store.PTTL(cmd.Args[0])
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
		Data:   strconv.FormatInt(pttl, 10),
	}
}
