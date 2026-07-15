package commands

import (
	"strconv"
	"testing"
	"time"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

func TestExpireTimeWithExpiration(t *testing.T) {
	// Arrange
	s := store.NewMemoryStore()
	h := NewExpireTimeHandler(s)

	future := time.Now().Add(3 * time.Second)
	s.Set("key", types.Value{
		Type:      types.StringType,
		Data:      "value",
		ExpiresAt: future,
	})

	// Act
	resp := h.Execute(types.Command{Name: "EXPIRETIME", Args: []string{"key"}})

	// Assert
	if resp.Status != types.StatusInteger {
		t.Fatalf("EXPIRETIME returned Status %v; want StatusInteger", resp.Status)
	}

	ts, err := strconv.ParseInt(resp.Data.(string), 10, 64)
	if err != nil {
		t.Fatalf("EXPIRETIME returned non-integer Data %q", resp.Data)
	}

	if ts != future.Unix() {
		t.Fatalf("EXPIRETIME returned %d; expected %d", ts, future.Unix())
	}
}

func TestExpireTimePersistentKey(t *testing.T) {
	// Arrange
	s := store.NewMemoryStore()
	h := NewExpireTimeHandler(s)

	s.Set("key", types.Value{
		Type: types.StringType,
		Data: "value",
	})

	// Act
	resp := h.Execute(types.Command{Name: "EXPIRETIME", Args: []string{"key"}})

	// Assert
	if resp.Status != types.StatusInteger {
		t.Fatalf("EXPIRETIME returned Status %v; want StatusInteger", resp.Status)
	}
	if resp.Data != "-1" {
		t.Fatalf("EXPIRETIME returned Data %v; want %q", resp.Data, "-1")
	}
}

func TestExpireTimeMissingKey(t *testing.T) {
	// Arrange
	s := store.NewMemoryStore()
	h := NewExpireTimeHandler(s)

	// Act
	resp := h.Execute(types.Command{Name: "EXPIRETIME", Args: []string{"missing"}})

	// Assert
	if resp.Status != types.StatusInteger {
		t.Fatalf("EXPIRETIME returned Status %v; want StatusInteger", resp.Status)
	}
	if resp.Data != "-2" {
		t.Fatalf("EXPIRETIME returned Data %v; want %q", resp.Data, "-2")
	}
}

func TestExpireTimeExpiredKey(t *testing.T) {
	// Arrange
	s := store.NewMemoryStore()
	h := NewExpireTimeHandler(s)

	s.Set("key", types.Value{
		Type:      types.StringType,
		Data:      "value",
		ExpiresAt: time.Now().Add(50 * time.Millisecond),
	})

	time.Sleep(100 * time.Millisecond)

	// Act
	resp := h.Execute(types.Command{Name: "EXPIRETIME", Args: []string{"key"}})

	// Assert
	if resp.Status != types.StatusInteger {
		t.Fatalf("EXPIRETIME returned Status %v; want StatusInteger", resp.Status)
	}
	if resp.Data != "-2" {
		t.Fatalf("EXPIRETIME returned Data %v; want %q", resp.Data, "-2")
	}

	if s.Exists("key") {
		t.Fatalf("expired key should have been lazily deleted by EXPIRETIME")
	}
}

func TestExpireTimeWrongArgumentCount(t *testing.T) {
	// Arrange
	s := store.NewMemoryStore()
	h := NewExpireTimeHandler(s)

	tests := []struct {
		name string
		args []string
	}{
		{"no arguments", []string{}},
		{"too many arguments", []string{"key", "extra"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			resp := h.Execute(types.Command{Name: "EXPIRETIME", Args: tt.args})

			// Assert
			if resp.Status != types.StatusError {
				t.Fatalf("EXPIRETIME with %s returned Status %v; want StatusError", tt.name, resp.Status)
			}
			if resp.Message != "wrong number of arguments" {
				t.Fatalf("EXPIRETIME with %s returned Message %q; want %q", tt.name, resp.Message, "wrong number of arguments")
			}
		})
	}
}

func TestDispatcherRoutesExpireTime(t *testing.T) {
	// Arrange
	s := store.NewMemoryStore()
	d := NewDispatcher(s)

	s.Set("key", types.Value{
		Type: types.StringType,
		Data: "value",
	})

	// Act
	resp := d.Dispatch(types.Command{Name: "EXPIRETIME", Args: []string{"key"}})

	// Assert
	if resp.Status != types.StatusInteger || resp.Data != "-1" {
		t.Fatalf("Dispatcher EXPIRETIME returned Status %v, Data %q; want StatusInteger, \"-1\"", resp.Status, resp.Data)
	}
}
