package utils

import (
	"context"
	"math"
	"sync"
	"time"
)

type tokenBucket struct {
	mu         sync.Mutex
	tokens     float64
	maxTokens  float64
	refillRate float64 // token per detik
	lastRefill time.Time
}

func newTokenBucket(ratePerSecond float64, burst float64) *tokenBucket {
	return &tokenBucket{
		tokens:     burst,
		maxTokens:  burst,
		refillRate: ratePerSecond,
		lastRefill: time.Now(),
	}
}

func (b *tokenBucket) refillLocked() {
	now := time.Now()
	elapsed := now.Sub(b.lastRefill).Seconds()
	if elapsed > 0 {
		b.tokens = math.Min(b.maxTokens, b.tokens+elapsed*b.refillRate)
		b.lastRefill = now
	}
}

func (b *tokenBucket) tryTake() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.refillLocked()
	if b.tokens >= 1 {
		b.tokens--
		return true
	}
	return false
}
func (b *tokenBucket) wait(ctx context.Context) error {
	if b.tryTake() {
		return nil
	}

	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if b.tryTake() {
				return nil
			}
		}
	}
}

var ildkiCredentialRateLimiter = newTokenBucket(6, 8)
var ildkiBackupRateLimiter = newTokenBucket(12, 15)
var ildkiRateLimitWaitTimeout = 30 * time.Second

func WaitForIldkiSlot(resourceType string) error {
	ctx, cancel := context.WithTimeout(context.Background(), ildkiRateLimitWaitTimeout)
	defer cancel()

	limiter := ildkiBackupRateLimiter
	if resourceType == "credential" {
		limiter = ildkiCredentialRateLimiter
	}
	return limiter.wait(ctx)
}

func SetIldkiCredentialRateLimiterForTest(ratePerSecond, burst float64, waitTimeout time.Duration) (restore func()) {
	origLimiter := ildkiCredentialRateLimiter
	origTimeout := ildkiRateLimitWaitTimeout

	ildkiCredentialRateLimiter = newTokenBucket(ratePerSecond, burst)
	ildkiRateLimitWaitTimeout = waitTimeout

	return func() {
		ildkiCredentialRateLimiter = origLimiter
		ildkiRateLimitWaitTimeout = origTimeout
	}
}

func SetIldkiBackupRateLimiterForTest(ratePerSecond, burst float64, waitTimeout time.Duration) (restore func()) {
	origLimiter := ildkiBackupRateLimiter
	origTimeout := ildkiRateLimitWaitTimeout

	ildkiBackupRateLimiter = newTokenBucket(ratePerSecond, burst)
	ildkiRateLimitWaitTimeout = waitTimeout

	return func() {
		ildkiBackupRateLimiter = origLimiter
		ildkiRateLimitWaitTimeout = origTimeout
	}
}
