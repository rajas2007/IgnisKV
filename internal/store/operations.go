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
func (s *MemoryStore) Get(key string) (types.Value, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	v, ok := s.data[key]
	if !ok {
		return types.Value{}, ErrKeyNotFound
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

// Exists reports whether the given key is present in the keyspace.
func (s *MemoryStore) Exists(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, ok := s.data[key]

	return ok
}
