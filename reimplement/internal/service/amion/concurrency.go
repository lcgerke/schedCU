package amion

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
)

var (
	// ErrQueueFull is returned when Submit() is called but the job queue is full.
	ErrQueueFull = errors.New("job queue is full")

	// ErrPoolClosed is returned when Submit() is called after the pool has been closed.
	ErrPoolClosed = errors.New("pool is closed")

	// ErrPoolShutdown is returned when Wait() is called after the pool has been shutdown.
	ErrPoolShutdown = errors.New("pool has been shutdown")
)

// Job is a function that performs work in the goroutine pool.
// It receives a context that can be used for cancellation and timeout.
// It returns an error if the work failed.
type Job func(ctx context.Context) error

// GoroutinePool manages a pool of worker goroutines that process jobs from a queue.
// It enforces a maximum concurrency level and implements backpressure via a limited queue.
//
// Key features:
//   - Max 5 concurrent workers by default (configurable)
//   - Max 100 pending jobs in queue (configurable)
//   - Thread-safe job submission with ErrQueueFull backpressure
//   - Graceful shutdown via Close()
//   - Context-based cancellation support
//   - Request deduplication (optional)
//
// Example usage:
//
//	pool := NewGoroutinePool(5) // 5 workers, 100 job queue
//
//	job := func(ctx context.Context) error {
//	    // Do work here
//	    return nil
//	}
//
//	err := pool.Submit(job)
//	if err != nil {
//	    log.Println("queue is full")
//	}
//
//	ctx := context.Background()
//	pool.Wait(ctx)
//	pool.Close()
type GoroutinePool struct {
	maxWorkers     int
	maxQueueSize   int
	jobQueue       chan Job
	wg             sync.WaitGroup
	mu             sync.Mutex
	closed         bool
	shutdown       bool
	workerStarted  atomic.Bool
	activeWorkers  int32
	duplicateCache map[string]bool // URL -> seen
	dupMu          sync.Mutex
}

// NewGoroutinePool creates a new GoroutinePool with the specified number of worker goroutines.
// It uses a default queue size of 100.
//
// Parameters:
//   - maxWorkers: Maximum number of concurrent worker goroutines (e.g., 5)
//
// Returns:
//   - *GoroutinePool: A new goroutine pool instance
//
// Example:
//
//	pool := NewGoroutinePool(5)
//	defer pool.Close()
func NewGoroutinePool(maxWorkers int) *GoroutinePool {
	return NewGoroutinePoolWithQueueSize(maxWorkers, 100)
}

// NewGoroutinePoolWithQueueSize creates a new GoroutinePool with custom queue size.
//
// Parameters:
//   - maxWorkers: Maximum number of concurrent worker goroutines
//   - maxQueueSize: Maximum number of pending jobs in the queue
//
// Returns:
//   - *GoroutinePool: A new goroutine pool instance
//
// Example:
//
//	pool := NewGoroutinePoolWithQueueSize(5, 200)
func NewGoroutinePoolWithQueueSize(maxWorkers, maxQueueSize int) *GoroutinePool {
	return &GoroutinePool{
		maxWorkers:     maxWorkers,
		maxQueueSize:   maxQueueSize,
		jobQueue:       make(chan Job, maxQueueSize),
		duplicateCache: make(map[string]bool),
	}
}

// Submit submits a job to the pool for execution. It returns immediately if the
// job was queued, or ErrQueueFull if the queue is at capacity, or ErrPoolClosed
// if the pool has been closed.
//
// The job will be executed by one of the worker goroutines. If the context in
// Wait() is cancelled before the job runs, it will still execute but the context
// will be cancelled.
//
// Parameters:
//   - job: The Job function to execute
//
// Returns:
//   - error: ErrQueueFull if queue is full, ErrPoolClosed if pool is closed, nil on success
//
// Example:
//
//	err := pool.Submit(func(ctx context.Context) error {
//	    // Do work
//	    return nil
//	})
//	if err == ErrQueueFull {
//	    log.Println("queue is full, try again later")
//	}
func (p *GoroutinePool) Submit(job Job) error {
	p.mu.Lock()

	if p.closed {
		p.mu.Unlock()
		return ErrPoolClosed
	}

	if p.shutdown {
		p.mu.Unlock()
		return ErrPoolShutdown
	}

	// Start workers on first submission
	if !p.workerStarted.Load() {
		p.workerStarted.Store(true)
		p.startWorkers()
	}

	p.mu.Unlock()

	// Non-blocking send with select for queue-full backpressure
	select {
	case p.jobQueue <- job:
		return nil
	default:
		return ErrQueueFull
	}
}

// Wait waits for all submitted jobs to complete. It respects the provided context
// for cancellation and timeout. It returns nil if all jobs completed successfully,
// or the context error if cancelled.
//
// Note: Wait() automatically closes the job queue to signal workers to stop accepting new jobs.
// After calling Wait(), further Submit() calls will fail with ErrPoolClosed.
//
// Parameters:
//   - ctx: Context for cancellation and timeout
//
// Returns:
//   - error: Context error if cancelled, nil if all jobs completed
//
// Example:
//
//	ctx := context.Background()
//	pool.Submit(job1)
//	pool.Submit(job2)
//	err := pool.Wait(ctx)
//	if err != nil {
//	    log.Printf("wait failed: %v", err)
//	}
func (p *GoroutinePool) Wait(ctx context.Context) error {
	// First, close the job queue to signal no more submissions
	p.mu.Lock()
	if !p.closed {
		p.closed = true
		p.mu.Unlock()
		// Close the queue without waiting
		close(p.jobQueue)
	} else {
		p.mu.Unlock()
	}

	// Now wait for all workers to finish with context support
	// Create a done channel to signal completion
	done := make(chan error, 1)
	go func() {
		p.wg.Wait()
		done <- nil
	}()

	// Wait for either completion or context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return err
	}
}

// Close closes the pool and waits for all pending jobs to complete.
// After Close(), Submit() will return ErrPoolClosed.
// This is a blocking call and should only be called after no more jobs will be submitted.
//
// Returns:
//   - error: Always nil (reserved for future use)
//
// Example:
//
//	defer pool.Close()
func (p *GoroutinePool) Close() error {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil // Already closed
	}
	p.closed = true
	p.mu.Unlock()

	// Close the job queue to signal workers to stop
	// This is safe to call multiple times in Go
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Channel already closed, that's fine
			}
		}()
		close(p.jobQueue)
	}()

	// Wait for all workers to finish
	p.wg.Wait()

	return nil
}

// startWorkers starts the configured number of worker goroutines.
// This is called internally on the first Submit() call.
func (p *GoroutinePool) startWorkers() {
	for i := 0; i < p.maxWorkers; i++ {
		p.wg.Add(1)
		go p.worker()
	}
}

// worker is the main loop for a worker goroutine. It reads jobs from the queue
// and executes them until the queue is closed.
func (p *GoroutinePool) worker() {
	defer p.wg.Done()

	for job := range p.jobQueue {
		atomic.AddInt32(&p.activeWorkers, 1)

		// Execute the job with a background context
		// The job should respect context cancellation if needed
		ctx := context.Background()
		_ = job(ctx) // Ignore error; jobs handle their own error reporting

		atomic.AddInt32(&p.activeWorkers, -1)
	}
}

// ActiveWorkers returns the current number of active worker goroutines.
// This can be useful for monitoring and debugging.
//
// Returns:
//   - int: The current number of active workers
//
// Example:
//
//	fmt.Printf("Active workers: %d\n", pool.ActiveWorkers())
func (p *GoroutinePool) ActiveWorkers() int {
	return int(atomic.LoadInt32(&p.activeWorkers))
}

// IsDuplicate checks if a URL has already been fetched in the current batch.
// Returns true if it's a duplicate, false if it's new.
//
// This is useful for preventing redundant requests.
//
// Parameters:
//   - url: The URL to check
//
// Returns:
//   - bool: True if duplicate, false if new
//
// Example:
//
//	if !pool.IsDuplicate(url) {
//	    pool.MarkSeen(url)
//	    // Fetch the URL
//	}
func (p *GoroutinePool) IsDuplicate(url string) bool {
	p.dupMu.Lock()
	defer p.dupMu.Unlock()

	return p.duplicateCache[url]
}

// MarkSeen marks a URL as seen to enable deduplication.
// This should be called before or after fetching a URL to track
// what's been processed.
//
// Parameters:
//   - url: The URL to mark as seen
//
// Example:
//
//	pool.MarkSeen(url)
func (p *GoroutinePool) MarkSeen(url string) {
	p.dupMu.Lock()
	defer p.dupMu.Unlock()

	p.duplicateCache[url] = true
}

// ClearDuplicateCache clears the duplicate detection cache. This is useful
// when starting a new batch of jobs and you want to allow URLs to be
// fetched again.
//
// Example:
//
//	pool.ClearDuplicateCache()
func (p *GoroutinePool) ClearDuplicateCache() {
	p.dupMu.Lock()
	defer p.dupMu.Unlock()

	p.duplicateCache = make(map[string]bool)
}

// QueueDepth returns the current number of pending jobs in the queue.
// This can be useful for monitoring backpressure.
//
// Returns:
//   - int: The current number of pending jobs
//
// Example:
//
//	fmt.Printf("Queue depth: %d\n", pool.QueueDepth())
func (p *GoroutinePool) QueueDepth() int {
	return len(p.jobQueue)
}
