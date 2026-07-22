package commands

import (
	"testing"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

func TestSRemHandler(t *testing.T) {
	s := store.NewMemoryStore()
	handler := NewSRemHandler(s)

	tests := []struct {
		name           string
		args           []string
		setup          func()
		expectedStatus types.ResponseStatus
		expectedData   interface{}
		expectedError  string
	}{
		{
			name:           "missing arguments",
			args:           []string{"key1"}, // needs key and at least 1 member
			setup:          func() {},
			expectedStatus: types.StatusError,
			expectedError:  "wrong number of arguments for 'srem' command",
		},
		{
			name:           "remove from missing key",
			args:           []string{"set1", "m1"},
			setup:          func() {},
			expectedStatus: types.StatusInteger,
			expectedData:   0,
		},
		{
			name: "remove single existing member",
			args: []string{"set1", "m1"},
			setup: func() {
				s.SAdd("set1", []string{"m1", "m2"})
			},
			expectedStatus: types.StatusInteger,
			expectedData:   1,
		},
		{
			name: "remove multiple members (mixed)",
			args: []string{"set1", "m1", "m3", "m2"},
			setup: func() {
				s.SAdd("set1", []string{"m1", "m2"})
			},
			expectedStatus: types.StatusInteger,
			expectedData:   2, // m1 and m2
		},
		{
			name: "wrong type",
			args: []string{"string_key", "m1"},
			setup: func() {
				s.Set("string_key", types.Value{Type: types.StringType, Data: "val"})
			},
			expectedStatus: types.StatusError,
			expectedError:  store.ErrWrongType.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s.Delete("set1")
			s.Delete("string_key")
			tt.setup()

			cmd := types.Command{
				Name: "srem",
				Args: tt.args,
			}

			response := handler.Execute(cmd)

			if response.Status != tt.expectedStatus {
				t.Fatalf("expected status %d, got %d (message: %s)", tt.expectedStatus, response.Status, response.Message)
			}

			if tt.expectedStatus == types.StatusError {
				if response.Message != tt.expectedError {
					t.Fatalf("expected error '%s', got '%s'", tt.expectedError, response.Message)
				}
			} else {
				if response.Data != tt.expectedData {
					t.Fatalf("expected data %v, got %v", tt.expectedData, response.Data)
				}
			}
		})
	}
}
