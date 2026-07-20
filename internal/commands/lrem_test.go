package commands

import (
	"os"
	"testing"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

func TestLRemHandler(t *testing.T) {
	s := store.NewMemoryStore()
	h := NewLRemHandler(s)

	// 1. Wrong number of arguments (0)
	resp := h.Execute(types.Command{
		Name: "LREM",
		Args: []string{},
	})
	if resp.Status != types.StatusError || resp.Message != "wrong number of arguments" {
		t.Fatalf("expected StatusError 'wrong number of arguments', got %v: %s", resp.Status, resp.Message)
	}

	// 2. Wrong number of arguments (2)
	resp = h.Execute(types.Command{
		Name: "LREM",
		Args: []string{"list", "1"},
	})
	if resp.Status != types.StatusError || resp.Message != "wrong number of arguments" {
		t.Fatalf("expected StatusError 'wrong number of arguments', got %v: %s", resp.Status, resp.Message)
	}

	// 3. Wrong number of arguments (4)
	resp = h.Execute(types.Command{
		Name: "LREM",
		Args: []string{"list", "1", "a", "extra"},
	})
	if resp.Status != types.StatusError || resp.Message != "wrong number of arguments" {
		t.Fatalf("expected StatusError 'wrong number of arguments', got %v: %s", resp.Status, resp.Message)
	}

	// 4. Invalid count
	resp = h.Execute(types.Command{
		Name: "LREM",
		Args: []string{"list", "abc", "a"},
	})
	if resp.Status != types.StatusError || resp.Message != "value is not an integer or out of range" {
		t.Fatalf("expected StatusError 'value is not an integer or out of range', got %v: %s", resp.Status, resp.Message)
	}

	// 5. Missing key
	resp = h.Execute(types.Command{
		Name: "LREM",
		Args: []string{"missing", "1", "a"},
	})
	if resp.Status != types.StatusInteger || resp.Data.(int64) != 0 {
		t.Fatalf("expected StatusInteger(0), got %v: %v", resp.Status, resp.Data)
	}

	// 6. count > 0
	s.RPush("list", "a", "b", "a", "c", "a")
	resp = h.Execute(types.Command{
		Name: "LREM",
		Args: []string{"list", "2", "a"},
	})
	if resp.Status != types.StatusInteger || resp.Data.(int64) != 2 {
		t.Fatalf("expected StatusInteger(2), got %v: %v", resp.Status, resp.Data)
	}
	listVal, _ := s.Get("list")
	list := listVal.Data.([]string)
	if len(list) != 3 || list[0] != "b" || list[1] != "c" || list[2] != "a" {
		t.Fatalf("expected list [b c a], got %v", list)
	}

	// 7. count < 0
	s.RPush("list2", "a", "b", "a", "c", "a")
	resp = h.Execute(types.Command{
		Name: "LREM",
		Args: []string{"list2", "-2", "a"},
	})
	if resp.Status != types.StatusInteger || resp.Data.(int64) != 2 {
		t.Fatalf("expected StatusInteger(2), got %v: %v", resp.Status, resp.Data)
	}
	listVal, _ = s.Get("list2")
	list = listVal.Data.([]string)
	if len(list) != 3 || list[0] != "a" || list[1] != "b" || list[2] != "c" {
		t.Fatalf("expected list [a b c], got %v", list)
	}

	// 8. count == 0
	s.RPush("list3", "a", "b", "a", "c", "a")
	resp = h.Execute(types.Command{
		Name: "LREM",
		Args: []string{"list3", "0", "a"},
	})
	if resp.Status != types.StatusInteger || resp.Data.(int64) != 3 {
		t.Fatalf("expected StatusInteger(3), got %v: %v", resp.Status, resp.Data)
	}
	listVal, _ = s.Get("list3")
	list = listVal.Data.([]string)
	if len(list) != 2 || list[0] != "b" || list[1] != "c" {
		t.Fatalf("expected list [b c], got %v", list)
	}

	// 9. WRONGTYPE
	s.Set("stringkey", types.Value{
		Type: types.StringType,
		Data: "val",
	})
	resp = h.Execute(types.Command{
		Name: "LREM",
		Args: []string{"stringkey", "1", "a"},
	})
	if resp.Status != types.StatusError || resp.Message != store.ErrWrongType.Error() {
		t.Fatalf("expected StatusError ErrWrongType, got %v: %s", resp.Status, resp.Message)
	}

	// 10. Persistence after success
	os.Remove(store.DefaultSnapshotFile)
	s.RPush("persist_list", "a", "a")
	os.Remove(store.DefaultSnapshotFile) // clear snapshot from RPUSH

	resp = h.Execute(types.Command{
		Name: "LREM",
		Args: []string{"persist_list", "1", "a"},
	})
	if resp.Status != types.StatusInteger || resp.Data.(int64) != 1 {
		t.Fatalf("expected StatusInteger(1), got %v: %v", resp.Status, resp.Data)
	}
	if _, err := os.Stat(store.DefaultSnapshotFile); os.IsNotExist(err) {
		t.Fatalf("expected snapshot file to exist after successful LREM, but it does not")
	}

	// 11. Persistence after zero removals
	os.Remove(store.DefaultSnapshotFile)
	resp = h.Execute(types.Command{
		Name: "LREM",
		Args: []string{"missing_persist", "1", "a"},
	})
	if resp.Status != types.StatusInteger || resp.Data.(int64) != 0 {
		t.Fatalf("expected StatusInteger(0), got %v: %v", resp.Status, resp.Data)
	}
	if _, err := os.Stat(store.DefaultSnapshotFile); os.IsNotExist(err) {
		t.Fatalf("expected snapshot file to exist after successful no-op LREM, but it does not")
	}

	// 12. Dispatcher routing
	d := NewDispatcher(s)
	d.handlers["LREM"] = h

	s.RPush("list4", "a", "b", "a")
	resp = d.Dispatch(types.Command{
		Name: "LREM",
		Args: []string{"list4", "1", "a"},
	})
	if resp.Status != types.StatusInteger || resp.Data.(int64) != 1 {
		t.Fatalf("expected StatusInteger(1) from dispatcher, got %v: %v", resp.Status, resp.Data)
	}
	listVal, _ = s.Get("list4")
	list = listVal.Data.([]string)
	if len(list) != 2 || list[0] != "b" || list[1] != "a" {
		t.Fatalf("expected list [b a], got %v", list)
	}

	// cleanup
	os.Remove(store.DefaultSnapshotFile)
}
