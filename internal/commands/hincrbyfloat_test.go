package commands

import (
	"testing"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

func TestHIncrByFloatHandler_Execute(t *testing.T) {
	s := store.NewMemoryStore()
	h := NewHIncrByFloatHandler(s)

	tests := []struct {
		name           string
		args           []string
		setup          func()
		expectedStatus types.ResponseStatus
		expectedMsg    string
		expectedData   string
	}{
		{
			name:           "missing key",
			args:           []string{"hash1", "f1", "5.5"},
			expectedStatus: types.StatusString,
			expectedData:   "5.5",
		},
		{
			name: "existing field",
			args: []string{"hash1", "f1", "10.2"},
			setup: func() {
				s.HSet("hash1", []string{"f1", "5.5"})
			},
			expectedStatus: types.StatusString,
			expectedData:   "15.7",
		},
		{
			name: "existing field negative",
			args: []string{"hash1", "f1", "-3.5"},
			setup: func() {
				s.HSet("hash1", []string{"f1", "5.5"})
			},
			expectedStatus: types.StatusString,
			expectedData:   "2",
		},
		{
			name: "zero increment",
			args: []string{"hash1", "f1", "0"},
			setup: func() {
				s.HSet("hash1", []string{"f1", "5.5"})
			},
			expectedStatus: types.StatusString,
			expectedData:   "5.5",
		},
		{
			name: "wrong type",
			args: []string{"string_key", "f1", "1.1"},
			setup: func() {
				s.Set("string_key", types.Value{Type: types.StringType, Data: "val"})
			},
			expectedStatus: types.StatusError,
			expectedMsg:    "WRONGTYPE Operation against a key holding the wrong kind of value",
		},
		{
			name:           "invalid increment arg",
			args:           []string{"hash1", "f1", "not-float"},
			expectedStatus: types.StatusError,
			expectedMsg:    "ERR value is not a valid float",
		},
		{
			name: "invalid hash value",
			args: []string{"hash1", "f1", "1.1"},
			setup: func() {
				s.HSet("hash1", []string{"f1", "not-float"})
			},
			expectedStatus: types.StatusError,
			expectedMsg:    "ERR hash value is not a float",
		},
		{
			name:           "NaN",
			args:           []string{"hash1", "f1", "NaN"},
			expectedStatus: types.StatusError,
			expectedMsg:    "ERR value is not a valid float", // strconv.ParseFloat parses "NaN", but let's see. Wait, "NaN" parsing might succeed and return math.NaN(). If so, it fails inside HIncrByFloat with ErrNaNOrInfinity. Let's let the test pass if it's either, or we can just expect ErrNaNOrInfinity if we test the inner store condition.
		},
		{
			name: "Infinity",
			args: []string{"hash1", "f1", "1e308"},
			setup: func() {
				s.HSet("hash1", []string{"f1", "1e308"})
			},
			expectedStatus: types.StatusError,
			expectedMsg:    "ERR increment would produce NaN or Infinity",
		},
		{
			name:           "invalid argument count",
			args:           []string{"hash1", "f1"},
			expectedStatus: types.StatusError,
			expectedMsg:    "wrong number of arguments for 'hincrbyfloat' command",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s = store.NewMemoryStore()
			h = NewHIncrByFloatHandler(s)
			if tt.setup != nil {
				tt.setup()
			}

			cmd := types.Command{
				Name: "HINCRBYFLOAT",
				Args: tt.args,
			}

			resp := h.Execute(cmd)

			if resp.Status != tt.expectedStatus {
				t.Errorf("expected status %v, got %v", tt.expectedStatus, resp.Status)
			}

			// For NaN, ParseFloat handles it, so we'll get ErrNaNOrInfinity from the store instead of arg parsing.
			if tt.name == "NaN" && resp.Message != "ERR increment would produce NaN or Infinity" {
				t.Errorf("expected NaN error message, got %q", resp.Message)
			} else if tt.name != "NaN" && tt.expectedMsg != "" && resp.Message != tt.expectedMsg {
				t.Errorf("expected message %q, got %q", tt.expectedMsg, resp.Message)
			}

			if tt.expectedStatus == types.StatusString {
				val, ok := resp.Data.(string)
				if !ok {
					t.Fatalf("expected string data, got %T", resp.Data)
				}
				if val != tt.expectedData {
					t.Errorf("expected %q, got %q", tt.expectedData, val)
				}
			}
		})
	}
}

func TestHIncrByFloatHandler_DispatcherIntegration(t *testing.T) {
	s := store.NewMemoryStore()
	d := NewDispatcher(s)

	cmd := types.Command{
		Name: "HINCRBYFLOAT",
		Args: []string{"hash1", "f1", "10.5"},
	}

	resp := d.Dispatch(cmd)
	if resp.Status != types.StatusString {
		t.Fatalf("expected StatusBulkString, got %v", resp.Status)
	}

	val, ok := resp.Data.(string)
	if !ok {
		t.Fatalf("expected string data, got %T", resp.Data)
	}
	if val != "10.5" {
		t.Errorf("expected 10.5, got %s", val)
	}
}
