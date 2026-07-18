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
