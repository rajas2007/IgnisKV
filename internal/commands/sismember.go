package commands

import (
	"errors"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

// SIsMemberHandler implements the SISMEMBER command.
type SIsMemberHandler struct {
	store *store.MemoryStore
}

// NewSIsMemberHandler returns a new SIsMemberHandler.
func NewSIsMemberHandler(store *store.MemoryStore) *SIsMemberHandler {
	return &SIsMemberHandler{store: store}
}

// Execute handles the SISMEMBER command.
func (h *SIsMemberHandler) Execute(cmd types.Command) types.Response {
	if len(cmd.Args) != 2 {
		return types.Response{
			Status:  types.StatusError,
			Message: "wrong number of arguments for 'sismember' command",
		}
	}

	key := cmd.Args[0]
	member := cmd.Args[1]

	isMember, err := h.store.SIsMember(key, member)
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

	val := 0
	if isMember {
		val = 1
	}

	return types.Response{
		Status: types.StatusInteger,
		Data:   val,
	}
}
