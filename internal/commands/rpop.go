package commands

import (
	"errors"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

// RPopHandler implements the RPOP command, which removes and returns
// the right-most element of a list. It coordinates the mutating storage
// operation with durable persistence, while correctly translating missing
// keys or type mismatches into protocol-independent Response values.
// RPOP mirrors LPOP, differing only in removal direction.
type RPopHandler struct {
	store *store.MemoryStore
}

// NewRPopHandler returns a new RPopHandler with the provided MemoryStore
// injected as its storage dependency.
func NewRPopHandler(store *store.MemoryStore) *RPopHandler {
	return &RPopHandler{store: store}
}

// Execute handles the RPOP command.
//
// It expects exactly one argument: the list key. If the argument count
// is incorrect, it returns a StatusError. Upon a successful pop, it triggers
// background persistence and returns a StatusString response with the popped
// element. If the key is missing or expired, it returns StatusNil.
// Any type mismatches or persistence failures result in a StatusError.
// RPOP mirrors LPOP, differing only in removal direction.
func (h *RPopHandler) Execute(cmd types.Command) types.Response {
	if len(cmd.Args) != 1 {
		return types.Response{
			Status:  types.StatusError,
			Message: "wrong number of arguments",
		}
	}

	key := cmd.Args[0]

	val, err := h.store.RPop(key)
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
