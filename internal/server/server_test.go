package server_test

import (
	"fmt"
	"net"
	"strings"
	"sync"
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

func TestServerConcurrentClients(t *testing.T) {
	// Arrange
	s := store.NewMemoryStore()
	dispatcher := commands.NewDispatcher(s)
	srv := server.New(dispatcher)

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to find free port: %v", err)
	}
	address := l.Addr().String()
	l.Close()

	go func() {
		_ = srv.Start(address)
	}()

	time.Sleep(100 * time.Millisecond)

	// Act
	numClients := 10
	errCh := make(chan error, numClients)
	var wg sync.WaitGroup
	wg.Add(numClients)

	for i := 0; i < numClients; i++ {
		go func(clientID int) {
			defer wg.Done()
			conn, err := net.Dial("tcp", address)
			if err != nil {
				errCh <- fmt.Errorf("client %d dial error: %v", clientID, err)
				return
			}
			defer conn.Close()

			buf := make([]byte, 1024)

			// Helper to send and verify
			sendAndVerify := func(cmd []byte, expected string) bool {
				if _, err := conn.Write(cmd); err != nil {
					errCh <- fmt.Errorf("client %d write error: %v", clientID, err)
					return false
				}
				n, err := conn.Read(buf)
				if err != nil {
					errCh <- fmt.Errorf("client %d read error: %v", clientID, err)
					return false
				}
				if string(buf[:n]) != expected {
					errCh <- fmt.Errorf("client %d expected %q, got %q", clientID, expected, string(buf[:n]))
					return false
				}
				return true
			}

			// PING
			if !sendAndVerify([]byte("*1\r\n$4\r\nPING\r\n"), "+PONG\r\n") {
				return
			}

			// SET 1
			key := fmt.Sprintf("key%d", clientID)
			val1 := fmt.Sprintf("val%d_1", clientID)
			setCmd1 := []byte(fmt.Sprintf("*3\r\n$3\r\nSET\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n", len(key), key, len(val1), val1))
			if !sendAndVerify(setCmd1, "+OK\r\n") {
				return
			}

			// GET 1
			getCmd := []byte(fmt.Sprintf("*2\r\n$3\r\nGET\r\n$%d\r\n%s\r\n", len(key), key))
			expectedGet1 := fmt.Sprintf("$%d\r\n%s\r\n", len(val1), val1)
			if !sendAndVerify(getCmd, expectedGet1) {
				return
			}

			// SET 2
			val2 := fmt.Sprintf("val%d_2", clientID)
			setCmd2 := []byte(fmt.Sprintf("*3\r\n$3\r\nSET\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n", len(key), key, len(val2), val2))
			if !sendAndVerify(setCmd2, "+OK\r\n") {
				return
			}

			// GET 2
			expectedGet2 := fmt.Sprintf("$%d\r\n%s\r\n", len(val2), val2)
			if !sendAndVerify(getCmd, expectedGet2) {
				return
			}

			// DEL
			delCmd := []byte(fmt.Sprintf("*2\r\n$3\r\nDEL\r\n$%d\r\n%s\r\n", len(key), key))
			if !sendAndVerify(delCmd, "+OK\r\n") {
				return
			}

			// GET 3 (Deleted)
			if !sendAndVerify(getCmd, "$-1\r\n") {
				return
			}

			// QUIT
			if !sendAndVerify([]byte("*1\r\n$4\r\nQUIT\r\n"), "+BYE\r\n") {
				return
			}

		}(i)
	}

	wg.Wait()
	close(errCh)

	// Assert
	for err := range errCh {
		t.Errorf("Concurrent client error: %v", err)
	}
}

func TestServerConcurrentSharedKey(t *testing.T) {
	// Arrange
	s := store.NewMemoryStore()
	dispatcher := commands.NewDispatcher(s)
	srv := server.New(dispatcher)

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to find free port: %v", err)
	}
	address := l.Addr().String()
	l.Close()

	go func() {
		_ = srv.Start(address)
	}()

	time.Sleep(100 * time.Millisecond)

	// Act
	numClients := 20
	errCh := make(chan error, numClients)
	var wg sync.WaitGroup
	wg.Add(numClients)

	sharedKey := "shared_counter"

	for i := 0; i < numClients; i++ {
		go func(clientID int) {
			defer wg.Done()
			conn, err := net.Dial("tcp", address)
			if err != nil {
				errCh <- fmt.Errorf("client %d dial error: %v", clientID, err)
				return
			}
			defer conn.Close()

			buf := make([]byte, 1024)

			// Helper to send and verify loosely
			sendAndVerify := func(cmd []byte, expected string, prefixMatch bool) bool {
				if _, err := conn.Write(cmd); err != nil {
					errCh <- fmt.Errorf("client %d write error: %v", clientID, err)
					return false
				}
				n, err := conn.Read(buf)
				if err != nil {
					errCh <- fmt.Errorf("client %d read error: %v", clientID, err)
					return false
				}
				response := string(buf[:n])
				if prefixMatch {
					if !strings.HasPrefix(response, expected) || !strings.HasSuffix(response, "\r\n") {
						errCh <- fmt.Errorf("client %d expected prefix %q and suffix '\\r\\n', got %q", clientID, expected, response)
						return false
					}
				} else {
					if response != expected {
						errCh <- fmt.Errorf("client %d expected %q, got %q", clientID, expected, response)
						return false
					}
				}
				return true
			}

			val := fmt.Sprintf("val%d", clientID)

			// SET shared
			setCmd := []byte(fmt.Sprintf("*3\r\n$3\r\nSET\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n", len(sharedKey), sharedKey, len(val), val))
			if !sendAndVerify(setCmd, "+OK\r\n", false) {
				return
			}

			// GET shared (Value could be anything written by any goroutine, so just ensure it's a bulk string response starting with $)
			getCmd := []byte(fmt.Sprintf("*2\r\n$3\r\nGET\r\n$%d\r\n%s\r\n", len(sharedKey), sharedKey))
			if !sendAndVerify(getCmd, "$", true) {
				return
			}

			// QUIT
			if !sendAndVerify([]byte("*1\r\n$4\r\nQUIT\r\n"), "+BYE\r\n", false) {
				return
			}

		}(i)
	}

	wg.Wait()
	close(errCh)

	// Assert
	for err := range errCh {
		t.Errorf("Concurrent shared key client error: %v", err)
	}
}
