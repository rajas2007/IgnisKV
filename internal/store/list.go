package store

import (
	"github.com/rajas2007/IgnisKV/internal/types"
)

// LIndex returns the element at the specified index in the list stored at key.
// Positive indices begin at 0. Negative indices count backward from the tail
// (-1 is the last element, -2 is the penultimate element).
//
// LIndex is a read-only operation. It never mutates collection state and
// never performs persistence.
//
// If the key does not exist or has been lazily expired, it returns an empty
// string and a nil error. If the normalized index is outside the bounds of
// the list, it returns an empty string and a nil error.
// If the stored value is not a list, ErrWrongType is returned.
func (s *MemoryStore) LIndex(key string, index int64) (string, error) {
	s.mu.RLock()
	v, ok := s.data[key]
	s.mu.RUnlock()

	if !ok {
		return "", nil
	}

	if isExpired(v) {
		s.lazyExpire(key)
		return "", nil
	}

	if v.Type != types.ListType {
		return "", ErrWrongType
	}

	list := v.Data.([]string)
	length := int64(len(list))

	if index < 0 {
		index = length + index
	}

	if index < 0 || index >= length {
		return "", nil
	}

	return list[index], nil
}
