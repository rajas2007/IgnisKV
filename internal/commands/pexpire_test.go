package commands

import (
	"os"
	"testing"
	"time"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

func TestPExpireValidDuration(t *testing.T) {
	// Arrange
	s := store.NewMemoryStore()
	h := NewPExpireHandler(s)

	s.Set("key", types.Value{
		Type: types.StringType,
		Data: "value",
	})

	// Act
	resp := h.Execute(types.Command{Name: "PEXPIRE", Args: []string{"key", "1500"}})

	// Assert
	if resp.Status != types.StatusInteger {
		t.Fatalf("PEXPIRE returned Status %v; want StatusInteger", resp.Status)
	}
	if resp.Data != "1" {
		t.Fatalf("PEXPIRE returned Data %v; want %q", resp.Data, "1")
	}

	val, err := s.Get("key")
	if err != nil {
		t.Fatalf("Get() returned unexpected error: %v", err)
	}
	if val.ExpiresAt.IsZero() {
		t.Fatalf("ExpiresAt was zero, expected a future time")
	}

	remaining := time.Until(val.ExpiresAt)
	if remaining < 900*time.Millisecond || remaining > 1500*time.Millisecond {
		t.Fatalf("ExpiresAt remaining time %v is out of expected bounds (900ms - 1500ms)", remaining)
	}
}

func TestPExpireMissingKey(t *testing.T) {
	// Arrange
	s := store.NewMemoryStore()
	h := NewPExpireHandler(s)

	// Act
	resp := h.Execute(types.Command{Name: "PEXPIRE", Args: []string{"missing", "1500"}})

	// Assert
	if resp.Status != types.StatusInteger {
		t.Fatalf("PEXPIRE returned Status %v; want StatusInteger", resp.Status)
	}
	if resp.Data != "0" {
		t.Fatalf("PEXPIRE returned Data %v; want %q", resp.Data, "0")
	}
}

func TestPExpireExpiredKey(t *testing.T) {
	// Arrange
	s := store.NewMemoryStore()
	h := NewPExpireHandler(s)

	s.Set("key", types.Value{
		Type:      types.StringType,
		Data:      "value",
		ExpiresAt: time.Now().Add(50 * time.Millisecond),
	})

	time.Sleep(100 * time.Millisecond)

	// Act
	resp := h.Execute(types.Command{Name: "PEXPIRE", Args: []string{"key", "1500"}})

	// Assert
	if resp.Status != types.StatusInteger {
		t.Fatalf("PEXPIRE returned Status %v; want StatusInteger", resp.Status)
	}
	if resp.Data != "0" {
		t.Fatalf("PEXPIRE returned Data %v; want %q", resp.Data, "0")
	}

	if s.Exists("key") {
		t.Fatalf("expired key should have been lazily deleted by PEXPIRE")
	}
}

func TestPExpireInvalidDuration(t *testing.T) {
	// Arrange
	s := store.NewMemoryStore()
	h := NewPExpireHandler(s)

	s.Set("key", types.Value{
		Type: types.StringType,
		Data: "value",
	})

	tests := []struct {
		name string
		args []string
	}{
		{"zero duration", []string{"key", "0"}},
		{"negative duration", []string{"key", "-100"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			resp := h.Execute(types.Command{Name: "PEXPIRE", Args: tt.args})

			// Assert
			if resp.Status != types.StatusError {
				t.Fatalf("PEXPIRE with %s returned Status %v; want StatusError", tt.name, resp.Status)
			}
			if resp.Message != store.ErrInvalidDuration.Error() {
				t.Fatalf("PEXPIRE with %s returned Message %q; want %q", tt.name, resp.Message, store.ErrInvalidDuration.Error())
			}
		})
	}
}

func TestPExpireMalformedDuration(t *testing.T) {
	// Arrange
	s := store.NewMemoryStore()
	h := NewPExpireHandler(s)

	s.Set("key", types.Value{
		Type: types.StringType,
		Data: "value",
	})

	// Act
	resp := h.Execute(types.Command{Name: "PEXPIRE", Args: []string{"key", "not-an-int"}})

	// Assert
	if resp.Status != types.StatusError {
		t.Fatalf("PEXPIRE returned Status %v; want StatusError", resp.Status)
	}
	if resp.Message != store.ErrInvalidDuration.Error() {
		t.Fatalf("PEXPIRE returned Message %q; want %q", resp.Message, store.ErrInvalidDuration.Error())
	}
}

func TestPExpireWrongArgumentCount(t *testing.T) {
	// Arrange
	s := store.NewMemoryStore()
	h := NewPExpireHandler(s)

	tests := []struct {
		name string
		args []string
	}{
		{"no arguments", []string{}},
		{"one argument", []string{"key"}},
		{"too many arguments", []string{"key", "1500", "extra"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			resp := h.Execute(types.Command{Name: "PEXPIRE", Args: tt.args})

			// Assert
			if resp.Status != types.StatusError {
				t.Fatalf("PEXPIRE with %s returned Status %v; want StatusError", tt.name, resp.Status)
			}
			if resp.Message != "wrong number of arguments" {
				t.Fatalf("PEXPIRE with %s returned Message %q; want %q", tt.name, resp.Message, "wrong number of arguments")
			}
		})
	}
}

func TestPExpirePersistenceAfterSuccess(t *testing.T) {
	// Arrange
	os.Remove(store.DefaultSnapshotFile)

	s := store.NewMemoryStore()
	h := NewPExpireHandler(s)

	s.Set("key", types.Value{
		Type: types.StringType,
		Data: "value",
	})

	// Act
	resp := h.Execute(types.Command{Name: "PEXPIRE", Args: []string{"key", "1500"}})

	// Assert
	if resp.Status != types.StatusInteger || resp.Data != "1" {
		t.Fatalf("PEXPIRE returned unexpected response: %+v", resp)
	}

	if _, err := os.Stat(store.DefaultSnapshotFile); os.IsNotExist(err) {
		t.Fatalf("expected snapshot file to be created on successful PEXPIRE")
	}
}

func TestPExpireNoPersistenceAfterFailure(t *testing.T) {
	// Arrange
	os.Remove(store.DefaultSnapshotFile)

	s := store.NewMemoryStore()
	h := NewPExpireHandler(s)

	// Act — missing key
	resp := h.Execute(types.Command{Name: "PEXPIRE", Args: []string{"missing", "1500"}})

	// Assert
	if resp.Status != types.StatusInteger || resp.Data != "0" {
		t.Fatalf("PEXPIRE on missing key returned unexpected response: %+v", resp)
	}

	if _, err := os.Stat(store.DefaultSnapshotFile); !os.IsNotExist(err) {
		t.Fatalf("expected snapshot file NOT to be created on failed PEXPIRE")
	}
}

func TestDispatcherRoutesPExpire(t *testing.T) {
	// Arrange
	s := store.NewMemoryStore()
	d := NewDispatcher(s)

	s.Set("key", types.Value{
		Type: types.StringType,
		Data: "value",
	})

	// Act
	resp := d.Dispatch(types.Command{Name: "PEXPIRE", Args: []string{"key", "1500"}})

	// Assert
	if resp.Status != types.StatusInteger || resp.Data != "1" {
		t.Fatalf("Dispatcher PEXPIRE returned Status %v, Data %q; want StatusInteger, \"1\"", resp.Status, resp.Data)
	}
}
