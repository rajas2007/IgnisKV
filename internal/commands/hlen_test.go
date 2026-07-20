package commands

import (
	"testing"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

func TestHLenHandler_Execute(t *testing.T) {
	s := store.NewMemoryStore()
	h := NewHLenHandler(s)

	tests := []struct {
		name     string
		args     []string
		setup    func()
		expected types.Response
	}{
		{
			name: "existing hash",
			args: []string{"hash1"},
			setup: func() {
				s.HSet("hash1", []string{"f1", "v1", "f2", "v2"})
			},
			expected: types.Response{
				Status: types.StatusInteger,
				Data:   "2",
			},
		},
		{
			name: "missing key",
			args: []string{"missing_hash"},
			expected: types.Response{
				Status: types.StatusInteger,
				Data:   "0",
			},
		},
		{
			name: "wrong type",
			args: []string{"string_key"},
			setup: func() {
				s.Set("string_key", types.Value{Type: types.StringType, Data: "val"})
			},
			expected: types.Response{
				Status:  types.StatusError,
				Message: "WRONGTYPE Operation against a key holding the wrong kind of value",
			},
		},
		{
			name: "invalid argument count - zero args",
			args: []string{},
			expected: types.Response{
				Status:  types.StatusError,
				Message: "wrong number of arguments for 'hlen' command",
			},
		},
		{
			name: "invalid argument count - too many args",
			args: []string{"hash1", "extra"},
			expected: types.Response{
				Status:  types.StatusError,
				Message: "wrong number of arguments for 'hlen' command",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s = store.NewMemoryStore()
			h = NewHLenHandler(s)
			if tt.setup != nil {
				tt.setup()
			}

			cmd := types.Command{
				Name: "HLEN",
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
