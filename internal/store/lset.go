package store

import (
	"errors"

	"github.com/rajas2007/IgnisKV/internal/types"
)

// ErrIndexOutOfRange is returned when a list index is out of bounds.
var ErrIndexOutOfRange = errors.New("index out of range")

// LSet replaces the element at the specified index in the list stored at key.
// Positive indices begin at 0. Negative indices count backward from the tail
// (-1 is the last element, -2 is the penultimate element).
//
// LSet is a mutating command. It acquires a write lock immediately.
// If the key does not exist or has been lazily expired, it returns ErrKeyNotFound.
// If the stored value is not a list, it returns ErrWrongType.
// If the normalized index is outside the bounds of the list, it returns ErrIndexOutOfRange.
//
// LSet replaces exactly one element. The list length and ordering remain unchanged.
// The empty collection invariant is unaffected because LSet neither creates nor removes elements.
// This command performs no persistence; persistence belongs in the handler.
func (s *MemoryStore) LSet(key string, index int64, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.data[key]
	if !ok {
		return ErrKeyNotFound
	}

	if isExpired(v) {
		s.deleteExpiredLocked(key)
		return ErrKeyNotFound
	}

	if v.Type != types.ListType {
		return ErrWrongType
	}

	list := v.Data.([]string)
	length := int64(len(list))

	if index < 0 {
		index = length + index
	}

	if index < 0 || index >= length {
		return ErrIndexOutOfRange
	}

	list[index] = value
	s.data[key] = v

	return nil
}
