package store

import (
	"github.com/rajas2007/IgnisKV/internal/types"
)

// SAdd adds the specified members to the set stored at key.
// Specified members that are already a member of this set are ignored.
// If key does not exist, a new key holding a set is created before adding the specified members.
// An ErrWrongType is returned when the value stored at key is not a set.
// It returns the number of elements that were added to the set, not including all the elements already present in the set.
func (s *MemoryStore) SAdd(key string, members []string) (int, error) {
	if len(members) == 0 {
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
			Type: types.SetType,
			Data: make(map[string]struct{}),
		}
	} else if v.Type != types.SetType {
		return 0, ErrWrongType
	}

	setMap := v.Data.(map[string]struct{})
	added := 0

	for _, member := range members {
		if _, exists := setMap[member]; !exists {
			setMap[member] = struct{}{}
			added++
		}
	}

	if added > 0 {
		v.Data = setMap
		s.data[key] = v
	}

	return added, nil
}

// SRem removes the specified members from the set stored at key.
// Specified members that are not a member of this set are ignored.
// If key does not exist, it is treated as an empty set and this command returns 0.
// An ErrWrongType is returned when the value stored at key is not a set.
// If the set becomes empty after removing the members, the key is deleted from the store.
// It returns the number of members that were removed from the set.
func (s *MemoryStore) SRem(key string, members []string) (int, error) {
	if len(members) == 0 {
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

	if v.Type != types.SetType {
		return 0, ErrWrongType
	}

	setMap := v.Data.(map[string]struct{})
	removed := 0

	for _, member := range members {
		if _, exists := setMap[member]; exists {
			delete(setMap, member)
			removed++
		}
	}

	if removed > 0 {
		if len(setMap) == 0 {
			delete(s.data, key)
		} else {
			v.Data = setMap
			s.data[key] = v
		}
	}

	return removed, nil
}

// SIsMember returns true if member is a member of the set stored at key.
// If key does not exist or if member is not present in the set, it returns false.
// An ErrWrongType is returned when the value stored at key is not a set.
func (s *MemoryStore) SIsMember(key, member string) (bool, error) {
	s.mu.RLock()
	v, ok := s.data[key]

	if !ok {
		s.mu.RUnlock()
		return false, nil
	}

	if isExpired(v) {
		s.mu.RUnlock()
		s.lazyExpire(key)
		return false, nil
	}

	if v.Type != types.SetType {
		s.mu.RUnlock()
		return false, ErrWrongType
	}

	setMap := v.Data.(map[string]struct{})
	_, exists := setMap[member]
	s.mu.RUnlock()

	return exists, nil
}
