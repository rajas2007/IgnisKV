package commands

import (
	"errors"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

// GetHandler implements the GET command, which retrieves the value stored under
// a given key. It translates storage-layer outcomes into protocol-independent
// Response values, distinguishing between a present value, an absent key, and
// an unexpected error.
type GetHandler struct {
	// store is the storage engine this handler reads values from.
	store *store.MemoryStore
}

// NewGetHandler returns a new GetHandler with the provided MemoryStore injected
// as its storage dependency.
func NewGetHandler(store *store.MemoryStore) *GetHandler {
	return &GetHandler{store: store}
}

// Execute handles the GET command.
//
// It expects exactly one argument: the key to retrieve. If the argument count
// is incorrect it returns a StatusError response. When the key exists and has
// not expired, it returns a StatusOK response with the stored value in the
// Data field. When the key does not exist, or exists but has expired (in which
// case it is lazily deleted), it returns a StatusNil response to distinguish
// the absence of a value from an error condition. Any other storage error is
// returned as a StatusError response.
func (h *GetHandler) Execute(cmd types.Command) types.Response {
	if len(cmd.Args) != 1 {
		return types.Response{
			Status:  types.StatusError,
			Message: "wrong number of arguments",
		}
	}

	value, err := h.store.Get(cmd.Args[0])
	if err != nil {
		if errors.Is(err, store.ErrKeyNotFound) || errors.Is(err, store.ErrKeyExpired) {
			return types.Response{
				Status: types.StatusNil,
			}
		}
		return types.Response{
			Status:  types.StatusError,
			Message: err.Error(),
		}
	}

	return types.Response{
		Status: types.StatusOK,
		Data:   value.Data,
	}
}
