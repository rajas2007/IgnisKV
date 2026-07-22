package commands

import (
	"testing"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

func TestHSetNXHandler_Execute(t *testing.T) {
	s := store.NewMemoryStore()
	h := NewHSetNXHandler(s)

	tests := []struct {
		name           string
		args           []string
		setup          func()
		expectedStatus types.ResponseStatus
		expectedMsg    string
		expectedData   int
	}{
		{
			name:           "missing key",
			args:           []string{"hash1", "f1", "v1"},
			expectedStatus: types.StatusInteger,
			expectedData:   1,
		},
		{
			name: "existing field",
			args: []string{"hash1", "f1", "v2"},
			setup: func() {
				s.HSet("hash1", []string{"f1", "v1"})
			},
			expectedStatus: types.StatusInteger,
			expectedData:   0,
		},
		{
			name: "missing field",
			args: []string{"hash1", "f2", "v2"},
			setup: func() {
				s.HSet("hash1", []string{"f1", "v1"})
			},
			expectedStatus: types.StatusInteger,
			expectedData:   1,
		},
		{
			name: "wrong type",
			args: []string{"string_key", "f1", "v1"},
			setup: func() {
				s.Set("string_key", types.Value{Type: types.StringType, Data: "val"})
			},
			expectedStatus: types.StatusError,
			expectedMsg:    "WRONGTYPE Operation against a key holding the wrong kind of value",
		},
		{
			name:           "invalid argument count - two args",
			args:           []string{"hash1", "f1"},
			expectedStatus: types.StatusError,
			expectedMsg:    "wrong number of arguments for 'hsetnx' command",
		},
		{
			name:           "invalid argument count - four args",
			args:           []string{"hash1", "f1", "v1", "extra"},
			expectedStatus: types.StatusError,
			expectedMsg:    "wrong number of arguments for 'hsetnx' command",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s = store.NewMemoryStore()
			h = NewHSetNXHandler(s)
			if tt.setup != nil {
				tt.setup()
			}

			cmd := types.Command{
				Name: "HSETNX",
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
				if val != tt.expectedData {
					t.Errorf("expected %d, got %d", tt.expectedData, val)
				}
			}
		})
	}
}

func TestHSetNXHandler_DispatcherIntegration(t *testing.T) {
	s := store.NewMemoryStore()
	d := NewDispatcher(s)

	cmd := types.Command{
		Name: "HSETNX",
		Args: []string{"hash1", "f1", "v1"},
	}

	resp := d.Dispatch(cmd)
	if resp.Status != types.StatusInteger {
		t.Fatalf("expected StatusInteger, got %v", resp.Status)
	}

	val, ok := resp.Data.(int)
	if !ok {
		t.Fatalf("expected int data, got %T", resp.Data)
	}
	if val != 1 {
		t.Errorf("expected 1, got %d", val)
	}
}
