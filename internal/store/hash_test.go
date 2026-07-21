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

func TestHGet(t *testing.T) {
	s := NewMemoryStore()

	// 1. Missing key
	_, err := s.HGet("missing_key", "field1")
	if err != ErrKeyNotFound {
		t.Fatalf("expected ErrKeyNotFound, got %v", err)
	}

	// Setup hash
	s.HSet("hash1", []string{"field1", "value1", "empty_val", "", "", "empty_field"})

	// 2. Existing field
	val, err := s.HGet("hash1", "field1")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if val != "value1" {
		t.Fatalf("expected value1, got %v", val)
	}

	// 3. Missing field
	_, err = s.HGet("hash1", "missing_field")
	if err != ErrFieldNotFound {
		t.Fatalf("expected ErrFieldNotFound, got %v", err)
	}

	// 4. Empty value
	val, err = s.HGet("hash1", "empty_val")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if val != "" {
		t.Fatalf("expected empty value, got %v", val)
	}

	// 5. Empty field name
	val, err = s.HGet("hash1", "")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if val != "empty_field" {
		t.Fatalf("expected empty_field, got %v", val)
	}

	// 6. Wrong type
	s.Set("string_key", types.Value{Type: types.StringType, Data: "val"})
	_, err = s.HGet("string_key", "field1")
	if err != ErrWrongType {
		t.Fatalf("expected ErrWrongType, got %v", err)
	}

	// 7. Lazy expiration
	s.Set("hash_expired", types.Value{
		Type:      types.HashType,
		Data:      map[string]string{"f": "v"},
		ExpiresAt: time.Now().Add(-1 * time.Minute),
	})
	_, err = s.HGet("hash_expired", "f")
	if err != ErrKeyNotFound {
		t.Fatalf("expected ErrKeyNotFound for expired key, got %v", err)
	}
	if s.Exists("hash_expired") {
		t.Fatalf("expected key to be physically deleted after lazy expiration")
	}
}

func TestHExists(t *testing.T) {
	s := NewMemoryStore()

	// 1. Missing key
	exists, err := s.HExists("missing_key", "field1")
	if err != ErrKeyNotFound {
		t.Fatalf("expected ErrKeyNotFound, got %v", err)
	}
	if exists {
		t.Fatalf("expected exists to be false")
	}

	// Setup hash
	s.HSet("hash1", []string{"field1", "value1", "empty_val", "", "", "empty_field"})

	// 2. Existing field
	exists, err = s.HExists("hash1", "field1")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !exists {
		t.Fatalf("expected exists to be true")
	}

	// 3. Missing field
	exists, err = s.HExists("hash1", "missing_field")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if exists {
		t.Fatalf("expected exists to be false")
	}

	// 4. Empty value
	exists, err = s.HExists("hash1", "empty_val")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !exists {
		t.Fatalf("expected exists to be true")
	}

	// 5. Empty field name
	exists, err = s.HExists("hash1", "")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !exists {
		t.Fatalf("expected exists to be true")
	}

	// 6. Wrong type
	s.Set("string_key", types.Value{Type: types.StringType, Data: "val"})
	exists, err = s.HExists("string_key", "field1")
	if err != ErrWrongType {
		t.Fatalf("expected ErrWrongType, got %v", err)
	}
	if exists {
		t.Fatalf("expected exists to be false")
	}

	// 7. Lazy expiration
	s.Set("hash_expired", types.Value{
		Type:      types.HashType,
		Data:      map[string]string{"f": "v"},
		ExpiresAt: time.Now().Add(-1 * time.Minute),
	})
	exists, err = s.HExists("hash_expired", "f")
	if err != ErrKeyNotFound {
		t.Fatalf("expected ErrKeyNotFound for expired key, got %v", err)
	}
	if exists {
		t.Fatalf("expected exists to be false")
	}
	if s.Exists("hash_expired") {
		t.Fatalf("expected key to be physically deleted after lazy expiration")
	}
}

func TestHDel(t *testing.T) {
	s := NewMemoryStore()

	// 1. Missing key
	deleted, err := s.HDel("missing_key", []string{"f1"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if deleted != 0 {
		t.Fatalf("expected 0 deleted, got %d", deleted)
	}

	// 2. Wrong type
	s.Set("string_key", types.Value{Type: types.StringType, Data: "val"})
	_, err = s.HDel("string_key", []string{"f1"})
	if err != ErrWrongType {
		t.Fatalf("expected ErrWrongType, got %v", err)
	}

	// 3. Single field deletion
	s.HSet("hash1", []string{"f1", "v1", "f2", "v2"})
	deleted, err = s.HDel("hash1", []string{"f1"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if deleted != 1 {
		t.Fatalf("expected 1 deleted, got %d", deleted)
	}

	// 4. Multiple field deletion and Mixed existing/missing fields
	s.HSet("hash2", []string{"f1", "v1", "f2", "v2", "f3", "v3", "f4", "v4"})
	deleted, err = s.HDel("hash2", []string{"f1", "f2", "missing1", "f3", "missing2"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if deleted != 3 {
		t.Fatalf("expected 3 deleted, got %d", deleted)
	}

	// 5. Duplicate field names (HDEL hash f1 f1 f1)
	s.HSet("hash3", []string{"f1", "v1", "f2", "v2"})
	deleted, err = s.HDel("hash3", []string{"f1", "f1", "f1"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	// "f1" is deleted once -> total 1
	if deleted != 1 {
		t.Fatalf("expected 1 deleted, got %d", deleted)
	}
	vHash3, _ := s.Get("hash3")
	m3 := vHash3.Data.(map[string]string)
	if _, ok := m3["f1"]; ok {
		t.Fatalf("expected f1 to be removed")
	}
	if val, ok := m3["f2"]; !ok || val != "v2" {
		t.Fatalf("expected f2 to remain unchanged, got %v", val)
	}

	// 6. Delete last field removes key (empty hash invariant)
	s.HSet("hash4", []string{"f1", "v1"})
	deleted, err = s.HDel("hash4", []string{"f1"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if deleted != 1 {
		t.Fatalf("expected 1 deleted, got %d", deleted)
	}
	if s.Exists("hash4") {
		t.Fatalf("expected key to be deleted after last field removed")
	}

	// 7. TTL preserved when hash remains
	now := time.Now()
	s.Set("hash5", types.Value{
		Type:      types.HashType,
		Data:      map[string]string{"f1": "v1", "f2": "v2"},
		ExpiresAt: now.Add(1 * time.Minute),
	})
	deleted, err = s.HDel("hash5", []string{"f1"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if deleted != 1 {
		t.Fatalf("expected 1 deleted, got %d", deleted)
	}
	v, _ := s.Get("hash5") // Wait, Get() returns a copy. This is fine to read ExpiresAt
	if v.ExpiresAt.IsZero() || v.ExpiresAt.Before(now) {
		t.Fatalf("expected TTL to be preserved, got %v", v.ExpiresAt)
	}

	// 8. Lazy expiration
	s.Set("hash_expired", types.Value{
		Type:      types.HashType,
		Data:      map[string]string{"f": "v"},
		ExpiresAt: time.Now().Add(-1 * time.Minute),
	})
	deleted, err = s.HDel("hash_expired", []string{"f"})
	if err != nil {
		t.Fatalf("expected nil error for expired key, got %v", err)
	}
	if deleted != 0 {
		t.Fatalf("expected 0 deleted, got %d", deleted)
	}
	if s.Exists("hash_expired") {
		t.Fatalf("expected key to be physically deleted after lazy expiration")
	}
}

func TestHLen(t *testing.T) {
	s := NewMemoryStore()

	// 1. Missing key
	length, err := s.HLen("missing_key")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if length != 0 {
		t.Fatalf("expected 0 length, got %d", length)
	}

	// 2. Wrong type
	s.Set("string_key", types.Value{Type: types.StringType, Data: "val"})
	_, err = s.HLen("string_key")
	if err != ErrWrongType {
		t.Fatalf("expected ErrWrongType, got %v", err)
	}

	// 3. Single field
	s.HSet("hash1", []string{"f1", "v1"})
	length, err = s.HLen("hash1")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if length != 1 {
		t.Fatalf("expected 1 length, got %d", length)
	}

	// 4. Multiple fields
	s.HSet("hash2", []string{"f1", "v1", "f2", "v2", "f3", "v3"})
	length, err = s.HLen("hash2")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if length != 3 {
		t.Fatalf("expected 3 length, got %d", length)
	}

	// 5. Empty hash (manually inserted since HSET/HDEL maintain invariant)
	s.Set("hash_empty", types.Value{Type: types.HashType, Data: make(map[string]string)})
	length, err = s.HLen("hash_empty")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if length != 0 {
		t.Fatalf("expected 0 length for empty hash, got %d", length)
	}

	// 6. Lazy expiration
	s.Set("hash_expired", types.Value{
		Type:      types.HashType,
		Data:      map[string]string{"f": "v"},
		ExpiresAt: time.Now().Add(-1 * time.Minute),
	})
	length, err = s.HLen("hash_expired")
	if err != nil {
		t.Fatalf("expected nil error for expired key, got %v", err)
	}
	if length != 0 {
		t.Fatalf("expected 0 length, got %d", length)
	}
	if s.Exists("hash_expired") {
		t.Fatalf("expected key to be physically deleted after lazy expiration")
	}
}

func TestHGetAll(t *testing.T) {
	s := NewMemoryStore()

	// 1. Missing key
	result, err := s.HGetAll("missing_key")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected empty slice, got %v", result)
	}

	// 2. Wrong type
	s.Set("string_key", types.Value{Type: types.StringType, Data: "val"})
	_, err = s.HGetAll("string_key")
	if err != ErrWrongType {
		t.Fatalf("expected ErrWrongType, got %v", err)
	}

	// 3. Single-field hash
	s.HSet("hash1", []string{"f1", "v1"})
	result, err = s.HGetAll("hash1")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(result))
	}
	got := sliceToMap(result)
	if got["f1"] != "v1" {
		t.Fatalf("expected f1=v1, got f1=%s", got["f1"])
	}

	// 4. Multiple fields - verify alternating field/value layout
	s.HSet("hash2", []string{"f1", "v1", "f2", "v2", "f3", "v3"})
	result, err = s.HGetAll("hash2")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(result) != 6 {
		t.Fatalf("expected 6 elements, got %d", len(result))
	}
	if len(result)%2 != 0 {
		t.Fatalf("expected even number of elements, got %d", len(result))
	}
	got = sliceToMap(result)
	expected := map[string]string{"f1": "v1", "f2": "v2", "f3": "v3"}
	for k, v := range expected {
		if got[k] != v {
			t.Fatalf("expected %s=%s, got %s=%s", k, v, k, got[k])
		}
	}

	// 5. Empty hash (manually constructed)
	s.Set("hash_empty", types.Value{Type: types.HashType, Data: make(map[string]string)})
	result, err = s.HGetAll("hash_empty")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected empty slice for empty hash, got %v", result)
	}

	// 6. Lazy expiration
	s.Set("hash_expired", types.Value{
		Type:      types.HashType,
		Data:      map[string]string{"f": "v"},
		ExpiresAt: time.Now().Add(-1 * time.Minute),
	})
	result, err = s.HGetAll("hash_expired")
	if err != nil {
		t.Fatalf("expected nil error for expired key, got %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected empty slice, got %v", result)
	}
	if s.Exists("hash_expired") {
		t.Fatalf("expected key to be physically deleted after lazy expiration")
	}
}

// sliceToMap converts a flat alternating field/value slice into a map.
func sliceToMap(s []string) map[string]string {
	m := make(map[string]string)
	for i := 0; i < len(s); i += 2 {
		m[s[i]] = s[i+1]
	}
	return m
}
