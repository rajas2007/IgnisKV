package store

import (
	"time"

	"github.com/rajas2007/IgnisKV/internal/types"
)

// isExpired reports whether a Value has passed its expiration deadline.
//
// A zero ExpiresAt field indicates a persistent key that never expires.
// This helper is the single authoritative definition of expiration used
// throughout the Store. It is pure: it has no side effects, performs no
// synchronization, and acquires no locks. Callers are responsible for
// holding appropriate locks before invoking this function.
func isExpired(v types.Value) bool {
	return !v.ExpiresAt.IsZero() && time.Now().After(v.ExpiresAt)
}
