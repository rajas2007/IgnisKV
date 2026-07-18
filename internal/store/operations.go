package store

import (
	"errors"
	"math"
	"time"

	"github.com/rajas2007/IgnisKV/internal/types"
)

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
		s.lazyExpire(key)
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

// TTL returns the remaining lifetime of the given key in whole seconds.
// It returns -1 if the key exists but has no associated expiration.
// It returns ErrKeyNotFound if the key does not exist.
//
// Sprint 12: TTL performs lazy expiration using the same check-then-act
// concurrency pattern established by Get(). If the key is found to be
// expired, it is deleted and ErrKeyExpired is returned.
func (s *MemoryStore) TTL(key string) (int64, error) {
	s.mu.RLock()
	v, ok := s.data[key]
	s.mu.RUnlock()

	if !ok {
		return 0, ErrKeyNotFound
	}

	if isExpired(v) {
		s.lazyExpire(key)
		return 0, ErrKeyExpired
	}

	if v.ExpiresAt.IsZero() {
		return -1, nil
	}

	remaining := v.ExpiresAt.Sub(time.Now())
	if remaining < 0 {
		remaining = 0
	}

	// Round up to ensure that a key with <1s remaining does not return 0,
	// which would misleadingly imply it has already expired.
	seconds := int64(math.Ceil(remaining.Seconds()))
	return seconds, nil
}

// PTTL returns the remaining lifetime of the given key in whole milliseconds.
// It returns -1 if the key exists but has no associated expiration.
// It returns ErrKeyNotFound if the key does not exist.
//
// Sprint 18: PTTL performs lazy expiration using the same check-then-act
// concurrency pattern established by TTL(). If the key is found to be
// expired, it is deleted and ErrKeyExpired is returned.
func (s *MemoryStore) PTTL(key string) (int64, error) {
	s.mu.RLock()
	v, ok := s.data[key]
	s.mu.RUnlock()

	if !ok {
		return 0, ErrKeyNotFound
	}

	if isExpired(v) {
		s.lazyExpire(key)
		return 0, ErrKeyExpired
	}

	if v.ExpiresAt.IsZero() {
		return -1, nil
	}

	remaining := v.ExpiresAt.Sub(time.Now())
	if remaining < 0 {
		remaining = 0
	}

	milliseconds := remaining.Milliseconds()
	return milliseconds, nil
}

// lazyExpire performs the check-then-act concurrency pattern to safely delete
// a key that has passed its expiration deadline. It acquires a write lock and
// re-verifies the key's state to prevent deleting a value that was updated by
// another goroutine between the read and write locks.
// It returns true if the key was deleted, false otherwise.
func (s *MemoryStore) lazyExpire(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	current, ok := s.data[key]
	if ok && isExpired(current) {
		s.deleteExpiredLocked(key)
		return true
	}
	return false
}

// deleteExpiredLocked removes the given key from the keyspace. It assumes
// the caller already holds the write lock and has verified the key is expired.
func (s *MemoryStore) deleteExpiredLocked(key string) {
	delete(s.data, key)
}

// Expire updates the expiration time of an existing key. It returns 1 if the
// expiration was successfully set. Otherwise returns 0 together with an
// appropriate Store error. Non-positive durations are rejected with
// ErrInvalidDuration.
//
// Sprint 13: Expire performs lazy expiration using the check-then-act
// concurrency pattern. It never modifies the stored value, only the ExpiresAt field.
func (s *MemoryStore) Expire(key string, seconds int64) (int64, error) {
	if seconds <= 0 {
		return 0, ErrInvalidDuration
	}

	s.mu.RLock()
	v, ok := s.data[key]
	s.mu.RUnlock()

	if !ok {
		return 0, ErrKeyNotFound
	}

	if isExpired(v) {
		s.lazyExpire(key)
		return 0, ErrKeyExpired
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Re-verify the key exists and hasn't expired since releasing the read lock
	current, ok := s.data[key]
	if !ok {
		return 0, ErrKeyNotFound
	}
	if isExpired(current) {
		s.deleteExpiredLocked(key)
		return 0, ErrKeyExpired
	}

	current.ExpiresAt = time.Now().Add(time.Duration(seconds) * time.Second)
	s.data[key] = current

	return 1, nil
}

// Persist removes the expiration from an existing key, converting it back into
// a persistent key. It returns 1 if an expiration was successfully removed.
// It returns 0 if the key does not exist, has already expired, or has no
// expiration to remove (idempotent). Non-existent keys return ErrKeyNotFound.
// Expired keys are lazily deleted and return ErrKeyExpired.
//
// Sprint 14: Persist performs lazy expiration using the check-then-act
// concurrency pattern. It never modifies the stored value, only the ExpiresAt field.
func (s *MemoryStore) Persist(key string) (int64, error) {
	s.mu.RLock()
	v, ok := s.data[key]
	s.mu.RUnlock()

	if !ok {
		return 0, ErrKeyNotFound
	}

	if isExpired(v) {
		s.lazyExpire(key)
		return 0, ErrKeyExpired
	}

	// Already persistent — no state change required
	if v.ExpiresAt.IsZero() {
		return 0, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Re-verify the key exists and hasn't expired since releasing the read lock
	current, ok := s.data[key]
	if !ok {
		return 0, ErrKeyNotFound
	}
	if isExpired(current) {
		s.deleteExpiredLocked(key)
		return 0, ErrKeyExpired
	}

	// Already persistent under write lock — idempotent
	if current.ExpiresAt.IsZero() {
		return 0, nil
	}

	current.ExpiresAt = time.Time{}
	s.data[key] = current

	return 1, nil
}

// ExpireAt updates the expiration time of an existing key to an absolute timestamp.
// It returns 1 if the expiration was successfully set. Otherwise returns 0 together
// with an appropriate Store error. Timestamps that are not in the future are rejected
// with ErrInvalidTimestamp.
//
// Sprint 15: ExpireAt performs lazy expiration using the check-then-act
// concurrency pattern. It never modifies the stored value, only the ExpiresAt field.
func (s *MemoryStore) ExpireAt(key string, t time.Time) (int64, error) {
	if !t.After(time.Now()) {
		return 0, ErrInvalidTimestamp
	}

	s.mu.RLock()
	v, ok := s.data[key]
	s.mu.RUnlock()

	if !ok {
		return 0, ErrKeyNotFound
	}

	if isExpired(v) {
		s.lazyExpire(key)
		return 0, ErrKeyExpired
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Re-verify the key exists and hasn't expired since releasing the read lock
	current, ok := s.data[key]
	if !ok {
		return 0, ErrKeyNotFound
	}
	if isExpired(current) {
		s.deleteExpiredLocked(key)
		return 0, ErrKeyExpired
	}

	current.ExpiresAt = t
	s.data[key] = current

	return 1, nil
}

// PExpire updates the expiration time of an existing key. It returns 1 if the
// expiration was successfully set. Otherwise returns 0 together with an
// appropriate Store error. Non-positive durations are rejected with
// ErrInvalidDuration.
//
// Sprint 16: PExpire performs lazy expiration using the check-then-act
// concurrency pattern. It never modifies the stored value, only the ExpiresAt field.
func (s *MemoryStore) PExpire(key string, d time.Duration) (int64, error) {
	if d <= 0 {
		return 0, ErrInvalidDuration
	}

	s.mu.RLock()
	v, ok := s.data[key]
	s.mu.RUnlock()

	if !ok {
		return 0, ErrKeyNotFound
	}

	if isExpired(v) {
		s.lazyExpire(key)
		return 0, ErrKeyExpired
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Re-verify the key exists and hasn't expired since releasing the read lock
	current, ok := s.data[key]
	if !ok {
		return 0, ErrKeyNotFound
	}
	if isExpired(current) {
		s.deleteExpiredLocked(key)
		return 0, ErrKeyExpired
	}

	current.ExpiresAt = time.Now().Add(d)
	s.data[key] = current

	return 1, nil
}

// PExpireAt updates the expiration time of an existing key to an absolute timestamp.
// It returns 1 if the expiration was successfully set. Otherwise returns 0 together
// with an appropriate Store error. Timestamps that are not in the future are rejected
// with ErrInvalidTimestamp.
//
// Sprint 17: PExpireAt performs lazy expiration using the check-then-act
// concurrency pattern. It never modifies the stored value, only the ExpiresAt field.
func (s *MemoryStore) PExpireAt(key string, t time.Time) (int64, error) {
	if !t.After(time.Now()) {
		return 0, ErrInvalidTimestamp
	}

	s.mu.RLock()
	v, ok := s.data[key]
	s.mu.RUnlock()

	if !ok {
		return 0, ErrKeyNotFound
	}

	if isExpired(v) {
		s.lazyExpire(key)
		return 0, ErrKeyExpired
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Re-verify the key exists and hasn't expired since releasing the read lock
	current, ok := s.data[key]
	if !ok {
		return 0, ErrKeyNotFound
	}
	if isExpired(current) {
		s.deleteExpiredLocked(key)
		return 0, ErrKeyExpired
	}

	current.ExpiresAt = t
	s.data[key] = current

	return 1, nil
}

// ExpireTime returns the absolute expiration timestamp of the given key in Unix seconds.
// It returns -1 if the key exists but has no associated expiration.
// It returns ErrKeyNotFound if the key does not exist.
//
// Sprint 19: ExpireTime performs lazy expiration using the same check-then-act
// concurrency pattern established by TTL(). If the key is found to be
// expired, it is deleted and ErrKeyExpired is returned.
func (s *MemoryStore) ExpireTime(key string) (int64, error) {
	s.mu.RLock()
	v, ok := s.data[key]
	s.mu.RUnlock()

	if !ok {
		return 0, ErrKeyNotFound
	}

	if isExpired(v) {
		s.lazyExpire(key)
		return 0, ErrKeyExpired
	}

	if v.ExpiresAt.IsZero() {
		return -1, nil
	}

	return v.ExpiresAt.Unix(), nil
}

// PExpireTime returns the absolute expiration timestamp of the given key in Unix milliseconds.
// It returns -1 if the key exists but has no associated expiration.
// It returns ErrKeyNotFound if the key does not exist.
//
// Sprint 20: PExpireTime performs lazy expiration using the same check-then-act
// concurrency pattern established by ExpireTime(). If the key is found to be
// expired, it is deleted and ErrKeyExpired is returned.
func (s *MemoryStore) PExpireTime(key string) (int64, error) {
	s.mu.RLock()
	v, ok := s.data[key]
	s.mu.RUnlock()

	if !ok {
		return 0, ErrKeyNotFound
	}

	if isExpired(v) {
		s.lazyExpire(key)
		return 0, ErrKeyExpired
	}

	if v.ExpiresAt.IsZero() {
		return -1, nil
	}

	return v.ExpiresAt.UnixMilli(), nil
}

// ErrInvalidArguments is returned when an operation is provided with missing or
// invalid arguments.
var ErrInvalidArguments = errors.New("invalid arguments")

// LPush prepends one or more values to a list. If the key does not exist,
// it creates a new list. It returns the new length of the list.
func (s *MemoryStore) LPush(key string, values ...string) (int64, error) {
	if len(values) == 0 {
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
			Type: types.ListType,
			Data: []string{},
		}
	} else if v.Type != types.ListType {
		return 0, ErrWrongType
	}

	list := v.Data.([]string)

	// Prepend left-to-right (last argument becomes the head)
	// Example: LPUSH mylist a b c -> [c, b, a, ...]
	for _, val := range values {
		list = append([]string{val}, list...)
	}

	v.Data = list
	s.data[key] = v

	return int64(len(list)), nil
}

// RPush appends one or more values to a list. If the key does not exist,
// it creates a new list. It returns the new length of the list.
func (s *MemoryStore) RPush(key string, values ...string) (int64, error) {
	if len(values) == 0 {
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
			Type: types.ListType,
			Data: []string{},
		}
	} else if v.Type != types.ListType {
		return 0, ErrWrongType
	}

	list := v.Data.([]string)

	// Append values in the order received
	// Example: RPUSH mylist c d e -> [... c, d, e]
	list = append(list, values...)

	v.Data = list
	s.data[key] = v

	return int64(len(list)), nil
}

// LLen returns the number of elements stored in a list. If the key does not
// exist, it returns 0.
func (s *MemoryStore) LLen(key string) (int64, error) {
	s.mu.RLock()
	v, ok := s.data[key]
	s.mu.RUnlock()

	if !ok {
		return 0, nil
	}

	if isExpired(v) {
		s.lazyExpire(key)
		return 0, nil
	}

	if v.Type != types.ListType {
		return 0, ErrWrongType
	}

	return int64(len(v.Data.([]string))), nil
}

// normalizeRange converts possibly negative, out-of-bounds start and stop indices
// into valid zero-indexed slice bounds. It returns the normalized start and stop
// (inclusive) indices, and a boolean indicating whether the range is valid.
func normalizeRange(start, stop int64, length int) (int, int, bool) {
	if start < 0 {
		start = int64(length) + start
	}
	if stop < 0 {
		stop = int64(length) + stop
	}

	if start < 0 {
		start = 0
	}
	if stop >= int64(length) {
		stop = int64(length) - 1
	}

	if stop < 0 {
		return 0, 0, false
	}

	if start >= int64(length) {
		return 0, 0, false
	}
	if start > stop {
		return 0, 0, false
	}

	return int(start), int(stop), true
}

// LRange returns a copy of the specified elements of the list stored at key.
// The offsets start and stop are zero-based indexes, with 0 being the first
// element of the list, 1 being the next element and so on.
// These offsets can also be negative numbers indicating offsets starting at
// the end of the list. For example, -1 is the last element of the list, -2
// the penultimate, and so on.
//
// LRANGE never mutates the underlying list and always returns a copy of the
// requested slice to preserve thread safety.
func (s *MemoryStore) LRange(key string, start, stop int64) ([]string, error) {
	s.mu.RLock()
	v, ok := s.data[key]
	s.mu.RUnlock()

	if !ok {
		return []string{}, nil
	}

	if isExpired(v) {
		s.lazyExpire(key)
		return []string{}, nil
	}

	if v.Type != types.ListType {
		return nil, ErrWrongType
	}

	list := v.Data.([]string)

	normStart, normStop, valid := normalizeRange(start, stop, len(list))
	if !valid {
		return []string{}, nil
	}

	// normStop is inclusive, so we slice up to normStop + 1
	// We return a copy of the slice to protect the internal storage from
	// concurrent modification after the RLock has been released.
	return append([]string(nil), list[normStart:normStop+1]...), nil
}

// LPop removes and returns the left-most element of the list stored at key.
// If the key does not exist or has expired, it returns an empty string and a nil error.
// The handler translates this empty result into a RESP Nil reply.
// If the key exists but is not a list, it returns ErrWrongType.
// LPop preserves the invariant that empty collections never exist: if removing
// the last element leaves the list empty, the key is deleted from the store.
func (s *MemoryStore) LPop(key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.data[key]
	if ok && isExpired(v) {
		s.deleteExpiredLocked(key)
		ok = false
	}

	if !ok {
		return "", nil
	}

	if v.Type != types.ListType {
		return "", ErrWrongType
	}

	list := v.Data.([]string)

	// Defensive check.
	// Empty collections should never exist because the collection
	// invariant deletes keys when the last element is removed.
	// If an empty list is encountered (for example, manually
	// injected during testing), remove the invalid key and treat
	// it as missing.
	if len(list) == 0 {
		delete(s.data, key)
		return "", nil
	}

	element := list[0]
	list = list[1:]

	if len(list) == 0 {
		delete(s.data, key)
	} else {
		v.Data = list
		s.data[key] = v
	}

	return element, nil
}

// RPop removes and returns the right-most element.
// Missing key returns empty string with nil error.
// The handler translates this into RESP Nil.
// RPOP mirrors LPOP, differing only in removal direction.
func (s *MemoryStore) RPop(key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.data[key]
	if ok && isExpired(v) {
		s.deleteExpiredLocked(key)
		ok = false
	}

	if !ok {
		return "", nil
	}

	if v.Type != types.ListType {
		return "", ErrWrongType
	}

	list := v.Data.([]string)

	// Defensive check.
	// Empty collections should never exist because the collection
	// invariant deletes keys when the last element is removed.
	// If an empty list is encountered (for example, manually
	// injected during testing), remove the invalid key and treat
	// it as missing.
	if len(list) == 0 {
		delete(s.data, key)
		return "", nil
	}

	lastIdx := len(list) - 1
	element := list[lastIdx]
	list = list[:lastIdx]

	if len(list) == 0 {
		delete(s.data, key)
	} else {
		v.Data = list
		s.data[key] = v
	}

	return element, nil
}
