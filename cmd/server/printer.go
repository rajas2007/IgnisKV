package main

import (
	"fmt"

	"github.com/rajas2007/IgnisKV/internal/types"
)

// PrintResponse writes a human-readable representation of resp to stdout.
//
// It is the final stage of the CLI request pipeline. Its only responsibility
// is presentation: translating a types.Response value into text that the user
// can read. It makes no decisions about control flow, does not know which
// command produced the response, and does not access the store or dispatcher.
//
// Separating presentation from execution means command handlers are testable
// without a terminal, and the output format can change without touching any
// handler or the REPL loop.
//
// StatusOK and StatusExit are treated identically for presentation purposes.
// If Message is non-empty it is printed. If Message is empty and Data is
// non-nil, Data is printed. Lifecycle decisions — such as whether StatusExit
// should terminate the session — belong exclusively to the REPL. The printer
// never calls os.Exit or returns a signal to the caller; it simply renders
// what it receives.
//
// StatusNil prints "(nil)" to clearly communicate the absence of a value.
// StatusError prints the Message field, which contains the error description.
func PrintResponse(resp types.Response) {
	switch resp.Status {
	case types.StatusOK, types.StatusExit:
		if resp.Message != "" {
			fmt.Println(resp.Message)
		} else if resp.Data != nil {
			fmt.Println(resp.Data)
		}
	case types.StatusNil:
		fmt.Println("(nil)")
	case types.StatusError:
		fmt.Println(resp.Message)
	}
}
