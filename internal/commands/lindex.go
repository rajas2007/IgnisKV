package commands

import (
	"errors"
	"strconv"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

// LIndexHandler handles the LINDEX command.
// LINDEX is a read-only lookup that returns a single element by index.
// It supports both positive and negative indices.
// Missing keys and out-of-range indices return Nil.
// It never performs persistence.
type LIndexHandler struct {
	store *store.MemoryStore
}

// NewLIndexHandler creates a new LIndexHandler.
func NewLIndexHandler(store *store.MemoryStore) *LIndexHandler {
	return &LIndexHandler{
		store: store,
	}
}

// Execute performs the LINDEX operation.
// It validates the argument count, parses the index, and delegates to the Store.
// WRONGTYPE errors are translated into StatusError responses.
// Missing keys and out-of-range indices are translated into StatusNil.
// No persistence is triggered because LINDEX is read-only.
func (h *LIndexHandler) Execute(cmd types.Command) types.Response {
	if len(cmd.Args) != 2 {
		return types.Response{
			Status:  types.StatusError,
			Message: "wrong number of arguments",
		}
	}

	key := cmd.Args[0]

	index, err := strconv.ParseInt(cmd.Args[1], 10, 64)
	if err != nil {
		return types.Response{
			Status:  types.StatusError,
			Message: "value is not an integer or out of range",
		}
	}

	val, err := h.store.LIndex(key, index)
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

	if val == "" {
		return types.Response{
			Status: types.StatusNil,
		}
	}

	return types.Response{
		Status: types.StatusString,
		Data:   val,
	}
}
