package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/rajas2007/IgnisKV/internal/commands"
	"github.com/rajas2007/IgnisKV/internal/types"
)

// RunREPL starts the interactive command-line interface and blocks until the
// session ends.
//
// It coordinates the three stages of the CLI pipeline — parsing, dispatching,
// and printing — without owning any of their logic. ParseCommand converts raw
// input into a types.Command. The Dispatcher routes that Command to the
// appropriate handler and returns a types.Response. PrintResponse renders the
// response to the terminal. RunREPL owns only the loop that connects them.
//
// The Dispatcher is injected by the caller rather than constructed here. This
// keeps RunREPL free of wiring decisions: it does not know which commands are
// registered, which store implementation is in use, or how handlers are built.
// Those decisions belong in main.
//
// Lifecycle is controlled entirely by ResponseStatus. The loop exits when a
// handler returns types.StatusExit, without inspecting the command name that
// produced it. This means any future command that should terminate the session
// simply returns StatusExit from its Execute method — RunREPL requires no
// changes.
func RunREPL(dispatcher *commands.Dispatcher) {
	fmt.Println("IgnisKV v0.1")
	fmt.Println("Type HELP for available commands.")

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("> ")

		if !scanner.Scan() {
			// EOF or scanner error — exit cleanly.
			break
		}

		line := scanner.Text()

		cmd, ok := ParseCommand(line)
		if !ok {
			continue
		}

		response := dispatcher.Dispatch(cmd)
		PrintResponse(response)

		if response.Status == types.StatusExit {
			break
		}
	}
}
