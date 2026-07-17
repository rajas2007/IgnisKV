package commands

import (
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

func TestLRangeHandler(t *testing.T) {
	s := store.NewMemoryStore()
	handler := NewLRangeHandler(s)

	// Helper to set up lists
	s.Set("list", types.Value{
		Type: types.ListType,
		Data: []string{"a", "b", "c", "d", "e"},
	})

	s.Set("stringkey", types.Value{
		Type: types.StringType,
		Data: "val",
	})

	tests := []struct {
		name     string
		args     []string
		expected types.Response
	}{
		{
			name: "wrong number of arguments (0)",
			args: []string{},
			expected: types.Response{
				Status:  types.StatusError,
				Message: "wrong number of arguments",
			},
		},
		{
			name: "wrong number of arguments (1)",
			args: []string{"list"},
			expected: types.Response{
				Status:  types.StatusError,
				Message: "wrong number of arguments",
			},
		},
		{
			name: "wrong number of arguments (2)",
			args: []string{"list", "0"},
			expected: types.Response{
				Status:  types.StatusError,
				Message: "wrong number of arguments",
			},
		},
		{
			name: "wrong number of arguments (4)",
			args: []string{"list", "0", "-1", "extra"},
			expected: types.Response{
				Status:  types.StatusError,
				Message: "wrong number of arguments",
			},
		},
		{
			name: "invalid start",
			args: []string{"list", "abc", "-1"},
			expected: types.Response{
				Status:  types.StatusError,
				Message: "value is not an integer or out of range",
			},
		},
		{
			name: "invalid stop",
			args: []string{"list", "0", "xyz"},
			expected: types.Response{
				Status:  types.StatusError,
				Message: "value is not an integer or out of range",
			},
		},
		{
			name: "missing key",
			args: []string{"missing", "0", "-1"},
			expected: types.Response{
				Status: types.StatusArray,
				Data:   []string{},
			},
		},
		{
			name: "existing list",
			args: []string{"list", "1", "3"},
			expected: types.Response{
				Status: types.StatusArray,
				Data:   []string{"b", "c", "d"},
			},
		},
		{
			name: "negative indices",
			args: []string{"list", "-2", "-1"},
			expected: types.Response{
				Status: types.StatusArray,
				Data:   []string{"d", "e"},
			},
		},
		{
			name: "wrong type",
			args: []string{"stringkey", "0", "-1"},
			expected: types.Response{
				Status:  types.StatusError,
				Message: store.ErrWrongType.Error(),
			},
		},
		{
			name: "normalized empty range",
			args: []string{"list", "100", "200"},
			expected: types.Response{
				Status: types.StatusArray,
				Data:   []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := types.Command{
				Name: "LRANGE",
				Args: tt.args,
			}
			res := handler.Execute(cmd)

			if res.Status != tt.expected.Status {
				t.Fatalf("expected status %v, got %v", tt.expected.Status, res.Status)
			}

			switch res.Status {
			case types.StatusError:
				if res.Message != tt.expected.Message {
					t.Fatalf("expected message %q, got %q", tt.expected.Message, res.Message)
				}
			case types.StatusArray:
				if !reflect.DeepEqual(res.Data, tt.expected.Data) {
					t.Fatalf("expected data %v, got %v", tt.expected.Data, res.Data)
				}
			default:
				t.Fatalf("unexpected status %v", res.Status)
			}
		})
	}
}

func TestLRangeDispatcherRouting(t *testing.T) {
	s := store.NewMemoryStore()
	s.Set("list", types.Value{
		Type: types.ListType,
		Data: []string{"a", "b", "c", "d", "e"},
	})

	dispatcher := NewDispatcher(s)
	cmd := types.Command{
		Name: "LRANGE",
		Args: []string{"list", "0", "-1"},
	}
	res := dispatcher.Dispatch(cmd)
	if res.Status != types.StatusArray {
		t.Fatalf("expected dispatch to return array status, got %v", res.Status)
	}
	expectedData := []string{"a", "b", "c", "d", "e"}
	if !reflect.DeepEqual(res.Data, expectedData) {
		t.Fatalf("expected data %v, got %v", expectedData, res.Data)
	}
}

func TestLRangeNoPersistence(t *testing.T) {
	_ = os.Remove(store.DefaultSnapshotFile)
	defer os.Remove(store.DefaultSnapshotFile)

	s := store.NewMemoryStore()
	h := NewLRangeHandler(s)

	_, _ = s.RPush("list", "a")

	// Ensure snapshot file doesn't exist before LRANGE
	_ = os.Remove(store.DefaultSnapshotFile)

	resp := h.Execute(types.Command{
		Name: "LRANGE",
		Args: []string{"list", "0", "-1"},
	})

	if resp.Status != types.StatusArray {
		t.Fatalf("expected StatusArray, got %v", resp.Status)
	}

	_, err := os.Stat(store.DefaultSnapshotFile)
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("LRANGE triggered persistence, snapshot file should not exist")
	}
}
