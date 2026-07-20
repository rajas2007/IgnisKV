package commands

import (
	"errors"
	"strings"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

// LInsertHandler implements the LINSERT command.
// It is a mutating command that inserts an element before or after a pivot.
// BEFORE and AFTER semantics are supported.
// Missing key returns the integer value 0.
// Pivot not found returns the integer value -1.
// Successful insertion returns the new list length.
// Position parsing is case-insensitive.
// Persistence responsibility belongs exclusively to the Handler.
// WRONGTYPE is explicitly handled and returned if the key is not a list.
type LInsertHandler struct {
	store *store.MemoryStore
}

// NewLInsertHandler allocates and returns a new LInsertHandler.
func NewLInsertHandler(s *store.MemoryStore) *LInsertHandler {
	return &LInsertHandler{store: s}
}

// Execute validates the command arguments, parses the position keyword, and delegates insertion to the Store.
// It translates store errors into StatusError responses.
// After every successful Store execution, it triggers persistence by saving a snapshot.
func (h *LInsertHandler) Execute(cmd types.Command) types.Response {
	if len(cmd.Args) != 4 {
		return types.Response{
			Status:  types.StatusError,
			Message: "wrong number of arguments",
		}
	}

	positionKeyword := strings.ToUpper(cmd.Args[0])
	key := cmd.Args[1]
	pivot := cmd.Args[2]
	value := cmd.Args[3]

	var before bool
	switch positionKeyword {
	case "BEFORE":
		before = true
	case "AFTER":
		before = false
	default:
		return types.Response{
			Status:  types.StatusError,
			Message: "syntax error",
		}
	}

	length, err := h.store.LInsert(key, before, pivot, value)
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

	if err := h.store.Save(store.DefaultSnapshotFile); err != nil {
		return types.Response{
			Status:  types.StatusError,
			Message: err.Error(),
		}
	}

	return types.Response{
		Status: types.StatusInteger,
		Data:   length,
	}
}
