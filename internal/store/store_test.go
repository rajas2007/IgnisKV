package store

import (
	"errors"
	"testing"
	"time"

	"github.com/rajas2007/IgnisKV/internal/types"
)

func TestNewMemoryStore(t *testing.T) {
	// Arrange / Act
	s := NewMemoryStore()

	// Assert
	if s == nil {
		t.Fatalf("NewMemoryStore() returned nil; expected a non-nil *MemoryStore")
	}

	if s.Exists("any-key") {
		t.Fatalf("expected newly created store to be empty; Exists() returned true")
	}
}

func TestSetAndGet(t *testing.T) {
	// Arrange
	s := NewMemoryStore()
	want := types.Value{
		Type:      types.StringType,
		Data:      "hello",
		ExpiresAt: time.Time{},
	}

	// Act
	s.Set("greeting", want)
	got, err := s.Get("greeting")

	// Assert
	if err != nil {
		t.Fatalf("Get() returned unexpected error: %v", err)
	}
	if got.Type != want.Type {
		t.Fatalf("Get() Type = %v; want %v", got.Type, want.Type)
	}
	if got.Data != want.Data {
		t.Fatalf("Get() Data = %v; want %v", got.Data, want.Data)
	}
	if !got.ExpiresAt.Equal(want.ExpiresAt) {
		t.Fatalf("Get() ExpiresAt = %v; want %v", got.ExpiresAt, want.ExpiresAt)
	}
}

func TestGetMissingKey(t *testing.T) {
	// Arrange
	s := NewMemoryStore()

	// Act
	_, err := s.Get("nonexistent")

	// Assert
	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("Get() on missing key returned %v; want ErrKeyNotFound", err)
	}
}

func TestOverwriteValue(t *testing.T) {
	// Arrange
	s := NewMemoryStore()
	original := types.Value{Type: types.StringType, Data: "original"}
	updated := types.Value{Type: types.StringType, Data: "updated"}

	// Act
	s.Set("key", original)
	s.Set("key", updated)
	got, err := s.Get("key")

	// Assert
	if err != nil {
		t.Fatalf("Get() returned unexpected error after overwrite: %v", err)
	}
	if got.Data != updated.Data {
		t.Fatalf("Get() after overwrite = %v; want %v", got.Data, updated.Data)
	}
}

func TestDeleteExistingKey(t *testing.T) {
	// Arrange
	s := NewMemoryStore()
	s.Set("key", types.Value{Type: types.StringType, Data: "value"})

	// Act
	err := s.Delete("key")

	// Assert
	if err != nil {
		t.Fatalf("Delete() on existing key returned unexpected error: %v", err)
	}
	if s.Exists("key") {
		t.Fatalf("Exists() returned true after Delete(); expected false")
	}
	_, getErr := s.Get("key")
	if !errors.Is(getErr, ErrKeyNotFound) {
		t.Fatalf("Get() after Delete() returned %v; want ErrKeyNotFound", getErr)
	}
}

func TestDeleteMissingKey(t *testing.T) {
	// Arrange
	s := NewMemoryStore()

	// Act
	err := s.Delete("nonexistent")

	// Assert
	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("Delete() on missing key returned %v; want ErrKeyNotFound", err)
	}
}

func TestExists(t *testing.T) {
	// Arrange
	s := NewMemoryStore()

	// Assert — missing key
	if s.Exists("key") {
		t.Fatalf("Exists() returned true for a key that was never set")
	}

	// Act — set key
	s.Set("key", types.Value{Type: types.StringType, Data: "value"})

	// Assert — present key
	if !s.Exists("key") {
		t.Fatalf("Exists() returned false for a key that was set")
	}

	// Act — delete key
	if err := s.Delete("key"); err != nil {
		t.Fatalf("Delete() returned unexpected error: %v", err)
	}

	// Assert — deleted key
	if s.Exists("key") {
		t.Fatalf("Exists() returned true for a key that was deleted")
	}
}
