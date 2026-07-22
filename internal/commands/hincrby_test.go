package commands

import (
	"math"
	"strconv"
	"testing"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

func TestHIncrByHandler_Execute(t *testing.T) {
	s := store.NewMemoryStore()
	h := NewHIncrByHandler(s)

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
			args:           []string{"hash1", "f1", "5"},
			expectedStatus: types.StatusInteger,
			expectedData:   5,
		},
		{
			name: "existing field",
			args: []string{"hash1", "f1", "10"},
			setup: func() {
				s.HSet("hash1", []string{"f1", "5"})
			},
			expectedStatus: types.StatusInteger,
			expectedData:   15,
		},
		{
			name: "existing field negative",
			args: []string{"hash1", "f1", "-3"},
			setup: func() {
				s.HSet("hash1", []string{"f1", "5"})
			},
			expectedStatus: types.StatusInteger,
			expectedData:   2,
		},
		{
			name: "zero increment",
			args: []string{"hash1", "f1", "0"},
			setup: func() {
				s.HSet("hash1", []string{"f1", "5"})
			},
			expectedStatus: types.StatusInteger,
			expectedData:   5,
		},
		{
			name: "wrong type",
			args: []string{"string_key", "f1", "1"},
			setup: func() {
				s.Set("string_key", types.Value{Type: types.StringType, Data: "val"})
			},
			expectedStatus: types.StatusError,
			expectedMsg:    "WRONGTYPE Operation against a key holding the wrong kind of value",
		},
		{
			name:           "invalid increment arg",
			args:           []string{"hash1", "f1", "not-int"},
			expectedStatus: types.StatusError,
			expectedMsg:    "ERR value is not an integer or out of range",
		},
		{
			name: "invalid hash value",
			args: []string{"hash1", "f1", "1"},
			setup: func() {
				s.HSet("hash1", []string{"f1", "not-int"})
			},
			expectedStatus: types.StatusError,
			expectedMsg:    "ERR hash value is not an integer",
		},
		{
			name: "overflow",
			args: []string{"hash1", "f1", "10"},
			setup: func() {
				s.HSet("hash1", []string{"f1", strconv.FormatInt(math.MaxInt64-5, 10)})
			},
			expectedStatus: types.StatusError,
			expectedMsg:    "ERR increment or decrement would overflow",
		},
		{
			name:           "invalid argument count",
			args:           []string{"hash1", "f1"},
			expectedStatus: types.StatusError,
			expectedMsg:    "wrong number of arguments for 'hincrby' command",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s = store.NewMemoryStore()
			h = NewHIncrByHandler(s)
			if tt.setup != nil {
				tt.setup()
			}

			cmd := types.Command{
				Name: "HINCRBY",
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

func TestHIncrByHandler_DispatcherIntegration(t *testing.T) {
	s := store.NewMemoryStore()
	d := NewDispatcher(s)

	cmd := types.Command{
		Name: "HINCRBY",
		Args: []string{"hash1", "f1", "10"},
	}

	resp := d.Dispatch(cmd)
	if resp.Status != types.StatusInteger {
		t.Fatalf("expected StatusInteger, got %v", resp.Status)
	}

	val, ok := resp.Data.(int)
	if !ok {
		t.Fatalf("expected int data, got %T", resp.Data)
	}
	if val != 10 {
		t.Errorf("expected 10, got %d", val)
	}
}
