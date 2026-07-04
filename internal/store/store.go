package store

import (
	"sync"

	"github.com/rajas2007/IgnisKV/internal/types"
)

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
