package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
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

func TestSaveAndLoadRoundtrip(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	snapshotFile := filepath.Join(tempDir, "igniskv.json")

	s1 := NewMemoryStore()
	s1.Set("city", types.Value{Type: types.StringType, Data: "Pune"})
	s1.Set("age", types.Value{Type: types.StringType, Data: "25"})

	// Act - Save
	if err := s1.Save(snapshotFile); err != nil {
		t.Fatalf("Save() returned unexpected error: %v", err)
	}

	// Act - Load into new store
	s2 := NewMemoryStore()
	if err := s2.Load(snapshotFile); err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}

	// Assert
	if val, _ := s2.Get("city"); val.Data != "Pune" {
		t.Errorf("Load() city = %v; want Pune", val.Data)
	}
	if val, _ := s2.Get("age"); val.Data != "25" {
		t.Errorf("Load() age = %v; want 25", val.Data)
	}
}

func TestSaveAndLoadEmpty(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	snapshotFile := filepath.Join(tempDir, "empty.json")

	s1 := NewMemoryStore()

	// Act - Save empty store
	if err := s1.Save(snapshotFile); err != nil {
		t.Fatalf("Save() on empty store returned unexpected error: %v", err)
	}

	// Act - Load into new store
	s2 := NewMemoryStore()
	if err := s2.Load(snapshotFile); err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}

	// Assert
	if s2.Exists("any-key") {
		t.Fatalf("Load() of empty store should be empty")
	}
}

func TestLoadMissingFile(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	snapshotFile := filepath.Join(tempDir, "missing.json")
	s := NewMemoryStore()

	// Act
	err := s.Load(snapshotFile)

	// Assert
	if err != nil {
		t.Fatalf("Load() for missing file returned unexpected error: %v; expected nil (first-run)", err)
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	snapshotFile := filepath.Join(tempDir, "corrupted.json")
	if err := os.WriteFile(snapshotFile, []byte("{ invalid json "), 0644); err != nil {
		t.Fatalf("Failed to create corrupted file: %v", err)
	}

	s := NewMemoryStore()

	// Act
	err := s.Load(snapshotFile)

	// Assert
	if err == nil {
		t.Fatalf("Load() for invalid JSON returned nil; expected an error")
	}
}

// ----- Expiration tests (Sprint 10) -----

func TestGetPersistentKey(t *testing.T) {
	// Arrange — no ExpiresAt set
	s := NewMemoryStore()
	s.Set("key", types.Value{Type: types.StringType, Data: "value"})

	// Act
	got, err := s.Get("key")

	// Assert
	if err != nil {
		t.Fatalf("Get() returned unexpected error for persistent key: %v", err)
	}
	if got.Data != "value" {
		t.Fatalf("Get() Data = %v; want %q", got.Data, "value")
	}
}

func TestGetExpiredKey(t *testing.T) {
	// Arrange — set a key that expires immediately
	s := NewMemoryStore()
	s.Set("expiring", types.Value{
		Type:      types.StringType,
		Data:      "gone",
		ExpiresAt: time.Now().Add(-1 * time.Millisecond),
	})

	// Act
	_, err := s.Get("expiring")

	// Assert
	if !errors.Is(err, ErrKeyExpired) {
		t.Fatalf("Get() returned %v; want ErrKeyExpired", err)
	}

	// Verify lazy deletion removed the key
	if s.Exists("expiring") {
		t.Fatalf("Get() should have lazily deleted the expired key")
	}
}

func TestGetDeletesExpiredKeyLazily(t *testing.T) {
	// Arrange
	s := NewMemoryStore()
	s.Set("ttl", types.Value{
		Type:      types.StringType,
		Data:      "temporary",
		ExpiresAt: time.Now().Add(50 * time.Millisecond),
	})

	// Assert key is alive
	if _, err := s.Get("ttl"); err != nil {
		t.Fatalf("Get() before expiry returned unexpected error: %v", err)
	}

	// Wait for expiration
	time.Sleep(60 * time.Millisecond)

	// Assert key is expired
	if _, err := s.Get("ttl"); !errors.Is(err, ErrKeyExpired) {
		t.Fatalf("Get() after expiry returned %v; want ErrKeyExpired", err)
	}

	// Assert lazy deletion occurred
	if s.Exists("ttl") {
		t.Fatalf("Expired key should have been lazily deleted")
	}
}

func TestSaveSkipsExpiredKeys(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	snapshotFile := filepath.Join(tempDir, "snap.json")

	s := NewMemoryStore()
	s.Set("alive", types.Value{Type: types.StringType, Data: "live"})
	s.Set("dead", types.Value{
		Type:      types.StringType,
		Data:      "expired",
		ExpiresAt: time.Now().Add(-1 * time.Millisecond),
	})

	// Act
	if err := s.Save(snapshotFile); err != nil {
		t.Fatalf("Save() returned unexpected error: %v", err)
	}

	// Reload into a fresh store and verify expired key is absent
	s2 := NewMemoryStore()
	if err := s2.Load(snapshotFile); err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}

	if val, _ := s2.Get("alive"); val.Data != "live" {
		t.Fatalf("Load() missing persistent key 'alive'")
	}
	if s2.Exists("dead") {
		t.Fatalf("Load() should not have restored expired key 'dead'")
	}
}

func TestLoadIgnoresExpiredEntries(t *testing.T) {
	// Arrange — bypass Save()'s filter by writing directly to the internal map,
	// then persist manually. This tests that Load() provides a second line of
	// defence for snapshots written by older builds that lacked the filter.
	tempDir := t.TempDir()
	snapshotFile := filepath.Join(tempDir, "snap.json")

	s1 := NewMemoryStore()
	// Directly inject an expired entry into the internal map so it appears in
	// the raw snapshot bytes, simulating a snapshot from a pre-Sprint-10 build.
	s1.data["fresh"] = types.Value{Type: types.StringType, Data: "keep"}
	s1.data["stale"] = types.Value{
		Type:      types.StringType,
		Data:      "remove",
		ExpiresAt: time.Now().Add(-1 * time.Second),
	}

	// Write raw JSON directly without the expiration filter.
	raw, _ := json.Marshal(s1.data)
	_ = os.WriteFile(snapshotFile, raw, 0o644)

	// Act — Load must filter the already-expired key.
	s2 := NewMemoryStore()
	if err := s2.Load(snapshotFile); err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}

	// Assert
	if val, _ := s2.Get("fresh"); val.Data != "keep" {
		t.Fatalf("Load() missing persistent key 'fresh'")
	}
	if s2.Exists("stale") {
		t.Fatalf("Load() should have filtered expired key 'stale'")
	}
}

func TestPersistentKeysSurviveSaveAndLoad(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	snapshotFile := filepath.Join(tempDir, "snap.json")

	s1 := NewMemoryStore()
	s1.Set("name", types.Value{Type: types.StringType, Data: "Rajas"})
	s1.Set("city", types.Value{Type: types.StringType, Data: "Pune"})

	// Act
	if err := s1.Save(snapshotFile); err != nil {
		t.Fatalf("Save() returned unexpected error: %v", err)
	}

	s2 := NewMemoryStore()
	if err := s2.Load(snapshotFile); err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}

	// Assert
	if val, _ := s2.Get("name"); val.Data != "Rajas" {
		t.Fatalf("Load() name = %v; want %q", val.Data, "Rajas")
	}
	if val, _ := s2.Get("city"); val.Data != "Pune" {
		t.Fatalf("Load() city = %v; want %q", val.Data, "Pune")
	}
}
