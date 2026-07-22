package commands

import (
	"errors"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

// HStrLenHandler implements the HSTRLEN command.
type HStrLenHandler struct {
	store *store.MemoryStore
}

// NewHStrLenHandler returns a new HStrLenHandler.
func NewHStrLenHandler(store *store.MemoryStore) *HStrLenHandler {
	return &HStrLenHandler{store: store}
}

// Execute handles the HSTRLEN command.
func (h *HStrLenHandler) Execute(cmd types.Command) types.Response {
	if len(cmd.Args) != 2 {
		return types.Response{
			Status:  types.StatusError,
			Message: "wrong number of arguments for 'hstrlen' command",
		}
	}

	key := cmd.Args[0]
	field := cmd.Args[1]

	length, err := h.store.HStrLen(key, field)
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
		Data:   length,
	}
}
