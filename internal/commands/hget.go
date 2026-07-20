package commands

import (
	"errors"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

// HGetHandler implements the HGET command.
type HGetHandler struct {
	store *store.MemoryStore
}

// NewHGetHandler returns a new HGetHandler.
func NewHGetHandler(store *store.MemoryStore) *HGetHandler {
	return &HGetHandler{store: store}
}

// Execute handles the HGET command.
func (h *HGetHandler) Execute(cmd types.Command) types.Response {
	if len(cmd.Args) != 2 {
		return types.Response{
			Status:  types.StatusError,
			Message: "wrong number of arguments for 'hget' command",
		}
	}

	key := cmd.Args[0]
	field := cmd.Args[1]

	val, err := h.store.HGet(key, field)
	if err != nil {
		if errors.Is(err, store.ErrKeyNotFound) || errors.Is(err, store.ErrKeyExpired) || errors.Is(err, store.ErrFieldNotFound) {
			return types.Response{
				Status: types.StatusNil,
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

	return types.Response{
		Status: types.StatusString,
		Data:   val,
	}
}
