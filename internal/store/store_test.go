package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
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

func TestPExpireTime(t *testing.T) {
	s := NewMemoryStore()

	// 1. Missing key
	ts, err := s.PExpireTime("missing")
	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("expected ErrKeyNotFound for missing key, got %v", err)
	}
	if ts != 0 {
		t.Fatalf("expected 0 for missing key, got %d", ts)
	}

	// 2. Persistent key
	s.Set("persistent", types.Value{Type: types.StringType, Data: "val"})
	ts, err = s.PExpireTime("persistent")
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

	ts, err = s.PExpireTime("expiring")
	if err != nil {
		t.Fatalf("expected nil error for expiring key, got %v", err)
	}
	if ts != future.UnixMilli() {
		t.Fatalf("expected %d, got %d", future.UnixMilli(), ts)
	}

	// 4. Expired key
	past := time.Now().Add(-5 * time.Second)
	s.Set("expired", types.Value{
		Type:      types.StringType,
		Data:      "val",
		ExpiresAt: past,
	})

	ts, err = s.PExpireTime("expired")
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

func TestLPush(t *testing.T) {
	s := NewMemoryStore()

	// 1. Missing key (creates new list)
	length, err := s.LPush("mylist", "a")
	if err != nil {
		t.Fatalf("expected nil error for missing key, got %v", err)
	}
	if length != 1 {
		t.Fatalf("expected length 1, got %d", length)
	}
	val, err := s.Get("mylist")
	if err != nil || val.Type != types.ListType || len(val.Data.([]string)) != 1 || val.Data.([]string)[0] != "a" {
		t.Fatalf("expected list [a], got %v", val.Data)
	}

	// 2. Existing list (multiple values, prepended left-to-right)
	// LPUSH mylist b c -> should result in [c b a]
	length, err = s.LPush("mylist", "b", "c")
	if err != nil {
		t.Fatalf("expected nil error for existing list, got %v", err)
	}
	if length != 3 {
		t.Fatalf("expected length 3, got %d", length)
	}
	val, _ = s.Get("mylist")
	list := val.Data.([]string)
	if len(list) != 3 || list[0] != "c" || list[1] != "b" || list[2] != "a" {
		t.Fatalf("expected list [c b a], got %v", list)
	}

	// 2.1 Multiple values insert ordering regression test
	length, err = s.LPush("mylist", "d", "e", "f")
	if err != nil {
		t.Fatalf("expected nil error for existing list, got %v", err)
	}
	if length != 6 {
		t.Fatalf("expected length 6, got %d", length)
	}
	val, _ = s.Get("mylist")
	list = val.Data.([]string)
	if len(list) != 6 || list[0] != "f" || list[1] != "e" || list[2] != "d" || list[3] != "c" || list[4] != "b" || list[5] != "a" {
		t.Fatalf("expected list [f e d c b a], got %v", list)
	}

	// 3. WRONGTYPE
	s.Set("mystring", types.Value{Type: types.StringType, Data: "val"})
	length, err = s.LPush("mystring", "a")
	if !errors.Is(err, ErrWrongType) {
		t.Fatalf("expected ErrWrongType, got %v", err)
	}
	if length != 0 {
		t.Fatalf("expected length 0 on error, got %d", length)
	}

	// 4. Expired list recreation
	past := time.Now().Add(-1 * time.Second)
	s.Set("expiredlist", types.Value{
		Type:      types.ListType,
		Data:      []string{"old"},
		ExpiresAt: past,
	})
	length, err = s.LPush("expiredlist", "new")
	if err != nil {
		t.Fatalf("expected nil error after expiring key, got %v", err)
	}
	if length != 1 {
		t.Fatalf("expected length 1 after recreation, got %d", length)
	}
	val, _ = s.Get("expiredlist")
	list = val.Data.([]string)
	if len(list) != 1 || list[0] != "new" {
		t.Fatalf("expected list [new], got %v", list)
	}

	// 5. No values
	length, err = s.LPush("mylist")
	if !errors.Is(err, ErrInvalidArguments) {
		t.Fatalf("expected ErrInvalidArguments, got %v", err)
	}
	if length != 0 {
		t.Fatalf("expected length 0, got %d", length)
	}
}

func TestRPush(t *testing.T) {
	s := NewMemoryStore()

	// 1. Missing key creates new list
	length, err := s.RPush("rlist", "a")
	if err != nil {
		t.Fatalf("expected nil error for missing key, got %v", err)
	}
	if length != 1 {
		t.Fatalf("expected length 1, got %d", length)
	}
	val, _ := s.Get("rlist")
	if val.Type != types.ListType || len(val.Data.([]string)) != 1 || val.Data.([]string)[0] != "a" {
		t.Fatalf("expected list [a], got %v", val.Data)
	}

	// 2. Existing list (appends values)
	length, err = s.RPush("rlist", "b")
	if err != nil {
		t.Fatalf("expected nil error for existing list, got %v", err)
	}
	if length != 2 {
		t.Fatalf("expected length 2, got %d", length)
	}
	val, _ = s.Get("rlist")
	list := val.Data.([]string)
	if len(list) != 2 || list[0] != "a" || list[1] != "b" {
		t.Fatalf("expected list [a b], got %v", list)
	}

	// 3. Multiple-value ordering
	length, err = s.RPush("rlist", "c", "d", "e")
	if err != nil {
		t.Fatalf("expected nil error for multiple values, got %v", err)
	}
	if length != 5 {
		t.Fatalf("expected length 5, got %d", length)
	}
	val, _ = s.Get("rlist")
	list = val.Data.([]string)
	if len(list) != 5 || list[0] != "a" || list[1] != "b" || list[2] != "c" || list[3] != "d" || list[4] != "e" {
		t.Fatalf("expected list [a b c d e], got %v", list)
	}

	// 3.1 Append again to protect against replacement regression
	length, err = s.RPush("rlist", "f", "g")
	if err != nil {
		t.Fatalf("expected nil error for subsequent multiple values, got %v", err)
	}
	if length != 7 {
		t.Fatalf("expected length 7, got %d", length)
	}
	val, _ = s.Get("rlist")
	list = val.Data.([]string)
	if len(list) != 7 || list[0] != "a" || list[1] != "b" || list[2] != "c" || list[3] != "d" || list[4] != "e" || list[5] != "f" || list[6] != "g" {
		t.Fatalf("expected list [a b c d e f g], got %v", list)
	}

	// 4. WRONGTYPE error
	s.Set("stringkey", types.Value{Type: types.StringType, Data: "val"})
	length, err = s.RPush("stringkey", "x")
	if !errors.Is(err, ErrWrongType) {
		t.Fatalf("expected ErrWrongType, got %v", err)
	}
	if length != 0 {
		t.Fatalf("expected length 0 on error, got %d", length)
	}

	// 5. Expired list recreation
	past := time.Now().Add(-1 * time.Second)
	s.Set("expiredlist", types.Value{
		Type:      types.ListType,
		Data:      []string{"old"},
		ExpiresAt: past,
	})
	length, err = s.RPush("expiredlist", "new")
	if err != nil {
		t.Fatalf("expected nil error after expiring key, got %v", err)
	}
	if length != 1 {
		t.Fatalf("expected length 1 after recreation, got %d", length)
	}
	val, _ = s.Get("expiredlist")
	list = val.Data.([]string)
	if len(list) != 1 || list[0] != "new" {
		t.Fatalf("expected list [new], got %v", list)
	}

	// 6. No values supplied
	length, err = s.RPush("rlist")
	if !errors.Is(err, ErrInvalidArguments) {
		t.Fatalf("expected ErrInvalidArguments, got %v", err)
	}
	if length != 0 {
		t.Fatalf("expected length 0, got %d", length)
	}
}

func TestLLen(t *testing.T) {
	s := NewMemoryStore()

	// 1. Missing key
	length, err := s.LLen("missing")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if length != 0 {
		t.Fatalf("expected length 0, got %d", length)
	}

	// 2. Existing single-element list
	s.Set("single", types.Value{
		Type: types.ListType,
		Data: []string{"a"},
	})
	length, err = s.LLen("single")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if length != 1 {
		t.Fatalf("expected length 1, got %d", length)
	}

	// 3. Existing multi-element list
	s.Set("multi", types.Value{
		Type: types.ListType,
		Data: []string{"a", "b", "c"},
	})
	length, err = s.LLen("multi")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if length != 3 {
		t.Fatalf("expected length 3, got %d", length)
	}

	// 4. WRONGTYPE error
	s.Set("stringkey", types.Value{
		Type: types.StringType,
		Data: "val",
	})
	length, err = s.LLen("stringkey")
	if !errors.Is(err, ErrWrongType) {
		t.Fatalf("expected ErrWrongType, got %v", err)
	}
	if length != 0 {
		t.Fatalf("expected length 0 on error, got %d", length)
	}

	// 5. Expired list
	past := time.Now().Add(-1 * time.Second)
	s.Set("expiredlist", types.Value{
		Type:      types.ListType,
		Data:      []string{"a", "b"},
		ExpiresAt: past,
	})
	length, err = s.LLen("expiredlist")
	if err != nil {
		t.Fatalf("expected nil error after expiring key, got %v", err)
	}
	if length != 0 {
		t.Fatalf("expected length 0 after expiration, got %d", length)
	}
	// Verify the key was lazily deleted
	s.mu.RLock()
	_, ok := s.data["expiredlist"]
	s.mu.RUnlock()
	if ok {
		t.Fatalf("expected expired list to be lazily deleted")
	}

	// 6. Public API integration test
	_, _ = s.RPush("publiclist", "a", "b", "c")
	length, err = s.LLen("publiclist")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if length != 3 {
		t.Fatalf("expected length 3, got %d", length)
	}
}

func TestLRange(t *testing.T) {
	s := NewMemoryStore()

	// 1. Missing key
	res, err := s.LRange("missing", 0, -1)
	if err != nil {
		t.Fatalf("expected nil error for missing key, got %v", err)
	}
	if !reflect.DeepEqual(res, []string{}) {
		t.Fatalf("expected empty slice for missing key, got %v", res)
	}
	if res == nil {
		t.Fatalf("expected non-nil empty slice for missing key")
	}

	// Set up list for next tests
	s.Set("list", types.Value{
		Type: types.ListType,
		Data: []string{"a", "b", "c", "d", "e"},
	})

	// 2. Existing list, full range
	res, err = s.LRange("list", 0, -1)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	expected := []string{"a", "b", "c", "d", "e"}
	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("expected %v, got %v", expected, res)
	}

	// 3. Existing list, partial range
	res, err = s.LRange("list", 1, 3)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	expected = []string{"b", "c", "d"}
	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("expected %v, got %v", expected, res)
	}

	// 4. Single element
	res, err = s.LRange("list", 2, 2)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	expected = []string{"c"}
	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("expected %v, got %v", expected, res)
	}

	// 5. Negative indices
	res, err = s.LRange("list", -2, -1)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	expected = []string{"d", "e"}
	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("expected %v, got %v", expected, res)
	}

	// 6. Start greater than stop
	res, err = s.LRange("list", 4, 2)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !reflect.DeepEqual(res, []string{}) {
		t.Fatalf("expected empty slice, got %v", res)
	}

	// 7. Stop beyond length
	res, err = s.LRange("list", 2, 100)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	expected = []string{"c", "d", "e"}
	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("expected %v, got %v", expected, res)
	}

	// 8. Start beyond length
	res, err = s.LRange("list", 100, 200)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !reflect.DeepEqual(res, []string{}) {
		t.Fatalf("expected empty slice, got %v", res)
	}

	// 9. Very negative indices
	res, err = s.LRange("list", -100, -1)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	expected = []string{"a", "b", "c", "d", "e"}
	if !reflect.DeepEqual(res, expected) {
		t.Fatalf("expected %v, got %v", expected, res)
	}

	// 10. Wrong type
	s.Set("stringkey", types.Value{
		Type: types.StringType,
		Data: "val",
	})
	_, err = s.LRange("stringkey", 0, -1)
	if !errors.Is(err, ErrWrongType) {
		t.Fatalf("expected ErrWrongType, got %v", err)
	}

	// 11. Expired key
	past := time.Now().Add(-1 * time.Second)
	s.Set("expired", types.Value{
		Type:      types.ListType,
		Data:      []string{"a", "b"},
		ExpiresAt: past,
	})
	res, err = s.LRange("expired", 0, -1)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !reflect.DeepEqual(res, []string{}) {
		t.Fatalf("expected empty slice, got %v", res)
	}
	s.mu.RLock()
	_, ok := s.data["expired"]
	s.mu.RUnlock()
	if ok {
		t.Fatalf("expected expired key to be lazily deleted")
	}

	// 12. Returned slice ownership
	s.Set("owner", types.Value{
		Type: types.ListType,
		Data: []string{"a", "b", "c"},
	})
	res, _ = s.LRange("owner", 0, -1)
	res[0] = "modified"

	res2, _ := s.LRange("owner", 0, -1)
	expected = []string{"a", "b", "c"}
	if !reflect.DeepEqual(res2, expected) {
		t.Fatalf("slice modification affected store, expected %v but got %v", expected, res2)
	}
}
