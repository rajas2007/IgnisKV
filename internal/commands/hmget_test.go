package commands

import (
	"testing"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

func TestHMGetHandler_Execute(t *testing.T) {
	s := store.NewMemoryStore()
	h := NewHMGetHandler(s)

	tests := []struct {
		name           string
		args           []string
		setup          func()
		expectedStatus types.ResponseStatus
		expectedMsg    string
		expectedLen    int
		checkPositions func(t *testing.T, data any)
	}{
		{
			name: "existing hash all fields present",
			args: []string{"hash1", "name", "age"},
			setup: func() {
				s.HSet("hash1", []string{"name", "Rajas", "age", "19"})
			},
			expectedStatus: types.StatusArray,
			expectedLen:    2,
			checkPositions: func(t *testing.T, data any) {
				arr := data.([]any)
				if arr[0] != "Rajas" {
					t.Errorf("expected Rajas at 0, got %v", arr[0])
				}
				if arr[1] != "19" {
					t.Errorf("expected 19 at 1, got %v", arr[1])
				}
			},
		},
		{
			name: "mixed present and missing fields",
			args: []string{"hash2", "name", "city", "age"},
			setup: func() {
				s.HSet("hash2", []string{"name", "Rajas", "age", "19"})
			},
			expectedStatus: types.StatusArray,
			expectedLen:    3,
			checkPositions: func(t *testing.T, data any) {
				arr := data.([]any)
				if arr[0] != "Rajas" {
					t.Errorf("expected Rajas at 0, got %v", arr[0])
				}
				if arr[1] != nil {
					t.Errorf("expected nil at 1, got %v", arr[1])
				}
				if arr[2] != "19" {
					t.Errorf("expected 19 at 2, got %v", arr[2])
				}
			},
		},
		{
			name:           "missing key returns all nil",
			args:           []string{"missing_hash", "f1", "f2"},
			expectedStatus: types.StatusArray,
			expectedLen:    2,
			checkPositions: func(t *testing.T, data any) {
				arr := data.([]any)
				for i, v := range arr {
					if v != nil {
						t.Errorf("expected nil at %d, got %v", i, v)
					}
				}
			},
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
			name:           "invalid argument count - too few",
			args:           []string{"hash1"},
			expectedStatus: types.StatusError,
			expectedMsg:    "wrong number of arguments for 'hmget' command",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s = store.NewMemoryStore()
			h = NewHMGetHandler(s)
			if tt.setup != nil {
				tt.setup()
			}

			cmd := types.Command{
				Name: "HMGET",
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
				arr, ok := resp.Data.([]any)
				if !ok {
					t.Fatalf("expected []any data, got %T", resp.Data)
				}
				if len(arr) != tt.expectedLen {
					t.Fatalf("expected %d elements, got %d", tt.expectedLen, len(arr))
				}
				if tt.checkPositions != nil {
					tt.checkPositions(t, resp.Data)
				}
			}
		})
	}
}

func TestHMGetHandler_DispatcherIntegration(t *testing.T) {
	s := store.NewMemoryStore()
	d := NewDispatcher(s)

	s.HSet("hash1", []string{"f1", "v1", "f2", "v2"})

	cmd := types.Command{
		Name: "HMGET",
		Args: []string{"hash1", "f1", "missing", "f2"},
	}

	resp := d.Dispatch(cmd)
	if resp.Status != types.StatusArray {
		t.Fatalf("expected StatusArray, got %v", resp.Status)
	}

	arr, ok := resp.Data.([]any)
	if !ok {
		t.Fatalf("expected []any data, got %T", resp.Data)
	}
	if len(arr) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(arr))
	}
	if arr[0] != "v1" {
		t.Errorf("expected v1 at 0, got %v", arr[0])
	}
	if arr[1] != nil {
		t.Errorf("expected nil at 1, got %v", arr[1])
	}
	if arr[2] != "v2" {
		t.Errorf("expected v2 at 2, got %v", arr[2])
	}
}
