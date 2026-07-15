package commands

import (
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

func TestPExpireAtValidTimestamp(t *testing.T) {
	// Arrange
	s := store.NewMemoryStore()
	h := NewPExpireAtHandler(s)

	s.Set("key", types.Value{
		Type: types.StringType,
		Data: "value",
	})

	futureTime := time.Now().Add(5 * time.Second)
	timestampMs := futureTime.UnixMilli()

	// Act
	resp := h.Execute(types.Command{Name: "PEXPIREAT", Args: []string{"key", strconv.FormatInt(timestampMs, 10)}})

	// Assert
	if resp.Status != types.StatusInteger {
		t.Fatalf("PEXPIREAT returned Status %v; want StatusInteger", resp.Status)
	}
	if resp.Data != "1" {
		t.Fatalf("PEXPIREAT returned Data %v; want %q", resp.Data, "1")
	}

	val, err := s.Get("key")
	if err != nil {
		t.Fatalf("Get() returned unexpected error: %v", err)
	}
	if val.ExpiresAt.IsZero() {
		t.Fatalf("ExpiresAt was zero, expected a future time")
	}

	// UnixMilli truncates the original time.Now() to milliseconds, so they should be very close.
	if val.ExpiresAt.UnixMilli() != timestampMs {
		t.Fatalf("ExpiresAt %v does not match expected timestamp %v", val.ExpiresAt.UnixMilli(), timestampMs)
	}
}

func TestPExpireAtMissingKey(t *testing.T) {
	// Arrange
	s := store.NewMemoryStore()
	h := NewPExpireAtHandler(s)

	futureTime := time.Now().Add(5 * time.Second)
	timestampMs := futureTime.UnixMilli()

	// Act
	resp := h.Execute(types.Command{Name: "PEXPIREAT", Args: []string{"missing", strconv.FormatInt(timestampMs, 10)}})

	// Assert
	if resp.Status != types.StatusInteger {
		t.Fatalf("PEXPIREAT returned Status %v; want StatusInteger", resp.Status)
	}
	if resp.Data != "0" {
		t.Fatalf("PEXPIREAT returned Data %v; want %q", resp.Data, "0")
	}
}

func TestPExpireAtExpiredKey(t *testing.T) {
	// Arrange
	s := store.NewMemoryStore()
	h := NewPExpireAtHandler(s)

	s.Set("key", types.Value{
		Type:      types.StringType,
		Data:      "value",
		ExpiresAt: time.Now().Add(50 * time.Millisecond),
	})

	time.Sleep(100 * time.Millisecond)

	futureTime := time.Now().Add(5 * time.Second)
	timestampMs := futureTime.UnixMilli()

	// Act
	resp := h.Execute(types.Command{Name: "PEXPIREAT", Args: []string{"key", strconv.FormatInt(timestampMs, 10)}})

	// Assert
	if resp.Status != types.StatusInteger {
		t.Fatalf("PEXPIREAT returned Status %v; want StatusInteger", resp.Status)
	}
	if resp.Data != "0" {
		t.Fatalf("PEXPIREAT returned Data %v; want %q", resp.Data, "0")
	}

	if s.Exists("key") {
		t.Fatalf("expired key should have been lazily deleted by PEXPIREAT")
	}
}

func TestPExpireAtPastTimestamp(t *testing.T) {
	// Arrange
	s := store.NewMemoryStore()
	h := NewPExpireAtHandler(s)

	s.Set("key", types.Value{
		Type: types.StringType,
		Data: "value",
	})

	pastTime := time.Now().Add(-1 * time.Hour)
	timestampMs := pastTime.UnixMilli()

	// Act
	resp := h.Execute(types.Command{Name: "PEXPIREAT", Args: []string{"key", strconv.FormatInt(timestampMs, 10)}})

	// Assert
	if resp.Status != types.StatusError {
		t.Fatalf("PEXPIREAT returned Status %v; want StatusError", resp.Status)
	}
	if resp.Message != store.ErrInvalidTimestamp.Error() {
		t.Fatalf("PEXPIREAT returned Message %q; want %q", resp.Message, store.ErrInvalidTimestamp.Error())
	}
}

func TestPExpireAtMalformedTimestamp(t *testing.T) {
	// Arrange
	s := store.NewMemoryStore()
	h := NewPExpireAtHandler(s)

	s.Set("key", types.Value{
		Type: types.StringType,
		Data: "value",
	})

	// Act
	resp := h.Execute(types.Command{Name: "PEXPIREAT", Args: []string{"key", "not-an-int"}})

	// Assert
	if resp.Status != types.StatusError {
		t.Fatalf("PEXPIREAT returned Status %v; want StatusError", resp.Status)
	}
	if resp.Message != store.ErrInvalidTimestamp.Error() {
		t.Fatalf("PEXPIREAT returned Message %q; want %q", resp.Message, store.ErrInvalidTimestamp.Error())
	}
}

func TestPExpireAtWrongArgumentCount(t *testing.T) {
	// Arrange
	s := store.NewMemoryStore()
	h := NewPExpireAtHandler(s)

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
			resp := h.Execute(types.Command{Name: "PEXPIREAT", Args: tt.args})

			// Assert
			if resp.Status != types.StatusError {
				t.Fatalf("PEXPIREAT with %s returned Status %v; want StatusError", tt.name, resp.Status)
			}
			if resp.Message != "wrong number of arguments" {
				t.Fatalf("PEXPIREAT with %s returned Message %q; want %q", tt.name, resp.Message, "wrong number of arguments")
			}
		})
	}
}

func TestPExpireAtPersistenceAfterSuccess(t *testing.T) {
	// Arrange
	os.Remove(store.DefaultSnapshotFile)

	s := store.NewMemoryStore()
	h := NewPExpireAtHandler(s)

	s.Set("key", types.Value{
		Type: types.StringType,
		Data: "value",
	})

	futureTime := time.Now().Add(5 * time.Second)
	timestampMs := futureTime.UnixMilli()

	// Act
	resp := h.Execute(types.Command{Name: "PEXPIREAT", Args: []string{"key", strconv.FormatInt(timestampMs, 10)}})

	// Assert
	if resp.Status != types.StatusInteger || resp.Data != "1" {
		t.Fatalf("PEXPIREAT returned unexpected response: %+v", resp)
	}

	if _, err := os.Stat(store.DefaultSnapshotFile); os.IsNotExist(err) {
		t.Fatalf("expected snapshot file to be created on successful PEXPIREAT")
	}
}

func TestPExpireAtNoPersistenceAfterFailure(t *testing.T) {
	// Arrange
	os.Remove(store.DefaultSnapshotFile)

	s := store.NewMemoryStore()
	h := NewPExpireAtHandler(s)

	futureTime := time.Now().Add(5 * time.Second)
	timestampMs := futureTime.UnixMilli()

	// Act — missing key
	resp := h.Execute(types.Command{Name: "PEXPIREAT", Args: []string{"missing", strconv.FormatInt(timestampMs, 10)}})

	// Assert
	if resp.Status != types.StatusInteger || resp.Data != "0" {
		t.Fatalf("PEXPIREAT on missing key returned unexpected response: %+v", resp)
	}

	if _, err := os.Stat(store.DefaultSnapshotFile); !os.IsNotExist(err) {
		t.Fatalf("expected snapshot file NOT to be created on failed PEXPIREAT")
	}
}

func TestDispatcherRoutesPExpireAt(t *testing.T) {
	// Arrange
	s := store.NewMemoryStore()
	d := NewDispatcher(s)

	s.Set("key", types.Value{
		Type: types.StringType,
		Data: "value",
	})

	futureTime := time.Now().Add(5 * time.Second)
	timestampMs := futureTime.UnixMilli()

	// Act
	resp := d.Dispatch(types.Command{Name: "PEXPIREAT", Args: []string{"key", strconv.FormatInt(timestampMs, 10)}})

	// Assert
	if resp.Status != types.StatusInteger || resp.Data != "1" {
		t.Fatalf("Dispatcher PEXPIREAT returned Status %v, Data %q; want StatusInteger, \"1\"", resp.Status, resp.Data)
	}
}
