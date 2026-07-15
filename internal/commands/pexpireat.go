package commands

import (
	"errors"
	"strconv"
	"time"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

// PExpireAtHandler implements the PEXPIREAT command, which sets the expiration
// time on a key to an absolute Unix timestamp in milliseconds.
type PExpireAtHandler struct {
	store *store.MemoryStore
}

// NewPExpireAtHandler returns a new PExpireAtHandler.
func NewPExpireAtHandler(store *store.MemoryStore) *PExpireAtHandler {
	return &PExpireAtHandler{store: store}
}

// Execute handles the PEXPIREAT command.
//
// It expects exactly two arguments: the key and the absolute Unix timestamp
// in milliseconds. If the argument count is incorrect, it returns a StatusError.
// If the timestamp is not a valid signed integer, it returns a StatusError.
//
// On success, it triggers automatic persistence via store.Save(store.DefaultSnapshotFile)
// and returns a StatusInteger response of "1". If the key is missing or already
// expired, it returns a StatusInteger response of "0" and does not trigger persistence.
func (h *PExpireAtHandler) Execute(cmd types.Command) types.Response {
	if len(cmd.Args) != 2 {
		return types.Response{
			Status:  types.StatusError,
			Message: "wrong number of arguments",
		}
	}

	key := cmd.Args[0]
	timestampStr := cmd.Args[1]

	timestampMs, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return types.Response{
			Status:  types.StatusError,
			Message: store.ErrInvalidTimestamp.Error(),
		}
	}

	t := time.UnixMilli(timestampMs)
	result, err := h.store.PExpireAt(key, t)

	if err != nil {
		if errors.Is(err, store.ErrKeyNotFound) || errors.Is(err, store.ErrKeyExpired) {
			return types.Response{
				Status: types.StatusInteger,
				Data:   "0",
			}
		}
		// ErrInvalidTimestamp or other internal store errors
		return types.Response{
			Status:  types.StatusError,
			Message: err.Error(),
		}
	}

	// Trigger persistence only on successful state modification
	if result == 1 {
		if err := h.store.Save(store.DefaultSnapshotFile); err != nil {
			return types.Response{
				Status:  types.StatusError,
				Message: err.Error(),
			}
		}
	}

	return types.Response{
		Status: types.StatusInteger,
		Data:   strconv.FormatInt(result, 10),
	}
}
