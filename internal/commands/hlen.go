package commands

import (
	"errors"
	"strconv"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

// HLenHandler implements the HLEN command.
type HLenHandler struct {
	store *store.MemoryStore
}

// NewHLenHandler returns a new HLenHandler.
func NewHLenHandler(store *store.MemoryStore) *HLenHandler {
	return &HLenHandler{store: store}
}

// Execute handles the HLEN command.
func (h *HLenHandler) Execute(cmd types.Command) types.Response {
	if len(cmd.Args) != 1 {
		return types.Response{
			Status:  types.StatusError,
			Message: "wrong number of arguments for 'hlen' command",
		}
	}

	key := cmd.Args[0]

	length, err := h.store.HLen(key)
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

	// Read-only command; no persistence required.

	return types.Response{
		Status: types.StatusInteger,
		Data:   strconv.Itoa(length),
	}
}
