package commands

import (
	"errors"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

// LPopHandler implements the LPOP command, which removes and returns
// the left-most element of a list. It coordinates the mutating storage
// operation with durable persistence, while correctly translating missing
// keys or type mismatches into protocol-independent Response values.
type LPopHandler struct {
	store *store.MemoryStore
}

// NewLPopHandler returns a new LPopHandler with the provided MemoryStore
// injected as its storage dependency.
func NewLPopHandler(store *store.MemoryStore) *LPopHandler {
	return &LPopHandler{store: store}
}

// Execute handles the LPOP command.
//
// It expects exactly one argument: the list key. If the argument count
// is incorrect, it returns a StatusError. Upon a successful pop, it triggers
// background persistence and returns a StatusOK response with the popped
// element. If the key is missing or expired, it returns StatusNil.
// Any type mismatches or persistence failures result in a StatusError.
func (h *LPopHandler) Execute(cmd types.Command) types.Response {
	if len(cmd.Args) != 1 {
		return types.Response{
			Status:  types.StatusError,
			Message: "wrong number of arguments",
		}
	}

	key := cmd.Args[0]

	val, err := h.store.LPop(key)
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

	if err := h.store.Save(store.DefaultSnapshotFile); err != nil {
		return types.Response{
			Status:  types.StatusError,
			Message: err.Error(),
		}
	}

	return types.Response{
		Status: types.StatusString,
		Data:   val,
	}
}
