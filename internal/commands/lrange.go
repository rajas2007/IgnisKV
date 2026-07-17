package commands

import (
	"errors"
	"strconv"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

// LRangeHandler handles the LRANGE command.
type LRangeHandler struct {
	store *store.MemoryStore
}

// NewLRangeHandler creates a new LRangeHandler.
func NewLRangeHandler(store *store.MemoryStore) *LRangeHandler {
	return &LRangeHandler{
		store: store,
	}
}

// Execute performs the LRANGE operation.
func (h *LRangeHandler) Execute(cmd types.Command) types.Response {
	if len(cmd.Args) != 3 {
		return types.Response{
			Status:  types.StatusError,
			Message: "wrong number of arguments",
		}
	}

	key := cmd.Args[0]

	start, err := strconv.ParseInt(cmd.Args[1], 10, 64)
	if err != nil {
		return types.Response{
			Status:  types.StatusError,
			Message: "value is not an integer or out of range",
		}
	}

	stop, err := strconv.ParseInt(cmd.Args[2], 10, 64)
	if err != nil {
		return types.Response{
			Status:  types.StatusError,
			Message: "value is not an integer or out of range",
		}
	}

	list, err := h.store.LRange(key, start, stop)
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

	return types.Response{
		Status: types.StatusArray,
		Data:   list,
	}
}
