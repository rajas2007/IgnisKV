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

// HExists reports whether the specified field exists in the hash stored at key.
// It returns false and ErrKeyNotFound if the key does not exist.
// It returns false and ErrWrongType if the key exists but is not a hash.
// It returns false and nil if the key exists but the field does not.
func (s *MemoryStore) HExists(key, field string) (bool, error) {
	s.mu.RLock()
	v, ok := s.data[key]

	if !ok {
		s.mu.RUnlock()
		return false, ErrKeyNotFound
	}

	if isExpired(v) {
		s.mu.RUnlock()
		s.lazyExpire(key)
		return false, ErrKeyNotFound
	}

	if v.Type != types.HashType {
		s.mu.RUnlock()
		return false, ErrWrongType
	}

	hashMap := v.Data.(map[string]string)
	_, exists := hashMap[field]
	s.mu.RUnlock()

	return exists, nil
}

// HDel removes the specified fields from the hash stored at key.
// It returns the number of fields that were removed from the hash, not including
// specified but non-existing fields.
// If the key does not exist, it returns 0.
// If the key exists but is not a hash, it returns ErrWrongType.
// If the hash becomes empty after deletion, the key is deleted from the store.
func (s *MemoryStore) HDel(key string, fields []string) (int, error) {
	if len(fields) == 0 {
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
		return 0, nil
	}

	if v.Type != types.HashType {
		return 0, ErrWrongType
	}

	hashMap := v.Data.(map[string]string)
	deleted := 0

	for _, field := range fields {
		if _, exists := hashMap[field]; exists {
			delete(hashMap, field)
			deleted++
		}
	}

	if len(hashMap) == 0 {
		delete(s.data, key)
	} else {
		// Update map
		v.Data = hashMap
		s.data[key] = v
	}

	return deleted, nil
}

// HLen returns the number of fields contained in the hash stored at key.
// If the key does not exist, it returns 0.
// If the key exists but is not a hash, it returns ErrWrongType.
func (s *MemoryStore) HLen(key string) (int, error) {
	s.mu.RLock()
	v, ok := s.data[key]
	if !ok {
		s.mu.RUnlock()
		return 0, nil
	}

	if isExpired(v) {
		s.mu.RUnlock()
		s.lazyExpire(key)
		return 0, nil
	}

	if v.Type != types.HashType {
		s.mu.RUnlock()
		return 0, ErrWrongType
	}

	hashMap := v.Data.(map[string]string)
	length := len(hashMap)
	s.mu.RUnlock()

	return length, nil
}

// HGetAll returns all fields and values of the hash stored at key as a flat
// slice of alternating field/value pairs: [field1, value1, field2, value2, ...].
// If the key does not exist, it returns an empty slice.
// If the key exists but is not a hash, it returns ErrWrongType.
func (s *MemoryStore) HGetAll(key string) ([]string, error) {
	s.mu.RLock()
	v, ok := s.data[key]
	if !ok {
		s.mu.RUnlock()
		return []string{}, nil
	}

	if isExpired(v) {
		s.mu.RUnlock()
		s.lazyExpire(key)
		return []string{}, nil
	}

	if v.Type != types.HashType {
		s.mu.RUnlock()
		return nil, ErrWrongType
	}

	hashMap := v.Data.(map[string]string)
	result := make([]string, 0, len(hashMap)*2)
	for field, value := range hashMap {
		result = append(result, field, value)
	}
	s.mu.RUnlock()

	return result, nil
}

// HMGet returns the values associated with the specified fields in the hash
// stored at key. For every field that does not exist in the hash, nil is
// returned in that position. If the key does not exist, a slice of nil values
// is returned whose length equals len(fields).
// If the key exists but is not a hash, it returns ErrWrongType.
// The returned slice preserves the order of the requested fields exactly.
func (s *MemoryStore) HMGet(key string, fields []string) ([]any, error) {
	s.mu.RLock()
	v, ok := s.data[key]
	if !ok {
		s.mu.RUnlock()
		result := make([]any, len(fields))
		return result, nil
	}

	if isExpired(v) {
		s.mu.RUnlock()
		s.lazyExpire(key)
		result := make([]any, len(fields))
		return result, nil
	}

	if v.Type != types.HashType {
		s.mu.RUnlock()
		return nil, ErrWrongType
	}

	hashMap := v.Data.(map[string]string)
	result := make([]any, len(fields))
	for i, field := range fields {
		if val, exists := hashMap[field]; exists {
			result[i] = val
		}
		// Missing fields remain nil (zero value of any).
	}
	s.mu.RUnlock()

	return result, nil
}

// HKeys returns all field names in the hash stored at key.
// If the key does not exist, it returns an empty slice.
// If the key exists but is not a hash, it returns ErrWrongType.
// The order of the returned field names is unspecified.
func (s *MemoryStore) HKeys(key string) ([]string, error) {
	s.mu.RLock()
	v, ok := s.data[key]
	if !ok {
		s.mu.RUnlock()
		return []string{}, nil
	}

	if isExpired(v) {
		s.mu.RUnlock()
		s.lazyExpire(key)
		return []string{}, nil
	}

	if v.Type != types.HashType {
		s.mu.RUnlock()
		return nil, ErrWrongType
	}

	hashMap := v.Data.(map[string]string)
	keys := make([]string, 0, len(hashMap))
	for field := range hashMap {
		keys = append(keys, field)
	}
	s.mu.RUnlock()

	return keys, nil
}

// HVals returns all values in the hash stored at key.
// If the key does not exist, it returns an empty slice.
// If the key exists but is not a hash, it returns ErrWrongType.
// The order of the returned values is unspecified.
func (s *MemoryStore) HVals(key string) ([]string, error) {
	s.mu.RLock()
	v, ok := s.data[key]
	if !ok {
		s.mu.RUnlock()
		return []string{}, nil
	}

	if isExpired(v) {
		s.mu.RUnlock()
		s.lazyExpire(key)
		return []string{}, nil
	}

	if v.Type != types.HashType {
		s.mu.RUnlock()
		return nil, ErrWrongType
	}

	hashMap := v.Data.(map[string]string)
	vals := make([]string, 0, len(hashMap))
	for _, val := range hashMap {
		vals = append(vals, val)
	}
	s.mu.RUnlock()

	return vals, nil
}

// HStrLen returns the string length of the value associated with field in the
// hash stored at key. If the key or the field do not exist, 0 is returned.
// If the key exists but is not a hash, it returns ErrWrongType.
// The length is calculated in bytes using Go's native len() function.
func (s *MemoryStore) HStrLen(key, field string) (int, error) {
	s.mu.RLock()
	v, ok := s.data[key]
	if !ok {
		s.mu.RUnlock()
		return 0, nil
	}

	if isExpired(v) {
		s.mu.RUnlock()
		s.lazyExpire(key)
		return 0, nil
	}

	if v.Type != types.HashType {
		s.mu.RUnlock()
		return 0, ErrWrongType
	}

	hashMap := v.Data.(map[string]string)
	val, exists := hashMap[field]
	s.mu.RUnlock()

	if !exists {
		return 0, nil
	}

	return len(val), nil
}
