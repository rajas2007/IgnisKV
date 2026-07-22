package store

import (
	"testing"
	"time"

	"github.com/rajas2007/IgnisKV/internal/types"
)

func TestSAdd(t *testing.T) {
	s := NewMemoryStore()

	// 1. Missing key
	added, err := s.SAdd("set1", []string{"m1", "m2"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if added != 2 {
		t.Fatalf("expected 2 added, got %d", added)
	}

	// Verify internals directly
	v, ok := s.data["set1"]
	if !ok {
		t.Fatalf("expected key to exist")
	}
	if v.Type != types.SetType {
		t.Fatalf("expected SetType, got %v", v.Type)
	}
	setMap := v.Data.(map[string]struct{})
	if len(setMap) != 2 {
		t.Fatalf("expected set map length 2, got %d", len(setMap))
	}
	if _, ok := setMap["m1"]; !ok {
		t.Fatalf("expected m1 to exist")
	}
	if _, ok := setMap["m2"]; !ok {
		t.Fatalf("expected m2 to exist")
	}

	// 2. Multiple inserts with duplicates
	added, err = s.SAdd("set1", []string{"m2", "m3", "m4", "m3"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	// "m2" already exists. "m3", "m4", "m3" -> 2 new elements added ("m3", "m4")
	if added != 2 {
		t.Fatalf("expected 2 added, got %d", added)
	}

	setMap = s.data["set1"].Data.(map[string]struct{})
	if len(setMap) != 4 {
		t.Fatalf("expected set map length 4, got %d", len(setMap))
	}

	// 3. Duplicate inserts (all already exist)
	added, err = s.SAdd("set1", []string{"m1", "m4"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if added != 0 {
		t.Fatalf("expected 0 added, got %d", added)
	}

	// 4. Duplicate members within same command for new key
	added, err = s.SAdd("set2", []string{"m1", "m1", "m1"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if added != 1 {
		t.Fatalf("expected 1 added, got %d", added)
	}

	// 5. Wrong type
	s.Set("string_key", types.Value{Type: types.StringType, Data: "val"})
	_, err = s.SAdd("string_key", []string{"m1"})
	if err != ErrWrongType {
		t.Fatalf("expected ErrWrongType, got %v", err)
	}

	// 6. Lazy expiration
	s.Set("set_expired", types.Value{
		Type:      types.SetType,
		Data:      map[string]struct{}{"old": {}},
		ExpiresAt: time.Now().Add(-1 * time.Minute),
	})
	added, err = s.SAdd("set_expired", []string{"new"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if added != 1 {
		t.Fatalf("expected 1 added for expired key, got %d", added)
	}

	// Ensure old is gone
	setMap = s.data["set_expired"].Data.(map[string]struct{})
	if _, ok := setMap["old"]; ok {
		t.Fatalf("expected old element to be removed during lazy expiration")
	}
}

func TestSRem(t *testing.T) {
	s := NewMemoryStore()

	// 1. Missing key
	removed, err := s.SRem("set1", []string{"m1"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if removed != 0 {
		t.Fatalf("expected 0 removed, got %d", removed)
	}

	// Setup for tests
	s.SAdd("set1", []string{"m1", "m2", "m3", "m4"})

	// 2. Remove existing (single)
	removed, err = s.SRem("set1", []string{"m1"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if removed != 1 {
		t.Fatalf("expected 1 removed, got %d", removed)
	}
	setMap := s.data["set1"].Data.(map[string]struct{})
	if _, ok := setMap["m1"]; ok {
		t.Fatalf("expected m1 to be removed")
	}
	if len(setMap) != 3 {
		t.Fatalf("expected len 3, got %d", len(setMap))
	}

	// 3. Remove missing
	removed, err = s.SRem("set1", []string{"missing"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if removed != 0 {
		t.Fatalf("expected 0 removed, got %d", removed)
	}

	// 4. Remove multiple (mixed existing/missing)
	removed, err = s.SRem("set1", []string{"m2", "missing", "m3"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if removed != 2 {
		t.Fatalf("expected 2 removed, got %d", removed)
	}

	// 5. Remove last member deletes key
	removed, err = s.SRem("set1", []string{"m4"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if removed != 1 {
		t.Fatalf("expected 1 removed, got %d", removed)
	}
	if _, ok := s.data["set1"]; ok {
		t.Fatalf("expected key to be physically deleted after removing last element")
	}

	// 6. Wrong type
	s.Set("string_key", types.Value{Type: types.StringType, Data: "val"})
	_, err = s.SRem("string_key", []string{"m1"})
	if err != ErrWrongType {
		t.Fatalf("expected ErrWrongType, got %v", err)
	}

	// 7. Lazy expiration
	s.Set("set_expired", types.Value{
		Type:      types.SetType,
		Data:      map[string]struct{}{"m1": {}},
		ExpiresAt: time.Now().Add(-1 * time.Minute),
	})
	removed, err = s.SRem("set_expired", []string{"m1"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if removed != 0 { // Key is expired, treated as missing, returns 0
		t.Fatalf("expected 0 removed on expired key, got %d", removed)
	}
	if s.Exists("set_expired") {
		t.Fatalf("expected key to be physically deleted after lazy expiration")
	}
}

func TestSIsMember(t *testing.T) {
	s := NewMemoryStore()

	// 1. Missing key
	isMember, err := s.SIsMember("set1", "m1")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if isMember {
		t.Fatalf("expected false for missing key")
	}

	// Setup
	s.SAdd("set1", []string{"m1", "m2"})

	// 2. Existing member
	isMember, err = s.SIsMember("set1", "m1")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !isMember {
		t.Fatalf("expected true for existing member")
	}

	// 3. Missing member
	isMember, err = s.SIsMember("set1", "m3")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if isMember {
		t.Fatalf("expected false for missing member")
	}

	// 4. Wrong type
	s.Set("string_key", types.Value{Type: types.StringType, Data: "val"})
	_, err = s.SIsMember("string_key", "m1")
	if err != ErrWrongType {
		t.Fatalf("expected ErrWrongType, got %v", err)
	}

	// 5. Lazy expiration
	s.Set("set_expired", types.Value{
		Type:      types.SetType,
		Data:      map[string]struct{}{"m1": {}},
		ExpiresAt: time.Now().Add(-1 * time.Minute),
	})
	isMember, err = s.SIsMember("set_expired", "m1")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if isMember {
		t.Fatalf("expected false for expired key")
	}
	if s.Exists("set_expired") {
		t.Fatalf("expected key to be physically deleted after lazy expiration")
	}
}
