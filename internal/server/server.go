package server

import (
	"fmt"
	"net"

	"github.com/rajas2007/IgnisKV/internal/commands"
)

// Server owns the TCP networking and client connection management.
type Server struct {
	dispatcher *commands.Dispatcher
}

// New initializes the Server with the required command Dispatcher dependency.
func New(dispatcher *commands.Dispatcher) *Server {
	return &Server{
		dispatcher: dispatcher,
	}
}

// Start binds to the given TCP address and begins accepting connections.
// It blocks indefinitely unless a fatal listener error occurs.
func (s *Server) Start(address string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to bind to address %s: %w", address, err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			return fmt.Errorf("failed to accept connection: %w", err)
		}

		// Sprint 7: Concurrent client support.
		// Each client is handled in its own dedicated goroutine.
		go s.handleConnection(conn)
	}
}
