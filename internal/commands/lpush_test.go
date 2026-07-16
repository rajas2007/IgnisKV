package commands

import (
	"os"
	"testing"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

func TestLPushHandler(t *testing.T) {
	s := store.NewMemoryStore()
	handler := NewLPushHandler(s)

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
				if len(list) != 2 || list[0] != "b" || list[1] != "a" {
					t.Fatalf("expected [b a], got %v", list)
				}
			},
		},
		{
			name:       "Multiple values LPUSH ordering",
			args:       []string{"multilist", "1", "2", "3"},
			wantStatus: types.StatusInteger,
			wantData:   "3",
			checkStore: func(t *testing.T) {
				val, _ := s.Get("multilist")
				list := val.Data.([]string)
				if len(list) != 3 || list[0] != "3" || list[1] != "2" || list[2] != "1" {
					t.Fatalf("expected [3 2 1], got %v", list)
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

			cmd := types.Command{Name: "LPUSH", Args: tt.args}
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

func TestLPushPersistenceAfterSuccess(t *testing.T) {
	s := store.NewMemoryStore()
	handler := NewLPushHandler(s)

	os.Remove(store.DefaultSnapshotFile)

	resp := handler.Execute(types.Command{Name: "LPUSH", Args: []string{"listkey", "val"}})
	if resp.Status != types.StatusInteger {
		t.Fatalf("LPUSH returned Status %v; want StatusInteger", resp.Status)
	}

	if _, err := os.Stat(store.DefaultSnapshotFile); os.IsNotExist(err) {
		t.Fatalf("expected snapshot file %q to be created after successful LPUSH", store.DefaultSnapshotFile)
	}
}

func TestLPushNoPersistenceAfterFailure(t *testing.T) {
	s := store.NewMemoryStore()
	handler := NewLPushHandler(s)

	os.Remove(store.DefaultSnapshotFile)

	// WRONGTYPE error
	s.Set("stringkey", types.Value{Type: types.StringType, Data: "val"})
	resp := handler.Execute(types.Command{Name: "LPUSH", Args: []string{"stringkey", "val"}})
	if resp.Status != types.StatusError {
		t.Fatalf("LPUSH returned Status %v; want StatusError", resp.Status)
	}

	if _, err := os.Stat(store.DefaultSnapshotFile); !os.IsNotExist(err) {
		t.Fatalf("did not expect snapshot file %q to be created after failed LPUSH", store.DefaultSnapshotFile)
	}
}

func TestDispatcherRoutesLPush(t *testing.T) {
	s := store.NewMemoryStore()
	d := NewDispatcher(s)

	cmd := types.Command{Name: "LPUSH", Args: []string{"list", "val"}}

	// We expect a successful execution of LPUSH resulting in a status integer 1
	// because LPUSH "list" "val" is valid.
	resp := d.Dispatch(cmd)
	if resp.Status != types.StatusInteger || resp.Data != "1" {
		t.Fatalf("Dispatcher did not correctly route LPUSH command: %+v", resp)
	}
}
