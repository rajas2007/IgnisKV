package commands

import (
	"errors"
	"strconv"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

// HIncrByFloatHandler implements the HINCRBYFLOAT command.
type HIncrByFloatHandler struct {
	store *store.MemoryStore
}

// NewHIncrByFloatHandler returns a new HIncrByFloatHandler.
func NewHIncrByFloatHandler(store *store.MemoryStore) *HIncrByFloatHandler {
	return &HIncrByFloatHandler{store: store}
}

// Execute handles the HINCRBYFLOAT command.
func (h *HIncrByFloatHandler) Execute(cmd types.Command) types.Response {
	if len(cmd.Args) != 3 {
		return types.Response{
			Status:  types.StatusError,
			Message: "wrong number of arguments for 'hincrbyfloat' command",
		}
	}

	key := cmd.Args[0]
	field := cmd.Args[1]
	deltaStr := cmd.Args[2]

	delta, err := strconv.ParseFloat(deltaStr, 64)
	if err != nil {
		return types.Response{
			Status:  types.StatusError,
			Message: "ERR value is not a valid float",
		}
	}

	newVal, err := h.store.HIncrByFloat(key, field, delta)
	if err != nil {
		if errors.Is(err, store.ErrWrongType) {
			return types.Response{
				Status:  types.StatusError,
				Message: store.ErrWrongType.Error(),
			}
		}
		// Passes ErrNotFloat and ErrNaNOrInfinity directly
		return types.Response{
			Status:  types.StatusError,
			Message: err.Error(),
		}
	}

	// Mutated, so we persist
	h.store.Save(store.DefaultSnapshotFile)

	// Format identically to the store
	strValue := strconv.FormatFloat(newVal, 'f', -1, 64)

	return types.Response{
		Status: types.StatusString, // Matching Redis requirement
		Data:   strValue,
	}
}
