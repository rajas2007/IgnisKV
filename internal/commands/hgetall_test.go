package commands

import (
	"testing"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

func TestHGetAllHandler_Execute(t *testing.T) {
	s := store.NewMemoryStore()
	h := NewHGetAllHandler(s)

	tests := []struct {
		name           string
		args           []string
		setup          func()
		expectedStatus types.ResponseStatus
		expectedMsg    string
		expectEmpty    bool
		expectedMap    map[string]string
	}{
		{
			name: "existing hash",
			args: []string{"hash1"},
			setup: func() {
				s.HSet("hash1", []string{"f1", "v1", "f2", "v2"})
			},
			expectedStatus: types.StatusArray,
			expectedMap:    map[string]string{"f1": "v1", "f2": "v2"},
		},
		{
			name:           "missing key",
			args:           []string{"missing_hash"},
			expectedStatus: types.StatusArray,
			expectEmpty:    true,
		},
		{
			name: "wrong type",
			args: []string{"string_key"},
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
			expectedMsg:    "wrong number of arguments for 'hgetall' command",
		},
		{
			name:           "invalid argument count - too many args",
			args:           []string{"hash1", "extra"},
			expectedStatus: types.StatusError,
			expectedMsg:    "wrong number of arguments for 'hgetall' command",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s = store.NewMemoryStore()
			h = NewHGetAllHandler(s)
			if tt.setup != nil {
				tt.setup()
			}

			cmd := types.Command{
				Name: "HGETALL",
				Args: tt.args,
			}

			resp := h.Execute(cmd)

			if resp.Status != tt.expectedStatus {
				t.Errorf("expected status %v, got %v", tt.expectedStatus, resp.Status)
			}
			if tt.expectedMsg != "" && resp.Message != tt.expectedMsg {
				t.Errorf("expected message %q, got %q", tt.expectedMsg, resp.Message)
			}

			if tt.expectedStatus == types.StatusArray {
				arr, ok := resp.Data.([]string)
				if !ok {
					t.Fatalf("expected []string data, got %T", resp.Data)
				}

				if tt.expectEmpty {
					if len(arr) != 0 {
						t.Errorf("expected empty array, got %v", arr)
					}
					return
				}

				if tt.expectedMap != nil {
					if len(arr)%2 != 0 {
						t.Fatalf("expected even number of elements, got %d", len(arr))
					}
					if len(arr)/2 != len(tt.expectedMap) {
						t.Fatalf("expected %d pairs, got %d", len(tt.expectedMap), len(arr)/2)
					}
					// Reconstruct map to compare without ordering dependency
					got := make(map[string]string)
					for i := 0; i < len(arr); i += 2 {
						got[arr[i]] = arr[i+1]
					}
					for k, v := range tt.expectedMap {
						if got[k] != v {
							t.Errorf("expected %s=%s, got %s=%s", k, v, k, got[k])
						}
					}
				}
			}
		})
	}
}

func TestHGetAllHandler_DispatcherIntegration(t *testing.T) {
	s := store.NewMemoryStore()
	d := NewDispatcher(s)

	s.HSet("hash1", []string{"f1", "v1"})

	cmd := types.Command{
		Name: "HGETALL",
		Args: []string{"hash1"},
	}

	resp := d.Dispatch(cmd)
	if resp.Status != types.StatusArray {
		t.Fatalf("expected StatusArray, got %v", resp.Status)
	}

	arr, ok := resp.Data.([]string)
	if !ok {
		t.Fatalf("expected []string data, got %T", resp.Data)
	}
	if len(arr) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(arr))
	}
}
