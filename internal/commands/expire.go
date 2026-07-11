package commands

import (
	"errors"
	"strconv"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

// ExpireHandler implements the EXPIRE command, which sets or updates the
// expiration time on a key.
type ExpireHandler struct {
	store *store.MemoryStore
}

// NewExpireHandler returns a new ExpireHandler.
func NewExpireHandler(store *store.MemoryStore) *ExpireHandler {
	return &ExpireHandler{store: store}
}

// Execute handles the EXPIRE command.
//
// It expects exactly two arguments: the key and the duration in seconds.
// If the argument count is incorrect, it returns a StatusError response.
// If the duration is not a valid signed integer or is non-positive, it returns
// a StatusError response.
// On success, it triggers automatic persistence via store.Save(store.DefaultSnapshotFile)
// and returns a StatusInteger response of "1" (where Data holds the integer string).
// If the key is missing or already expired, it returns a StatusInteger response of "0".
func (h *ExpireHandler) Execute(cmd types.Command) types.Response {
	if len(cmd.Args) != 2 {
		return types.Response{
			Status:  types.StatusError,
			Message: "wrong number of arguments",
		}
	}

	key := cmd.Args[0]
	durationStr := cmd.Args[1]

	seconds, err := strconv.ParseInt(durationStr, 10, 64)
	if err != nil {
		return types.Response{
			Status:  types.StatusError,
			Message: store.ErrInvalidDuration.Error(),
		}
	}

	_, err = h.store.Expire(key, seconds)
	if err != nil {
		if errors.Is(err, store.ErrKeyNotFound) || errors.Is(err, store.ErrKeyExpired) {
			return types.Response{
				Status: types.StatusInteger, // Need to define StatusInteger if not present
				Data:   "0",
			}
		}
		// ErrInvalidDuration or other internal store errors
		return types.Response{
			Status:  types.StatusError,
			Message: err.Error(),
		}
	}

	// Trigger persistence after successful expiration update
	if err := h.store.Save(store.DefaultSnapshotFile); err != nil {
		return types.Response{
			Status:  types.StatusError,
			Message: err.Error(),
		}
	}

	return types.Response{
		Status: types.StatusInteger,
		Data:   "1",
	}
}
