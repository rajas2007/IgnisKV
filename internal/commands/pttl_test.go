package commands

import (
	"strconv"
	"testing"
	"time"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

func TestPTTLWithExpiration(t *testing.T) {
	// Arrange
	s := store.NewMemoryStore()
	h := NewPTTLHandler(s)

	s.Set("key", types.Value{
		Type:      types.StringType,
		Data:      "value",
		ExpiresAt: time.Now().Add(3 * time.Second),
	})

	// Act
	resp := h.Execute(types.Command{Name: "PTTL", Args: []string{"key"}})

	// Assert
	if resp.Status != types.StatusInteger {
		t.Fatalf("PTTL returned Status %v; want StatusInteger", resp.Status)
	}

	ms, err := strconv.ParseInt(resp.Data.(string), 10, 64)
	if err != nil {
		t.Fatalf("PTTL returned non-integer Data %q", resp.Data)
	}

	if ms < 1800 || ms > 3000 {
		t.Fatalf("PTTL returned %d ms; expected 1800–3000", ms)
	}
}

func TestPTTLPersistentKey(t *testing.T) {
	// Arrange
	s := store.NewMemoryStore()
	h := NewPTTLHandler(s)

	s.Set("key", types.Value{
		Type: types.StringType,
		Data: "value",
	})

	// Act
	resp := h.Execute(types.Command{Name: "PTTL", Args: []string{"key"}})

	// Assert
	if resp.Status != types.StatusInteger {
		t.Fatalf("PTTL returned Status %v; want StatusInteger", resp.Status)
	}
	if resp.Data != "-1" {
		t.Fatalf("PTTL returned Data %v; want %q", resp.Data, "-1")
	}
}

func TestPTTLMissingKey(t *testing.T) {
	// Arrange
	s := store.NewMemoryStore()
	h := NewPTTLHandler(s)

	// Act
	resp := h.Execute(types.Command{Name: "PTTL", Args: []string{"missing"}})

	// Assert
	if resp.Status != types.StatusInteger {
		t.Fatalf("PTTL returned Status %v; want StatusInteger", resp.Status)
	}
	if resp.Data != "-2" {
		t.Fatalf("PTTL returned Data %v; want %q", resp.Data, "-2")
	}
}

func TestPTTLExpiredKey(t *testing.T) {
	// Arrange
	s := store.NewMemoryStore()
	h := NewPTTLHandler(s)

	s.Set("key", types.Value{
		Type:      types.StringType,
		Data:      "value",
		ExpiresAt: time.Now().Add(50 * time.Millisecond),
	})

	time.Sleep(100 * time.Millisecond)

	// Act
	resp := h.Execute(types.Command{Name: "PTTL", Args: []string{"key"}})

	// Assert
	if resp.Status != types.StatusInteger {
		t.Fatalf("PTTL returned Status %v; want StatusInteger", resp.Status)
	}
	if resp.Data != "-2" {
		t.Fatalf("PTTL returned Data %v; want %q", resp.Data, "-2")
	}

	if s.Exists("key") {
		t.Fatalf("expired key should have been lazily deleted by PTTL")
	}
}

func TestPTTLWrongArgumentCount(t *testing.T) {
	// Arrange
	s := store.NewMemoryStore()
	h := NewPTTLHandler(s)

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
			resp := h.Execute(types.Command{Name: "PTTL", Args: tt.args})

			// Assert
			if resp.Status != types.StatusError {
				t.Fatalf("PTTL with %s returned Status %v; want StatusError", tt.name, resp.Status)
			}
			if resp.Message != "wrong number of arguments" {
				t.Fatalf("PTTL with %s returned Message %q; want %q", tt.name, resp.Message, "wrong number of arguments")
			}
		})
	}
}

func TestDispatcherRoutesPTTL(t *testing.T) {
	// Arrange
	s := store.NewMemoryStore()
	d := NewDispatcher(s)

	s.Set("key", types.Value{
		Type: types.StringType,
		Data: "value",
	})

	// Act
	resp := d.Dispatch(types.Command{Name: "PTTL", Args: []string{"key"}})

	// Assert
	if resp.Status != types.StatusInteger || resp.Data != "-1" {
		t.Fatalf("Dispatcher PTTL returned Status %v, Data %q; want StatusInteger, \"-1\"", resp.Status, resp.Data)
	}
}

func TestPTTLMillisecondPrecision(t *testing.T) {
	// Arrange
	s := store.NewMemoryStore()
	h := NewPTTLHandler(s)

	s.Set("key", types.Value{
		Type:      types.StringType,
		Data:      "value",
		ExpiresAt: time.Now().Add(1500 * time.Millisecond),
	})

	// Act
	resp := h.Execute(types.Command{Name: "PTTL", Args: []string{"key"}})

	// Assert
	if resp.Status != types.StatusInteger {
		t.Fatalf("PTTL returned Status %v; want StatusInteger", resp.Status)
	}

	ms, err := strconv.ParseInt(resp.Data.(string), 10, 64)
	if err != nil {
		t.Fatalf("PTTL returned non-integer Data %q", resp.Data)
	}

	if ms < 1000 || ms > 1500 {
		t.Fatalf("PTTL returned %d ms; expected 1000–1500", ms)
	}
}
