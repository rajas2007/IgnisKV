package store

import (
	"sort"
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

func TestHMGet(t *testing.T) {
	s := NewMemoryStore()

	// Setup hash: name=Rajas, age=19, empty_val=""
	s.HSet("hash1", []string{"name", "Rajas", "age", "19", "empty_val", "", "", "empty_field"})

	// 1. All fields present
	result, err := s.HMGet("hash1", []string{"name", "age"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 results, got %d", len(result))
	}
	if result[0] != "Rajas" {
		t.Fatalf("expected Rajas, got %v", result[0])
	}
	if result[1] != "19" {
		t.Fatalf("expected 19, got %v", result[1])
	}

	// 2. Mixed present/missing fields - order preserved
	result, err = s.HMGet("hash1", []string{"name", "city", "age"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 results, got %d", len(result))
	}
	if result[0] != "Rajas" {
		t.Fatalf("expected Rajas at position 0, got %v", result[0])
	}
	if result[1] != nil {
		t.Fatalf("expected nil at position 1, got %v", result[1])
	}
	if result[2] != "19" {
		t.Fatalf("expected 19 at position 2, got %v", result[2])
	}

	// 3. Missing key - all nil
	result, err = s.HMGet("missing_key", []string{"f1", "f2", "f3"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 results, got %d", len(result))
	}
	for i, v := range result {
		if v != nil {
			t.Fatalf("expected nil at position %d, got %v", i, v)
		}
	}

	// 4. Wrong type
	s.Set("string_key", types.Value{Type: types.StringType, Data: "val"})
	_, err = s.HMGet("string_key", []string{"f1"})
	if err != ErrWrongType {
		t.Fatalf("expected ErrWrongType, got %v", err)
	}

	// 5. Empty string value
	result, err = s.HMGet("hash1", []string{"empty_val"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if result[0] != "" {
		t.Fatalf("expected empty string, got %v", result[0])
	}

	// 6. Empty field name
	result, err = s.HMGet("hash1", []string{""})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if result[0] != "empty_field" {
		t.Fatalf("expected empty_field, got %v", result[0])
	}

	// 7. Lazy expiration - all nil
	s.Set("hash_expired", types.Value{
		Type:      types.HashType,
		Data:      map[string]string{"f": "v"},
		ExpiresAt: time.Now().Add(-1 * time.Minute),
	})
	result, err = s.HMGet("hash_expired", []string{"f", "missing"})
	if err != nil {
		t.Fatalf("expected nil error for expired key, got %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 results, got %d", len(result))
	}
	for i, v := range result {
		if v != nil {
			t.Fatalf("expected nil at position %d, got %v", i, v)
		}
	}
	if s.Exists("hash_expired") {
		t.Fatalf("expected key to be physically deleted after lazy expiration")
	}
}

func TestHKeys(t *testing.T) {
	s := NewMemoryStore()

	// 1. Missing key
	keys, err := s.HKeys("missing_key")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(keys) != 0 {
		t.Fatalf("expected empty slice, got %v", keys)
	}

	// 2. Wrong type
	s.Set("string_key", types.Value{Type: types.StringType, Data: "val"})
	_, err = s.HKeys("string_key")
	if err != ErrWrongType {
		t.Fatalf("expected ErrWrongType, got %v", err)
	}

	// 3. Single-field hash
	s.HSet("hash1", []string{"f1", "v1"})
	keys, err = s.HKeys("hash1")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(keys) != 1 {
		t.Fatalf("expected 1 key, got %d", len(keys))
	}
	if keys[0] != "f1" {
		t.Fatalf("expected f1, got %s", keys[0])
	}

	// 4. Multiple fields - sort before comparing
	s.HSet("hash2", []string{"f1", "v1", "f2", "v2", "f3", "v3"})
	keys, err = s.HKeys("hash2")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(keys) != 3 {
		t.Fatalf("expected 3 keys, got %d", len(keys))
	}
	sort.Strings(keys)
	expected := []string{"f1", "f2", "f3"}
	for i, k := range keys {
		if k != expected[i] {
			t.Fatalf("expected %s at position %d, got %s", expected[i], i, k)
		}
	}

	// 5. Empty hash (manually constructed)
	s.Set("hash_empty", types.Value{Type: types.HashType, Data: make(map[string]string)})
	keys, err = s.HKeys("hash_empty")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(keys) != 0 {
		t.Fatalf("expected empty slice for empty hash, got %v", keys)
	}

	// 6. Lazy expiration
	s.Set("hash_expired", types.Value{
		Type:      types.HashType,
		Data:      map[string]string{"f": "v"},
		ExpiresAt: time.Now().Add(-1 * time.Minute),
	})
	keys, err = s.HKeys("hash_expired")
	if err != nil {
		t.Fatalf("expected nil error for expired key, got %v", err)
	}
	if len(keys) != 0 {
		t.Fatalf("expected empty slice, got %v", keys)
	}
	if s.Exists("hash_expired") {
		t.Fatalf("expected key to be physically deleted after lazy expiration")
	}
}

func TestHVals(t *testing.T) {
	s := NewMemoryStore()

	// 1. Missing key
	vals, err := s.HVals("missing_key")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(vals) != 0 {
		t.Fatalf("expected empty slice, got %v", vals)
	}

	// 2. Wrong type
	s.Set("string_key", types.Value{Type: types.StringType, Data: "val"})
	_, err = s.HVals("string_key")
	if err != ErrWrongType {
		t.Fatalf("expected ErrWrongType, got %v", err)
	}

	// 3. Single-value hash
	s.HSet("hash1", []string{"f1", "v1"})
	vals, err = s.HVals("hash1")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(vals) != 1 {
		t.Fatalf("expected 1 value, got %d", len(vals))
	}
	if vals[0] != "v1" {
		t.Fatalf("expected v1, got %s", vals[0])
	}

	// 4. Multiple values - sort before comparing
	s.HSet("hash2", []string{"f1", "v1", "f2", "v2", "f3", "v3"})
	vals, err = s.HVals("hash2")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(vals) != 3 {
		t.Fatalf("expected 3 values, got %d", len(vals))
	}
	sort.Strings(vals)
	expected := []string{"v1", "v2", "v3"}
	for i, v := range vals {
		if v != expected[i] {
			t.Fatalf("expected %s at position %d, got %s", expected[i], i, v)
		}
	}

	// 5. Duplicate values
	s.HSet("hash_dup", []string{"f1", "val", "f2", "val", "f3", "other"})
	vals, err = s.HVals("hash_dup")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(vals) != 3 {
		t.Fatalf("expected 3 values, got %d", len(vals))
	}
	sort.Strings(vals)
	expectedDup := []string{"other", "val", "val"}
	for i, v := range vals {
		if v != expectedDup[i] {
			t.Fatalf("expected %s at position %d, got %s", expectedDup[i], i, v)
		}
	}

	// 6. Empty hash (manually constructed)
	s.Set("hash_empty", types.Value{Type: types.HashType, Data: make(map[string]string)})
	vals, err = s.HVals("hash_empty")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(vals) != 0 {
		t.Fatalf("expected empty slice for empty hash, got %v", vals)
	}

	// 7. Lazy expiration
	s.Set("hash_expired", types.Value{
		Type:      types.HashType,
		Data:      map[string]string{"f": "v"},
		ExpiresAt: time.Now().Add(-1 * time.Minute),
	})
	vals, err = s.HVals("hash_expired")
	if err != nil {
		t.Fatalf("expected nil error for expired key, got %v", err)
	}
	if len(vals) != 0 {
		t.Fatalf("expected empty slice, got %v", vals)
	}
	if s.Exists("hash_expired") {
		t.Fatalf("expected key to be physically deleted after lazy expiration")
	}
}

func TestHStrLen(t *testing.T) {
	s := NewMemoryStore()

	// 1. Existing field
	s.HSet("hash1", []string{"f1", "value", "f2", "é"}) // "é" is 2 bytes in UTF-8
	length, err := s.HStrLen("hash1", "f1")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if length != 5 {
		t.Fatalf("expected length 5, got %d", length)
	}

	// UTF-8 byte counting verification
	length, err = s.HStrLen("hash1", "f2")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if length != 2 {
		t.Fatalf("expected length 2 (bytes) for 'é', got %d", length)
	}

	// 2. Missing field
	length, err = s.HStrLen("hash1", "missing_field")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if length != 0 {
		t.Fatalf("expected length 0 for missing field, got %d", length)
	}

	// 3. Missing key
	length, err = s.HStrLen("missing_key", "f1")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if length != 0 {
		t.Fatalf("expected length 0 for missing key, got %d", length)
	}

	// 4. Wrong type
	s.Set("string_key", types.Value{Type: types.StringType, Data: "val"})
	_, err = s.HStrLen("string_key", "f1")
	if err != ErrWrongType {
		t.Fatalf("expected ErrWrongType, got %v", err)
	}

	// 5. Empty string value
	s.HSet("hash2", []string{"f1", ""})
	length, err = s.HStrLen("hash2", "f1")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if length != 0 {
		t.Fatalf("expected length 0 for empty string, got %d", length)
	}

	// 6. Empty field name
	s.HSet("hash3", []string{"", "value"})
	length, err = s.HStrLen("hash3", "")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if length != 5 {
		t.Fatalf("expected length 5 for empty field name, got %d", length)
	}

	// 7. Lazy expiration
	s.Set("hash_expired", types.Value{
		Type:      types.HashType,
		Data:      map[string]string{"f": "v"},
		ExpiresAt: time.Now().Add(-1 * time.Minute),
	})
	length, err = s.HStrLen("hash_expired", "f")
	if err != nil {
		t.Fatalf("expected nil error for expired key, got %v", err)
	}
	if length != 0 {
		t.Fatalf("expected length 0 for expired key, got %d", length)
	}
	if s.Exists("hash_expired") {
		t.Fatalf("expected key to be physically deleted after lazy expiration")
	}
}
