package commands

import (
	"os"
	"testing"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

func TestRPushHandler(t *testing.T) {
	// The commands package TestMain creates a temporary directory and changes
	// into it, so store.DefaultSnapshotFile ("igniskv.json") is written safely
	// into the isolated temp dir.

	s := store.NewMemoryStore()
	handler := NewRPushHandler(s)

	tests := []struct {
		name       string
		args       []string
		setup      func()
		wantStatus types.ResponseStatus
		wantMsg    string
		wantData   string
		checkStore func(*testing.T)
	}{
		{
			name:       "Wrong number of arguments (no args)",
			args:       []string{},
			wantStatus: types.StatusError,
			wantMsg:    "wrong number of arguments",
		},
		{
			name:       "Wrong number of arguments (only key)",
			args:       []string{"mylist"},
			wantStatus: types.StatusError,
			wantMsg:    "wrong number of arguments",
		},
		{
			name:       "Creates new list",
			args:       []string{"newlist", "a"},
			wantStatus: types.StatusInteger,
			wantData:   "1",
			checkStore: func(t *testing.T) {
				val, err := s.Get("newlist")
				if err != nil || val.Type != types.ListType {
					t.Fatalf("expected list")
				}
				list := val.Data.([]string)
				if len(list) != 1 || list[0] != "a" {
					t.Fatalf("expected [a]")
				}
			},
		},
		{
			name:       "Pushes to existing list",
			args:       []string{"newlist", "b"},
			wantStatus: types.StatusInteger,
			wantData:   "2",
			checkStore: func(t *testing.T) {
				val, _ := s.Get("newlist")
				list := val.Data.([]string)
				if len(list) != 2 || list[0] != "a" || list[1] != "b" {
					t.Fatalf("expected [a b], got %v", list)
				}
			},
		},
		{
			name: "Multiple values RPUSH ordering",
			args: []string{"multilist", "c", "d", "e"},
			setup: func() {
				_, _ = s.RPush("multilist", "a", "b")
			},
			wantStatus: types.StatusInteger,
			wantData:   "5",
			checkStore: func(t *testing.T) {
				val, _ := s.Get("multilist")
				list := val.Data.([]string)
				if len(list) != 5 || list[0] != "a" || list[1] != "b" || list[2] != "c" || list[3] != "d" || list[4] != "e" {
					t.Fatalf("expected [a b c d e], got %v", list)
				}
			},
		},
		{
			name: "WRONGTYPE error",
			setup: func() {
				s.Set("stringkey", types.Value{Type: types.StringType, Data: "val"})
			},
			args:       []string{"stringkey", "a"},
			wantStatus: types.StatusError,
			wantMsg:    store.ErrWrongType.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			cmd := types.Command{Name: "RPUSH", Args: tt.args}
			resp := handler.Execute(cmd)

			if resp.Status != tt.wantStatus {
				t.Fatalf("expected status %v, got %v", tt.wantStatus, resp.Status)
			}
			if tt.wantMsg != "" && resp.Message != tt.wantMsg {
				t.Fatalf("expected message %q, got %q", tt.wantMsg, resp.Message)
			}
			if tt.wantData != "" && resp.Data != tt.wantData {
				t.Fatalf("expected data %q, got %q", tt.wantData, resp.Data)
			}

			if tt.checkStore != nil {
				tt.checkStore(t)
			}
		})
	}
}

func TestRPushPersistenceAfterSuccess(t *testing.T) {
	s := store.NewMemoryStore()
	handler := NewRPushHandler(s)

	os.Remove(store.DefaultSnapshotFile)

	resp := handler.Execute(types.Command{Name: "RPUSH", Args: []string{"listkey", "val"}})
	if resp.Status != types.StatusInteger {
		t.Fatalf("RPUSH returned Status %v; want StatusInteger", resp.Status)
	}

	if _, err := os.Stat(store.DefaultSnapshotFile); os.IsNotExist(err) {
		t.Fatalf("expected snapshot file %q to be created after successful RPUSH", store.DefaultSnapshotFile)
	}
}

func TestRPushNoPersistenceAfterFailure(t *testing.T) {
	s := store.NewMemoryStore()
	handler := NewRPushHandler(s)

	os.Remove(store.DefaultSnapshotFile)

	// WRONGTYPE error
	s.Set("stringkey", types.Value{Type: types.StringType, Data: "val"})
	resp := handler.Execute(types.Command{Name: "RPUSH", Args: []string{"stringkey", "val"}})
	if resp.Status != types.StatusError {
		t.Fatalf("RPUSH returned Status %v; want StatusError", resp.Status)
	}

	if _, err := os.Stat(store.DefaultSnapshotFile); !os.IsNotExist(err) {
		t.Fatalf("did not expect snapshot file %q to be created after failed RPUSH", store.DefaultSnapshotFile)
	}
}

func TestDispatcherRoutesRPush(t *testing.T) {
	s := store.NewMemoryStore()
	d := NewDispatcher(s)

	cmd := types.Command{Name: "RPUSH", Args: []string{"list", "val"}}

	// We expect a successful execution of RPUSH resulting in a status integer 1
	// because RPUSH "list" "val" is valid.
	resp := d.Dispatch(cmd)
	if resp.Status != types.StatusInteger || resp.Data != "1" {
		t.Fatalf("Dispatcher did not correctly route RPUSH command: %+v", resp)
	}
}
