package commands

import (
	"testing"
	"time"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

func TestHGetHandler_Execute(t *testing.T) {
	s := store.NewMemoryStore()
	h := NewHGetHandler(s)

	tests := []struct {
		name     string
		args     []string
		setup    func()
		expected types.Response
	}{
		{
			name: "successful HGET",
			args: []string{"hash1", "field1"},
			setup: func() {
				s.HSet("hash1", []string{"field1", "value1"})
			},
			expected: types.Response{
				Status: types.StatusString,
				Data:   "value1",
			},
		},
		{
			name: "missing key",
			args: []string{"missing_hash", "field1"},
			expected: types.Response{
				Status: types.StatusNil,
			},
		},
		{
			name: "missing field",
			args: []string{"hash1", "missing_field"},
			setup: func() {
				s.HSet("hash1", []string{"field1", "value1"})
			},
			expected: types.Response{
				Status: types.StatusNil,
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
				Message: "wrong number of arguments for 'hget' command",
			},
		},
		{
			name: "invalid argument count - too many",
			args: []string{"hash1", "field1", "field2"},
			expected: types.Response{
				Status:  types.StatusError,
				Message: "wrong number of arguments for 'hget' command",
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
				Status: types.StatusNil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s = store.NewMemoryStore()
			h = NewHGetHandler(s)
			if tt.setup != nil {
				tt.setup()
			}

			cmd := types.Command{
				Name: "HGET",
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
