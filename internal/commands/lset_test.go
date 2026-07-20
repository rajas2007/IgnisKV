package commands

import (
	"os"
	"testing"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

func TestLSetHandler(t *testing.T) {
	_ = os.Remove(store.DefaultSnapshotFile)
	defer os.Remove(store.DefaultSnapshotFile)

	s := store.NewMemoryStore()
	handler := NewLSetHandler(s)
	dispatcher := NewDispatcher(s)
	// Register it for the dispatcher test (case 12)
	dispatcher.handlers["LSET"] = handler

	// 1. Wrong number of arguments (0)
	t.Run("wrong_number_of_arguments_(0)", func(t *testing.T) {
		res := handler.Execute(types.Command{
			Name: "LSET",
			Args: []string{},
		})
		if res.Status != types.StatusError {
			t.Fatalf("expected StatusError, got %v", res.Status)
		}
		if res.Message != "wrong number of arguments" {
			t.Fatalf("expected 'wrong number of arguments', got %q", res.Message)
		}
	})

	// 2. Wrong number of arguments (2)
	t.Run("wrong_number_of_arguments_(2)", func(t *testing.T) {
		res := handler.Execute(types.Command{
			Name: "LSET",
			Args: []string{"list", "0"},
		})
		if res.Status != types.StatusError {
			t.Fatalf("expected StatusError, got %v", res.Status)
		}
		if res.Message != "wrong number of arguments" {
			t.Fatalf("expected 'wrong number of arguments', got %q", res.Message)
		}
	})

	// 3. Wrong number of arguments (4)
	t.Run("wrong_number_of_arguments_(4)", func(t *testing.T) {
		res := handler.Execute(types.Command{
			Name: "LSET",
			Args: []string{"list", "0", "x", "extra"},
		})
		if res.Status != types.StatusError {
			t.Fatalf("expected StatusError, got %v", res.Status)
		}
		if res.Message != "wrong number of arguments" {
			t.Fatalf("expected 'wrong number of arguments', got %q", res.Message)
		}
	})

	// 4. Invalid index
	t.Run("invalid_index", func(t *testing.T) {
		res := handler.Execute(types.Command{
			Name: "LSET",
			Args: []string{"list", "abc", "x"},
		})
		if res.Status != types.StatusError {
			t.Fatalf("expected StatusError, got %v", res.Status)
		}
		if res.Message != "value is not an integer or out of range" {
			t.Fatalf("expected 'value is not an integer or out of range', got %q", res.Message)
		}
	})

	// 5. Missing key
	t.Run("missing_key", func(t *testing.T) {
		res := handler.Execute(types.Command{
			Name: "LSET",
			Args: []string{"missing", "0", "x"},
		})
		if res.Status != types.StatusError {
			t.Fatalf("expected StatusError, got %v", res.Status)
		}
		if res.Message != store.ErrKeyNotFound.Error() {
			t.Fatalf("expected ErrKeyNotFound, got %q", res.Message)
		}
	})

	// 6. Existing list (first element)
	t.Run("existing_list_(first_element)", func(t *testing.T) {
		if _, err := s.RPush("list", "a", "b", "c", "d", "e"); err != nil {
			t.Fatalf("setup failed: %v", err)
		}
		res := handler.Execute(types.Command{
			Name: "LSET",
			Args: []string{"list", "0", "x"},
		})
		if res.Status != types.StatusOK {
			t.Fatalf("expected StatusOK, got %v: %v", res.Status, res.Message)
		}
		val, _ := s.Get("list")
		list := val.Data.([]string)
		if len(list) != 5 || list[0] != "x" || list[1] != "b" || list[2] != "c" || list[3] != "d" || list[4] != "e" {
			t.Fatalf("expected list [x b c d e], got %v", list)
		}
	})

	// 7. Existing list (negative index)
	t.Run("existing_list_(negative_index)", func(t *testing.T) {
		res := handler.Execute(types.Command{
			Name: "LSET",
			Args: []string{"list", "-1", "last"},
		})
		if res.Status != types.StatusOK {
			t.Fatalf("expected StatusOK, got %v: %v", res.Status, res.Message)
		}
		val, _ := s.Get("list")
		list := val.Data.([]string)
		if len(list) != 5 || list[0] != "x" || list[1] != "b" || list[2] != "c" || list[3] != "d" || list[4] != "last" {
			t.Fatalf("expected list [x b c d last], got %v", list)
		}
	})

	// 8. Out-of-range
	t.Run("out-of-range", func(t *testing.T) {
		res := handler.Execute(types.Command{
			Name: "LSET",
			Args: []string{"list", "100", "x"},
		})
		if res.Status != types.StatusError {
			t.Fatalf("expected StatusError, got %v", res.Status)
		}
		if res.Message != store.ErrIndexOutOfRange.Error() {
			t.Fatalf("expected ErrIndexOutOfRange, got %q", res.Message)
		}
		val, _ := s.Get("list")
		list := val.Data.([]string)
		if len(list) != 5 || list[0] != "x" || list[1] != "b" || list[2] != "c" || list[3] != "d" || list[4] != "last" {
			t.Fatalf("expected list to remain unchanged, got %v", list)
		}
	})

	// 9. WRONGTYPE
	t.Run("WRONGTYPE", func(t *testing.T) {
		s.Set("stringkey", types.Value{
			Type: types.StringType,
			Data: "val",
		})
		res := handler.Execute(types.Command{
			Name: "LSET",
			Args: []string{"stringkey", "0", "x"},
		})
		if res.Status != types.StatusError {
			t.Fatalf("expected StatusError, got %v", res.Status)
		}
		if res.Message != store.ErrWrongType.Error() {
			t.Fatalf("expected ErrWrongType, got %q", res.Message)
		}
	})

	// 10. Persistence after success
	t.Run("persistence_after_success", func(t *testing.T) {
		_ = os.Remove(store.DefaultSnapshotFile)
		res := handler.Execute(types.Command{
			Name: "LSET",
			Args: []string{"list", "0", "x"},
		})
		if res.Status != types.StatusOK {
			t.Fatalf("expected StatusOK, got %v: %v", res.Status, res.Message)
		}
		if _, err := os.Stat(store.DefaultSnapshotFile); os.IsNotExist(err) {
			t.Fatal("expected snapshot file to exist after successful mutation")
		}
	})

	// 11. No persistence after failure
	t.Run("no_persistence_after_failure", func(t *testing.T) {
		_ = os.Remove(store.DefaultSnapshotFile)
		res := handler.Execute(types.Command{
			Name: "LSET",
			Args: []string{"missing", "0", "x"},
		})
		if res.Status != types.StatusError {
			t.Fatalf("expected StatusError, got %v", res.Status)
		}
		if _, err := os.Stat(store.DefaultSnapshotFile); !os.IsNotExist(err) {
			t.Fatal("expected snapshot file to NOT exist after failed mutation")
		}
	})

	// 12. Dispatcher routing
	t.Run("dispatcher_routing", func(t *testing.T) {
		res := dispatcher.Dispatch(types.Command{
			Name: "LSET",
			Args: []string{"list", "1", "updated"},
		})
		if res.Status != types.StatusOK {
			t.Fatalf("expected StatusOK, got %v: %v", res.Status, res.Message)
		}
		val, _ := s.Get("list")
		list := val.Data.([]string)
		if len(list) != 5 || list[0] != "x" || list[1] != "updated" || list[2] != "c" || list[3] != "d" || list[4] != "last" {
			t.Fatalf("expected list [x updated c d last], got %v", list)
		}
	})
}
