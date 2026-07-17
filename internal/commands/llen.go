package commands

import (
	"errors"
	"strconv"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

// LLenHandler implements the LLEN command.
type LLenHandler struct {
	store *store.MemoryStore
}

// NewLLenHandler creates a new LLenHandler.
func NewLLenHandler(store *store.MemoryStore) *LLenHandler {
	return &LLenHandler{store: store}
}

// Execute processes the LLEN command.
func (h *LLenHandler) Execute(cmd types.Command) types.Response {
	if len(cmd.Args) != 1 {
		return types.Response{
			Status:  types.StatusError,
			Message: "wrong number of arguments",
		}
	}

	key := cmd.Args[0]

	length, err := h.store.LLen(key)
	if err != nil {
		if errors.Is(err, store.ErrWrongType) {
			return types.Response{
				Status:  types.StatusError,
				Message: store.ErrWrongType.Error(),
			}
		}
		// Handle unexpected errors conservatively
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
