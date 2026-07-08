package server_test

import (
	"net"
	"testing"
	"time"

	"github.com/rajas2007/IgnisKV/internal/commands"
	"github.com/rajas2007/IgnisKV/internal/server"
	"github.com/rajas2007/IgnisKV/internal/store"
)

func TestServerEndToEnd(t *testing.T) {
	// Arrange
	s := store.NewMemoryStore()
	dispatcher := commands.NewDispatcher(s)
	srv := server.New(dispatcher)

	// Use an OS-assigned free port (127.0.0.1:0) to avoid conflicts
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to find free port: %v", err)
	}
	address := l.Addr().String()
	l.Close() // Release the port so the server can bind to it

	go func() {
		_ = srv.Start(address)
	}()

	// Temporary synchronization: Give the listener a moment to bind.
	// Future milestones will implement graceful startup synchronization.
	time.Sleep(100 * time.Millisecond)

	// Act
	conn, err := net.Dial("tcp", address)
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	// 1. Test PING
	pingCmd := []byte("*1\r\n$4\r\nPING\r\n")
	_, err = conn.Write(pingCmd)
	if err != nil {
		t.Fatalf("Failed to write PING: %v", err)
	}

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("Failed to read PING response: %v", err)
	}
	if string(buf[:n]) != "+PONG\r\n" {
		t.Fatalf("Expected +PONG\\r\\n, got %q", string(buf[:n]))
	}

	// 2. Test SET
	setCmd := []byte("*3\r\n$3\r\nSET\r\n$4\r\nname\r\n$5\r\nRajas\r\n")
	_, err = conn.Write(setCmd)
	if err != nil {
		t.Fatalf("Failed to write SET: %v", err)
	}

	n, err = conn.Read(buf)
	if err != nil {
		t.Fatalf("Failed to read SET response: %v", err)
	}
	if string(buf[:n]) != "+OK\r\n" {
		t.Fatalf("Expected +OK\\r\\n, got %q", string(buf[:n]))
	}

	// 3. Test GET
	getCmd := []byte("*2\r\n$3\r\nGET\r\n$4\r\nname\r\n")
	_, err = conn.Write(getCmd)
	if err != nil {
		t.Fatalf("Failed to write GET: %v", err)
	}

	n, err = conn.Read(buf)
	if err != nil {
		t.Fatalf("Failed to read GET response: %v", err)
	}
	if string(buf[:n]) != "$5\r\nRajas\r\n" {
		t.Fatalf("Expected $5\\r\\nRajas\\r\\n, got %q", string(buf[:n]))
	}

	// 4. Test QUIT
	quitCmd := []byte("*1\r\n$4\r\nQUIT\r\n")
	_, err = conn.Write(quitCmd)
	if err != nil {
		t.Fatalf("Failed to write QUIT: %v", err)
	}

	n, err = conn.Read(buf)
	if err != nil {
		t.Fatalf("Failed to read QUIT response: %v", err)
	}
	if string(buf[:n]) != "+BYE\r\n" {
		t.Fatalf("Expected +BYE\\r\\n, got %q", string(buf[:n]))
	}
}
