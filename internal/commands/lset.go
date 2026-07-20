package commands

import (
	"errors"
	"strconv"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

// LSetHandler implements the LSET command.
// It is a mutating command that replaces an existing element at a specified index.
// It supports positive indices (starting at 0) and negative indices (counting backward from the tail).
// Missing keys return an error (ErrKeyNotFound).
// Out-of-range indices return an error (ErrIndexOutOfRange).
// Non-list keys return an error (ErrWrongType).
// It performs persistence only after a successful mutation.
type LSetHandler struct {
	store *store.MemoryStore
}

// NewLSetHandler allocates and returns a new LSetHandler.
func NewLSetHandler(s *store.MemoryStore) *LSetHandler {
	return &LSetHandler{store: s}
}

// Execute validates the command arguments, parses the index, and delegates replacement to the Store.
// It translates store errors into StatusError responses.
// Upon successful replacement, it triggers persistence by saving a snapshot.
func (h *LSetHandler) Execute(cmd types.Command) types.Response {
	if len(cmd.Args) != 3 {
		return types.Response{
			Status:  types.StatusError,
			Message: "wrong number of arguments",
		}
	}

	key := cmd.Args[0]
	indexStr := cmd.Args[1]
	value := cmd.Args[2]

	index, err := strconv.ParseInt(indexStr, 10, 64)
	if err != nil {
		return types.Response{
			Status:  types.StatusError,
			Message: "value is not an integer or out of range",
		}
	}

	err = h.store.LSet(key, index, value)
	if err != nil {
		if errors.Is(err, store.ErrKeyNotFound) {
			return types.Response{
				Status:  types.StatusError,
				Message: store.ErrKeyNotFound.Error(),
			}
		}
		if errors.Is(err, store.ErrIndexOutOfRange) {
			return types.Response{
				Status:  types.StatusError,
				Message: store.ErrIndexOutOfRange.Error(),
			}
		}
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

	if err := h.store.Save(store.DefaultSnapshotFile); err != nil {
		return types.Response{
			Status:  types.StatusError,
			Message: err.Error(),
		}
	}

	return types.Response{
		Status: types.StatusOK,
	}
}
