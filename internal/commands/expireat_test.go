package commands

import (
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/rajas2007/IgnisKV/internal/store"
	"github.com/rajas2007/IgnisKV/internal/types"
)

func TestExpireAtValidTimestamp(t *testing.T) {
	// Arrange
	s := store.NewMemoryStore()
	h := NewExpireAtHandler(s)

	s.Set("key", types.Value{
		Type: types.StringType,
		Data: "value",
	})

	futureTime := time.Now().Add(1 * time.Hour).Unix()
	timestampStr := strconv.FormatInt(futureTime, 10)

	// Act
	resp := h.Execute(types.Command{Name: "EXPIREAT", Args: []string{"key", timestampStr}})

	// Assert
	if resp.Status != types.StatusInteger {
		t.Fatalf("EXPIREAT returned Status %v; want StatusInteger", resp.Status)
	}
	if resp.Data != "1" {
		t.Fatalf("EXPIREAT returned Data %v; want %q", resp.Data, "1")
	}

	val, err := s.Get("key")
	if err != nil {
		t.Fatalf("Get() returned unexpected error: %v", err)
	}
	if val.ExpiresAt.Unix() != futureTime {
		t.Fatalf("ExpiresAt was %d, expected %d", val.ExpiresAt.Unix(), futureTime)
	}
}

func TestExpireAtMissingKey(t *testing.T) {
	// Arrange
	s := store.NewMemoryStore()
	h := NewExpireAtHandler(s)

	futureTime := time.Now().Add(1 * time.Hour).Unix()
	timestampStr := strconv.FormatInt(futureTime, 10)

	// Act
	resp := h.Execute(types.Command{Name: "EXPIREAT", Args: []string{"missing", timestampStr}})

	// Assert
	if resp.Status != types.StatusInteger {
		t.Fatalf("EXPIREAT returned Status %v; want StatusInteger", resp.Status)
	}
	if resp.Data != "0" {
		t.Fatalf("EXPIREAT returned Data %v; want %q", resp.Data, "0")
	}
}

func TestExpireAtExpiredKey(t *testing.T) {
	// Arrange
	s := store.NewMemoryStore()
	h := NewExpireAtHandler(s)

	s.Set("key", types.Value{
		Type:      types.StringType,
		Data:      "value",
		ExpiresAt: time.Now().Add(50 * time.Millisecond),
	})

	time.Sleep(100 * time.Millisecond)

	futureTime := time.Now().Add(1 * time.Hour).Unix()
	timestampStr := strconv.FormatInt(futureTime, 10)

	// Act
	resp := h.Execute(types.Command{Name: "EXPIREAT", Args: []string{"key", timestampStr}})

	// Assert
	if resp.Status != types.StatusInteger {
		t.Fatalf("EXPIREAT returned Status %v; want StatusInteger", resp.Status)
	}
	if resp.Data != "0" {
		t.Fatalf("EXPIREAT returned Data %v; want %q", resp.Data, "0")
	}

	if s.Exists("key") {
		t.Fatalf("expired key should have been lazily deleted by EXPIREAT")
	}
}

func TestExpireAtPastTimestamp(t *testing.T) {
	// Arrange
	s := store.NewMemoryStore()
	h := NewExpireAtHandler(s)

	s.Set("key", types.Value{
		Type: types.StringType,
		Data: "value",
	})

	pastTime := time.Now().Add(-1 * time.Hour).Unix()
	timestampStr := strconv.FormatInt(pastTime, 10)

	// Act
	resp := h.Execute(types.Command{Name: "EXPIREAT", Args: []string{"key", timestampStr}})

	// Assert
	if resp.Status != types.StatusError {
		t.Fatalf("EXPIREAT returned Status %v; want StatusError", resp.Status)
	}
	if resp.Message != store.ErrInvalidTimestamp.Error() {
		t.Fatalf("EXPIREAT returned Message %q; want %q", resp.Message, store.ErrInvalidTimestamp.Error())
	}
}

func TestExpireAtMalformedTimestamp(t *testing.T) {
	// Arrange
	s := store.NewMemoryStore()
	h := NewExpireAtHandler(s)

	s.Set("key", types.Value{
		Type: types.StringType,
		Data: "value",
	})

	// Act
	resp := h.Execute(types.Command{Name: "EXPIREAT", Args: []string{"key", "not-an-int"}})

	// Assert
	if resp.Status != types.StatusError {
		t.Fatalf("EXPIREAT returned Status %v; want StatusError", resp.Status)
	}
	if resp.Message != store.ErrInvalidTimestamp.Error() {
		t.Fatalf("EXPIREAT returned Message %q; want %q", resp.Message, store.ErrInvalidTimestamp.Error())
	}
}

func TestExpireAtWrongArgumentCount(t *testing.T) {
	// Arrange
	s := store.NewMemoryStore()
	h := NewExpireAtHandler(s)

	tests := []struct {
		name string
		args []string
	}{
		{"no arguments", []string{}},
		{"one argument", []string{"key"}},
		{"too many arguments", []string{"key", "123", "extra"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			resp := h.Execute(types.Command{Name: "EXPIREAT", Args: tt.args})

			// Assert
			if resp.Status != types.StatusError {
				t.Fatalf("EXPIREAT with %s returned Status %v; want StatusError", tt.name, resp.Status)
			}
			if resp.Message != "wrong number of arguments" {
				t.Fatalf("EXPIREAT with %s returned Message %q; want %q", tt.name, resp.Message, "wrong number of arguments")
			}
		})
	}
}

func TestExpireAtPersistenceAfterSuccess(t *testing.T) {
	// Arrange
	os.Remove(store.DefaultSnapshotFile)

	s := store.NewMemoryStore()
	h := NewExpireAtHandler(s)

	s.Set("key", types.Value{
		Type: types.StringType,
		Data: "value",
	})

	futureTime := time.Now().Add(1 * time.Hour).Unix()
	timestampStr := strconv.FormatInt(futureTime, 10)

	// Act
	resp := h.Execute(types.Command{Name: "EXPIREAT", Args: []string{"key", timestampStr}})

	// Assert
	if resp.Status != types.StatusInteger || resp.Data != "1" {
		t.Fatalf("EXPIREAT returned unexpected response: %+v", resp)
	}

	if _, err := os.Stat(store.DefaultSnapshotFile); os.IsNotExist(err) {
		t.Fatalf("expected snapshot file to be created on successful EXPIREAT")
	}
}

func TestExpireAtNoPersistenceAfterFailure(t *testing.T) {
	// Arrange
	os.Remove(store.DefaultSnapshotFile)

	s := store.NewMemoryStore()
	h := NewExpireAtHandler(s)

	futureTime := time.Now().Add(1 * time.Hour).Unix()
	timestampStr := strconv.FormatInt(futureTime, 10)

	// Act — missing key
	resp := h.Execute(types.Command{Name: "EXPIREAT", Args: []string{"missing", timestampStr}})

	// Assert
	if resp.Status != types.StatusInteger || resp.Data != "0" {
		t.Fatalf("EXPIREAT on missing key returned unexpected response: %+v", resp)
	}

	if _, err := os.Stat(store.DefaultSnapshotFile); !os.IsNotExist(err) {
		t.Fatalf("expected snapshot file NOT to be created on failed EXPIREAT")
	}
}
