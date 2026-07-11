package commands

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

// SetHandler implements the SET command, which stores a string value under a
// given key in the database. It holds a reference to the MemoryStore through
// which all write operations are performed.
type SetHandler struct {
	// store is the storage engine this handler writes values to.
	store *store.MemoryStore
}

// NewSetHandler returns a new SetHandler with the provided MemoryStore injected
// as its storage dependency.
func NewSetHandler(store *store.MemoryStore) *SetHandler {
	return &SetHandler{store: store}
}

// Execute handles the SET command.
//
// It accepts two forms:
//
//	SET key value
//	SET key value EX seconds
//
// If the argument count is incorrect it returns a StatusError response.
// When a valid EX option is provided, the key is stored with an absolute
// expiration timestamp. Invalid EX values (non-numeric, zero, or negative)
// are rejected with a StatusError before the Store is modified.
func (h *SetHandler) Execute(cmd types.Command) types.Response {
	if len(cmd.Args) != 2 && len(cmd.Args) != 4 {
		return types.Response{
			Status:  types.StatusError,
			Message: "wrong number of arguments",
		}
	}

	value := types.Value{
		Type: types.StringType,
		Data: cmd.Args[1],
	}

	if len(cmd.Args) == 4 {
		if !strings.EqualFold(cmd.Args[2], "EX") {
			return types.Response{
				Status:  types.StatusError,
				Message: fmt.Sprintf("unsupported option: %s", cmd.Args[2]),
			}
		}

		secs, err := strconv.ParseInt(cmd.Args[3], 10, 64)
		if err != nil || secs <= 0 {
			return types.Response{
				Status:  types.StatusError,
				Message: "invalid expire time in SET",
			}
		}

		value.ExpiresAt = time.Now().Add(time.Duration(secs) * time.Second)
	}

	h.store.Set(cmd.Args[0], value)

	if err := h.store.Save(store.DefaultSnapshotFile); err != nil {
		return types.Response{
			Status:  types.StatusError,
			Message: err.Error(),
		}
	}

	return types.Response{
		Status:  types.StatusOK,
		Message: "OK",
	}
}
