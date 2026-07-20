package commands

import (
	"testing"
	"time"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

func TestHSetHandler_Execute(t *testing.T) {
	s := store.NewMemoryStore()
	h := NewHSetHandler(s)

	tests := []struct {
		name     string
		args     []string
		setup    func()
		expected types.Response
	}{
		{
			name: "missing key single pair",
			args: []string{"hash1", "field1", "value1"},
			expected: types.Response{
				Status: types.StatusInteger,
				Data:   "1",
			},
		},
		{
			name: "existing key new pair",
			args: []string{"hash1", "field2", "value2"},
			setup: func() {
				s.HSet("hash1", []string{"field1", "value1"})
			},
			expected: types.Response{
				Status: types.StatusInteger,
				Data:   "1",
			},
		},
		{
			name: "update existing pair",
			args: []string{"hash1", "field1", "value1_new"},
			setup: func() {
				s.HSet("hash1", []string{"field1", "value1"})
			},
			expected: types.Response{
				Status: types.StatusInteger,
				Data:   "0",
			},
		},
		{
			name: "wrong number of arguments - too few",
			args: []string{"hash1", "field1"},
			expected: types.Response{
				Status:  types.StatusError,
				Message: "wrong number of arguments for 'hset' command",
			},
		},
		{
			name: "wrong number of arguments - odd pairs",
			args: []string{"hash1", "field1", "value1", "field2"},
			expected: types.Response{
				Status:  types.StatusError,
				Message: "wrong number of arguments for 'hset' command",
			},
		},
		{
			name: "wrong type",
			args: []string{"string_key", "field", "value"},
			setup: func() {
				s.Set("string_key", types.Value{Type: types.StringType, Data: "val"})
			},
			expected: types.Response{
				Status:  types.StatusError,
				Message: "WRONGTYPE Operation against a key holding the wrong kind of value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh store for each test to avoid interference
			s = store.NewMemoryStore()
			h = NewHSetHandler(s)
			if tt.setup != nil {
				tt.setup()
			}

			cmd := types.Command{
				Name: "HSET",
				Args: tt.args,
			}

			resp := h.Execute(cmd)

			if resp.Status != tt.expected.Status {
				t.Errorf("expected status %v, got %v", tt.expected.Status, resp.Status)
			}
			if resp.Data != tt.expected.Data {
				t.Errorf("expected data %v, got %v", tt.expected.Data, resp.Data)
			}
			if resp.Message != tt.expected.Message {
				t.Errorf("expected message %q, got %q", tt.expected.Message, resp.Message)
			}
		})
	}
}

func TestHSetHandler_Persistence(t *testing.T) {
	s := store.NewMemoryStore()
	h := NewHSetHandler(s)

	cmd := types.Command{
		Name: "HSET",
		Args: []string{"hash", "field", "value"},
	}

	// Capture initial modification time if file exists
	// But it shouldn't exist in a fresh test, so we just run the command
	resp := h.Execute(cmd)
	if resp.Status != types.StatusInteger {
		t.Fatalf("expected integer status, got %v", resp.Status)
	}

	// Attempting to run with wrong arguments should not trigger panic or persist failure
	cmdErr := types.Command{
		Name: "HSET",
		Args: []string{"hash"},
	}
	respErr := h.Execute(cmdErr)
	if respErr.Status != types.StatusError {
		t.Fatalf("expected error status, got %v", respErr.Status)
	}
}

func TestHSetHandler_TTLPreservation(t *testing.T) {
	s := store.NewMemoryStore()
	h := NewHSetHandler(s)

	// Set a key with TTL
	s.Set("hash_ttl", types.Value{
		Type:      types.HashType,
		Data:      make(map[string]string),
		ExpiresAt: time.Now().Add(5 * time.Minute),
	})

	cmd := types.Command{
		Name: "HSET",
		Args: []string{"hash_ttl", "field", "value"},
	}

	resp := h.Execute(cmd)
	if resp.Status != types.StatusInteger || resp.Data != "1" {
		t.Fatalf("expected 1 added, got %v", resp.Data)
	}

	val, _ := s.Get("hash_ttl")
	if val.ExpiresAt.IsZero() {
		t.Fatalf("expected TTL to be preserved, got zero time")
	}
}
