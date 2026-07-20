package commands

import (
	"errors"
	"os"
	"testing"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

func TestLIndexHandler(t *testing.T) {
	s := store.NewMemoryStore()
	h := NewLIndexHandler(s)

	// Setup: [a b c d e]
	if _, err := s.RPush("list", "a", "b", "c", "d", "e"); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	s.Set("stringkey", types.Value{
		Type: types.StringType,
		Data: "val",
	})

	t.Run("wrong number of arguments (0)", func(t *testing.T) {
		res := h.Execute(types.Command{Args: []string{}})
		if res.Status != types.StatusError {
			t.Fatalf("expected StatusError, got %v", res.Status)
		}
		if res.Message != "wrong number of arguments" {
			t.Fatalf("expected 'wrong number of arguments', got %q", res.Message)
		}
	})

	t.Run("wrong number of arguments (1)", func(t *testing.T) {
		res := h.Execute(types.Command{Args: []string{"list"}})
		if res.Status != types.StatusError {
			t.Fatalf("expected StatusError, got %v", res.Status)
		}
		if res.Message != "wrong number of arguments" {
			t.Fatalf("expected 'wrong number of arguments', got %q", res.Message)
		}
	})

	t.Run("wrong number of arguments (3)", func(t *testing.T) {
		res := h.Execute(types.Command{Args: []string{"list", "0", "extra"}})
		if res.Status != types.StatusError {
			t.Fatalf("expected StatusError, got %v", res.Status)
		}
		if res.Message != "wrong number of arguments" {
			t.Fatalf("expected 'wrong number of arguments', got %q", res.Message)
		}
	})

	t.Run("invalid index", func(t *testing.T) {
		res := h.Execute(types.Command{Args: []string{"list", "abc"}})
		if res.Status != types.StatusError {
			t.Fatalf("expected StatusError, got %v", res.Status)
		}
		if res.Message != "value is not an integer or out of range" {
			t.Fatalf("expected 'value is not an integer or out of range', got %q", res.Message)
		}
	})

	t.Run("missing key", func(t *testing.T) {
		res := h.Execute(types.Command{Args: []string{"missing", "0"}})
		if res.Status != types.StatusNil {
			t.Fatalf("expected StatusNil, got %v", res.Status)
		}
	})

	t.Run("existing list (first element)", func(t *testing.T) {
		res := h.Execute(types.Command{Args: []string{"list", "0"}})
		if res.Status != types.StatusString {
			t.Fatalf("expected StatusString, got %v", res.Status)
		}
		if res.Data != "a" {
			t.Fatalf("expected 'a', got %v", res.Data)
		}
	})

	t.Run("existing list (middle element)", func(t *testing.T) {
		res := h.Execute(types.Command{Args: []string{"list", "2"}})
		if res.Status != types.StatusString {
			t.Fatalf("expected StatusString, got %v", res.Status)
		}
		if res.Data != "c" {
			t.Fatalf("expected 'c', got %v", res.Data)
		}
	})

	t.Run("negative index", func(t *testing.T) {
		res := h.Execute(types.Command{Args: []string{"list", "-1"}})
		if res.Status != types.StatusString {
			t.Fatalf("expected StatusString, got %v", res.Status)
		}
		if res.Data != "e" {
			t.Fatalf("expected 'e', got %v", res.Data)
		}
	})

	t.Run("out-of-range", func(t *testing.T) {
		res := h.Execute(types.Command{Args: []string{"list", "100"}})
		if res.Status != types.StatusNil {
			t.Fatalf("expected StatusNil, got %v", res.Status)
		}
	})

	t.Run("WRONGTYPE", func(t *testing.T) {
		res := h.Execute(types.Command{Args: []string{"stringkey", "0"}})
		if res.Status != types.StatusError {
			t.Fatalf("expected StatusError, got %v", res.Status)
		}
		if res.Message != store.ErrWrongType.Error() {
			t.Fatalf("expected %q, got %q", store.ErrWrongType.Error(), res.Message)
		}
	})

	t.Run("read-only guarantee", func(t *testing.T) {
		res := h.Execute(types.Command{Args: []string{"list", "1"}})
		if res.Status != types.StatusString {
			t.Fatalf("expected StatusString, got %v", res.Status)
		}
		if res.Data != "b" {
			t.Fatalf("expected 'b', got %v", res.Data)
		}

		val, _ := s.Get("list")
		list := val.Data.([]string)
		if len(list) != 5 || list[0] != "a" || list[1] != "b" || list[2] != "c" || list[3] != "d" || list[4] != "e" {
			t.Fatalf("expected list [a b c d e], got %v", list)
		}
	})
}

func TestLIndexDispatcherRouting(t *testing.T) {
	s := store.NewMemoryStore()
	if _, err := s.RPush("list", "a", "b", "c", "d", "e"); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	d := NewDispatcher(s)

	res := d.Dispatch(types.Command{Name: "LINDEX", Args: []string{"list", "0"}})
	if res.Status != types.StatusString {
		t.Fatalf("expected StatusString, got %v", res.Status)
	}
	if res.Data != "a" {
		t.Fatalf("expected 'a', got %v", res.Data)
	}
}

func TestLIndexNoPersistence(t *testing.T) {
	_ = os.Remove(store.DefaultSnapshotFile)
	defer os.Remove(store.DefaultSnapshotFile)

	s := store.NewMemoryStore()
	h := NewLIndexHandler(s)

	if _, err := s.RPush("list", "a"); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// Ensure snapshot file doesn't exist before LINDEX
	_ = os.Remove(store.DefaultSnapshotFile)

	resp := h.Execute(types.Command{
		Name: "LINDEX",
		Args: []string{"list", "0"},
	})

	if resp.Status != types.StatusString {
		t.Fatalf("expected StatusString, got %v", resp.Status)
	}

	_, err := os.Stat(store.DefaultSnapshotFile)
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("LINDEX triggered persistence, snapshot file should not exist")
	}
}
