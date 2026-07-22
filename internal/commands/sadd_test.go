package commands

import (
	"testing"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

func TestSAddHandler(t *testing.T) {
	s := store.NewMemoryStore()
	handler := NewSAddHandler(s)

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
			expectedError:  "wrong number of arguments for 'sadd' command",
		},
		{
			name:           "add single member",
			args:           []string{"set1", "m1"},
			setup:          func() {},
			expectedStatus: types.StatusInteger,
			expectedData:   1, // 1 inserted
		},
		{
			name:           "add multiple members with duplicates",
			args:           []string{"set2", "m1", "m2", "m1", "m3"},
			setup:          func() {},
			expectedStatus: types.StatusInteger,
			expectedData:   3, // m1, m2, m3
		},
		{
			name: "add to existing set",
			args: []string{"set3", "m1", "m2"},
			setup: func() {
				s.SAdd("set3", []string{"m1"})
			},
			expectedStatus: types.StatusInteger,
			expectedData:   1, // only m2 is new
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
			s.Delete("set2")
			s.Delete("set3")
			s.Delete("string_key")
			tt.setup()

			cmd := types.Command{
				Name: "sadd",
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
