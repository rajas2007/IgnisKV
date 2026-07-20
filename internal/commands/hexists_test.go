package commands

import (
	"testing"
	"time"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

func TestHExistsHandler_Execute(t *testing.T) {
	s := store.NewMemoryStore()
	h := NewHExistsHandler(s)

	tests := []struct {
		name     string
		args     []string
		setup    func()
		expected types.Response
	}{
		{
			name: "existing field",
			args: []string{"hash1", "field1"},
			setup: func() {
				s.HSet("hash1", []string{"field1", "value1"})
			},
			expected: types.Response{
				Status: types.StatusInteger,
				Data:   "1",
			},
		},
		{
			name: "missing field",
			args: []string{"hash1", "missing_field"},
			setup: func() {
				s.HSet("hash1", []string{"field1", "value1"})
			},
			expected: types.Response{
				Status: types.StatusInteger,
				Data:   "0",
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
				Message: "wrong number of arguments for 'hexists' command",
			},
		},
		{
			name: "invalid argument count - too many",
			args: []string{"hash1", "field1", "field2"},
			expected: types.Response{
				Status:  types.StatusError,
				Message: "wrong number of arguments for 'hexists' command",
			},
		},
		{
			name: "lazy expiration",
			args: []string{"hash2", "field1"},
			setup: func() {
				s.Set("hash2", types.Value{
					Type:      types.HashType,
					Data:      map[string]string{"field1": "value1"},
					ExpiresAt: time.Now().Add(-1 * time.Minute),
				})
			},
			expected: types.Response{
				Status: types.StatusInteger,
				Data:   "0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s = store.NewMemoryStore()
			h = NewHExistsHandler(s)
			if tt.setup != nil {
				tt.setup()
			}

			cmd := types.Command{
				Name: "HEXISTS",
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
