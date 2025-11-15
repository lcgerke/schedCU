package amion

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestRateLimiterWaitEnforcesMinimumDelay tests that Wait() blocks until
// at least 1 second has elapsed since the last request.
func TestRateLimiterWaitEnforcesMinimumDelay(t *testing.T) {
	limiter := NewRateLimiter(1 * time.Second)

	// First request should complete immediately
	start := time.Now()
	limiter.Wait()
	duration := time.Since(start)
	assert.Less(t, duration, 100*time.Millisecond, "first Wait() should be immediate")

	// Second request should wait ~1 second
	start = time.Now()
	limiter.Wait()
	duration = time.Since(start)
	assert.GreaterOrEqual(t, duration, 900*time.Millisecond, "second Wait() should wait ~1 second")
	assert.Less(t, duration, 1200*time.Millisecond, "second Wait() should not wait much longer than 1 second")
}

// TestRateLimiterMultipleRequests verifies consistent 1-second delays between requests.
func TestRateLimiterMultipleRequests(t *testing.T) {
	limiter := NewRateLimiter(1 * time.Second)

	start := time.Now()
	for i := 0; i < 5; i++ {
		limiter.Wait()
	}
	totalDuration := time.Since(start)

	// Should have waited approximately 4 seconds for 5 requests (first is immediate)
	expectedMin := 3800 * time.Millisecond
	expectedMax := 5500 * time.Millisecond
	assert.GreaterOrEqual(t, totalDuration, expectedMin, "5 requests should take at least ~4 seconds")
	assert.Less(t, totalDuration, expectedMax, "5 requests should not take much longer than ~4 seconds")
}

// TestRateLimiterThreadSafety verifies the rate limiter works correctly with concurrent access.
func TestRateLimiterThreadSafety(t *testing.T) {
	limiter := NewRateLimiter(100 * time.Millisecond)

	var wg sync.WaitGroup
	var requestCount int32
	var lastRequestTime int64

	// Launch 10 goroutines, each making 3 requests
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 3; j++ {
				limiter.Wait()
				atomic.AddInt32(&requestCount, 1)
				now := time.Now().UnixNano()
				atomic.StoreInt64(&lastRequestTime, now)
			}
		}()
	}

	wg.Wait()
	assert.Equal(t, int32(30), requestCount, "should have completed all 30 requests")
}

// TestRateLimiterResetAfterWait verifies that the rate limiter can be reset.
func TestRateLimiterResetAfterWait(t *testing.T) {
	limiter := NewRateLimiter(100 * time.Millisecond)

	limiter.Wait()
	limiter.Reset()

	// After reset, next Wait() should be immediate
	start := time.Now()
	limiter.Wait()
	duration := time.Since(start)
	assert.Less(t, duration, 50*time.Millisecond, "after reset, Wait() should be immediate")
}

// TestRateLimiterDifferentIntervals tests the rate limiter with different intervals.
func TestRateLimiterDifferentIntervals(t *testing.T) {
	tests := []struct {
		name     string
		interval time.Duration
	}{
		{"100ms", 100 * time.Millisecond},
		{"500ms", 500 * time.Millisecond},
		{"1s", 1 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter := NewRateLimiter(tt.interval)

			// First wait is immediate
			start := time.Now()
			limiter.Wait()
			duration := time.Since(start)
			assert.Less(t, duration, 50*time.Millisecond)

			// Second wait should respect the interval
			start = time.Now()
			limiter.Wait()
			duration = time.Since(start)
			expectedMin := tt.interval - 50*time.Millisecond
			assert.GreaterOrEqual(t, duration, expectedMin)
		})
	}
}
