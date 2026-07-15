package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
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

func TestActiveExpirationDeletesKeys(t *testing.T) {
	// Arrange — use a very aggressive cleanup interval for testing
	s := newMemoryStoreWithInterval(10 * time.Millisecond)

	s.Set("ephemeral", types.Value{
		Type:      types.StringType,
		Data:      "short-lived",
		ExpiresAt: time.Now().Add(50 * time.Millisecond),
	})

	// Assert key exists physically initially
	if !s.Exists("ephemeral") {
		t.Fatalf("expected key to exist physically before expiration")
	}

	// Act — Wait for the key to expire and the cleanup ticker to fire a few times
	time.Sleep(100 * time.Millisecond)

	// Assert key has been deleted physically by the background goroutine
	// without any Get() call triggering lazy deletion.
	if s.Exists("ephemeral") {
		t.Fatalf("expected key to be physically deleted by background cleanup")
	}
}

func TestConcurrentAccessDuringCleanup(t *testing.T) {
	// Arrange — aggressively run cleanup to maximize contention
	s := newMemoryStoreWithInterval(1 * time.Millisecond)

	// Add a persistent key
	s.Set("persistent", types.Value{Type: types.StringType, Data: "forever"})

	// Act — Run a tight loop of Sets and Gets while the cleanup goroutine
	// continuously acquires the write lock and scans the keyspace.
	done := make(chan struct{})
	go func() {
		for i := 0; i < 1000; i++ {
			s.Set("temp", types.Value{
				Type:      types.StringType,
				Data:      "temp",
				ExpiresAt: time.Now().Add(1 * time.Millisecond),
			})
			s.Get("persistent")
		}
		close(done)
	}()

	select {
	case <-done:
		// Success — no deadlocks or panics
	case <-time.After(2 * time.Second):
		t.Fatalf("Test timed out, possible deadlock")
	}
}

func TestTTL(t *testing.T) {
	s := NewMemoryStore()

	// 1. Missing key
	_, err := s.TTL("missing")
	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("expected ErrKeyNotFound for missing key, got %v", err)
	}

	// 2. Persistent key
	s.Set("persistent", types.Value{Type: types.StringType, Data: "forever"})
	ttl, err := s.TTL("persistent")
	if err != nil {
		t.Fatalf("expected nil error for persistent key, got %v", err)
	}
	if ttl != -1 {
		t.Fatalf("expected TTL -1 for persistent key, got %d", ttl)
	}

	// 3. Expiring key (future)
	s.Set("expiring", types.Value{
		Type:      types.StringType,
		Data:      "short-lived",
		ExpiresAt: time.Now().Add(5 * time.Second),
	})
	ttl, err = s.TTL("expiring")
	if err != nil {
		t.Fatalf("expected nil error for expiring key, got %v", err)
	}
	if ttl < 4 || ttl > 5 {
		t.Fatalf("expected TTL between 4 and 5 seconds, got %d", ttl)
	}

	// 4. Expired key (past)
	s.Set("expired", types.Value{
		Type:      types.StringType,
		Data:      "dead",
		ExpiresAt: time.Now().Add(-1 * time.Second), // Expired 1 second ago
	})
	ttl, err = s.TTL("expired")
	if !errors.Is(err, ErrKeyExpired) {
		t.Fatalf("expected ErrKeyExpired for expired key, got %v", err)
	}
	if ttl != 0 {
		t.Fatalf("expected TTL 0 for expired key, got %d", ttl)
	}
	// Verify lazy deletion occurred
	if s.Exists("expired") {
		t.Fatalf("expected expired key to be lazily deleted by TTL")
	}
}

func TestTTLPersistsAcrossSaveLoad(t *testing.T) {
	s1 := NewMemoryStore()

	s1.Set("expiring_key", types.Value{
		Type:      types.StringType,
		Data:      "some-data",
		ExpiresAt: time.Now().Add(5 * time.Second),
	})

	snapshotFile := filepath.Join(t.TempDir(), "ttl_persistence_test.json")

	if err := s1.Save(snapshotFile); err != nil {
		t.Fatalf("Save() returned unexpected error: %v", err)
	}

	s2 := NewMemoryStore()
	if err := s2.Load(snapshotFile); err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}

	ttl, err := s2.TTL("expiring_key")
	if err != nil {
		t.Fatalf("TTL() returned unexpected error: %v", err)
	}

	// 5 seconds rounded up might be exactly 5, but after a few ms it could still be 5 because of math.Ceil
	// Just ensure it is 4 or 5
	if ttl < 4 || ttl > 5 {
		t.Fatalf("expected TTL between 4 and 5 seconds, got %d", ttl)
	}
}

func TestExpire(t *testing.T) {
	s := NewMemoryStore()

	// 1. Missing key
	res, err := s.Expire("missing", 5)
	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("expected ErrKeyNotFound, got %v", err)
	}
	if res != 0 {
		t.Fatalf("expected 0 result for missing key, got %d", res)
	}

	// 2. Invalid duration
	s.Set("key1", types.Value{Type: types.StringType, Data: "val1"})
	res, err = s.Expire("key1", 0)
	if !errors.Is(err, ErrInvalidDuration) {
		t.Fatalf("expected ErrInvalidDuration for 0, got %v", err)
	}
	res, err = s.Expire("key1", -5)
	if !errors.Is(err, ErrInvalidDuration) {
		t.Fatalf("expected ErrInvalidDuration for negative, got %v", err)
	}

	// 3. Expire existing persistent key
	res, err = s.Expire("key1", 10)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if res != 1 {
		t.Fatalf("expected 1 result, got %d", res)
	}

	// Verify with TTL
	ttl, _ := s.TTL("key1")
	if ttl < 9 || ttl > 10 {
		t.Fatalf("expected TTL between 9 and 10, got %d", ttl)
	}

	// 4. Replace existing expiration
	res, err = s.Expire("key1", 20)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if res != 1 {
		t.Fatalf("expected 1 result, got %d", res)
	}

	ttl, _ = s.TTL("key1")
	if ttl < 19 || ttl > 20 {
		t.Fatalf("expected TTL between 19 and 20, got %d", ttl)
	}

	// 5. Expire already expired key
	s.Set("expired_key", types.Value{
		Type:      types.StringType,
		Data:      "val",
		ExpiresAt: time.Now().Add(-1 * time.Second),
	})

	res, err = s.Expire("expired_key", 5)
	if !errors.Is(err, ErrKeyExpired) {
		t.Fatalf("expected ErrKeyExpired, got %v", err)
	}
	if res != 0 {
		t.Fatalf("expected 0 result, got %d", res)
	}

	// Verify it was lazily deleted
	if s.Exists("expired_key") {
		t.Fatalf("expected expired_key to be physically deleted")
	}

	// 6. Expire followed by Get
	s.Set("key2", types.Value{Type: types.StringType, Data: "val2"})
	s.Expire("key2", 1)
	val, err := s.Get("key2")
	if err != nil {
		t.Fatalf("expected nil error on Get, got %v", err)
	}
	if val.Data != "val2" {
		t.Fatalf("expected val2, got %v", val.Data)
	}
}

func TestConcurrentExpire(t *testing.T) {
	s := newMemoryStoreWithInterval(1 * time.Hour) // no background cleanup needed
	s.Set("concurrent_key", types.Value{Type: types.StringType, Data: "val"})

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(dur int64) {
			defer wg.Done()
			if _, err := s.Expire("concurrent_key", dur); err != nil && !errors.Is(err, ErrKeyExpired) {
				t.Errorf("Concurrent Expire failed: %v", err)
			}
		}(int64((i % 5) + 1)) // Durations 1 to 5
	}
	wg.Wait()

	// Verify key still exists and has some TTL
	ttl, err := s.TTL("concurrent_key")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if ttl < 1 || ttl > 5 {
		t.Fatalf("expected TTL between 1 and 5, got %d", ttl)
	}
}

func TestExpirePersistsAcrossSaveLoad(t *testing.T) {
	s1 := NewMemoryStore()

	s1.Set("persistent_key", types.Value{
		Type: types.StringType,
		Data: "some-data",
	})

	// Set expiration
	res, err := s1.Expire("persistent_key", 5)
	if err != nil {
		t.Fatalf("Expire() returned unexpected error: %v", err)
	}
	if res != 1 {
		t.Fatalf("expected 1 result, got %d", res)
	}

	snapshotFile := filepath.Join(t.TempDir(), "expire_persistence_test.json")

	if err := s1.Save(snapshotFile); err != nil {
		t.Fatalf("Save() returned unexpected error: %v", err)
	}

	s2 := NewMemoryStore()
	if err := s2.Load(snapshotFile); err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}

	ttl, err := s2.TTL("persistent_key")
	if err != nil {
		t.Fatalf("TTL() returned unexpected error: %v", err)
	}

	if ttl < 4 || ttl > 5 {
		t.Fatalf("expected TTL between 4 and 5 seconds, got %d", ttl)
	}
}

func TestExpireTime(t *testing.T) {
	s := NewMemoryStore()

	// 1. Missing key
	ts, err := s.ExpireTime("missing")
	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("expected ErrKeyNotFound for missing key, got %v", err)
	}
	if ts != 0 {
		t.Fatalf("expected 0 for missing key, got %d", ts)
	}

	// 2. Persistent key
	s.Set("persistent", types.Value{Type: types.StringType, Data: "val"})
	ts, err = s.ExpireTime("persistent")
	if err != nil {
		t.Fatalf("expected nil error for persistent key, got %v", err)
	}
	if ts != -1 {
		t.Fatalf("expected -1 for persistent key, got %d", ts)
	}

	// 3. Expiring key
	future := time.Now().Add(5 * time.Second)
	s.Set("expiring", types.Value{
		Type:      types.StringType,
		Data:      "val",
		ExpiresAt: future,
	})

	ts, err = s.ExpireTime("expiring")
	if err != nil {
		t.Fatalf("expected nil error for expiring key, got %v", err)
	}
	if ts != future.Unix() {
		t.Fatalf("expected %d, got %d", future.Unix(), ts)
	}

	// 4. Expired key
	past := time.Now().Add(-5 * time.Second)
	s.Set("expired", types.Value{
		Type:      types.StringType,
		Data:      "val",
		ExpiresAt: past,
	})

	ts, err = s.ExpireTime("expired")
	if !errors.Is(err, ErrKeyExpired) {
		t.Fatalf("expected ErrKeyExpired for expired key, got %v", err)
	}
	if ts != 0 {
		t.Fatalf("expected 0 for expired key, got %d", ts)
	}
	// Verify lazy deletion occurred
	if s.Exists("expired") {
		t.Fatalf("expected expired key to be lazily deleted")
	}
}
