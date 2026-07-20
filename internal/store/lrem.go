package store

import (
	"github.com/rajas2007/IgnisKV/internal/types"
)

// LRem removes elements matching a specified value from a list.
// It implements Redis count semantics:
//   - count > 0 removes up to count matches from head to tail
//   - count < 0 removes up to abs(count) matches from tail to head
//   - count == 0 removes all matching elements
//
// It is a mutating command that operates under an immediate write lock.
// It preserves the relative ordering of all remaining elements.
// Missing keys or keys containing no matches are treated as successful no-ops and return (0, nil).
// If the stored value is not a list, it returns (0, ErrWrongType).
// If all elements are removed, the key itself is deleted to preserve the empty collection invariant.
// Persistence is not handled here; it is the responsibility of the calling handler.
func (s *MemoryStore) LRem(key string, count int64, value string) (int64, error) {
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
	if len(list) == 0 {
		return 0, nil
	}

	var removedCount int64
	var absCount int64
	if count > 0 {
		absCount = count
	} else if count < 0 {
		absCount = -count
	}

	var newList []string

	if count > 0 {
		for _, elem := range list {
			if elem == value && removedCount < absCount {
				removedCount++
			} else {
				newList = append(newList, elem)
			}
		}
	} else if count < 0 {
		toRemove := make(map[int]bool)
		for i := len(list) - 1; i >= 0; i-- {
			if list[i] == value && removedCount < absCount {
				toRemove[i] = true
				removedCount++
			}
		}
		for i, elem := range list {
			if !toRemove[i] {
				newList = append(newList, elem)
			}
		}
	} else {
		for _, elem := range list {
			if elem == value {
				removedCount++
			} else {
				newList = append(newList, elem)
			}
		}
	}

	if removedCount == 0 {
		return 0, nil
	}

	if len(newList) == 0 {
		delete(s.data, key)
	} else {
		v.Data = newList
		s.data[key] = v
	}

	return removedCount, nil
}
