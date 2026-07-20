package commands

import (
	"errors"
	"strconv"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

// LRemHandler implements the LREM command.
//
// LREM removes elements matching a specified value from a list.
// It implements Redis count semantics:
//   - count > 0 removes up to count matches from head to tail
//   - count < 0 removes up to abs(count) matches from tail to head
//   - count == 0 removes all matching elements
//
// LREM is a mutating command. It returns an integer response containing
// the number of elements removed.
//
// A removed count of zero is a successful execution. This includes when
// the key does not exist, or when the key exists but contains no matching values.
// WRONGTYPE is the only specific error condition handled.
//
// Successful executions always trigger persistence, even if zero elements
// were removed, as it is classified as a mutating command.
type LRemHandler struct {
	store *store.MemoryStore
}

// NewLRemHandler creates a new LRemHandler with the provided store.
func NewLRemHandler(s *store.MemoryStore) *LRemHandler {
	return &LRemHandler{store: s}
}

// Execute validates arguments, parses the count, executes the removal,
// triggers persistence, and returns the removed count as an integer response.
func (h *LRemHandler) Execute(cmd types.Command) types.Response {
	if len(cmd.Args) != 3 {
		return types.Response{
			Status:  types.StatusError,
			Message: "wrong number of arguments",
		}
	}

	key := cmd.Args[0]
	countStr := cmd.Args[1]
	value := cmd.Args[2]

	count, err := strconv.ParseInt(countStr, 10, 64)
	if err != nil {
		return types.Response{
			Status:  types.StatusError,
			Message: "value is not an integer or out of range",
		}
	}

	removedCount, err := h.store.LRem(key, count, value)
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
		Data:   removedCount,
	}
}
