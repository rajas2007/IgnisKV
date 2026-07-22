package store

import (
	"math"
	"strconv"

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

// HSetNX sets field in the hash stored at key to value, only if field does not yet exist.
// If key does not exist, a new key holding a hash is created.
// If field already exists, this operation has no effect.
// It returns true if the field was set, and false if it already existed.
// It returns ErrWrongType if the key exists but is not a hash.
func (s *MemoryStore) HSetNX(key, field, value string) (bool, error) {
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
		return false, ErrWrongType
	}

	hashMap := v.Data.(map[string]string)

	if _, exists := hashMap[field]; exists {
		return false, nil // Do not overwrite, no persistence needed
	}

	hashMap[field] = value
	v.Data = hashMap
	s.data[key] = v

	return true, nil
}

// HIncrBy increments the number stored at field in the hash stored at key by delta.
// If key does not exist, a new key holding a hash is created.
// If field does not exist the value is set to 0 before the operation is performed.
// Returns the value of field after the increment operation.
// Returns ErrWrongType if the key exists but is not a hash.
// Returns ErrNotInteger if the hash value cannot be parsed as an integer.
// Returns ErrOverflow if the operation would overflow a 64-bit signed integer.
func (s *MemoryStore) HIncrBy(key, field string, delta int64) (int64, error) {
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

	hashMap := v.Data.(map[string]string)

	var current int64 = 0
	valStr, exists := hashMap[field]
	if exists {
		parsed, err := strconv.ParseInt(valStr, 10, 64)
		if err != nil {
			return 0, ErrNotInteger
		}
		current = parsed
	}

	// Detect overflow before assignment
	if (delta > 0 && current > math.MaxInt64-delta) || (delta < 0 && current < math.MinInt64-delta) {
		return 0, ErrOverflow
	}

	newValue := current + delta
	hashMap[field] = strconv.FormatInt(newValue, 10)
	v.Data = hashMap
	s.data[key] = v

	return newValue, nil
}

// HIncrByFloat increments the specified field of a hash stored at key, and representing a floating point number, by the specified increment.
// If the key does not exist, a new key holding a hash is created.
// If the field does not exist, it is set to 0 before the operation is performed.
// Returns the value of field after the increment operation.
// Returns ErrWrongType if the key exists but is not a hash.
// Returns ErrNotFloat if the hash value cannot be parsed as a float.
// Returns ErrNaNOrInfinity if the operation would produce NaN or Infinity.
func (s *MemoryStore) HIncrByFloat(key, field string, delta float64) (float64, error) {
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

	hashMap := v.Data.(map[string]string)

	var current float64 = 0
	valStr, exists := hashMap[field]
	if exists {
		parsed, err := strconv.ParseFloat(valStr, 64)
		if err != nil {
			return 0, ErrNotFloat
		}
		current = parsed
	}

	newValue := current + delta

	if math.IsNaN(newValue) || math.IsInf(newValue, 0) {
		return 0, ErrNaNOrInfinity
	}

	// Format matching Redis (which removes trailing zeros and often uses %g)
	// Go's strconv.FormatFloat with 'f', -1, 64 handles this, but some specific formatting might be needed.
	// Actually, Redis drops trailing zeros. FormatFloat does this with 'f', -1.
	// Wait, 'f' can result in very long strings. 'g' is more standard for Redis float responses, but let's use 'f' with trailing zero stripping which is what Redis does.
	// Actually, `strconv.FormatFloat(newValue, 'f', -1, 64)` doesn't output exponential notation for small/large numbers by default but might.
	// We'll use 'f', -1, 64 which drops trailing zeros naturally.
	strValue := strconv.FormatFloat(newValue, 'f', -1, 64)

	hashMap[field] = strValue
	v.Data = hashMap
	s.data[key] = v

	return newValue, nil
}
