package commands

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

func TestLLenHandler(t *testing.T) {
	s := store.NewMemoryStore()
	h := NewLLenHandler(s)

	tests := []struct {
		name     string
		setup    func()
		cmd      types.Command
		expected types.Response
	}{
		{
			name: "Wrong number of arguments (0)",
			cmd: types.Command{
				Name: "LLEN",
				Args: []string{},
			},
			expected: types.Response{
				Status:  types.StatusError,
				Message: "wrong number of arguments",
			},
		},
		{
			name: "Wrong number of arguments (2)",
			cmd: types.Command{
				Name: "LLEN",
				Args: []string{"list", "extra"},
			},
			expected: types.Response{
				Status:  types.StatusError,
				Message: "wrong number of arguments",
			},
		},
		{
			name: "Missing key",
			cmd: types.Command{
				Name: "LLEN",
				Args: []string{"missing"},
			},
			expected: types.Response{
				Status: types.StatusInteger,
				Data:   "0",
			},
		},
		{
			name: "Existing single-element list",
			setup: func() {
				_, _ = s.LPush("single", "a")
			},
			cmd: types.Command{
				Name: "LLEN",
				Args: []string{"single"},
			},
			expected: types.Response{
				Status: types.StatusInteger,
				Data:   "1",
			},
		},
		{
			name: "Existing multi-element list",
			setup: func() {
				_, _ = s.RPush("multi", "a", "b", "c")
			},
			cmd: types.Command{
				Name: "LLEN",
				Args: []string{"multi"},
			},
			expected: types.Response{
				Status: types.StatusInteger,
				Data:   "3",
			},
		},
		{
			name: "WRONGTYPE",
			setup: func() {
				s.Set("stringkey", types.Value{
					Type: types.StringType,
					Data: "val",
				})
			},
			cmd: types.Command{
				Name: "LLEN",
				Args: []string{"stringkey"},
			},
			expected: types.Response{
				Status:  types.StatusError,
				Message: store.ErrWrongType.Error(),
			},
		},
		{
			name: "Expired key",
			setup: func() {
				s.Set("expired", types.Value{
					Type:      types.ListType,
					Data:      []string{"a", "b"},
					ExpiresAt: time.Now().Add(-1 * time.Second),
				})
			},
			cmd: types.Command{
				Name: "LLEN",
				Args: []string{"expired"},
			},
			expected: types.Response{
				Status: types.StatusInteger,
				Data:   "0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			resp := h.Execute(tt.cmd)
			if resp.Status != tt.expected.Status {
				t.Fatalf("expected status %v, got %v", tt.expected.Status, resp.Status)
			}
			if resp.Message != tt.expected.Message {
				t.Fatalf("expected message %q, got %q", tt.expected.Message, resp.Message)
			}
			if resp.Data != tt.expected.Data {
				t.Fatalf("expected data %q, got %q", tt.expected.Data, resp.Data)
			}
		})
	}
}

func TestLLenDispatcherRouting(t *testing.T) {
	s := store.NewMemoryStore()
	d := NewDispatcher(s)

	_, _ = s.RPush("list", "a", "b")

	resp := d.Dispatch(types.Command{
		Name: "LLEN",
		Args: []string{"list"},
	})

	if resp.Status != types.StatusInteger {
		t.Fatalf("expected StatusInteger, got %v", resp.Status)
	}
	if resp.Data != "2" {
		t.Fatalf("expected length 2, got %q", resp.Data)
	}
}

func TestLLenNoPersistence(t *testing.T) {
	_ = os.Remove(store.DefaultSnapshotFile)
	defer os.Remove(store.DefaultSnapshotFile)

	s := store.NewMemoryStore()
	h := NewLLenHandler(s)

	_, _ = s.RPush("list", "a")

	// Ensure snapshot file doesn't exist before LLEN
	_ = os.Remove(store.DefaultSnapshotFile)

	resp := h.Execute(types.Command{
		Name: "LLEN",
		Args: []string{"list"},
	})

	if resp.Status != types.StatusInteger {
		t.Fatalf("expected StatusInteger, got %v", resp.Status)
	}

	_, err := os.Stat(store.DefaultSnapshotFile)
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("LLEN triggered persistence, snapshot file should not exist")
	}
}
