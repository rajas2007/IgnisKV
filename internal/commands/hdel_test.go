package commands

import (
	"testing"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

func TestHDelHandler_Execute(t *testing.T) {
	s := store.NewMemoryStore()
	h := NewHDelHandler(s)

	tests := []struct {
		name     string
		args     []string
		setup    func()
		expected types.Response
	}{
		{
			name: "successful deletion of single field",
			args: []string{"hash1", "field1"},
			setup: func() {
				s.HSet("hash1", []string{"field1", "value1", "field2", "value2"})
			},
			expected: types.Response{
				Status: types.StatusInteger,
				Data:   "1",
			},
		},
		{
			name: "successful deletion of multiple fields",
			args: []string{"hash2", "field1", "field2", "missing"},
			setup: func() {
				s.HSet("hash2", []string{"field1", "value1", "field2", "value2"})
			},
			expected: types.Response{
				Status: types.StatusInteger,
				Data:   "2",
			},
		},
		{
			name: "duplicate fields in command",
			args: []string{"hash_dup", "f1", "f1", "f1"},
			setup: func() {
				s.HSet("hash_dup", []string{"f1", "v1", "f2", "v2"})
			},
			expected: types.Response{
				Status: types.StatusInteger,
				Data:   "1", // Returns 1 because f1 is deleted only once
			},
		},
		{
			name: "missing key",
			args: []string{"missing_hash", "field1"},
			expected: types.Response{
				Status: types.StatusInteger,
				Data:   "0",
			},
		},
		{
			name: "zero deletions",
			args: []string{"hash3", "missing_field"},
			setup: func() {
				s.HSet("hash3", []string{"field1", "value1"})
			},
			expected: types.Response{
				Status: types.StatusInteger,
				Data:   "0",
			},
		},
		{
			name: "wrong type",
			args: []string{"string_key", "field1"},
			setup: func() {
				s.Set("string_key", types.Value{Type: types.StringType, Data: "val"})
			},
			expected: types.Response{
				Status:  types.StatusError,
				Message: "WRONGTYPE Operation against a key holding the wrong kind of value",
			},
		},
		{
			name: "invalid argument count - too few",
			args: []string{"hash1"},
			expected: types.Response{
				Status:  types.StatusError,
				Message: "wrong number of arguments for 'hdel' command",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s = store.NewMemoryStore()
			h = NewHDelHandler(s)
			if tt.setup != nil {
				tt.setup()
			}

			cmd := types.Command{
				Name: "HDEL",
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

func TestHDelHandler_Persistence(t *testing.T) {
	s := store.NewMemoryStore()
	h := NewHDelHandler(s)
	s.HSet("hash1", []string{"f1", "v1"})

	cmd := types.Command{
		Name: "HDEL",
		Args: []string{"hash1", "f1"},
	}

	resp := h.Execute(cmd)
	if resp.Status != types.StatusInteger {
		t.Fatalf("expected integer status, got %v", resp.Status)
	}

	cmdErr := types.Command{
		Name: "HDEL",
		Args: []string{"hash1"},
	}
	respErr := h.Execute(cmdErr)
	if respErr.Status != types.StatusError {
		t.Fatalf("expected error status, got %v", respErr.Status)
	}
}
