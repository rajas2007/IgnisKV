package store

import "github.com/rajas2007/IgnisKV/internal/types"

// Set stores a value under the given key in the keyspace. If the key already
// exists its value is overwritten.
func (s *MemoryStore) Set(key string, value types.Value) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[key] = value
}

// Get retrieves the value associated with the given key. It returns
// ErrKeyNotFound if the key does not exist in the keyspace.
//
// Sprint 10: Get performs lazy expiration. If the key is found but has
// passed its expiration deadline, the key is deleted and ErrKeyExpired
// is returned. A double-check pattern is used to avoid deleting a value
// that was updated by another goroutine between the read and write locks.
func (s *MemoryStore) Get(key string) (types.Value, error) {
	s.mu.RLock()
	v, ok := s.data[key]
	s.mu.RUnlock()

	if !ok {
		return types.Value{}, ErrKeyNotFound
	}

	if isExpired(v) {
		// Re-acquire as a write lock and re-verify before deleting.
		// Another goroutine may have overwritten the key in the window
		// between releasing the read lock and acquiring the write lock.
		s.mu.Lock()
		current, ok := s.data[key]
		if ok && isExpired(current) {
			delete(s.data, key)
		}
		s.mu.Unlock()

		return types.Value{}, ErrKeyExpired
	}

	return v, nil
}

// Delete removes the given key from the keyspace. It returns ErrKeyNotFound
// if the key does not exist.
func (s *MemoryStore) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.data[key]; !ok {
		return ErrKeyNotFound
	}

	delete(s.data, key)

	return nil
}

// Exists reports whether the given key is physically present in the in-memory
// map, regardless of its expiration status.
//
// Sprint 10 intentionally does not perform expiration checks inside Exists.
// A key that has passed its ExpiresAt deadline may still return true until
// it is discovered and lazily deleted by a subsequent Get call. This behavior
// is intentional and not a bug. Expiration-aware existence checks will be
// introduced in a future expiration milestone.
func (s *MemoryStore) Exists(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, ok := s.data[key]

	return ok
}
