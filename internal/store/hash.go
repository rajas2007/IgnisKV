package store

import (
	"github.com/rajas2007/IgnisKV/internal/types"
)

// HSet sets the specified fields to their respective values in the hash stored at key.
// If key does not exist, a new key holding a hash is created with a zero ExpiresAt.
// If the key already exists and is a hash, its ExpiresAt is preserved.
// If the key exists but is not a hash, ErrWrongType is returned.
// It returns the number of fields that were added (not updated).
// The pairs argument must contain an even number of elements (field, value, field, value...).
func (s *MemoryStore) HSet(key string, pairs []string) (int, error) {
	if len(pairs)%2 != 0 || len(pairs) == 0 {
		return 0, ErrInvalidArguments
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.data[key]
	if ok && isExpired(v) {
		s.deleteExpiredLocked(key)
		ok = false
	}

	if !ok {
		v = types.Value{
			Type: types.HashType,
			Data: make(map[string]string),
		}
	} else if v.Type != types.HashType {
		return 0, ErrWrongType
	}

	// Safe type assertion: v.Type == types.HashType guarantees that v.Data
	// holds a map[string]string without any wrapper structs.
	hashMap := v.Data.(map[string]string)
	added := 0

	for i := 0; i < len(pairs); i += 2 {
		field := pairs[i]
		value := pairs[i+1]

		if _, exists := hashMap[field]; !exists {
			added++
		}
		hashMap[field] = value
	}

	v.Data = hashMap
	s.data[key] = v

	return added, nil
}

// HGet returns the value associated with field in the hash stored at key.
// It returns ErrKeyNotFound if the key does not exist.
// It returns ErrWrongType if the key exists but is not a hash.
// It returns ErrFieldNotFound if the key exists but the field does not.
func (s *MemoryStore) HGet(key, field string) (string, error) {
	s.mu.RLock()
	v, ok := s.data[key]

	if !ok {
		s.mu.RUnlock()
		return "", ErrKeyNotFound
	}

	if isExpired(v) {
		s.mu.RUnlock()
		s.lazyExpire(key)
		return "", ErrKeyNotFound
	}

	if v.Type != types.HashType {
		s.mu.RUnlock()
		return "", ErrWrongType
	}

	hashMap := v.Data.(map[string]string)
	val, ok := hashMap[field]
	s.mu.RUnlock()

	if !ok {
		return "", ErrFieldNotFound
	}

	return val, nil
}
