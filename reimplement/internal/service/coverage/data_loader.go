// Package coverage provides coverage calculation services for schedule management.
package coverage

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/schedcu/reimplement/internal/entity"
)

var (
	// ErrRepositoryFailed indicates a repository operation failed.
	ErrRepositoryFailed = errors.New("repository operation failed")

	// ErrNilRepository indicates a nil repository was passed.
	ErrNilRepository = errors.New("repository cannot be nil")

	// ErrInvalidScheduleVersion indicates an invalid schedule version ID.
	ErrInvalidScheduleVersion = errors.New("invalid schedule version ID")
)

// ShiftInstanceRepositoryLoader defines the interface for loading shift instances.
// This is a minimal interface focused on the batch loading operation needed for coverage calculation.
type ShiftInstanceRepositoryLoader interface {
	// GetByScheduleVersion retrieves all shift instances for a given schedule version.
	// This is the batch query method that prevents N+1 queries.
	// Returns an empty slice if no shifts exist (not an error).
	GetByScheduleVersion(ctx context.Context, scheduleVersionID uuid.UUID) ([]*entity.ShiftInstance, error)
}

// CoverageDataLoader loads assignment data for coverage calculation.
// It implements the batch query pattern to prevent N+1 queries when loading
// all assignments for a schedule version before passing to the coverage algorithm.
//
// Design principles:
// - Single query per load operation (no N+1 pattern)
// - Linear time complexity O(n) in assignment count
// - Minimal memory allocation through direct repository access
// - Reuses repository connections provided by caller
// - No internal caching (caller controls caching strategy)
type CoverageDataLoader struct {
	// repository provides the batch query method
	repository ShiftInstanceRepositoryLoader
}

// NewCoverageDataLoader creates a new coverage data loader.
// The repository must be non-nil and implement GetByScheduleVersion with batch query semantics.
//
// Example usage:
//
//	loader := NewCoverageDataLoader(myRepository)
//	shifts, err := loader.LoadAssignmentsForScheduleVersion(ctx, scheduleVersionID)
//	if err != nil {
//		return fmt.Errorf("failed to load assignments: %w", err)
//	}
//	// Pass shifts to coverage algorithm
//	metrics := CalculateCoverage(shifts, requirements)
func NewCoverageDataLoader(repository ShiftInstanceRepositoryLoader) *CoverageDataLoader {
	return &CoverageDataLoader{
		repository: repository,
	}
}

// LoadAssignmentsForScheduleVersion loads all assignments for a schedule version in one query.
// This implements the batch query pattern requirement from work package [1.14].
//
// Guarantees:
// - Exactly 1 database query executed (no N+1 pattern)
// - O(n) time complexity where n = number of assignments
// - Returns assignments in repository order (typically ID or creation order)
// - Empty slice with nil error if no assignments exist
// - Respects context cancellation
//
// Arguments:
//   - ctx: context.Context for cancellation support
//   - scheduleVersionID: the schedule version to load assignments for
//
// Returns:
//   - []*entity.ShiftInstance: all shifts for the schedule version
//   - error: if schedule version is invalid, repository fails, or context is cancelled
//
// Example:
//
//	ctx := context.Background()
//	shifts, err := loader.LoadAssignmentsForScheduleVersion(ctx, scheduleVersionID)
//	if err != nil {
//		return fmt.Errorf("load failed: %w", err)
//	}
//	// shifts now contains all assignments ready for coverage calculation
func (l *CoverageDataLoader) LoadAssignmentsForScheduleVersion(
	ctx context.Context,
	scheduleVersionID uuid.UUID,
) ([]*entity.ShiftInstance, error) {
	// Validate inputs
	if l.repository == nil {
		return nil, ErrNilRepository
	}

	if scheduleVersionID == uuid.Nil {
		return nil, ErrInvalidScheduleVersion
	}

	// Single batch query - this is the core requirement of [1.14]
	shifts, err := l.repository.GetByScheduleVersion(ctx, scheduleVersionID)
	if err != nil {
		return nil, fmt.Errorf("failed to load assignments for schedule version %s: %w", scheduleVersionID, err)
	}

	// No post-processing needed - repository returns data ready for algorithm
	// Time complexity: O(n) where n = len(shifts)
	// Space complexity: O(1) - we don't allocate additional structures
	// Query count: exactly 1 (guaranteed by GetByScheduleVersion using IN clause or similar)

	return shifts, nil
}
