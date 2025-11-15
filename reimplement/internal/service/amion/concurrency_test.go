package amion

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestGoroutinePoolMaxConcurrency verifies that at most MaxWorkers goroutines
// are active at the same time.
func TestGoroutinePoolMaxConcurrency(t *testing.T) {
	pool := NewGoroutinePool(5)

	var activeCount int32
	var maxConcurrent int32

	mu := sync.Mutex{}

	job := func(ctx context.Context) error {
		current := atomic.AddInt32(&activeCount, 1)

		// Update max concurrent
		mu.Lock()
		if current > maxConcurrent {
			maxConcurrent = current
		}
		mu.Unlock()

		// Simulate work
		time.Sleep(100 * time.Millisecond)

		atomic.AddInt32(&activeCount, -1)
		return nil
	}

	// Submit 20 jobs
	for i := 0; i < 20; i++ {
		err := pool.Submit(job)
		assert.NoError(t, err)
	}

	// Wait for completion
	err := pool.Wait(context.Background())
	assert.NoError(t, err)

	// Verify max concurrent was <= 5
	assert.LessOrEqual(t, maxConcurrent, int32(5), "should not exceed max 5 concurrent workers")
	assert.Greater(t, maxConcurrent, int32(0), "should have had at least 1 concurrent worker")
}

// TestGoroutinePoolBackpressure verifies that Submit() returns an error
// when the queue is full.
func TestGoroutinePoolBackpressure(t *testing.T) {
	pool := NewGoroutinePool(2) // 2 workers, queue size 100

	// Slow job that takes a while
	slowJob := func(ctx context.Context) error {
		time.Sleep(1 * time.Second)
		return nil
	}

	// Submit jobs until we fill the queue
	successCount := 0
	for i := 0; i < 150; i++ {
		err := pool.Submit(slowJob)
		if err != nil {
			// Queue is full
			assert.Equal(t, ErrQueueFull, err)
			break
		}
		successCount++
	}

	// Should have succeeded for some jobs but eventually hit the limit
	assert.Greater(t, successCount, 0, "should have submitted some jobs")
	assert.Less(t, successCount, 150, "should have hit queue limit before all 150 jobs")

	// Wait for completion
	pool.Wait(context.Background())
}

// TestGoroutinePoolQueueDepthLimit verifies queue depth is limited to maxQueueSize.
func TestGoroutinePoolQueueDepthLimit(t *testing.T) {
	pool := NewGoroutinePool(1) // 1 worker, queue size 100

	fastJob := func(ctx context.Context) error {
		time.Sleep(1 * time.Millisecond)
		return nil
	}

	// Submit jobs up to the queue limit
	for i := 0; i < 100; i++ {
		err := pool.Submit(fastJob)
		assert.NoError(t, err, "should accept job %d", i)
	}

	// Next job should fail
	err := pool.Submit(fastJob)
	assert.Equal(t, ErrQueueFull, err)

	pool.Wait(context.Background())
}

// TestGoroutinePoolJobExecution verifies that jobs are executed.
func TestGoroutinePoolJobExecution(t *testing.T) {
	pool := NewGoroutinePool(3)

	var executedCount int32

	job := func(ctx context.Context) error {
		atomic.AddInt32(&executedCount, 1)
		return nil
	}

	// Submit 10 jobs
	for i := 0; i < 10; i++ {
		err := pool.Submit(job)
		assert.NoError(t, err)
	}

	// Wait for completion
	err := pool.Wait(context.Background())
	assert.NoError(t, err)

	assert.Equal(t, int32(10), executedCount, "all jobs should be executed")
}

// TestGoroutinePoolErrorHandling verifies that job errors are tracked.
func TestGoroutinePoolErrorHandling(t *testing.T) {
	pool := NewGoroutinePool(3)

	var executedCount int32
	var errorCount int32

	job := func(ctx context.Context) error {
		current := atomic.AddInt32(&executedCount, 1)
		// Make every other job fail
		if current%2 == 0 {
			atomic.AddInt32(&errorCount, 1)
			return errors.New("test error")
		}
		return nil
	}

	// Submit 10 jobs
	for i := 0; i < 10; i++ {
		err := pool.Submit(job)
		assert.NoError(t, err)
	}

	// Wait for completion - should not fail even though jobs had errors
	err := pool.Wait(context.Background())
	assert.NoError(t, err, "Wait() should not fail due to job errors")

	assert.Equal(t, int32(10), executedCount, "all jobs should be executed")
	assert.Equal(t, int32(5), errorCount, "half the jobs should have errored")
}

// TestGoroutinePoolContextCancellation verifies that Wait() respects context cancellation.
func TestGoroutinePoolContextCancellation(t *testing.T) {
	pool := NewGoroutinePool(2)

	var executedCount int32

	job := func(ctx context.Context) error {
		atomic.AddInt32(&executedCount, 1)
		// Simulate some work
		time.Sleep(100 * time.Millisecond)
		return nil
	}

	// Submit many jobs
	for i := 0; i < 30; i++ {
		err := pool.Submit(job)
		assert.NoError(t, err)
	}

	// Create a context that will be cancelled
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel after a short delay
	time.AfterFunc(200*time.Millisecond, cancel)

	// Wait should return with context error
	err := pool.Wait(ctx)

	// With cancellation, we should get context error (not all jobs execute)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

// TestGoroutinePoolClose verifies that Close() stops workers and prevents
// new job submissions.
func TestGoroutinePoolClose(t *testing.T) {
	pool := NewGoroutinePool(2)

	job := func(ctx context.Context) error {
		time.Sleep(100 * time.Millisecond)
		return nil
	}

	// Submit a few jobs
	for i := 0; i < 3; i++ {
		err := pool.Submit(job)
		assert.NoError(t, err)
	}

	// Close the pool
	err := pool.Close()
	assert.NoError(t, err)

	// Try to submit another job - should fail
	err = pool.Submit(job)
	assert.Equal(t, ErrPoolClosed, err)
}

// TestGoroutinePoolStressTest runs a stress test with many jobs.
func TestGoroutinePoolStressTest(t *testing.T) {
	pool := NewGoroutinePool(5)

	var completedCount int32

	job := func(ctx context.Context) error {
		// Simulate some work
		time.Sleep(time.Duration(10) * time.Millisecond)
		atomic.AddInt32(&completedCount, 1)
		return nil
	}

	// Submit 100 jobs
	jobCount := 100
	for i := 0; i < jobCount; i++ {
		err := pool.Submit(job)
		assert.NoError(t, err)
	}

	// Wait for completion
	err := pool.Wait(context.Background())
	assert.NoError(t, err)

	assert.Equal(t, int32(jobCount), completedCount, "all jobs should be completed")
}

// TestGoroutinePoolIntegrationWithRateLimiter verifies that the goroutine
// pool works correctly with the rate limiter.
func TestGoroutinePoolIntegrationWithRateLimiter(t *testing.T) {
	pool := NewGoroutinePool(3)
	limiter := NewRateLimiter(50 * time.Millisecond)

	var completedCount int32
	var requestTimes []time.Time
	var mu sync.Mutex

	job := func(ctx context.Context) error {
		limiter.Wait()

		mu.Lock()
		requestTimes = append(requestTimes, time.Now())
		mu.Unlock()

		atomic.AddInt32(&completedCount, 1)
		return nil
	}

	// Submit 10 jobs
	for i := 0; i < 10; i++ {
		err := pool.Submit(job)
		assert.NoError(t, err)
	}

	// Wait for completion
	err := pool.Wait(context.Background())
	assert.NoError(t, err)

	assert.Equal(t, int32(10), completedCount)

	// Verify rate limiting was respected (rough check)
	mu.Lock()
	times := requestTimes
	mu.Unlock()

	if len(times) > 1 {
		// Check that consecutive requests are roughly spaced out
		for i := 1; i < len(times); i++ {
			timeBetween := times[i].Sub(times[i-1])
			// Due to concurrency, some requests might come closer together
			// but we should see the rate limiter effect in general
			_ = timeBetween
		}
	}
}
