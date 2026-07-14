package commands

import (
	"os"
	"testing"
	"time"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

func TestPersistExpiringKey(t *testing.T) {
	// Arrange — create a key with an expiration
	s := store.NewMemoryStore()
	h := NewPersistHandler(s)

	s.Set("session", types.Value{
		Type:      types.StringType,
		Data:      "abc",
		ExpiresAt: time.Now().Add(10 * time.Second),
	})

	// Act
	resp := h.Execute(types.Command{Name: "PERSIST", Args: []string{"session"}})

	// Assert
	if resp.Status != types.StatusInteger {
		t.Fatalf("PERSIST returned Status %v; want StatusInteger", resp.Status)
	}
	if resp.Data != "1" {
		t.Fatalf("PERSIST returned Data %v; want %q", resp.Data, "1")
	}

	// Verify the key is now persistent
	ttl, err := s.TTL("session")
	if err != nil {
		t.Fatalf("TTL() after PERSIST returned unexpected error: %v", err)
	}
	if ttl != -1 {
		t.Fatalf("TTL() after PERSIST returned %d; want -1", ttl)
	}
}

func TestPersistAlreadyPersistentKey(t *testing.T) {
	// Arrange — create a key without expiration
	s := store.NewMemoryStore()
	h := NewPersistHandler(s)

	s.Set("permanent", types.Value{
		Type: types.StringType,
		Data: "value",
	})

	// Act
	resp := h.Execute(types.Command{Name: "PERSIST", Args: []string{"permanent"}})

	// Assert
	if resp.Status != types.StatusInteger {
		t.Fatalf("PERSIST on persistent key returned Status %v; want StatusInteger", resp.Status)
	}
	if resp.Data != "0" {
		t.Fatalf("PERSIST on persistent key returned Data %v; want %q", resp.Data, "0")
	}
}

func TestPersistMissingKey(t *testing.T) {
	// Arrange
	s := store.NewMemoryStore()
	h := NewPersistHandler(s)

	// Act
	resp := h.Execute(types.Command{Name: "PERSIST", Args: []string{"nonexistent"}})

	// Assert
	if resp.Status != types.StatusInteger {
		t.Fatalf("PERSIST on missing key returned Status %v; want StatusInteger", resp.Status)
	}
	if resp.Data != "0" {
		t.Fatalf("PERSIST on missing key returned Data %v; want %q", resp.Data, "0")
	}
}

func TestPersistExpiredKey(t *testing.T) {
	// Arrange — create a key with a very short expiration
	s := store.NewMemoryStore()
	h := NewPersistHandler(s)

	s.Set("ephemeral", types.Value{
		Type:      types.StringType,
		Data:      "temp",
		ExpiresAt: time.Now().Add(50 * time.Millisecond),
	})

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Act
	resp := h.Execute(types.Command{Name: "PERSIST", Args: []string{"ephemeral"}})

	// Assert
	if resp.Status != types.StatusInteger {
		t.Fatalf("PERSIST on expired key returned Status %v; want StatusInteger", resp.Status)
	}
	if resp.Data != "0" {
		t.Fatalf("PERSIST on expired key returned Data %v; want %q", resp.Data, "0")
	}

	// Verify the key was lazily deleted
	if s.Exists("ephemeral") {
		t.Fatalf("expired key should have been lazily deleted by PERSIST")
	}
}

func TestPersistWrongArgumentCount(t *testing.T) {
	// Arrange
	s := store.NewMemoryStore()
	h := NewPersistHandler(s)

	tests := []struct {
		name string
		args []string
	}{
		{"no arguments", []string{}},
		{"too many arguments", []string{"key1", "key2"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			resp := h.Execute(types.Command{Name: "PERSIST", Args: tt.args})

			// Assert
			if resp.Status != types.StatusError {
				t.Fatalf("PERSIST with %s returned Status %v; want StatusError", tt.name, resp.Status)
			}
			if resp.Message != "wrong number of arguments" {
				t.Fatalf("PERSIST with %s returned Message %q; want %q", tt.name, resp.Message, "wrong number of arguments")
			}
		})
	}
}

func TestPersistPersistenceAfterSuccess(t *testing.T) {
	// Arrange
	os.Remove(store.DefaultSnapshotFile)

	s := store.NewMemoryStore()
	h := NewPersistHandler(s)

	s.Set("key", types.Value{
		Type:      types.StringType,
		Data:      "value",
		ExpiresAt: time.Now().Add(10 * time.Second),
	})

	// Act — successful PERSIST should trigger Save()
	resp := h.Execute(types.Command{Name: "PERSIST", Args: []string{"key"}})

	// Assert
	if resp.Status != types.StatusInteger || resp.Data != "1" {
		t.Fatalf("PERSIST returned unexpected response: %+v", resp)
	}

	if _, err := os.Stat(store.DefaultSnapshotFile); os.IsNotExist(err) {
		t.Fatalf("expected snapshot file to be created on successful PERSIST")
	}
}

func TestPersistNoPersistenceAfterFailure(t *testing.T) {
	// Arrange
	os.Remove(store.DefaultSnapshotFile)

	s := store.NewMemoryStore()
	h := NewPersistHandler(s)

	// Act — PERSIST on missing key should NOT trigger Save()
	resp := h.Execute(types.Command{Name: "PERSIST", Args: []string{"missing"}})

	// Assert
	if resp.Status != types.StatusInteger || resp.Data != "0" {
		t.Fatalf("PERSIST on missing key returned unexpected response: %+v", resp)
	}

	if _, err := os.Stat(store.DefaultSnapshotFile); !os.IsNotExist(err) {
		t.Fatalf("expected snapshot file NOT to be created on failed PERSIST")
	}
}
