package commands

import "github.com/rajas2007/IgnisKV/internal/types"

// PingHandler implements the PING command. It does not interact with the
// storage engine and exists solely to verify that the command execution
// pipeline is operational.
type PingHandler struct{}

// NewPingHandler returns a new PingHandler ready for registration with the Dispatcher.
func NewPingHandler() *PingHandler {
	return &PingHandler{}
}

// Execute handles the PING command.
//
// With no arguments it returns a StatusOK response with the message "PONG".
// With one argument it returns a StatusOK response whose Data field contains
// that argument, echoing it back to the caller.
// With more than one argument it returns a StatusError response.
func (h *PingHandler) Execute(cmd types.Command) types.Response {
	switch len(cmd.Args) {
	case 0:
		return types.Response{
			Status:  types.StatusOK,
			Message: "PONG",
		}
	case 1:
		return types.Response{
			Status: types.StatusOK,
			Data:   cmd.Args[0],
		}
	default:
		return types.Response{
			Status:  types.StatusError,
			Message: "wrong number of arguments",
		}
	}
}
