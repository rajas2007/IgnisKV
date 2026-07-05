package main

import (
	"github.com/rajas2007/IgnisKV/internal/commands"
	"github.com/rajas2007/IgnisKV/internal/store"
)

// main is the application entry point. It constructs the storage engine and
// the command dispatcher, wires them together, and transfers control to the
// REPL. It contains no business logic. All database, execution, and
// presentation concerns belong in the packages those components own.
func main() {
	s := store.NewMemoryStore()
	dispatcher := commands.NewDispatcher(s)
	RunREPL(dispatcher)
}
