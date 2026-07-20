package store

import (
	"github.com/rajas2007/IgnisKV/internal/types"
)

// LInsert inserts an element into a list either immediately before or after a pivot element.
//
// LINSERT is a mutating list command.
// Missing key returns 0 (indicating the target list does not exist).
// Pivot not found returns -1 (indicating the list exists but no matching pivot element was found).
// It inserts exactly one element at the first matching pivot.
// The existing element ordering is always preserved.
//
// LINSERT acquires the write lock immediately because it mutates state.
// LINSERT never performs persistence; persistence belongs exclusively to the Handler.
//
// Complexity: O(n) worst-case since the list is scanned until the first matching pivot.
func (s *MemoryStore) LInsert(key string, before bool, pivot string, value string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, exists := s.data[key]
	if !exists {
		return 0, nil
	}

	if isExpired(v) {
		s.deleteExpiredLocked(key)
		return 0, nil
	}

	if v.Type != types.ListType {
		return 0, ErrWrongType
	}

	list := v.Data.([]string)

	idx := -1
	for i, elem := range list {
		if elem == pivot {
			idx = i
			break
		}
	}

	if idx == -1 {
		return -1, nil
	}

	var newList []string
	if before {
		newList = append(list[:idx], append([]string{value}, list[idx:]...)...)
	} else {
		newList = append(list[:idx+1], append([]string{value}, list[idx+1:]...)...)
	}

	v.Data = newList
	s.data[key] = v

	return int64(len(newList)), nil
}
