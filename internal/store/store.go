package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"

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
// keyspace ready to accept database operations.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		data: make(map[string]types.Value),
	}
}

// Save serializes the complete logical database state to the specified file
// as a JSON snapshot. It acquires a read lock to block concurrent writers
// while allowing concurrent readers to continue.
func (s *MemoryStore) Save(filename string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := json.Marshal(s.data)
	if err != nil {
		return fmt.Errorf("failed to serialize database state: %w", err)
	}

	tempFilename := filename + ".tmp"
	if err := os.WriteFile(tempFilename, data, snapshotFilePermission); err != nil {
		return fmt.Errorf("failed to write snapshot file: %w", err)
	}

	if err := os.Rename(tempFilename, filename); err != nil {
		return fmt.Errorf("failed to commit snapshot file: %w", err)
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

	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = parsedData

	return nil
}
