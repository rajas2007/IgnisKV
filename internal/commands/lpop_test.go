package commands

import (
	"errors"
	"os"
	"testing"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

func TestLPopHandler(t *testing.T) {
	s := store.NewMemoryStore()
	h := NewLPopHandler(s)

	t.Run("wrong number of arguments (0)", func(t *testing.T) {
		res := h.Execute(types.Command{Args: []string{}})
		if res.Status != types.StatusError {
			t.Fatalf("expected StatusError, got %v", res.Status)
		}
		if res.Message != "wrong number of arguments" {
			t.Fatalf("expected 'wrong number of arguments', got %q", res.Message)
		}
	})

	t.Run("wrong number of arguments (2)", func(t *testing.T) {
		res := h.Execute(types.Command{Args: []string{"list", "extra"}})
		if res.Status != types.StatusError {
			t.Fatalf("expected StatusError, got %v", res.Status)
		}
		if res.Message != "wrong number of arguments" {
			t.Fatalf("expected 'wrong number of arguments', got %q", res.Message)
		}
	})

	t.Run("missing key", func(t *testing.T) {
		res := h.Execute(types.Command{Args: []string{"missing"}})
		if res.Status != types.StatusNil {
			t.Fatalf("expected StatusNil, got %v", res.Status)
		}
	})

	t.Run("existing multi-element list", func(t *testing.T) {
		if _, err := s.RPush("list1", "a", "b", "c"); err != nil {
			t.Fatalf("setup failed: %v", err)
		}
		res := h.Execute(types.Command{Args: []string{"list1"}})
		if res.Status != types.StatusString {
			t.Fatalf("expected StatusString, got %v", res.Status)
		}
		if res.Data != "a" {
			t.Fatalf("expected 'a', got %v", res.Data)
		}

		val, _ := s.Get("list1")
		list := val.Data.([]string)
		if len(list) != 2 || list[0] != "b" || list[1] != "c" {
			t.Fatalf("expected [b c], got %v", list)
		}
	})

	t.Run("single-element list", func(t *testing.T) {
		if _, err := s.RPush("single", "a"); err != nil {
			t.Fatalf("setup failed: %v", err)
		}
		res := h.Execute(types.Command{Args: []string{"single"}})
		if res.Status != types.StatusString {
			t.Fatalf("expected StatusString, got %v", res.Status)
		}
		if res.Data != "a" {
			t.Fatalf("expected 'a', got %v", res.Data)
		}

		_, err := s.Get("single")
		if !errors.Is(err, store.ErrKeyNotFound) {
			t.Fatalf("expected ErrKeyNotFound, got %v", err)
		}
	})

	t.Run("WRONGTYPE", func(t *testing.T) {
		s.Set("stringkey", types.Value{
			Type: types.StringType,
			Data: "val",
		})
		res := h.Execute(types.Command{Args: []string{"stringkey"}})
		if res.Status != types.StatusError {
			t.Fatalf("expected StatusError, got %v", res.Status)
		}
		if res.Message != store.ErrWrongType.Error() {
			t.Fatalf("expected %q, got %q", store.ErrWrongType.Error(), res.Message)
		}
	})

	t.Run("persistence after success", func(t *testing.T) {
		os.Remove(store.DefaultSnapshotFile)
		if _, err := s.RPush("persist", "a"); err != nil {
			t.Fatalf("setup failed: %v", err)
		}
		res := h.Execute(types.Command{Args: []string{"persist"}})
		if res.Status != types.StatusString {
			t.Fatalf("expected StatusString, got %v", res.Status)
		}

		if _, err := os.Stat(store.DefaultSnapshotFile); os.IsNotExist(err) {
			t.Fatalf("expected snapshot file to exist")
		}
		os.Remove(store.DefaultSnapshotFile)
	})

	t.Run("no persistence after missing key", func(t *testing.T) {
		os.Remove(store.DefaultSnapshotFile)
		res := h.Execute(types.Command{Args: []string{"missing"}})
		if res.Status != types.StatusNil {
			t.Fatalf("expected StatusNil, got %v", res.Status)
		}

		if _, err := os.Stat(store.DefaultSnapshotFile); !os.IsNotExist(err) {
			t.Fatalf("expected snapshot file to not exist")
		}
	})

	t.Run("no persistence after WRONGTYPE", func(t *testing.T) {
		os.Remove(store.DefaultSnapshotFile)
		s.Set("wrong", types.Value{
			Type: types.StringType,
			Data: "val",
		})
		os.Remove(store.DefaultSnapshotFile) // Remove it after SET created it

		res := h.Execute(types.Command{Args: []string{"wrong"}})
		if res.Status != types.StatusError {
			t.Fatalf("expected StatusError, got %v", res.Status)
		}

		if _, err := os.Stat(store.DefaultSnapshotFile); !os.IsNotExist(err) {
			t.Fatalf("expected snapshot file to not exist")
		}
	})
}

func TestLPopDispatcherRouting(t *testing.T) {
	s := store.NewMemoryStore()
	if _, err := s.RPush("list", "popped_val"); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	d := NewDispatcher(s)

	res := d.Dispatch(types.Command{Name: "LPOP", Args: []string{"list"}})
	if res.Status != types.StatusString {
		t.Fatalf("expected StatusString, got %v", res.Status)
	}
	if res.Data != "popped_val" {
		t.Fatalf("expected 'popped_val', got %v", res.Data)
	}
}
