package commands

import (
	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

// SaveHandler implements the SAVE command, which triggers a manual persistence snapshot.
type SaveHandler struct {
	store *store.MemoryStore
}

// NewSaveHandler returns a new SaveHandler.
func NewSaveHandler(store *store.MemoryStore) *SaveHandler {
	return &SaveHandler{store: store}
}

// Execute handles the SAVE command.
func (h *SaveHandler) Execute(cmd types.Command) types.Response {
	if len(cmd.Args) != 0 {
		return types.Response{
			Status:  types.StatusError,
			Message: "wrong number of arguments for 'save' command",
		}
	}

	err := h.store.Save(store.DefaultSnapshotFile)
	if err != nil {
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
