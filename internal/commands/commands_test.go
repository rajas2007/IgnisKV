package commands

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

func TestMain(m *testing.M) {
	tempDir, err := os.MkdirTemp("", "igniskv-test-*")
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

func TestNewDispatcher(t *testing.T) {
	// Arrange
	s := store.NewMemoryStore()

	// Act
	d := NewDispatcher(s)

	// Assert
	if d == nil {
		t.Fatalf("NewDispatcher() returned nil; expected a non-nil *Dispatcher")
	}

	// Verify built-in handlers are registered by dispatching PING.
	resp := d.Dispatch(types.Command{Name: "PING"})
	if resp.Status != types.StatusOK {
		t.Fatalf("PING after NewDispatcher() returned Status %v; want StatusOK", resp.Status)
	}
}

func TestDispatcherUnknownCommand(t *testing.T) {
	// Arrange
	s := store.NewMemoryStore()
	d := NewDispatcher(s)

	// Act
	resp := d.Dispatch(types.Command{Name: "UNKNOWN"})

	// Assert
	if resp.Status != types.StatusError {
		t.Fatalf("unknown command returned Status %v; want StatusError", resp.Status)
	}
	if resp.Message != "unknown command" {
		t.Fatalf("unknown command returned Message %q; want %q", resp.Message, "unknown command")
	}
}

func TestPingNoArguments(t *testing.T) {
	// Arrange
	s := store.NewMemoryStore()
	d := NewDispatcher(s)

	// Act
	resp := d.Dispatch(types.Command{Name: "PING"})

	// Assert
	if resp.Status != types.StatusOK {
		t.Fatalf("PING returned Status %v; want StatusOK", resp.Status)
	}
	if resp.Message != "PONG" {
		t.Fatalf("PING returned Message %q; want %q", resp.Message, "PONG")
	}
	if resp.Data != nil {
		t.Fatalf("PING returned Data %v; want nil", resp.Data)
	}
}

func TestPingOneArgument(t *testing.T) {
	// Arrange
	s := store.NewMemoryStore()
	d := NewDispatcher(s)

	// Act
	resp := d.Dispatch(types.Command{Name: "PING", Args: []string{"hello"}})

	// Assert
	if resp.Status != types.StatusOK {
		t.Fatalf("PING hello returned Status %v; want StatusOK", resp.Status)
	}
	if resp.Data != "hello" {
		t.Fatalf("PING hello returned Data %v; want %q", resp.Data, "hello")
	}
	if resp.Message != "" {
		t.Fatalf("PING hello returned Message %q; want empty string", resp.Message)
	}
}

func TestPingTooManyArguments(t *testing.T) {
	// Arrange
	s := store.NewMemoryStore()
	d := NewDispatcher(s)

	// Act
	resp := d.Dispatch(types.Command{Name: "PING", Args: []string{"a", "b"}})

	// Assert
	if resp.Status != types.StatusError {
		t.Fatalf("PING a b returned Status %v; want StatusError", resp.Status)
	}
	if resp.Message != "wrong number of arguments" {
		t.Fatalf("PING a b returned Message %q; want %q", resp.Message, "wrong number of arguments")
	}
}

// ----- SET tests -----

func TestSetValid(t *testing.T) {
	// Arrange
	s := store.NewMemoryStore()
	d := NewDispatcher(s)

	// Act
	resp := d.Dispatch(types.Command{Name: "SET", Args: []string{"name", "Rajas"}})

	// Assert — response
	if resp.Status != types.StatusOK {
		t.Fatalf("SET name Rajas returned Status %v; want StatusOK", resp.Status)
	}
	if resp.Message != "OK" {
		t.Fatalf("SET name Rajas returned Message %q; want %q", resp.Message, "OK")
	}

	// Assert — value persisted in store
	value, err := s.Get("name")
	if err != nil {
		t.Fatalf("store.Get(\"name\") after SET returned unexpected error: %v", err)
	}
	if value.Data != "Rajas" {
		t.Fatalf("store.Get(\"name\") returned Data %v; want %q", value.Data, "Rajas")
	}
}

func TestSetWrongArgumentCount(t *testing.T) {
	// Arrange
	s := store.NewMemoryStore()
	d := NewDispatcher(s)

	invalidCmds := []types.Command{
		{Name: "SET", Args: []string{}},
		{Name: "SET", Args: []string{"key"}},
		{Name: "SET", Args: []string{"key", "value", "extra"}},
	}

	for _, cmd := range invalidCmds {
		// Act
		resp := d.Dispatch(cmd)

		// Assert
		if resp.Status != types.StatusError {
			t.Fatalf("SET with %d args returned Status %v; want StatusError", len(cmd.Args), resp.Status)
		}
		if resp.Message != "wrong number of arguments" {
			t.Fatalf("SET with %d args returned Message %q; want %q", len(cmd.Args), resp.Message, "wrong number of arguments")
		}
	}
}

// ----- GET tests -----

func TestGetExistingKey(t *testing.T) {
	// Arrange
	s := store.NewMemoryStore()
	d := NewDispatcher(s)
	d.Dispatch(types.Command{Name: "SET", Args: []string{"city", "Bangalore"}})

	// Act
	resp := d.Dispatch(types.Command{Name: "GET", Args: []string{"city"}})

	// Assert
	if resp.Status != types.StatusOK {
		t.Fatalf("GET city returned Status %v; want StatusOK", resp.Status)
	}
	if resp.Data != "Bangalore" {
		t.Fatalf("GET city returned Data %v; want %q", resp.Data, "Bangalore")
	}
}

func TestGetMissingKey(t *testing.T) {
	// Arrange
	s := store.NewMemoryStore()
	d := NewDispatcher(s)

	// Act
	resp := d.Dispatch(types.Command{Name: "GET", Args: []string{"missing"}})

	// Assert
	if resp.Status != types.StatusNil {
		t.Fatalf("GET missing returned Status %v; want StatusNil", resp.Status)
	}
	if resp.Message != "" {
		t.Fatalf("GET missing returned Message %q; want empty string", resp.Message)
	}
	if resp.Data != nil {
		t.Fatalf("GET missing returned Data %v; want nil", resp.Data)
	}
}

func TestGetWrongArgumentCount(t *testing.T) {
	// Arrange
	s := store.NewMemoryStore()
	d := NewDispatcher(s)

	invalidCmds := []types.Command{
		{Name: "GET", Args: []string{}},
		{Name: "GET", Args: []string{"a", "b"}},
	}

	for _, cmd := range invalidCmds {
		// Act
		resp := d.Dispatch(cmd)

		// Assert
		if resp.Status != types.StatusError {
			t.Fatalf("GET with %d args returned Status %v; want StatusError", len(cmd.Args), resp.Status)
		}
		if resp.Message != "wrong number of arguments" {
			t.Fatalf("GET with %d args returned Message %q; want %q", len(cmd.Args), resp.Message, "wrong number of arguments")
		}
	}
}

// ----- DEL tests -----

func TestDelExistingKey(t *testing.T) {
	// Arrange
	s := store.NewMemoryStore()
	d := NewDispatcher(s)
	d.Dispatch(types.Command{Name: "SET", Args: []string{"lang", "Go"}})

	// Act
	resp := d.Dispatch(types.Command{Name: "DEL", Args: []string{"lang"}})

	// Assert — response
	if resp.Status != types.StatusOK {
		t.Fatalf("DEL lang returned Status %v; want StatusOK", resp.Status)
	}
	if resp.Message != "OK" {
		t.Fatalf("DEL lang returned Message %q; want %q", resp.Message, "OK")
	}

	// Assert — key no longer exists in store
	_, err := s.Get("lang")
	if !errors.Is(err, store.ErrKeyNotFound) {
		t.Fatalf("store.Get(\"lang\") after DEL returned %v; want ErrKeyNotFound", err)
	}
}

func TestDelMissingKey(t *testing.T) {
	// Arrange
	s := store.NewMemoryStore()
	d := NewDispatcher(s)

	// Act
	resp := d.Dispatch(types.Command{Name: "DEL", Args: []string{"missing"}})

	// Assert
	if resp.Status != types.StatusNil {
		t.Fatalf("DEL missing returned Status %v; want StatusNil", resp.Status)
	}
}

func TestDelWrongArgumentCount(t *testing.T) {
	// Arrange
	s := store.NewMemoryStore()
	d := NewDispatcher(s)

	invalidCmds := []types.Command{
		{Name: "DEL", Args: []string{}},
		{Name: "DEL", Args: []string{"a", "b"}},
	}

	for _, cmd := range invalidCmds {
		// Act
		resp := d.Dispatch(cmd)

		// Assert
		if resp.Status != types.StatusError {
			t.Fatalf("DEL with %d args returned Status %v; want StatusError", len(cmd.Args), resp.Status)
		}
		if resp.Message != "wrong number of arguments" {
			t.Fatalf("DEL with %d args returned Message %q; want %q", len(cmd.Args), resp.Message, "wrong number of arguments")
		}
	}
}

// ----- QUIT tests -----

func TestQuitValid(t *testing.T) {
	// Arrange
	s := store.NewMemoryStore()
	d := NewDispatcher(s)

	// Act
	resp := d.Dispatch(types.Command{Name: "QUIT"})

	// Assert
	if resp.Status != types.StatusExit {
		t.Fatalf("QUIT returned Status %v; want StatusExit", resp.Status)
	}
	if resp.Message != "BYE" {
		t.Fatalf("QUIT returned Message %q; want %q", resp.Message, "BYE")
	}
	if resp.Data != nil {
		t.Fatalf("QUIT returned Data %v; want nil", resp.Data)
	}
}

func TestQuitWrongArgumentCount(t *testing.T) {
	// Arrange
	s := store.NewMemoryStore()
	d := NewDispatcher(s)

	// Act
	resp := d.Dispatch(types.Command{Name: "QUIT", Args: []string{"now"}})

	// Assert
	if resp.Status != types.StatusError {
		t.Fatalf("QUIT now returned Status %v; want StatusError", resp.Status)
	}
	if resp.Message != "wrong number of arguments" {
		t.Fatalf("QUIT now returned Message %q; want %q", resp.Message, "wrong number of arguments")
	}
}

// ----- Dispatcher integration test -----

func TestDispatcherRoutesCommands(t *testing.T) {
	// Arrange
	s := store.NewMemoryStore()
	d := NewDispatcher(s)

	// PING
	resp := d.Dispatch(types.Command{Name: "PING"})
	if resp.Status != types.StatusOK || resp.Message != "PONG" {
		t.Fatalf("PING returned Status %v, Message %q; want StatusOK, \"PONG\"", resp.Status, resp.Message)
	}

	// SET
	resp = d.Dispatch(types.Command{Name: "SET", Args: []string{"key", "value"}})
	if resp.Status != types.StatusOK || resp.Message != "OK" {
		t.Fatalf("SET returned Status %v, Message %q; want StatusOK, \"OK\"", resp.Status, resp.Message)
	}

	// Verify SET persisted the value.
	stored, err := s.Get("key")
	if err != nil {
		t.Fatalf("store.Get(\"key\") after SET returned unexpected error: %v", err)
	}
	if stored.Data != "value" {
		t.Fatalf("store.Get(\"key\") returned Data %v; want %q", stored.Data, "value")
	}

	// GET
	resp = d.Dispatch(types.Command{Name: "GET", Args: []string{"key"}})
	if resp.Status != types.StatusOK {
		t.Fatalf("GET key returned Status %v; want StatusOK", resp.Status)
	}
	if resp.Data != "value" {
		t.Fatalf("GET key returned Data %v; want %q", resp.Data, "value")
	}

	// DEL
	resp = d.Dispatch(types.Command{Name: "DEL", Args: []string{"key"}})
	if resp.Status != types.StatusOK || resp.Message != "OK" {
		t.Fatalf("DEL key returned Status %v, Message %q; want StatusOK, \"OK\"", resp.Status, resp.Message)
	}

	// Verify DEL removed the value.
	resp = d.Dispatch(types.Command{Name: "GET", Args: []string{"key"}})
	if resp.Status != types.StatusNil {
		t.Fatalf("GET after DEL returned Status %v; want StatusNil", resp.Status)
	}

	// QUIT
	resp = d.Dispatch(types.Command{Name: "QUIT"})
	if resp.Status != types.StatusExit || resp.Message != "BYE" {
		t.Fatalf("QUIT returned Status %v, Message %q; want StatusExit, \"BYE\"", resp.Status, resp.Message)
	}
}

// ----- HELP tests -----

func TestHelpValid(t *testing.T) {
	// Arrange
	s := store.NewMemoryStore()
	d := NewDispatcher(s)

	// Act
	resp := d.Dispatch(types.Command{Name: "HELP"})

	// Assert
	if resp.Status != types.StatusOK {
		t.Fatalf("HELP returned Status %v; want StatusOK", resp.Status)
	}

	for _, cmd := range []string{"PING", "SET", "GET", "DEL", "HELP", "QUIT"} {
		if !strings.Contains(resp.Message, cmd) {
			t.Fatalf("HELP message does not contain %q", cmd)
		}
	}
}

func TestHelpWrongArgumentCount(t *testing.T) {
	// Arrange
	s := store.NewMemoryStore()
	d := NewDispatcher(s)

	// Act
	resp := d.Dispatch(types.Command{Name: "HELP", Args: []string{"extra"}})

	// Assert
	if resp.Status != types.StatusError {
		t.Fatalf("HELP extra returned Status %v; want StatusError", resp.Status)
	}
	if resp.Message != "wrong number of arguments" {
		t.Fatalf("HELP extra returned Message %q; want %q", resp.Message, "wrong number of arguments")
	}
}

// ----- Persistence tests -----

func TestAutomaticPersistence(t *testing.T) {
	// Arrange
	// Ensure a clean state for this specific test inside the shared TestMain tempdir
	os.Remove(store.DefaultSnapshotFile)

	s1 := store.NewMemoryStore()
	d1 := NewDispatcher(s1)

	// Act & Assert — GET does not create a snapshot
	d1.Dispatch(types.Command{Name: "GET", Args: []string{"missing"}})
	if _, err := os.Stat(store.DefaultSnapshotFile); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("GET created a snapshot file")
	}

	// Act & Assert — SET creates a snapshot
	d1.Dispatch(types.Command{Name: "SET", Args: []string{"persistent", "value"}})
	if _, err := os.Stat(store.DefaultSnapshotFile); err != nil {
		t.Fatalf("SET failed to create a snapshot file: %v", err)
	}

	// Verify the snapshot can be loaded properly
	s2 := store.NewMemoryStore()
	if err := s2.Load(store.DefaultSnapshotFile); err != nil {
		t.Fatalf("Failed to load snapshot: %v", err)
	}
	if val, _ := s2.Get("persistent"); val.Data != "value" {
		t.Fatalf("Loaded snapshot missing SET value")
	}

	// Act & Assert — DEL updates the snapshot
	d1.Dispatch(types.Command{Name: "DEL", Args: []string{"persistent"}})

	s3 := store.NewMemoryStore()
	if err := s3.Load(store.DefaultSnapshotFile); err != nil {
		t.Fatalf("Failed to load snapshot: %v", err)
	}
	if s3.Exists("persistent") {
		t.Fatalf("Loaded snapshot contained deleted value")
	}
}
