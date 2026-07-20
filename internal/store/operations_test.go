package store

import (
	"errors"
	"testing"
	"time"

	"github.com/rajas2007/IgnisKV/internal/types"
)

func TestLPop(t *testing.T) {
	s := NewMemoryStore()

	// 1. Missing key
	val, err := s.LPop("missing")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if val != "" {
		t.Fatalf("expected empty string, got %q", val)
	}

	// 2. Existing multi-element list
	if _, err := s.RPush("list1", "a", "b", "c"); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	val, err = s.LPop("list1")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if val != "a" {
		t.Fatalf("expected 'a', got %q", val)
	}
	listVal, _ := s.Get("list1")
	list := listVal.Data.([]string)
	if len(list) != 2 || list[0] != "b" || list[1] != "c" {
		t.Fatalf("expected list [b c], got %v", list)
	}

	// 3. Sequential pops
	if _, err := s.RPush("list2", "a", "b", "c"); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	val1, _ := s.LPop("list2")
	val2, _ := s.LPop("list2")
	val3, _ := s.LPop("list2")
	if val1 != "a" || val2 != "b" || val3 != "c" {
		t.Fatalf("expected 'a', 'b', 'c', got %q, %q, %q", val1, val2, val3)
	}
	_, err = s.Get("list2")
	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("expected ErrKeyNotFound, got %v", err)
	}

	// 4. Single-element list
	if _, err := s.RPush("list3", "a"); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	val, err = s.LPop("list3")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if val != "a" {
		t.Fatalf("expected 'a', got %q", val)
	}
	_, err = s.Get("list3")
	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("expected ErrKeyNotFound, got %v", err)
	}

	// 5. WRONGTYPE
	s.Set("stringkey", types.Value{
		Type: types.StringType,
		Data: "val",
	})
	_, err = s.LPop("stringkey")
	if !errors.Is(err, ErrWrongType) {
		t.Fatalf("expected ErrWrongType, got %v", err)
	}

	// 6. Expired key
	if _, err := s.RPush("expiredlist", "a"); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	s.data["expiredlist"] = types.Value{
		Type:      types.ListType,
		Data:      []string{"a"},
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}
	val, err = s.LPop("expiredlist")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if val != "" {
		t.Fatalf("expected empty string, got %q", val)
	}
	_, err = s.Get("expiredlist")
	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("expected key to be deleted, got %v", err)
	}

	// 7. Defensive invariant
	s.data["invalidlist"] = types.Value{
		Type: types.ListType,
		Data: []string{},
	}
	val, err = s.LPop("invalidlist")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if val != "" {
		t.Fatalf("expected empty string, got %q", val)
	}
	_, err = s.Get("invalidlist")
	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("expected invalid key to be deleted, got %v", err)
	}

	// 8. Collection invariant
	if _, err := s.RPush("list4", "x"); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	if _, err := s.LPop("list4"); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if _, exists := s.data["list4"]; exists {
		t.Fatalf("expected key to be completely removed from store data")
	}
}

func TestRPop(t *testing.T) {
	s := NewMemoryStore()

	// 1. Missing key
	val, err := s.RPop("missing")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if val != "" {
		t.Fatalf("expected empty string, got %q", val)
	}

	// 2. Existing multi-element list
	if _, err := s.RPush("list1", "a", "b", "c"); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	val, err = s.RPop("list1")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if val != "c" {
		t.Fatalf("expected 'c', got %q", val)
	}
	listVal, _ := s.Get("list1")
	list := listVal.Data.([]string)
	if len(list) != 2 || list[0] != "a" || list[1] != "b" {
		t.Fatalf("expected list [a b], got %v", list)
	}

	// 3. Sequential pops
	if _, err := s.RPush("list2", "a", "b", "c"); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	val1, _ := s.RPop("list2")
	val2, _ := s.RPop("list2")
	val3, _ := s.RPop("list2")
	if val1 != "c" || val2 != "b" || val3 != "a" {
		t.Fatalf("expected 'c', 'b', 'a', got %q, %q, %q", val1, val2, val3)
	}
	_, err = s.Get("list2")
	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("expected ErrKeyNotFound, got %v", err)
	}

	// 4. Single-element list
	if _, err := s.RPush("list3", "a"); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	val, err = s.RPop("list3")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if val != "a" {
		t.Fatalf("expected 'a', got %q", val)
	}
	_, err = s.Get("list3")
	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("expected ErrKeyNotFound, got %v", err)
	}

	// 5. WRONGTYPE
	s.Set("stringkey", types.Value{
		Type: types.StringType,
		Data: "val",
	})
	_, err = s.RPop("stringkey")
	if !errors.Is(err, ErrWrongType) {
		t.Fatalf("expected ErrWrongType, got %v", err)
	}

	// 6. Expired key
	if _, err := s.RPush("expiredlist", "a"); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	s.data["expiredlist"] = types.Value{
		Type:      types.ListType,
		Data:      []string{"a"},
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}
	val, err = s.RPop("expiredlist")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if val != "" {
		t.Fatalf("expected empty string, got %q", val)
	}
	_, err = s.Get("expiredlist")
	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("expected key to be deleted, got %v", err)
	}

	// 7. Defensive invariant
	s.data["invalidlist"] = types.Value{
		Type: types.ListType,
		Data: []string{},
	}
	val, err = s.RPop("invalidlist")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if val != "" {
		t.Fatalf("expected empty string, got %q", val)
	}
	_, err = s.Get("invalidlist")
	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("expected invalid key to be deleted, got %v", err)
	}

	// 8. Collection invariant
	if _, err := s.RPush("list4", "x"); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	if _, err := s.RPop("list4"); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if _, exists := s.data["list4"]; exists {
		t.Fatalf("expected key to be completely removed from store data")
	}
}

func TestLIndex(t *testing.T) {
	s := NewMemoryStore()

	// 1. Missing key
	val, err := s.LIndex("missing", 0)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if val != "" {
		t.Fatalf("expected empty string, got %q", val)
	}

	// Setup: [a b c d e]
	if _, err := s.RPush("list", "a", "b", "c", "d", "e"); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// 2. First element
	val, err = s.LIndex("list", 0)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if val != "a" {
		t.Fatalf("expected 'a', got %q", val)
	}

	// 3. Middle element
	val, err = s.LIndex("list", 2)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if val != "c" {
		t.Fatalf("expected 'c', got %q", val)
	}

	// 4. Last element
	val, err = s.LIndex("list", 4)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if val != "e" {
		t.Fatalf("expected 'e', got %q", val)
	}

	// 5. Negative index (-1)
	val, err = s.LIndex("list", -1)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if val != "e" {
		t.Fatalf("expected 'e', got %q", val)
	}

	// 6. Negative index (-2)
	val, err = s.LIndex("list", -2)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if val != "d" {
		t.Fatalf("expected 'd', got %q", val)
	}

	// 7. Positive out-of-range
	val, err = s.LIndex("list", 100)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if val != "" {
		t.Fatalf("expected empty string, got %q", val)
	}

	// 8. Negative out-of-range
	val, err = s.LIndex("list", -100)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if val != "" {
		t.Fatalf("expected empty string, got %q", val)
	}

	// 9. WRONGTYPE
	s.Set("stringkey", types.Value{
		Type: types.StringType,
		Data: "val",
	})
	_, err = s.LIndex("stringkey", 0)
	if !errors.Is(err, ErrWrongType) {
		t.Fatalf("expected ErrWrongType, got %v", err)
	}

	// 10. Expired key
	if _, err := s.RPush("expiredlist", "a"); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	s.data["expiredlist"] = types.Value{
		Type:      types.ListType,
		Data:      []string{"a"},
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}
	val, err = s.LIndex("expiredlist", 0)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if val != "" {
		t.Fatalf("expected empty string, got %q", val)
	}
	_, err = s.Get("expiredlist")
	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("expected key to be deleted, got %v", err)
	}

	// 11. Read-only guarantee
	if _, err := s.RPush("readonly", "a", "b", "c"); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	val, err = s.LIndex("readonly", 1)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if val != "b" {
		t.Fatalf("expected 'b', got %q", val)
	}
	listVal, _ := s.Get("readonly")
	list := listVal.Data.([]string)
	if len(list) != 3 || list[0] != "a" || list[1] != "b" || list[2] != "c" {
		t.Fatalf("expected list [a b c], got %v", list)
	}

	// 12. Collection invariant — LINDEX never deletes valid keys or modifies contents
	val, err = s.LIndex("list", 0)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if val != "a" {
		t.Fatalf("expected 'a', got %q", val)
	}
	listVal, _ = s.Get("list")
	list = listVal.Data.([]string)
	if len(list) != 5 || list[0] != "a" || list[1] != "b" || list[2] != "c" || list[3] != "d" || list[4] != "e" {
		t.Fatalf("expected list [a b c d e], got %v", list)
	}
}

func TestLSet(t *testing.T) {
	s := NewMemoryStore()

	// 1. Missing key
	err := s.LSet("missing", 0, "x")
	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("expected ErrKeyNotFound, got %v", err)
	}

	// Setup: [a b c d e]
	if _, err := s.RPush("list", "a", "b", "c", "d", "e"); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	// 2. Existing list (first element)
	err = s.LSet("list", 0, "x")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	listVal, _ := s.Get("list")
	list := listVal.Data.([]string)
	if len(list) != 5 || list[0] != "x" || list[1] != "b" || list[2] != "c" || list[3] != "d" || list[4] != "e" {
		t.Fatalf("expected list [x b c d e], got %v", list)
	}

	// 3. Existing list (middle element)
	err = s.LSet("list", 2, "y")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	listVal, _ = s.Get("list")
	list = listVal.Data.([]string)
	if len(list) != 5 || list[0] != "x" || list[1] != "b" || list[2] != "y" || list[3] != "d" || list[4] != "e" {
		t.Fatalf("expected list [x b y d e], got %v", list)
	}

	// 4. Existing list (last element)
	err = s.LSet("list", 4, "z")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	listVal, _ = s.Get("list")
	list = listVal.Data.([]string)
	if len(list) != 5 || list[0] != "x" || list[1] != "b" || list[2] != "y" || list[3] != "d" || list[4] != "z" {
		t.Fatalf("expected list [x b y d z], got %v", list)
	}

	// 5. Negative index (-1)
	err = s.LSet("list", -1, "last")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	listVal, _ = s.Get("list")
	list = listVal.Data.([]string)
	if len(list) != 5 || list[0] != "x" || list[1] != "b" || list[2] != "y" || list[3] != "d" || list[4] != "last" {
		t.Fatalf("expected list [x b y d last], got %v", list)
	}

	// 6. Negative index (-2)
	err = s.LSet("list", -2, "tail")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	listVal, _ = s.Get("list")
	list = listVal.Data.([]string)
	if len(list) != 5 || list[0] != "x" || list[1] != "b" || list[2] != "y" || list[3] != "tail" || list[4] != "last" {
		t.Fatalf("expected list [x b y tail last], got %v", list)
	}

	// 7. Positive out-of-range
	err = s.LSet("list", 100, "x")
	if !errors.Is(err, ErrIndexOutOfRange) {
		t.Fatalf("expected ErrIndexOutOfRange, got %v", err)
	}
	listVal, _ = s.Get("list")
	list = listVal.Data.([]string)
	if len(list) != 5 || list[0] != "x" || list[1] != "b" || list[2] != "y" || list[3] != "tail" || list[4] != "last" {
		t.Fatalf("expected list [x b y tail last], got %v", list)
	}

	// 8. Negative out-of-range
	err = s.LSet("list", -100, "x")
	if !errors.Is(err, ErrIndexOutOfRange) {
		t.Fatalf("expected ErrIndexOutOfRange, got %v", err)
	}
	listVal, _ = s.Get("list")
	list = listVal.Data.([]string)
	if len(list) != 5 || list[0] != "x" || list[1] != "b" || list[2] != "y" || list[3] != "tail" || list[4] != "last" {
		t.Fatalf("expected list [x b y tail last], got %v", list)
	}

	// 9. WRONGTYPE
	s.Set("stringkey", types.Value{
		Type: types.StringType,
		Data: "val",
	})
	err = s.LSet("stringkey", 0, "x")
	if !errors.Is(err, ErrWrongType) {
		t.Fatalf("expected ErrWrongType, got %v", err)
	}

	// 10. Expired key
	if _, err := s.RPush("expiredlist", "a"); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	s.data["expiredlist"] = types.Value{
		Type:      types.ListType,
		Data:      []string{"a"},
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}
	err = s.LSet("expiredlist", 0, "x")
	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("expected ErrKeyNotFound, got %v", err)
	}
	_, err = s.Get("expiredlist")
	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("expected key to be deleted, got %v", err)
	}

	// 11. Length invariant
	length, err := s.LLen("list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if length != 5 {
		t.Fatalf("expected length 5, got %d", length)
	}

	err = s.LSet("list", 0, "first")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	err = s.LSet("list", -1, "last_again")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	newLength, err := s.LLen("list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newLength != 5 {
		t.Fatalf("expected length to remain 5, got %d", newLength)
	}

	// 12. Ordering invariant
	listVal, _ = s.Get("list")
	list = listVal.Data.([]string)
	if len(list) != 5 || list[0] != "first" || list[1] != "b" || list[2] != "y" || list[3] != "tail" || list[4] != "last_again" {
		t.Fatalf("expected list [first b y tail last_again], got %v", list)
	}
}
