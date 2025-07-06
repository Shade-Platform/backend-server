package trust

import (
	"sync"
	"time"
)

type FailedLoginTracker struct {
	attempts   map[string][]time.Time
	lock       sync.RWMutex
	expiry     time.Duration
	threshold  int
	penalty    int
}

// NewFailedLoginTracker initializes the tracker
func NewFailedLoginTracker(threshold int, expiry time.Duration, penalty int) *FailedLoginTracker {
	return &FailedLoginTracker{
		attempts:  make(map[string][]time.Time),
		expiry:    expiry,
		threshold: threshold,
		penalty:   penalty,
	}
}

// RecordFailure logs a failed attempt for an IP or user ID
func (t *FailedLoginTracker) RecordFailure(key string) int {
	t.lock.Lock()
	defer t.lock.Unlock()

	now := time.Now()
	t.cleanupOldAttempts(key, now)

	t.attempts[key] = append(t.attempts[key], now)
	return len(t.attempts[key]) // Return current count
}

// ShouldPenalize checks if penalty should be applied
func (t *FailedLoginTracker) ShouldPenalize(key string) (bool, int) {
	t.lock.RLock()
	defer t.lock.RUnlock()

	now := time.Now()
	t.cleanupOldAttempts(key, now)

	currentCount := len(t.attempts[key])
	if currentCount >= t.threshold {
		return true, t.penalty
	}
	return false, 0
}

// GetFailureCount returns current failure count for a key
func (t *FailedLoginTracker) GetFailureCount(key string) int {
	t.lock.RLock()
	defer t.lock.RUnlock()

	t.cleanupOldAttempts(key, time.Now())
	return len(t.attempts[key])
}

// ResetFailures clears the failed attempts for a key
func (t *FailedLoginTracker) ResetFailures(key string) {
	t.lock.Lock()
	defer t.lock.Unlock()
	delete(t.attempts, key)
}

// GetTimeUntilReset returns duration until oldest failure expires
func (t *FailedLoginTracker) GetTimeUntilReset(key string) time.Duration {
	t.lock.RLock()
	defer t.lock.RUnlock()

	now := time.Now()
	attempts := t.attempts[key]
	if len(attempts) == 0 {
		return 0
	}

	// Find oldest attempt
	oldest := attempts[0]
	for _, ts := range attempts {
		if ts.Before(oldest) {
			oldest = ts
		}
	}

	expiryTime := oldest.Add(t.expiry)
	return expiryTime.Sub(now)
}

// cleanupOldAttempts removes expired attempts
func (t *FailedLoginTracker) cleanupOldAttempts(key string, now time.Time) {
	attempts, exists := t.attempts[key]
	if !exists {
		return
	}

	filtered := attempts[:0]
	for _, ts := range attempts {
		if now.Sub(ts) <= t.expiry {
			filtered = append(filtered, ts)
		}
	}
	t.attempts[key] = filtered
}