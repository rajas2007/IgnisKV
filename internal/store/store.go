package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/rajas2007/IgnisKV/internal/types"
)

// DefaultSnapshotFile is the standard filename used for JSON persistence.
const DefaultSnapshotFile = "igniskv.json"

const snapshotFilePermission = 0o644

// MemoryStore is the core in-memory storage engine of IgnisKV.
// It manages the full database keyspace and ensures that all concurrent
// reads and writes are safely coordinated. It is the single authoritative
// owner of stored data and must be accessed exclusively through its methods.
type MemoryStore struct {
	// mu guards all access to the data map, allowing concurrent reads while
	// serialising writes to prevent data races.
	mu sync.RWMutex

	// data is the internal keyspace, mapping every database key to its
	// associated value record. It must never be accessed directly from
	// outside this package.
	data map[string]types.Value
}

// NewMemoryStore allocates and initialises a new MemoryStore with an empty
// keyspace ready to accept database operations. It automatically starts the
// background cleanup goroutine for active expiration.
func NewMemoryStore() *MemoryStore {
	return newMemoryStoreWithInterval(5 * time.Minute)
}

// newMemoryStoreWithInterval allows tests to instantiate a MemoryStore with
// a short cleanup interval without mutating global state.
func newMemoryStoreWithInterval(interval time.Duration) *MemoryStore {
	s := &MemoryStore{
		data: make(map[string]types.Value),
	}
	go s.startCleanupGoroutine(interval)
	return s
}

// startCleanupGoroutine runs a continuous loop that periodically scans the
// entire keyspace and deletes expired keys. It runs for the lifetime of the
// application and provides eventual cleanup of keys that are never accessed.
//
// Sprint 11: The entire scan runs under a single write lock to prioritize
// correctness and simplicity over scalability.
func (s *MemoryStore) startCleanupGoroutine(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		s.mu.Lock()
		for key, value := range s.data {
			if isExpired(value) {
				delete(s.data, key)
			}
		}
		s.mu.Unlock()
	}
}

// Save serializes the complete logical database state to the specified file
// as a JSON snapshot. It acquires a read lock to block concurrent writers
// while allowing concurrent readers to continue.
//
// Sprint 10: Save skips expired keys. This prevents expired data from being
// revived after an application restart and keeps snapshots representative of
// the logical database state rather than the physical in-memory map contents.
func (s *MemoryStore) Save(filename string) error {
	s.mu.RLock()
	live := make(map[string]types.Value, len(s.data))
	for k, v := range s.data {
		if !isExpired(v) {
			live[k] = v
		}
	}
	s.mu.RUnlock()

	data, err := json.Marshal(live)
	if err != nil {
		return fmt.Errorf("failed to serialize database state: %w", err)
	}

	dir := filepath.Dir(filename)
	base := filepath.Base(filename)
	tmpFile, err := os.CreateTemp(dir, base+".*.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tempFilename := tmpFile.Name()

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		os.Remove(tempFilename)
		return fmt.Errorf("failed to write snapshot file: %w", err)
	}
	tmpFile.Close()

	if err := os.Chmod(tempFilename, snapshotFilePermission); err != nil {
		os.Remove(tempFilename)
		return fmt.Errorf("failed to set snapshot permissions: %w", err)
	}

	// Retry rename to handle Windows concurrent "Access is denied" errors
	// caused by multiple threads replacing the file simultaneously.
	var renameErr error
	for i := 0; i < 50; i++ {
		renameErr = os.Rename(tempFilename, filename)
		if renameErr == nil {
			break
		}
		time.Sleep(1 * time.Millisecond)
	}

	if renameErr != nil {
		os.Remove(tempFilename)
		return fmt.Errorf("failed to commit snapshot file: %w", renameErr)
	}

	return nil
}

// Load restores the database state from the specified JSON snapshot file.
// If the file does not exist, it returns nil, treating it as a first-run scenario.
// It acquires a write lock to ensure no client observes partially restored data.
func (s *MemoryStore) Load(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil // Missing file is an expected first-run scenario
		}
		return fmt.Errorf("failed to read snapshot file: %w", err)
	}

	var parsedData map[string]types.Value
	if err := json.Unmarshal(data, &parsedData); err != nil {
		return fmt.Errorf("failed to deserialize snapshot: %w", err)
	}

	if parsedData == nil {
		parsedData = make(map[string]types.Value)
	}

	// Filter expired entries discovered inside the snapshot.
	// A key may have expired between the time the snapshot was written
	// and the time it is being restored.
	for k, v := range parsedData {
		if isExpired(v) {
			delete(parsedData, k)
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = parsedData

	return nil
}
