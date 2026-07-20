package store

import (
	"testing"
	"time"

	"github.com/rajas2007/IgnisKV/internal/types"
)

func TestHSet(t *testing.T) {
	s := NewMemoryStore()

	// 1. Missing key, single pair
	added, err := s.HSet("hash1", []string{"field1", "value1"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if added != 1 {
		t.Fatalf("expected 1 added, got %d", added)
	}

	val, _ := s.Get("hash1")
	if val.Type != types.HashType {
		t.Fatalf("expected HashType, got %v", val.Type)
	}
	hash := val.Data.(map[string]string)
	if hash["field1"] != "value1" {
		t.Fatalf("expected value1, got %v", hash["field1"])
	}
	if !val.ExpiresAt.IsZero() {
		t.Fatalf("expected zero ExpiresAt for new hash")
	}

	// 2. Existing key, multiple pairs, some new, some update
	added, err = s.HSet("hash1", []string{"field1", "value1_updated", "field2", "value2"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if added != 1 {
		t.Fatalf("expected 1 added (field2), got %d", added)
	}

	val, _ = s.Get("hash1")
	hash = val.Data.(map[string]string)
	if hash["field1"] != "value1_updated" || hash["field2"] != "value2" {
		t.Fatalf("unexpected hash state: %v", hash)
	}

	// 3. Duplicate fields in the same command
	added, err = s.HSet("hash2", []string{"f1", "v1", "f1", "v2"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if added != 1 {
		t.Fatalf("expected 1 added, got %d", added)
	}
	val, _ = s.Get("hash2")
	hash = val.Data.(map[string]string)
	if hash["f1"] != "v2" {
		t.Fatalf("expected v2 for duplicate field update, got %v", hash["f1"])
	}

	// 4. Sequential duplicate fields in the same command
	added, err = s.HSet("hash3", []string{"f2", "v1", "f2", "v2", "f2", "v3"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	// "f2" is added once; final value should be v3.
	if added != 1 {
		t.Fatalf("expected 1 added, got %d", added)
	}
	val, _ = s.Get("hash3")
	hash = val.Data.(map[string]string)
	if hash["f2"] != "v3" {
		t.Fatalf("expected v3 for sequential duplicate field update, got %v", hash["f2"])
	}

	// 5. Wrong type
	s.Set("string_key", types.Value{Type: types.StringType, Data: "val"})
	_, err = s.HSet("string_key", []string{"field", "value"})
	if err != ErrWrongType {
		t.Fatalf("expected ErrWrongType, got %v", err)
	}

	// 6. Invalid arguments (handled in HSet properly now)
	_, err = s.HSet("hash4", []string{"field"})
	if err != ErrInvalidArguments {
		t.Fatalf("expected ErrInvalidArguments, got %v", err)
	}
	_, err = s.HSet("hash4", []string{})
	if err != ErrInvalidArguments {
		t.Fatalf("expected ErrInvalidArguments, got %v", err)
	}

	// 7. TTL preservation
	expireTime := time.Now().Add(10 * time.Minute)
	s.Set("hash4", types.Value{Type: types.HashType, Data: make(map[string]string), ExpiresAt: expireTime})
	_, err = s.HSet("hash4", []string{"f", "v"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	val, _ = s.Get("hash4")
	if !val.ExpiresAt.Equal(expireTime) {
		t.Fatalf("expected ExpiresAt to be preserved")
	}

	// 7. Lazy expiration triggers before mutation
	s.Set("hash5", types.Value{Type: types.HashType, Data: map[string]string{"old": "val"}, ExpiresAt: time.Now().Add(-1 * time.Minute)})
	added, err = s.HSet("hash5", []string{"new", "val"})
	if err != nil {
		t.Fatalf("expected nil error on expired key, got %v", err)
	}
	if added != 1 {
		t.Fatalf("expected 1 added for expired key acting as new, got %d", added)
	}
	val, _ = s.Get("hash5")
	hash = val.Data.(map[string]string)
	if _, ok := hash["old"]; ok {
		t.Fatalf("expected old field to be expired/deleted")
	}
	if !val.ExpiresAt.IsZero() {
		t.Fatalf("expected zero ExpiresAt for new hash replacing expired one")
	}
}
