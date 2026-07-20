package commands

import (
	"os"
	"testing"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

func TestLInsertHandler(t *testing.T) {
	_ = os.Remove(store.DefaultSnapshotFile)
	defer os.Remove(store.DefaultSnapshotFile)

	s := store.NewMemoryStore()
	handler := NewLInsertHandler(s)

	// 1. Wrong number of arguments (0)
	t.Run("wrong_number_of_arguments_(0)", func(t *testing.T) {
		res := handler.Execute(types.Command{
			Name: "LINSERT",
			Args: []string{},
		})
		if res.Status != types.StatusError {
			t.Fatalf("expected StatusError, got %v", res.Status)
		}
		if res.Message != "wrong number of arguments" {
			t.Fatalf("expected 'wrong number of arguments', got %q", res.Message)
		}
	})

	// 2. Wrong number of arguments (3)
	t.Run("wrong_number_of_arguments_(3)", func(t *testing.T) {
		res := handler.Execute(types.Command{
			Name: "LINSERT",
			Args: []string{"BEFORE", "list", "pivot"},
		})
		if res.Status != types.StatusError {
			t.Fatalf("expected StatusError, got %v", res.Status)
		}
		if res.Message != "wrong number of arguments" {
			t.Fatalf("expected 'wrong number of arguments', got %q", res.Message)
		}
	})

	// 3. Wrong number of arguments (5)
	t.Run("wrong_number_of_arguments_(5)", func(t *testing.T) {
		res := handler.Execute(types.Command{
			Name: "LINSERT",
			Args: []string{"BEFORE", "list", "pivot", "value", "extra"},
		})
		if res.Status != types.StatusError {
			t.Fatalf("expected StatusError, got %v", res.Status)
		}
		if res.Message != "wrong number of arguments" {
			t.Fatalf("expected 'wrong number of arguments', got %q", res.Message)
		}
	})

	// 4. Invalid position keyword
	t.Run("invalid_position_keyword", func(t *testing.T) {
		_ = os.Remove(store.DefaultSnapshotFile)
		res := handler.Execute(types.Command{
			Name: "LINSERT",
			Args: []string{"BETWEEN", "list", "pivot", "value"},
		})
		if res.Status != types.StatusError {
			t.Fatalf("expected StatusError, got %v", res.Status)
		}
		if res.Message != "syntax error" {
			t.Fatalf("expected 'syntax error', got %q", res.Message)
		}
		if _, err := os.Stat(store.DefaultSnapshotFile); !os.IsNotExist(err) {
			t.Fatalf("expected no snapshot file to be created, but it exists")
		}
	})

	// 5. Missing key
	t.Run("missing_key", func(t *testing.T) {
		_ = os.Remove(store.DefaultSnapshotFile)
		res := handler.Execute(types.Command{
			Name: "LINSERT",
			Args: []string{"BEFORE", "missing", "pivot", "value"},
		})
		if res.Status != types.StatusInteger {
			t.Fatalf("expected StatusInteger, got %v", res.Status)
		}
		if res.Data.(int64) != 0 {
			t.Fatalf("expected 0, got %v", res.Data)
		}
		if _, err := os.Stat(store.DefaultSnapshotFile); os.IsNotExist(err) {
			t.Fatalf("expected snapshot file to be created on successful no-op mutation")
		}
	})

	// 6. BEFORE insertion
	t.Run("BEFORE_insertion", func(t *testing.T) {
		s.Delete("list")
		if _, err := s.RPush("list", "a", "b", "c"); err != nil {
			t.Fatalf("setup failed: %v", err)
		}
		res := handler.Execute(types.Command{
			Name: "LINSERT",
			Args: []string{"BEFORE", "list", "b", "x"},
		})
		if res.Status != types.StatusInteger {
			t.Fatalf("expected StatusInteger, got %v", res.Status)
		}
		if res.Data.(int64) != 4 {
			t.Fatalf("expected 4, got %v", res.Data)
		}
		listVal, _ := s.Get("list")
		list := listVal.Data.([]string)
		if len(list) != 4 || list[0] != "a" || list[1] != "x" || list[2] != "b" || list[3] != "c" {
			t.Fatalf("expected [a x b c], got %v", list)
		}
	})

	// 7. AFTER insertion
	t.Run("AFTER_insertion", func(t *testing.T) {
		s.Delete("list")
		if _, err := s.RPush("list", "a", "b", "c"); err != nil {
			t.Fatalf("setup failed: %v", err)
		}
		res := handler.Execute(types.Command{
			Name: "LINSERT",
			Args: []string{"AFTER", "list", "b", "x"},
		})
		if res.Status != types.StatusInteger {
			t.Fatalf("expected StatusInteger, got %v", res.Status)
		}
		if res.Data.(int64) != 4 {
			t.Fatalf("expected 4, got %v", res.Data)
		}
		listVal, _ := s.Get("list")
		list := listVal.Data.([]string)
		if len(list) != 4 || list[0] != "a" || list[1] != "b" || list[2] != "x" || list[3] != "c" {
			t.Fatalf("expected [a b x c], got %v", list)
		}
	})

	// 8. Pivot not found
	t.Run("pivot_not_found", func(t *testing.T) {
		s.Delete("list")
		if _, err := s.RPush("list", "a", "b", "c"); err != nil {
			t.Fatalf("setup failed: %v", err)
		}
		_ = os.Remove(store.DefaultSnapshotFile)
		res := handler.Execute(types.Command{
			Name: "LINSERT",
			Args: []string{"BEFORE", "list", "missing", "x"},
		})
		if res.Status != types.StatusInteger {
			t.Fatalf("expected StatusInteger, got %v", res.Status)
		}
		if res.Data.(int64) != -1 {
			t.Fatalf("expected -1, got %v", res.Data)
		}
		listVal, _ := s.Get("list")
		list := listVal.Data.([]string)
		if len(list) != 3 || list[0] != "a" || list[1] != "b" || list[2] != "c" {
			t.Fatalf("expected [a b c], got %v", list)
		}
		if _, err := os.Stat(store.DefaultSnapshotFile); os.IsNotExist(err) {
			t.Fatalf("expected snapshot file to be created on successful pivot not found")
		}
	})

	// 9. WRONGTYPE
	t.Run("WRONGTYPE", func(t *testing.T) {
		_ = os.Remove(store.DefaultSnapshotFile)
		s.Set("stringkey", types.Value{
			Type: types.StringType,
			Data: "val",
		})
		res := handler.Execute(types.Command{
			Name: "LINSERT",
			Args: []string{"BEFORE", "stringkey", "pivot", "value"},
		})
		if res.Status != types.StatusError {
			t.Fatalf("expected StatusError, got %v", res.Status)
		}
		if res.Message != store.ErrWrongType.Error() {
			t.Fatalf("expected %q, got %q", store.ErrWrongType.Error(), res.Message)
		}
		if _, err := os.Stat(store.DefaultSnapshotFile); !os.IsNotExist(err) {
			t.Fatalf("expected no snapshot file to be created for WRONGTYPE")
		}
	})

	// 10. Persistence after successful insertion
	t.Run("persistence_after_successful_insertion", func(t *testing.T) {
		s.Delete("persist_list")
		if _, err := s.RPush("persist_list", "a"); err != nil {
			t.Fatalf("setup failed: %v", err)
		}
		_ = os.Remove(store.DefaultSnapshotFile)
		res := handler.Execute(types.Command{
			Name: "LINSERT",
			Args: []string{"BEFORE", "persist_list", "a", "x"},
		})
		if res.Status != types.StatusInteger {
			t.Fatalf("expected StatusInteger, got %v", res.Status)
		}
		if _, err := os.Stat(store.DefaultSnapshotFile); os.IsNotExist(err) {
			t.Fatalf("expected snapshot file to be created")
		}
	})

	// 11. Persistence after successful no-op
	t.Run("persistence_after_successful_no_op", func(t *testing.T) {
		_ = os.Remove(store.DefaultSnapshotFile)
		res := handler.Execute(types.Command{
			Name: "LINSERT",
			Args: []string{"BEFORE", "missing_persist", "pivot", "value"},
		})
		if res.Status != types.StatusInteger {
			t.Fatalf("expected StatusInteger, got %v", res.Status)
		}
		if res.Data.(int64) != 0 {
			t.Fatalf("expected 0, got %v", res.Data)
		}
		if _, err := os.Stat(store.DefaultSnapshotFile); os.IsNotExist(err) {
			t.Fatalf("expected snapshot file to be created for missing key")
		}

		_ = os.Remove(store.DefaultSnapshotFile)
		s.Delete("persist_list2")
		if _, err := s.RPush("persist_list2", "a"); err != nil {
			t.Fatalf("setup failed: %v", err)
		}
		res = handler.Execute(types.Command{
			Name: "LINSERT",
			Args: []string{"BEFORE", "persist_list2", "missing_pivot", "value"},
		})
		if res.Status != types.StatusInteger {
			t.Fatalf("expected StatusInteger, got %v", res.Status)
		}
		if res.Data.(int64) != -1 {
			t.Fatalf("expected -1, got %v", res.Data)
		}
		if _, err := os.Stat(store.DefaultSnapshotFile); os.IsNotExist(err) {
			t.Fatalf("expected snapshot file to be created for missing pivot")
		}
	})
}

func TestLInsertDispatcherRouting(t *testing.T) {
	_ = os.Remove(store.DefaultSnapshotFile)
	defer os.Remove(store.DefaultSnapshotFile)

	s := store.NewMemoryStore()
	dispatcher := NewDispatcher(s)

	if _, err := s.RPush("list", "a", "b", "c"); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	res := dispatcher.Dispatch(types.Command{
		Name: "LINSERT",
		Args: []string{"BEFORE", "list", "b", "x"},
	})
	if res.Status != types.StatusInteger {
		t.Fatalf("expected StatusInteger, got %v", res.Status)
	}
	if res.Data.(int64) != 4 {
		t.Fatalf("expected 4, got %v", res.Data)
	}
	listVal, _ := s.Get("list")
	list := listVal.Data.([]string)
	if len(list) != 4 || list[0] != "a" || list[1] != "x" || list[2] != "b" || list[3] != "c" {
		t.Fatalf("expected [a x b c], got %v", list)
	}
}
