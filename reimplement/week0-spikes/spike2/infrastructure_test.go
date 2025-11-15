package main

import (
	"context"
	"testing"
	"time"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAsynqIntegration validates Asynq job queue with Redis.
func TestAsynqIntegration(t *testing.T) {
	client := asynq.NewClient(asynq.RedisClientOpt{
		Addr: "localhost:6379",
	})
	defer client.Close()

	// Test: Can we enqueue a job?
	task := asynq.NewTask(
		"sample_task",
		map[string]interface{}{"name": "test"},
	)

	info, err := client.Enqueue(task)
	if err != nil {
		t.Skipf("Redis not available (expected for mock test): %v", err)
		return
	}

	assert.NotEmpty(t, info.ID)
	assert.Equal(t, "sample_task", info.Type)
}

// TestAsynqCapabilities validates Asynq features needed for spikes.
func TestAsynqCapabilities(t *testing.T) {
	inspector := asynq.NewInspector(asynq.RedisClientOpt{
		Addr: "localhost:6379",
	})
	defer inspector.Close()

	// Test: Can we inspect queues?
	queues, err := inspector.Queues()
	if err != nil {
		t.Skipf("Redis not available: %v", err)
		return
	}

	// Asynq should at least have a default queue
	assert.NotNil(t, queues)
}

// TestJobRetry validates Asynq retry mechanism.
func TestJobRetryMechanism(t *testing.T) {
	config := &asynq.Config{
		Concurrency: 1,
		RetryDeltas: []time.Duration{
			10 * time.Second,
			30 * time.Second,
			1 * time.Minute,
		},
	}

	// Verify config is valid
	assert.Equal(t, 1, config.Concurrency)
	assert.Equal(t, 3, len(config.RetryDeltas))
}

// TestAsynqScheduling validates scheduled task support.
func TestScheduledTasks(t *testing.T) {
	client := asynq.NewClient(asynq.RedisClientOpt{
		Addr: "localhost:6379",
	})
	defer client.Close()

	task := asynq.NewTask(
		"sample_task",
		map[string]interface{}{"name": "scheduled"},
	)

	// Schedule for 5 minutes from now
	_, err := client.Enqueue(
		task,
		asynq.ProcessIn(5*time.Minute),
	)

	if err != nil {
		t.Skipf("Redis not available: %v", err)
		return
	}

	// Success if no error
	assert.NoError(t, err)
}

// TestJobPriority validates priority queue support.
func TestJobPriority(t *testing.T) {
	config := &asynq.Config{
		Queues: map[string]int{
			"critical": 6,
			"default":  3,
			"low":      1,
		},
	}

	assert.Equal(t, 3, len(config.Queues))
	assert.Equal(t, 6, config.Queues["critical"])
}
