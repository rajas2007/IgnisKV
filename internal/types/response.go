package types

// ResponseStatus represents the outcome category of a command execution.
// Handlers use this to communicate success, failure, or absence of a value
// to the output layer without exposing protocol-specific codes.
type ResponseStatus int

const (
	// StatusOK indicates that the command executed successfully and produced
	// a meaningful result. Data may or may not be populated.
	StatusOK ResponseStatus = iota

	// StatusError indicates that the command failed due to an invalid operation,
	// incorrect arguments, or an internal engine error.
	StatusError

	// StatusNil indicates a successful operation that produced no value.
	// This is the expected outcome for commands like GET on a key that does not exist,
	// distinct from an error condition.
	StatusNil
)

// Response represents the result of a command handler's execution.
// It is returned by every command handler and consumed by the output layer,
// which is responsible for formatting it into the appropriate client representation.
type Response struct {
	// Status communicates the outcome category of the command execution.
	// The output layer uses this to determine how to format and present the result.
	Status ResponseStatus

	// Data holds any value returned by a successful command (e.g. the string
	// retrieved by a GET). It is nil when the command produces no return value.
	Data any

	// Message contains a human-readable description of the outcome.
	// For errors this describes what went wrong; for successful commands
	// it may carry a confirmation string such as "OK".
	Message string
}
