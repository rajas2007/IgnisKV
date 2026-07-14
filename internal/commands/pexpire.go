package commands

import (
	"errors"
	"strconv"
	"time"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

// PExpireHandler implements the PEXPIRE command, which sets or updates the
// expiration time on a key using milliseconds.
type PExpireHandler struct {
	store *store.MemoryStore
}

// NewPExpireHandler returns a new PExpireHandler.
func NewPExpireHandler(store *store.MemoryStore) *PExpireHandler {
	return &PExpireHandler{store: store}
}

// Execute handles the PEXPIRE command.
//
// It expects exactly two arguments: the key and the duration in milliseconds.
// If the argument count is incorrect, it returns a StatusError response.
// If the duration is not a valid signed integer or is non-positive, it returns
// a StatusError response.
//
// On success, it triggers automatic persistence via store.Save(store.DefaultSnapshotFile)
// and returns a StatusInteger response of "1". If the key is missing or already
// expired, it returns a StatusInteger response of "0" and does not trigger persistence.
func (h *PExpireHandler) Execute(cmd types.Command) types.Response {
	if len(cmd.Args) != 2 {
		return types.Response{
			Status:  types.StatusError,
			Message: "wrong number of arguments",
		}
	}

	key := cmd.Args[0]
	durationStr := cmd.Args[1]

	ms, err := strconv.ParseInt(durationStr, 10, 64)
	if err != nil {
		return types.Response{
			Status:  types.StatusError,
			Message: store.ErrInvalidDuration.Error(),
		}
	}

	d := time.Duration(ms) * time.Millisecond
	result, err := h.store.PExpire(key, d)

	if err != nil {
		if errors.Is(err, store.ErrKeyNotFound) || errors.Is(err, store.ErrKeyExpired) {
			return types.Response{
				Status: types.StatusInteger,
				Data:   "0",
			}
		}
		// ErrInvalidDuration or other internal store errors
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
