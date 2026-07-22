package commands

import (
	"testing"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

func TestHStrLenHandler_Execute(t *testing.T) {
	s := store.NewMemoryStore()
	h := NewHStrLenHandler(s)

	tests := []struct {
		name           string
		args           []string
		setup          func()
		expectedStatus types.ResponseStatus
		expectedMsg    string
		expectedLen    int
	}{
		{
			name: "existing field",
			args: []string{"hash1", "f1"},
			setup: func() {
				s.HSet("hash1", []string{"f1", "value"})
			},
			expectedStatus: types.StatusInteger,
			expectedLen:    5,
		},
		{
			name: "existing field with UTF-8 byte length",
			args: []string{"hash1", "f2"},
			setup: func() {
				s.HSet("hash1", []string{"f2", "é"}) // "é" is 2 bytes
			},
			expectedStatus: types.StatusInteger,
			expectedLen:    2,
		},
		{
			name: "missing field",
			args: []string{"hash1", "missing"},
			setup: func() {
				s.HSet("hash1", []string{"f1", "v1"})
			},
			expectedStatus: types.StatusInteger,
			expectedLen:    0,
		},
		{
			name:           "missing key",
			args:           []string{"missing_hash", "f1"},
			expectedStatus: types.StatusInteger,
			expectedLen:    0,
		},
		{
			name: "wrong type",
			args: []string{"string_key", "f1"},
			setup: func() {
				s.Set("string_key", types.Value{Type: types.StringType, Data: "val"})
			},
			expectedStatus: types.StatusError,
			expectedMsg:    "WRONGTYPE Operation against a key holding the wrong kind of value",
		},
		{
			name:           "invalid argument count - zero args",
			args:           []string{},
			expectedStatus: types.StatusError,
			expectedMsg:    "wrong number of arguments for 'hstrlen' command",
		},
		{
			name:           "invalid argument count - one arg",
			args:           []string{"hash1"},
			expectedStatus: types.StatusError,
			expectedMsg:    "wrong number of arguments for 'hstrlen' command",
		},
		{
			name:           "invalid argument count - too many args",
			args:           []string{"hash1", "f1", "f2"},
			expectedStatus: types.StatusError,
			expectedMsg:    "wrong number of arguments for 'hstrlen' command",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s = store.NewMemoryStore()
			h = NewHStrLenHandler(s)
			if tt.setup != nil {
				tt.setup()
			}

			cmd := types.Command{
				Name: "HSTRLEN",
				Args: tt.args,
			}

			resp := h.Execute(cmd)

			if resp.Status != tt.expectedStatus {
				t.Errorf("expected status %v, got %v", tt.expectedStatus, resp.Status)
			}
			if tt.expectedMsg != "" && resp.Message != tt.expectedMsg {
				t.Errorf("expected message %q, got %q", tt.expectedMsg, resp.Message)
			}

			if tt.expectedStatus == types.StatusInteger {
				val, ok := resp.Data.(int)
				if !ok {
					t.Fatalf("expected int data, got %T", resp.Data)
				}
				if val != tt.expectedLen {
					t.Errorf("expected %d, got %d", tt.expectedLen, val)
				}
			}
		})
	}
}

func TestHStrLenHandler_DispatcherIntegration(t *testing.T) {
	s := store.NewMemoryStore()
	d := NewDispatcher(s)

	s.HSet("hash1", []string{"f1", "value"})

	cmd := types.Command{
		Name: "HSTRLEN",
		Args: []string{"hash1", "f1"},
	}

	resp := d.Dispatch(cmd)
	if resp.Status != types.StatusInteger {
		t.Fatalf("expected StatusInteger, got %v", resp.Status)
	}

	val, ok := resp.Data.(int)
	if !ok {
		t.Fatalf("expected int data, got %T", resp.Data)
	}
	if val != 5 {
		t.Errorf("expected 5, got %d", val)
	}
}
