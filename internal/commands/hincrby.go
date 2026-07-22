package commands

import (
	"errors"
	"strconv"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

// HIncrByHandler implements the HINCRBY command.
type HIncrByHandler struct {
	store *store.MemoryStore
}

// NewHIncrByHandler returns a new HIncrByHandler.
func NewHIncrByHandler(store *store.MemoryStore) *HIncrByHandler {
	return &HIncrByHandler{store: store}
}

// Execute handles the HINCRBY command.
func (h *HIncrByHandler) Execute(cmd types.Command) types.Response {
	if len(cmd.Args) != 3 {
		return types.Response{
			Status:  types.StatusError,
			Message: "wrong number of arguments for 'hincrby' command",
		}
	}

	key := cmd.Args[0]
	field := cmd.Args[1]
	deltaStr := cmd.Args[2]

	delta, err := strconv.ParseInt(deltaStr, 10, 64)
	if err != nil {
		return types.Response{
			Status:  types.StatusError,
			Message: "ERR value is not an integer or out of range",
		}
	}

	newVal, err := h.store.HIncrBy(key, field, delta)
	if err != nil {
		if errors.Is(err, store.ErrWrongType) {
			return types.Response{
				Status:  types.StatusError,
				Message: store.ErrWrongType.Error(),
			}
		}
		// Passes ErrNotInteger and ErrOverflow directly
		return types.Response{
			Status:  types.StatusError,
			Message: err.Error(),
		}
	}

	// Mutated, so we persist
	h.store.Save(store.DefaultSnapshotFile)
	return types.Response{
		Status: types.StatusInteger,
		Data:   int(newVal), // The protocol writer uses fmt.Sprintf("%d", data) so casting to int or int64 is fine if the protocol handles it. Let's pass int64 to be safe. Wait, previous implementations often returned int, but int in Go is 64-bit on 64-bit systems. We'll pass int64.
	}
}
