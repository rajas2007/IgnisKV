package server_test

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/rajas2007/IgnisKV/internal/commands"
	"github.com/rajas2007/IgnisKV/internal/server"
	"github.com/rajas2007/IgnisKV/internal/store"
)

func TestMain(m *testing.M) {
	tempDir, err := os.MkdirTemp("", "igniskv-server-test-*")
	if err != nil {
		os.Exit(1)
	}
	originalDir, _ := os.Getwd()
	os.Chdir(tempDir)

	code := m.Run()

	os.Chdir(originalDir)
	os.RemoveAll(tempDir)
	os.Exit(code)
}

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

func TestIntegrationExpire(t *testing.T) {
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

	conn, err := net.Dial("tcp", address)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	buf := make([]byte, 1024)
	sendAndVerify := func(cmd string, expected string) string {
		if _, err := conn.Write([]byte(cmd)); err != nil {
			t.Fatalf("Write error: %v", err)
		}
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatalf("Read error: %v", err)
		}
		response := string(buf[:n])
		if expected != "" && response != expected {
			t.Fatalf("Expected %q, got %q", expected, response)
		}
		return response
	}

	// SET key1 val1
	sendAndVerify("*3\r\n$3\r\nSET\r\n$4\r\nkey1\r\n$4\r\nval1\r\n", "+OK\r\n")

	// EXPIRE key1 5
	sendAndVerify("*3\r\n$6\r\nEXPIRE\r\n$4\r\nkey1\r\n$1\r\n5\r\n", ":1\r\n")

	// TTL key1
	response := sendAndVerify("*2\r\n$3\r\nTTL\r\n$4\r\nkey1\r\n", "")
	if !strings.HasPrefix(response, ":") || !strings.HasSuffix(response, "\r\n") {
		t.Fatalf("Expected integer response, got %q", response)
	}
	ttlStr := response[1 : len(response)-2]
	ttl, err := strconv.Atoi(ttlStr)
	if err != nil {
		t.Fatalf("Expected integer TTL, got %q", ttlStr)
	}
	if ttl < 4 || ttl > 5 {
		t.Fatalf("Expected TTL 4 or 5, got %d", ttl)
	}

	// EXPIRE missing_key 5
	sendAndVerify("*3\r\n$6\r\nEXPIRE\r\n$11\r\nmissing_key\r\n$1\r\n5\r\n", ":0\r\n")
}

func TestIntegrationPersist(t *testing.T) {
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

	conn, err := net.Dial("tcp", address)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	buf := make([]byte, 1024)
	sendAndVerify := func(cmd string, expected string) string {
		if _, err := conn.Write([]byte(cmd)); err != nil {
			t.Fatalf("Write error: %v", err)
		}
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatalf("Read error: %v", err)
		}
		response := string(buf[:n])
		if expected != "" && response != expected {
			t.Fatalf("Expected %q, got %q", expected, response)
		}
		return response
	}

	// 1. SET key value
	sendAndVerify("*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n", "+OK\r\n")

	// 2. EXPIRE key 5
	sendAndVerify("*3\r\n$6\r\nEXPIRE\r\n$3\r\nkey\r\n$1\r\n5\r\n", ":1\r\n")

	// 3. TTL key → 5 or 4
	response := sendAndVerify("*2\r\n$3\r\nTTL\r\n$3\r\nkey\r\n", "")
	ttlStr := response[1 : len(response)-2]
	ttl, err := strconv.Atoi(ttlStr)
	if err != nil {
		t.Fatalf("Expected integer TTL, got %q", ttlStr)
	}
	if ttl < 3 || ttl > 5 {
		t.Fatalf("Expected TTL 3-5, got %d", ttl)
	}

	// 4. PERSIST key → :1
	sendAndVerify("*2\r\n$7\r\nPERSIST\r\n$3\r\nkey\r\n", ":1\r\n")

	// 5. TTL key → :-1 (now persistent)
	sendAndVerify("*2\r\n$3\r\nTTL\r\n$3\r\nkey\r\n", ":-1\r\n")

	// 6. PERSIST key again → :0 (already persistent)
	sendAndVerify("*2\r\n$7\r\nPERSIST\r\n$3\r\nkey\r\n", ":0\r\n")

	// 7. GET key → value still exists
	response = sendAndVerify("*2\r\n$3\r\nGET\r\n$3\r\nkey\r\n", "")
	if response != "$5\r\nvalue\r\n" {
		t.Fatalf("Expected value after PERSIST, got %q", response)
	}

	// 8. PERSIST missing_key → :0
	sendAndVerify("*2\r\n$7\r\nPERSIST\r\n$11\r\nmissing_key\r\n", ":0\r\n")
}

func TestIntegrationExpireAt(t *testing.T) {
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

	conn, err := net.Dial("tcp", address)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	buf := make([]byte, 1024)
	sendAndVerify := func(cmd string, expected string) string {
		if _, err := conn.Write([]byte(cmd)); err != nil {
			t.Fatalf("Write error: %v", err)
		}
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatalf("Read error: %v", err)
		}
		response := string(buf[:n])
		if expected != "" && response != expected {
			t.Fatalf("Expected %q, got %q", expected, response)
		}
		return response
	}

	// 1. SET key value
	sendAndVerify("*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n", "+OK\r\n")

	// 2. EXPIREAT key (now + 5 seconds)
	futureTime := time.Now().Add(5 * time.Second).Unix()
	timestampStr := strconv.FormatInt(futureTime, 10)
	tsLenStr := strconv.Itoa(len(timestampStr))
	expireAtCmd := "*3\r\n$8\r\nEXPIREAT\r\n$3\r\nkey\r\n$" + tsLenStr + "\r\n" + timestampStr + "\r\n"
	sendAndVerify(expireAtCmd, ":1\r\n")

	// 3. TTL key → 5 or 4
	response := sendAndVerify("*2\r\n$3\r\nTTL\r\n$3\r\nkey\r\n", "")
	ttlStr := response[1 : len(response)-2]
	ttl, err := strconv.Atoi(ttlStr)
	if err != nil {
		t.Fatalf("Expected integer TTL, got %q", ttlStr)
	}
	if ttl < 3 || ttl > 5 {
		t.Fatalf("Expected TTL 3-5, got %d", ttl)
	}

	// 4. EXPIREAT missing_key (now + 5 seconds) → :0
	expireAtMissingCmd := "*3\r\n$8\r\nEXPIREAT\r\n$11\r\nmissing_key\r\n$" + tsLenStr + "\r\n" + timestampStr + "\r\n"
	sendAndVerify(expireAtMissingCmd, ":0\r\n")

	// 5. EXPIREAT key past_timestamp → -ERR invalid timestamp
	pastTime := time.Now().Add(-1 * time.Hour).Unix()
	pastTimestampStr := strconv.FormatInt(pastTime, 10)
	pastLenStr := strconv.Itoa(len(pastTimestampStr))
	expireAtPastCmd := "*3\r\n$8\r\nEXPIREAT\r\n$3\r\nkey\r\n$" + pastLenStr + "\r\n" + pastTimestampStr + "\r\n"
	sendAndVerify(expireAtPastCmd, "-ERR invalid timestamp\r\n")
}

func TestIntegrationPExpire(t *testing.T) {
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

	conn, err := net.Dial("tcp", address)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	buf := make([]byte, 1024)
	sendAndVerify := func(cmd string, expected string) string {
		if _, err := conn.Write([]byte(cmd)); err != nil {
			t.Fatalf("Write error: %v", err)
		}
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatalf("Read error: %v", err)
		}
		response := string(buf[:n])
		if expected != "" && response != expected {
			t.Fatalf("Expected %q, got %q", expected, response)
		}
		return response
	}

	// 1. SET key value
	sendAndVerify("*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n", "+OK\r\n")

	// 2. PEXPIRE key 1500
	sendAndVerify("*3\r\n$7\r\nPEXPIRE\r\n$3\r\nkey\r\n$4\r\n1500\r\n", ":1\r\n")

	// 3. TTL key → expect 1 or 2 seconds
	response := sendAndVerify("*2\r\n$3\r\nTTL\r\n$3\r\nkey\r\n", "")
	ttlStr := response[1 : len(response)-2]
	ttl, err := strconv.Atoi(ttlStr)
	if err != nil {
		t.Fatalf("Expected integer TTL, got %q", ttlStr)
	}
	if ttl < 0 || ttl > 2 {
		t.Fatalf("Expected TTL 0-2, got %d", ttl)
	}

	// 4. GET key → value still exists
	response = sendAndVerify("*2\r\n$3\r\nGET\r\n$3\r\nkey\r\n", "")
	if response != "$5\r\nvalue\r\n" {
		t.Fatalf("Expected value after PEXPIRE, got %q", response)
	}

	// 5. PEXPIRE missing_key 1500 → :0
	sendAndVerify("*3\r\n$7\r\nPEXPIRE\r\n$11\r\nmissing_key\r\n$4\r\n1500\r\n", ":0\r\n")

	// 6. PEXPIRE key 0 → -ERR invalid duration
	sendAndVerify("*3\r\n$7\r\nPEXPIRE\r\n$3\r\nkey\r\n$1\r\n0\r\n", "-ERR invalid duration\r\n")
}

func TestIntegrationPExpireAt(t *testing.T) {
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

	conn, err := net.Dial("tcp", address)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	buf := make([]byte, 1024)
	sendAndVerify := func(cmd string, expected string) string {
		if _, err := conn.Write([]byte(cmd)); err != nil {
			t.Fatalf("Write error: %v", err)
		}
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatalf("Read error: %v", err)
		}
		response := string(buf[:n])
		if expected != "" && response != expected {
			t.Fatalf("Expected %q, got %q", expected, response)
		}
		return response
	}

	// 1. SET key value
	sendAndVerify("*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n", "+OK\r\n")

	// 2. PEXPIREAT key (now + 5000 ms)
	futureTime := time.Now().Add(5000 * time.Millisecond).UnixMilli()
	timestampStr := strconv.FormatInt(futureTime, 10)
	tsLenStr := strconv.Itoa(len(timestampStr))
	expireAtCmd := "*3\r\n$9\r\nPEXPIREAT\r\n$3\r\nkey\r\n$" + tsLenStr + "\r\n" + timestampStr + "\r\n"
	sendAndVerify(expireAtCmd, ":1\r\n")

	// 3. TTL key → expect 3-5 seconds
	response := sendAndVerify("*2\r\n$3\r\nTTL\r\n$3\r\nkey\r\n", "")
	ttlStr := response[1 : len(response)-2]
	ttl, err := strconv.Atoi(ttlStr)
	if err != nil {
		t.Fatalf("Expected integer TTL, got %q", ttlStr)
	}
	if ttl < 2 || ttl > 5 {
		t.Fatalf("Expected TTL 2-5, got %d", ttl)
	}

	// 4. GET key → value still exists
	response = sendAndVerify("*2\r\n$3\r\nGET\r\n$3\r\nkey\r\n", "")
	if response != "$5\r\nvalue\r\n" {
		t.Fatalf("Expected value after PEXPIREAT, got %q", response)
	}

	// 5. PEXPIREAT missing_key (same future timestamp) → :0
	expireAtMissingCmd := "*3\r\n$9\r\nPEXPIREAT\r\n$11\r\nmissing_key\r\n$" + tsLenStr + "\r\n" + timestampStr + "\r\n"
	sendAndVerify(expireAtMissingCmd, ":0\r\n")

	// 6. PEXPIREAT key past_timestamp → -ERR invalid timestamp
	pastTime := time.Now().Add(-1000 * time.Millisecond).UnixMilli()
	pastTimestampStr := strconv.FormatInt(pastTime, 10)
	pastLenStr := strconv.Itoa(len(pastTimestampStr))
	expireAtPastCmd := "*3\r\n$9\r\nPEXPIREAT\r\n$3\r\nkey\r\n$" + pastLenStr + "\r\n" + pastTimestampStr + "\r\n"
	sendAndVerify(expireAtPastCmd, "-ERR invalid timestamp\r\n")
}

func TestIntegrationPTTL(t *testing.T) {
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

	conn, err := net.Dial("tcp", address)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	buf := make([]byte, 1024)
	sendAndVerify := func(cmd string, expected string) string {
		if _, err := conn.Write([]byte(cmd)); err != nil {
			t.Fatalf("Write error: %v", err)
		}
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatalf("Read error: %v", err)
		}
		response := string(buf[:n])
		if expected != "" && response != expected {
			t.Fatalf("Expected %q, got %q", expected, response)
		}
		return response
	}

	// 1. SET key value
	sendAndVerify("*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n", "+OK\r\n")

	// 2. PEXPIRE key 3000
	sendAndVerify("*3\r\n$7\r\nPEXPIRE\r\n$3\r\nkey\r\n$4\r\n3000\r\n", ":1\r\n")

	// 3. PTTL key → expect between 1800 and 3000 milliseconds
	response := sendAndVerify("*2\r\n$4\r\nPTTL\r\n$3\r\nkey\r\n", "")
	pttlStr := response[1 : len(response)-2]
	pttl, err := strconv.Atoi(pttlStr)
	if err != nil {
		t.Fatalf("Expected integer PTTL, got %q", pttlStr)
	}
	if pttl < 1800 || pttl > 3000 {
		t.Fatalf("Expected PTTL 1800-3000, got %d", pttl)
	}

	// 4. GET key → value still exists
	response = sendAndVerify("*2\r\n$3\r\nGET\r\n$3\r\nkey\r\n", "")
	if response != "$5\r\nvalue\r\n" {
		t.Fatalf("Expected value after PTTL, got %q", response)
	}

	// 5. PTTL missing_key → :-2
	sendAndVerify("*2\r\n$4\r\nPTTL\r\n$11\r\nmissing_key\r\n", ":-2\r\n")
}

func TestIntegrationExpireTime(t *testing.T) {
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

	conn, err := net.Dial("tcp", address)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	buf := make([]byte, 1024)
	sendAndVerify := func(cmd string, expected string) string {
		if _, err := conn.Write([]byte(cmd)); err != nil {
			t.Fatalf("Write error: %v", err)
		}
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatalf("Read error: %v", err)
		}
		response := string(buf[:n])
		if expected != "" && response != expected {
			t.Fatalf("Expected %q, got %q", expected, response)
		}
		return response
	}

	// 1. SET key value
	sendAndVerify("*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n", "+OK\r\n")

	// 2. EXPIRE key 5
	sendAndVerify("*3\r\n$6\r\nEXPIRE\r\n$3\r\nkey\r\n$1\r\n5\r\n", ":1\r\n")

	// 3. EXPIRETIME key
	response := sendAndVerify("*2\r\n$10\r\nEXPIRETIME\r\n$3\r\nkey\r\n", "")
	tsStr := response[1 : len(response)-2]
	ts, err := strconv.ParseInt(tsStr, 10, 64)
	if err != nil {
		t.Fatalf("Expected integer EXPIRETIME, got %q", tsStr)
	}

	expectedTs := time.Now().Add(5 * time.Second).Unix()
	if ts < expectedTs-1 || ts > expectedTs+1 {
		t.Fatalf("Expected EXPIRETIME approx %d, got %d", expectedTs, ts)
	}

	// 4. EXPIRETIME missing_key → :-2
	sendAndVerify("*2\r\n$10\r\nEXPIRETIME\r\n$11\r\nmissing_key\r\n", ":-2\r\n")

	// 5. EXPIRETIME persistent_key
	sendAndVerify("*3\r\n$3\r\nSET\r\n$14\r\npersistent_key\r\n$5\r\nvalue\r\n", "+OK\r\n")
	sendAndVerify("*2\r\n$10\r\nEXPIRETIME\r\n$14\r\npersistent_key\r\n", ":-1\r\n")
}

func TestIntegrationPExpireTime(t *testing.T) {
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

	conn, err := net.Dial("tcp", address)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	buf := make([]byte, 1024)
	sendAndVerify := func(cmd string, expected string) string {
		if _, err := conn.Write([]byte(cmd)); err != nil {
			t.Fatalf("Write error: %v", err)
		}
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatalf("Read error: %v", err)
		}
		response := string(buf[:n])
		if expected != "" && response != expected {
			t.Fatalf("Expected %q, got %q", expected, response)
		}
		return response
	}

	// 1. SET key value
	sendAndVerify("*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n", "+OK\r\n")

	// 2. PEXPIRE key 5000
	sendAndVerify("*3\r\n$7\r\nPEXPIRE\r\n$3\r\nkey\r\n$4\r\n5000\r\n", ":1\r\n")

	// 3. PEXPIRETIME key
	response := sendAndVerify("*2\r\n$11\r\nPEXPIRETIME\r\n$3\r\nkey\r\n", "")
	tsStr := response[1 : len(response)-2]
	ts, err := strconv.ParseInt(tsStr, 10, 64)
	if err != nil {
		t.Fatalf("Expected integer PEXPIRETIME, got %q", tsStr)
	}

	expectedTs := time.Now().Add(5 * time.Second).UnixMilli()
	if ts < expectedTs-1000 || ts > expectedTs+1000 {
		t.Fatalf("Expected PEXPIRETIME approx %d, got %d", expectedTs, ts)
	}

	// 4. PEXPIRETIME missing_key → :-2
	sendAndVerify("*2\r\n$11\r\nPEXPIRETIME\r\n$11\r\nmissing_key\r\n", ":-2\r\n")

	// 5. PEXPIRETIME persistent_key
	sendAndVerify("*3\r\n$3\r\nSET\r\n$14\r\npersistent_key\r\n$5\r\nvalue\r\n", "+OK\r\n")
	sendAndVerify("*2\r\n$11\r\nPEXPIRETIME\r\n$14\r\npersistent_key\r\n", ":-1\r\n")
}

func TestIntegrationLPush(t *testing.T) {
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

	conn, err := net.Dial("tcp", address)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	buf := make([]byte, 1024)
	sendAndVerify := func(cmd string, expected string) string {
		if _, err := conn.Write([]byte(cmd)); err != nil {
			t.Fatalf("Write error: %v", err)
		}
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatalf("Read error: %v", err)
		}
		response := string(buf[:n])
		if expected != "" && response != expected {
			t.Fatalf("Expected %q, got %q", expected, response)
		}
		return response
	}

	// 1. LPUSH mylist a
	sendAndVerify("*3\r\n$5\r\nLPUSH\r\n$6\r\nmylist\r\n$1\r\na\r\n", ":1\r\n")

	// 2. LPUSH mylist b c
	sendAndVerify("*4\r\n$5\r\nLPUSH\r\n$6\r\nmylist\r\n$1\r\nb\r\n$1\r\nc\r\n", ":3\r\n")

	// 3. GET mylist
	sendAndVerify("*2\r\n$3\r\nGET\r\n$6\r\nmylist\r\n", "-ERR WRONGTYPE Operation against a key holding the wrong kind of value\r\n")

	// 4. LPUSH newlist x y z
	sendAndVerify("*5\r\n$5\r\nLPUSH\r\n$7\r\nnewlist\r\n$1\r\nx\r\n$1\r\ny\r\n$1\r\nz\r\n", ":3\r\n")

	// 5. GET newlist
	sendAndVerify("*2\r\n$3\r\nGET\r\n$7\r\nnewlist\r\n", "-ERR WRONGTYPE Operation against a key holding the wrong kind of value\r\n")

	// 6. LPUSH with missing value
	sendAndVerify("*2\r\n$5\r\nLPUSH\r\n$6\r\nmylist\r\n", "-ERR wrong number of arguments\r\n")

	// 7. LPUSH onto a String
	sendAndVerify("*3\r\n$3\r\nSET\r\n$9\r\nstringkey\r\n$5\r\nvalue\r\n", "+OK\r\n")
	sendAndVerify("*3\r\n$5\r\nLPUSH\r\n$9\r\nstringkey\r\n$1\r\nx\r\n", "-ERR WRONGTYPE Operation against a key holding the wrong kind of value\r\n")
}

func TestIntegrationRPush(t *testing.T) {
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

	conn, err := net.Dial("tcp", address)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	buf := make([]byte, 1024)
	sendAndVerify := func(cmd string, expected string) string {
		if _, err := conn.Write([]byte(cmd)); err != nil {
			t.Fatalf("Write error: %v", err)
		}
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatalf("Read error: %v", err)
		}
		response := string(buf[:n])
		if expected != "" && response != expected {
			t.Fatalf("Expected %q, got %q", expected, response)
		}
		return response
	}

	// 1. RPUSH mylist a
	sendAndVerify("*3\r\n$5\r\nRPUSH\r\n$6\r\nmylist\r\n$1\r\na\r\n", ":1\r\n")

	// 2. RPUSH mylist b c
	sendAndVerify("*4\r\n$5\r\nRPUSH\r\n$6\r\nmylist\r\n$1\r\nb\r\n$1\r\nc\r\n", ":3\r\n")

	// 3. GET mylist
	sendAndVerify("*2\r\n$3\r\nGET\r\n$6\r\nmylist\r\n", "-ERR WRONGTYPE Operation against a key holding the wrong kind of value\r\n")

	// 4. RPUSH newlist x y z
	sendAndVerify("*5\r\n$5\r\nRPUSH\r\n$7\r\nnewlist\r\n$1\r\nx\r\n$1\r\ny\r\n$1\r\nz\r\n", ":3\r\n")

	// 5. GET newlist
	sendAndVerify("*2\r\n$3\r\nGET\r\n$7\r\nnewlist\r\n", "-ERR WRONGTYPE Operation against a key holding the wrong kind of value\r\n")

	// 6. RPUSH with only the key
	sendAndVerify("*2\r\n$5\r\nRPUSH\r\n$6\r\nmylist\r\n", "-ERR wrong number of arguments\r\n")

	// 7. SET stringkey value
	sendAndVerify("*3\r\n$3\r\nSET\r\n$9\r\nstringkey\r\n$5\r\nvalue\r\n", "+OK\r\n")

	// 8. RPUSH stringkey x
	sendAndVerify("*3\r\n$5\r\nRPUSH\r\n$9\r\nstringkey\r\n$1\r\nx\r\n", "-ERR WRONGTYPE Operation against a key holding the wrong kind of value\r\n")
}

func TestIntegrationLLen(t *testing.T) {
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

	conn, err := net.Dial("tcp", address)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	buf := make([]byte, 1024)
	sendAndVerify := func(cmd string, expected string) string {
		if _, err := conn.Write([]byte(cmd)); err != nil {
			t.Fatalf("Write error: %v", err)
		}
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatalf("Read error: %v", err)
		}
		response := string(buf[:n])
		if expected != "" && response != expected {
			t.Fatalf("Expected %q, got %q", expected, response)
		}
		return response
	}

	// 1. RPUSH mylist a
	sendAndVerify("*3\r\n$5\r\nRPUSH\r\n$6\r\nmylist\r\n$1\r\na\r\n", ":1\r\n")

	// 2. RPUSH mylist b c
	sendAndVerify("*4\r\n$5\r\nRPUSH\r\n$6\r\nmylist\r\n$1\r\nb\r\n$1\r\nc\r\n", ":3\r\n")

	// 3. LLEN mylist
	sendAndVerify("*2\r\n$4\r\nLLEN\r\n$6\r\nmylist\r\n", ":3\r\n")

	// 4. LLEN missing
	sendAndVerify("*2\r\n$4\r\nLLEN\r\n$7\r\nmissing\r\n", ":0\r\n")

	// 5. SET stringkey value
	sendAndVerify("*3\r\n$3\r\nSET\r\n$9\r\nstringkey\r\n$5\r\nvalue\r\n", "+OK\r\n")

	// 6. LLEN stringkey
	sendAndVerify("*2\r\n$4\r\nLLEN\r\n$9\r\nstringkey\r\n", "-ERR WRONGTYPE Operation against a key holding the wrong kind of value\r\n")

	// 7. LLEN with no key
	sendAndVerify("*1\r\n$4\r\nLLEN\r\n", "-ERR wrong number of arguments\r\n")
}

func TestIntegrationLRange(t *testing.T) {
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

	conn, err := net.Dial("tcp", address)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	buf := make([]byte, 1024)
	sendAndVerify := func(cmd string, expected string) string {
		if _, err := conn.Write([]byte(cmd)); err != nil {
			t.Fatalf("Write error: %v", err)
		}
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatalf("Read error: %v", err)
		}
		response := string(buf[:n])
		if expected != "" && response != expected {
			t.Fatalf("Expected %q, got %q", expected, response)
		}
		return response
	}

	// 1. RPUSH mylist a b c d e
	sendAndVerify(
		"*7\r\n"+
			"$5\r\nRPUSH\r\n"+
			"$6\r\nmylist\r\n"+
			"$1\r\na\r\n"+
			"$1\r\nb\r\n"+
			"$1\r\nc\r\n"+
			"$1\r\nd\r\n"+
			"$1\r\ne\r\n",
		":5\r\n",
	)

	// 2. LRANGE mylist 0 -1
	sendAndVerify(
		"*4\r\n"+
			"$6\r\nLRANGE\r\n"+
			"$6\r\nmylist\r\n"+
			"$1\r\n0\r\n"+
			"$2\r\n-1\r\n",
		"*5\r\n"+
			"$1\r\na\r\n"+
			"$1\r\nb\r\n"+
			"$1\r\nc\r\n"+
			"$1\r\nd\r\n"+
			"$1\r\ne\r\n",
	)

	// 3. LRANGE mylist 1 3
	sendAndVerify("*4\r\n$6\r\nLRANGE\r\n$6\r\nmylist\r\n$1\r\n1\r\n$1\r\n3\r\n", "*3\r\n$1\r\nb\r\n$1\r\nc\r\n$1\r\nd\r\n")

	// 4. LRANGE mylist -2 -1
	sendAndVerify("*4\r\n$6\r\nLRANGE\r\n$6\r\nmylist\r\n$2\r\n-2\r\n$2\r\n-1\r\n", "*2\r\n$1\r\nd\r\n$1\r\ne\r\n")

	// 5. LRANGE mylist 100 200
	sendAndVerify("*4\r\n$6\r\nLRANGE\r\n$6\r\nmylist\r\n$3\r\n100\r\n$3\r\n200\r\n", "*0\r\n")

	// 6. LRANGE missing 0 -1
	sendAndVerify("*4\r\n$6\r\nLRANGE\r\n$7\r\nmissing\r\n$1\r\n0\r\n$2\r\n-1\r\n", "*0\r\n")

	// 7. SET stringkey value
	sendAndVerify("*3\r\n$3\r\nSET\r\n$9\r\nstringkey\r\n$5\r\nvalue\r\n", "+OK\r\n")

	// 8. LRANGE stringkey 0 -1
	sendAndVerify("*4\r\n$6\r\nLRANGE\r\n$9\r\nstringkey\r\n$1\r\n0\r\n$2\r\n-1\r\n", "-ERR WRONGTYPE Operation against a key holding the wrong kind of value\r\n")

	// 9. LRANGE with missing arguments
	sendAndVerify("*2\r\n$6\r\nLRANGE\r\n$6\r\nmylist\r\n", "-ERR wrong number of arguments\r\n")

	// 10. Invalid start
	sendAndVerify("*4\r\n$6\r\nLRANGE\r\n$6\r\nmylist\r\n$3\r\nabc\r\n$2\r\n-1\r\n", "-ERR value is not an integer or out of range\r\n")

	// 11. Invalid stop
	sendAndVerify("*4\r\n$6\r\nLRANGE\r\n$6\r\nmylist\r\n$1\r\n0\r\n$3\r\nxyz\r\n", "-ERR value is not an integer or out of range\r\n")
}

func TestIntegrationLPop(t *testing.T) {
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

	conn, err := net.Dial("tcp", address)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	buf := make([]byte, 1024)
	sendAndVerify := func(cmd string, expected string) string {
		if _, err := conn.Write([]byte(cmd)); err != nil {
			t.Fatalf("Write error: %v", err)
		}
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatalf("Read error: %v", err)
		}
		response := string(buf[:n])
		if expected != "" && response != expected {
			t.Fatalf("Expected %q, got %q", expected, response)
		}
		return response
	}

	// 1. RPUSH mylist a b c
	sendAndVerify(
		"*5\r\n"+
			"$5\r\nRPUSH\r\n"+
			"$6\r\nmylist\r\n"+
			"$1\r\na\r\n"+
			"$1\r\nb\r\n"+
			"$1\r\nc\r\n",
		":3\r\n",
	)

	// 2. LPOP mylist
	sendAndVerify(
		"*2\r\n"+
			"$4\r\nLPOP\r\n"+
			"$6\r\nmylist\r\n",
		"$1\r\na\r\n",
	)

	// 3. LPOP mylist again
	sendAndVerify(
		"*2\r\n"+
			"$4\r\nLPOP\r\n"+
			"$6\r\nmylist\r\n",
		"$1\r\nb\r\n",
	)

	// 4. LPOP mylist (last element)
	sendAndVerify(
		"*2\r\n"+
			"$4\r\nLPOP\r\n"+
			"$6\r\nmylist\r\n",
		"$1\r\nc\r\n",
	)

	// 5. LPOP mylist again (key should no longer exist)
	sendAndVerify(
		"*2\r\n"+
			"$4\r\nLPOP\r\n"+
			"$6\r\nmylist\r\n",
		"$-1\r\n",
	)

	// 6. LPOP missing key
	sendAndVerify(
		"*2\r\n"+
			"$4\r\nLPOP\r\n"+
			"$7\r\nmissing\r\n",
		"$-1\r\n",
	)

	// 7. SET stringkey value
	sendAndVerify(
		"*3\r\n"+
			"$3\r\nSET\r\n"+
			"$9\r\nstringkey\r\n"+
			"$5\r\nvalue\r\n",
		"+OK\r\n",
	)

	// 8. LPOP stringkey
	sendAndVerify(
		"*2\r\n"+
			"$4\r\nLPOP\r\n"+
			"$9\r\nstringkey\r\n",
		"-ERR WRONGTYPE Operation against a key holding the wrong kind of value\r\n",
	)

	// 9. LPOP with missing arguments
	sendAndVerify(
		"*1\r\n"+
			"$4\r\nLPOP\r\n",
		"-ERR wrong number of arguments\r\n",
	)
}

func TestIntegrationRPop(t *testing.T) {
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

	conn, err := net.Dial("tcp", address)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	buf := make([]byte, 1024)
	sendAndVerify := func(cmd string, expected string) string {
		if _, err := conn.Write([]byte(cmd)); err != nil {
			t.Fatalf("Write error: %v", err)
		}
		n, err := conn.Read(buf)
		if err != nil {
			t.Fatalf("Read error: %v", err)
		}
		response := string(buf[:n])
		if expected != "" && response != expected {
			t.Fatalf("Expected %q, got %q", expected, response)
		}
		return response
	}

	// 1. RPUSH mylist a b c
	sendAndVerify(
		"*5\r\n"+
			"$5\r\nRPUSH\r\n"+
			"$6\r\nmylist\r\n"+
			"$1\r\na\r\n"+
			"$1\r\nb\r\n"+
			"$1\r\nc\r\n",
		":3\r\n",
	)

	// 2. RPOP mylist
	sendAndVerify(
		"*2\r\n"+
			"$4\r\nRPOP\r\n"+
			"$6\r\nmylist\r\n",
		"$1\r\nc\r\n",
	)

	// 3. RPOP mylist again
	sendAndVerify(
		"*2\r\n"+
			"$4\r\nRPOP\r\n"+
			"$6\r\nmylist\r\n",
		"$1\r\nb\r\n",
	)

	// 4. RPOP mylist (last element)
	sendAndVerify(
		"*2\r\n"+
			"$4\r\nRPOP\r\n"+
			"$6\r\nmylist\r\n",
		"$1\r\na\r\n",
	)

	// 5. RPOP mylist again (key should no longer exist)
	sendAndVerify(
		"*2\r\n"+
			"$4\r\nRPOP\r\n"+
			"$6\r\nmylist\r\n",
		"$-1\r\n",
	)

	// 6. RPOP missing key
	sendAndVerify(
		"*2\r\n"+
			"$4\r\nRPOP\r\n"+
			"$7\r\nmissing\r\n",
		"$-1\r\n",
	)

	// 7. SET stringkey value
	sendAndVerify(
		"*3\r\n"+
			"$3\r\nSET\r\n"+
			"$9\r\nstringkey\r\n"+
			"$5\r\nvalue\r\n",
		"+OK\r\n",
	)

	// 8. RPOP stringkey
	sendAndVerify(
		"*2\r\n"+
			"$4\r\nRPOP\r\n"+
			"$9\r\nstringkey\r\n",
		"-ERR WRONGTYPE Operation against a key holding the wrong kind of value\r\n",
	)

	// 9. RPOP with missing arguments
	sendAndVerify(
		"*1\r\n"+
			"$4\r\nRPOP\r\n",
		"-ERR wrong number of arguments\r\n",
	)
}
