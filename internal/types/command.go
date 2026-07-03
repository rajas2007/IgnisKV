package types

// Command represents a parsed database request containing the target operation
// and its associated arguments. It is passed from the parser through to the dispatcher
// for handler routing.
type Command struct {
	// Name specifies the database operation to perform (e.g., "SET", "GET", "DEL").
	// It is used by the dispatcher to route execution to the correct handler.
	Name string

	// Args contains the slice of arguments passed along with the command.
	// The individual command handlers are responsible for parsing and validating these.
	Args []string
}
