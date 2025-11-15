package amion

import (
	"sync"
	"time"
)

// RateLimiter implements a simple token bucket algorithm to enforce
// a minimum interval between requests. It is thread-safe and can be used
// concurrently by multiple goroutines.
//
// The rate limiter ensures that Wait() blocks until at least the specified
// interval has elapsed since the last successful Wait() call.
//
// Example usage:
//
//	limiter := NewRateLimiter(1 * time.Second)
//	limiter.Wait() // Returns immediately on first call
//	limiter.Wait() // Waits ~1 second
//	limiter.Wait() // Waits ~1 second
type RateLimiter struct {
	mu               sync.Mutex
	lastRequestTime  time.Time
	minInterval      time.Duration
}

// NewRateLimiter creates a new RateLimiter with the specified minimum
// interval between requests.
//
// Parameters:
//   - minInterval: The minimum time to wait between requests
//
// Returns:
//   - *RateLimiter: A new rate limiter instance
//
// Example:
//
//	limiter := NewRateLimiter(1 * time.Second)
func NewRateLimiter(minInterval time.Duration) *RateLimiter {
	return &RateLimiter{
		minInterval:     minInterval,
		lastRequestTime: time.Now().Add(-minInterval), // Allow first request immediately
	}
}

// Wait blocks until at least minInterval has elapsed since the last request.
// It is safe to call from multiple goroutines concurrently.
//
// The first call returns immediately. Subsequent calls will block for
// approximately minInterval duration.
//
// Example:
//
//	limiter := NewRateLimiter(1 * time.Second)
//	limiter.Wait() // Returns immediately
//	limiter.Wait() // Blocks for ~1 second
func (rl *RateLimiter) Wait() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastRequestTime)

	if elapsed < rl.minInterval {
		// Sleep for the remaining time
		sleepDuration := rl.minInterval - elapsed
		rl.mu.Unlock()
		time.Sleep(sleepDuration)
		rl.mu.Lock()
	}

	// Update the last request time
	rl.lastRequestTime = time.Now()
}

// Reset resets the rate limiter so that the next Wait() call will return
// immediately. This can be useful if you want to restart the rate limiting
// window.
//
// Example:
//
//	limiter := NewRateLimiter(1 * time.Second)
//	limiter.Wait()
//	limiter.Reset()
//	limiter.Wait() // Returns immediately due to reset
func (rl *RateLimiter) Reset() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.lastRequestTime = time.Now().Add(-rl.minInterval)
}
