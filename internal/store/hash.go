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
