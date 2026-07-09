package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/rajas2007/IgnisKV/internal/commands"
	"github.com/rajas2007/IgnisKV/internal/server"
	"github.com/rajas2007/IgnisKV/internal/store"
)

// main is the application entry point. It constructs the storage engine and
// the command dispatcher, wires them together, and transfers control to either
// the REPL or the TCP Server based on runtime flags.
func main() {
	serverMode := flag.Bool("server", false, "Start the TCP server instead of the local REPL")
	port := flag.Int("port", 6379, "Port to listen on when in server mode")
	flag.Parse()

	s := store.NewMemoryStore()
	if err := s.Load(store.DefaultSnapshotFile); err != nil {
		log.Printf("Warning: failed to load snapshot: %v", err)
	}

	dispatcher := commands.NewDispatcher(s)

	if *serverMode {
		addr := fmt.Sprintf(":%d", *port)
		log.Printf("Starting IgnisKV TCP server on %s", addr)
		srv := server.New(dispatcher)
		if err := srv.Start(addr); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	} else {
		RunREPL(dispatcher)
	}
}
