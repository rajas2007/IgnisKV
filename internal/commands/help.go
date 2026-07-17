package commands

import "github.com/rajas2007/IgnisKV/internal/types"

// helpMessage is the full list of supported commands displayed by HELP.
const helpMessage = `Available commands:

PING                    Check server health
SET <key> <value>       Store a value
SET <key> <value> EX <s> Store a value with expiration
LPUSH <key> <value> [value ...] Prepend one or more values to a list
RPUSH <key> <value> [value ...] Append one or more values to a list
LLEN <key>              Get the length of a list
LRANGE <key> <start> <stop> Get a range of elements from a list
GET <key>               Retrieve a value
DEL <key>               Delete a key
TTL <key>               Get remaining lifetime in seconds
PTTL <key>              Get remaining lifetime in milliseconds
EXPIRETIME <key>        Get absolute expiration timestamp in seconds
PEXPIRETIME <key>       Get absolute expiration timestamp in milliseconds
EXPIRE <key> <seconds>  Set or update expiration
EXPIREAT <key> <ts>     Set absolute expiration timestamp
PEXPIRE <key> <milliseconds> Set expiration in ms
PEXPIREAT <key> <unix-ms-timestamp> Set absolute expiration in ms
PERSIST <key>           Remove expiration
SAVE                    Manual snapshot to disk
HELP                    Show available commands
QUIT                    Exit IgnisKV`

// HelpHandler implements the HELP command, which prints a summary of all
// supported commands. It does not interact with the storage engine.
type HelpHandler struct{}

// NewHelpHandler returns a new HelpHandler ready for registration with the
// Dispatcher.
func NewHelpHandler() *HelpHandler {
	return &HelpHandler{}
}

// Execute handles the HELP command.
//
// It accepts no arguments. If any arguments are supplied it returns a
// StatusError response. When called with no arguments it returns a StatusOK
// response whose Message field contains the full list of supported commands,
// formatted for terminal display.
func (h *HelpHandler) Execute(cmd types.Command) types.Response {
	if len(cmd.Args) != 0 {
		return types.Response{
			Status:  types.StatusError,
			Message: "wrong number of arguments",
		}
	}

	return types.Response{
		Status:  types.StatusOK,
		Message: helpMessage,
	}
}
