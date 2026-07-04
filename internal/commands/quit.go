package commands

import "github.com/rajas2007/IgnisKV/internal/types"

// QuitHandler implements the QUIT command, which signals that the client wishes
// to end the current session. It does not manage connection lifecycle; it only
// returns a Response that the server layer will later interpret and act upon.
type QuitHandler struct{}

// NewQuitHandler returns a new QuitHandler ready for registration with the Dispatcher.
func NewQuitHandler() *QuitHandler {
	return &QuitHandler{}
}

// Execute handles the QUIT command.
//
// It accepts no arguments. If any arguments are supplied it returns a StatusError
// response. When called with no arguments it returns a StatusOK response with the
// message "BYE", signalling the client's intention to quit. It does not close
// sockets, exit the process, or access the storage engine.
func (h *QuitHandler) Execute(cmd types.Command) types.Response {
	if len(cmd.Args) != 0 {
		return types.Response{
			Status:  types.StatusError,
			Message: "wrong number of arguments",
		}
	}

	return types.Response{
		Status:  types.StatusOK,
		Message: "BYE",
	}
}
