package commands

import (
	"sort"
	"testing"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

func TestHKeysHandler_Execute(t *testing.T) {
	s := store.NewMemoryStore()
	h := NewHKeysHandler(s)

	tests := []struct {
		name           string
		args           []string
		setup          func()
		expectedStatus types.ResponseStatus
		expectedMsg    string
		expectEmpty    bool
		expectedKeys   []string
	}{
		{
			name: "existing hash",
			args: []string{"hash1"},
			setup: func() {
				s.HSet("hash1", []string{"f1", "v1", "f2", "v2", "f3", "v3"})
			},
			expectedStatus: types.StatusArray,
			expectedKeys:   []string{"f1", "f2", "f3"},
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
			expectedMsg:    "wrong number of arguments for 'hkeys' command",
		},
		{
			name:           "invalid argument count - too many args",
			args:           []string{"hash1", "extra"},
			expectedStatus: types.StatusError,
			expectedMsg:    "wrong number of arguments for 'hkeys' command",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s = store.NewMemoryStore()
			h = NewHKeysHandler(s)
			if tt.setup != nil {
				tt.setup()
			}

			cmd := types.Command{
				Name: "HKEYS",
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

				if tt.expectedKeys != nil {
					if len(arr) != len(tt.expectedKeys) {
						t.Fatalf("expected %d keys, got %d", len(tt.expectedKeys), len(arr))
					}
					// Sort both slices before comparing
					sortedGot := make([]string, len(arr))
					copy(sortedGot, arr)
					sort.Strings(sortedGot)

					sortedExpected := make([]string, len(tt.expectedKeys))
					copy(sortedExpected, tt.expectedKeys)
					sort.Strings(sortedExpected)

					for i := range sortedGot {
						if sortedGot[i] != sortedExpected[i] {
							t.Errorf("expected %s at position %d, got %s", sortedExpected[i], i, sortedGot[i])
						}
					}
				}
			}
		})
	}
}

func TestHKeysHandler_DispatcherIntegration(t *testing.T) {
	s := store.NewMemoryStore()
	d := NewDispatcher(s)

	s.HSet("hash1", []string{"f1", "v1", "f2", "v2"})

	cmd := types.Command{
		Name: "HKEYS",
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

	sort.Strings(arr)
	if arr[0] != "f1" || arr[1] != "f2" {
		t.Errorf("expected [f1, f2], got %v", arr)
	}
}
