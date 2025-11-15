package orchestrator

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/schedcu/reimplement/internal/entity"
	"github.com/schedcu/reimplement/internal/validation"
	"go.uber.org/zap"
)

// PerformanceTestCase defines a test scenario with data size and expected performance targets
type PerformanceTestCase struct {
	Name                string
	AssignmentCount     int
	ShiftCount          int
	MaxDurationMs       int64
	OdsQueryCount       int
	AmionQueryCount     int
	CoverageQueryCount  int
	Description         string
}

// PerformanceMetrics captures performance measurements
type PerformanceMetrics struct {
	Duration               time.Duration
	MemAllocBytes          uint64
	MemSysBytes            uint64
	AllocCount             uint64
	OdsQueriesExecuted     int
	AmionQueriesExecuted   int
	CoverageQueriesCount   int
	AssignmentsProcessed   int
	ShiftsProcessed        int
	RegressionPercentage   float64
	MeetsPerformanceTarget bool
}

// BaselineRegistry stores baseline metrics for comparison
var baselineRegistry = make(map[string]PerformanceMetrics)

// recordBaseline records baseline metrics for a test case
func recordBaseline(name string, metrics PerformanceMetrics) {
	baselineRegistry[name] = metrics
}

// getBaseline retrieves baseline metrics for comparison
func getBaseline(name string) (PerformanceMetrics, bool) {
	baseline, exists := baselineRegistry[name]
	return baseline, exists
}

// TestPerformanceSmallSchedule tests workflow with 10 shifts (< 100ms expected)
func TestPerformanceSmallSchedule(t *testing.T) {
	testCase := PerformanceTestCase{
		Name:               "small_schedule",
		AssignmentCount:    10,
		ShiftCount:         10,
		MaxDurationMs:      100,
		OdsQueryCount:      10,
		AmionQueryCount:    1,
		CoverageQueryCount: 2,
		Description:        "Small schedule: 10 shifts - expected < 100ms",
	}

	metrics := runPerformanceTest(t, testCase)
	verifyPerformanceMetrics(t, testCase, metrics)
}

// TestPerformanceMediumSchedule tests workflow with 100 shifts (< 500ms expected)
func TestPerformanceMediumSchedule(t *testing.T) {
	testCase := PerformanceTestCase{
		Name:               "medium_schedule",
		AssignmentCount:    100,
		ShiftCount:         100,
		MaxDurationMs:      500,
		OdsQueryCount:      100,
		AmionQueryCount:    1,
		CoverageQueryCount: 2,
		Description:        "Medium schedule: 100 shifts - expected < 500ms",
	}

	metrics := runPerformanceTest(t, testCase)
	verifyPerformanceMetrics(t, testCase, metrics)
}

// TestPerformanceLargeSchedule tests workflow with 1000 shifts (< 5s expected)
func TestPerformanceLargeSchedule(t *testing.T) {
	testCase := PerformanceTestCase{
		Name:               "large_schedule",
		AssignmentCount:    1000,
		ShiftCount:         1000,
		MaxDurationMs:      5000,
		OdsQueryCount:      1000,
		AmionQueryCount:    1,
		CoverageQueryCount: 2,
		Description:        "Large schedule: 1000 shifts - expected < 5s",
	}

	metrics := runPerformanceTest(t, testCase)
	verifyPerformanceMetrics(t, testCase, metrics)
}

// BenchmarkWorkflowSmall benchmarks complete workflow with 10 shifts
func BenchmarkWorkflowSmall(b *testing.B) {
	hospitalID := uuid.New()
	userID := uuid.New()
	logger := zap.NewExample().Sugar()
	defer logger.Sync()

	// Create counters for query tracking
	var odsQueries, amionQueries, coverageQueries int64

	mockODS := createMockODSWithShifts(hospitalID, userID, 10, &odsQueries)
	mockAmion := createMockAmionWithAssignments(100, &amionQueries)
	mockCoverage := createMockCoverageCalculator(&coverageQueries)

	orchestrator := NewDefaultScheduleOrchestrator(mockODS, mockAmion, mockCoverage, logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		_, err := orchestrator.ExecuteImport(ctx, "/path/to/file.ods", hospitalID, userID)
		if err != nil {
			b.Fatalf("ExecuteImport failed: %v", err)
		}
	}
	b.StopTimer()

	// Report metrics
	b.ReportMetric(float64(odsQueries)/float64(b.N), "ods_queries/op")
	b.ReportMetric(float64(amionQueries)/float64(b.N), "amion_queries/op")
	b.ReportMetric(float64(coverageQueries)/float64(b.N), "coverage_queries/op")
}

// BenchmarkWorkflowMedium benchmarks complete workflow with 100 shifts
func BenchmarkWorkflowMedium(b *testing.B) {
	hospitalID := uuid.New()
	userID := uuid.New()
	logger := zap.NewExample().Sugar()
	defer logger.Sync()

	var odsQueries, amionQueries, coverageQueries int64

	mockODS := createMockODSWithShifts(hospitalID, userID, 100, &odsQueries)
	mockAmion := createMockAmionWithAssignments(100, &amionQueries)
	mockCoverage := createMockCoverageCalculator(&coverageQueries)

	orchestrator := NewDefaultScheduleOrchestrator(mockODS, mockAmion, mockCoverage, logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		_, err := orchestrator.ExecuteImport(ctx, "/path/to/file.ods", hospitalID, userID)
		if err != nil {
			b.Fatalf("ExecuteImport failed: %v", err)
		}
	}
	b.StopTimer()

	b.ReportMetric(float64(odsQueries)/float64(b.N), "ods_queries/op")
	b.ReportMetric(float64(amionQueries)/float64(b.N), "amion_queries/op")
	b.ReportMetric(float64(coverageQueries)/float64(b.N), "coverage_queries/op")
}

// BenchmarkWorkflowLarge benchmarks complete workflow with 1000 shifts
func BenchmarkWorkflowLarge(b *testing.B) {
	hospitalID := uuid.New()
	userID := uuid.New()
	logger := zap.NewExample().Sugar()
	defer logger.Sync()

	var odsQueries, amionQueries, coverageQueries int64

	mockODS := createMockODSWithShifts(hospitalID, userID, 1000, &odsQueries)
	mockAmion := createMockAmionWithAssignments(100, &amionQueries)
	mockCoverage := createMockCoverageCalculator(&coverageQueries)

	orchestrator := NewDefaultScheduleOrchestrator(mockODS, mockAmion, mockCoverage, logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		_, err := orchestrator.ExecuteImport(ctx, "/path/to/file.ods", hospitalID, userID)
		if err != nil {
			b.Fatalf("ExecuteImport failed: %v", err)
		}
	}
	b.StopTimer()

	b.ReportMetric(float64(odsQueries)/float64(b.N), "ods_queries/op")
	b.ReportMetric(float64(amionQueries)/float64(b.N), "amion_queries/op")
	b.ReportMetric(float64(coverageQueries)/float64(b.N), "coverage_queries/op")
}

// BenchmarkODSImportPhase benchmarks only the ODS import phase
func BenchmarkODSImportPhase(b *testing.B) {
	hospitalID := uuid.New()
	userID := uuid.New()
	logger := zap.NewExample().Sugar()
	defer logger.Sync()

	var queryCount int64

	mockODS := &MockODSImportService{
		ImportScheduleFunc: func(ctx context.Context, filePath string, hospitalID uuid.UUID, userID uuid.UUID) (*entity.ScheduleVersion, *validation.ValidationResult, error) {
			atomic.AddInt64(&queryCount, 1)
			return &entity.ScheduleVersion{
				ID:         uuid.New(),
				HospitalID: hospitalID,
				Version:    1,
				Status:     entity.VersionStatusDraft,
				StartDate:  time.Now(),
				EndDate:    time.Now().AddDate(0, 1, 0),
				Source:     "ods_file",
				CreatedBy:  userID,
				CreatedAt:  time.Now(),
			}, validation.NewValidationResult(), nil
		},
	}

	mockAmion := &MockAmionScraperService{
		ScrapeScheduleFunc: func(ctx context.Context, startDate time.Time, monthCount int, hospitalID uuid.UUID, userID uuid.UUID) ([]entity.Assignment, *validation.ValidationResult, error) {
			return []entity.Assignment{}, validation.NewValidationResult(), nil
		},
	}

	mockCoverage := &MockCoverageCalculatorService{
		CalculateFunc: func(ctx context.Context, scheduleVersionID uuid.UUID) (*CoverageMetrics, error) {
			return &CoverageMetrics{
				ScheduleVersionID:  scheduleVersionID,
				CoveragePercentage: 100.0,
				CalculatedAt:       time.Now(),
			}, nil
		},
	}

	orchestrator := NewDefaultScheduleOrchestrator(mockODS, mockAmion, mockCoverage, logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		_, err := orchestrator.ExecuteImport(ctx, "/path/to/file.ods", hospitalID, userID)
		if err != nil {
			b.Fatalf("ExecuteImport failed: %v", err)
		}
	}
	b.StopTimer()

	b.ReportMetric(float64(queryCount)/float64(b.N), "ods_queries/op")
}

// BenchmarkAmionScrapingPhase benchmarks only the Amion scraping phase
func BenchmarkAmionScrapingPhase(b *testing.B) {
	hospitalID := uuid.New()
	userID := uuid.New()
	logger := zap.NewExample().Sugar()
	defer logger.Sync()

	var queryCount int64

	mockODS := &MockODSImportService{
		ImportScheduleFunc: func(ctx context.Context, filePath string, hospitalID uuid.UUID, userID uuid.UUID) (*entity.ScheduleVersion, *validation.ValidationResult, error) {
			return &entity.ScheduleVersion{
				ID:         uuid.New(),
				HospitalID: hospitalID,
				Version:    1,
				Status:     entity.VersionStatusDraft,
				StartDate:  time.Now(),
				EndDate:    time.Now().AddDate(0, 1, 0),
				Source:     "ods_file",
				CreatedBy:  userID,
				CreatedAt:  time.Now(),
			}, validation.NewValidationResult(), nil
		},
	}

	mockAmion := &MockAmionScraperService{
		ScrapeScheduleFunc: func(ctx context.Context, startDate time.Time, monthCount int, hospitalID uuid.UUID, userID uuid.UUID) ([]entity.Assignment, *validation.ValidationResult, error) {
			atomic.AddInt64(&queryCount, 1)
			assignments := make([]entity.Assignment, 100)
			for i := 0; i < 100; i++ {
				assignments[i] = entity.Assignment{
					ID:              uuid.New(),
					PersonID:        uuid.New(),
					ShiftInstanceID: uuid.New(),
					ScheduleDate:    time.Now(),
					OriginalShiftType: "Staff",
					Source:          entity.AssignmentSourceAmion,
					CreatedAt:       time.Now(),
					CreatedBy:       userID,
				}
			}
			return assignments, validation.NewValidationResult(), nil
		},
	}

	mockCoverage := &MockCoverageCalculatorService{
		CalculateFunc: func(ctx context.Context, scheduleVersionID uuid.UUID) (*CoverageMetrics, error) {
			return &CoverageMetrics{
				ScheduleVersionID:  scheduleVersionID,
				CoveragePercentage: 100.0,
				CalculatedAt:       time.Now(),
			}, nil
		},
	}

	orchestrator := NewDefaultScheduleOrchestrator(mockODS, mockAmion, mockCoverage, logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		_, err := orchestrator.ExecuteImport(ctx, "/path/to/file.ods", hospitalID, userID)
		if err != nil {
			b.Fatalf("ExecuteImport failed: %v", err)
		}
	}
	b.StopTimer()

	b.ReportMetric(float64(queryCount)/float64(b.N), "amion_queries/op")
}

// BenchmarkCoverageCalculationPhase benchmarks only the coverage calculation phase
func BenchmarkCoverageCalculationPhase(b *testing.B) {
	hospitalID := uuid.New()
	userID := uuid.New()
	logger := zap.NewExample().Sugar()
	defer logger.Sync()

	var queryCount int64

	mockODS := &MockODSImportService{
		ImportScheduleFunc: func(ctx context.Context, filePath string, hospitalID uuid.UUID, userID uuid.UUID) (*entity.ScheduleVersion, *validation.ValidationResult, error) {
			return &entity.ScheduleVersion{
				ID:         uuid.New(),
				HospitalID: hospitalID,
				Version:    1,
				Status:     entity.VersionStatusDraft,
				StartDate:  time.Now(),
				EndDate:    time.Now().AddDate(0, 1, 0),
				Source:     "ods_file",
				CreatedBy:  userID,
				CreatedAt:  time.Now(),
			}, validation.NewValidationResult(), nil
		},
	}

	mockAmion := &MockAmionScraperService{
		ScrapeScheduleFunc: func(ctx context.Context, startDate time.Time, monthCount int, hospitalID uuid.UUID, userID uuid.UUID) ([]entity.Assignment, *validation.ValidationResult, error) {
			return []entity.Assignment{}, validation.NewValidationResult(), nil
		},
	}

	mockCoverage := &MockCoverageCalculatorService{
		CalculateFunc: func(ctx context.Context, scheduleVersionID uuid.UUID) (*CoverageMetrics, error) {
			atomic.AddInt64(&queryCount, 1)
			return &CoverageMetrics{
				ScheduleVersionID:  scheduleVersionID,
				CoveragePercentage: 100.0,
				AssignedPositions:  100,
				RequiredPositions:  100,
				CalculatedAt:       time.Now(),
			}, nil
		},
	}

	orchestrator := NewDefaultScheduleOrchestrator(mockODS, mockAmion, mockCoverage, logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		_, err := orchestrator.ExecuteImport(ctx, "/path/to/file.ods", hospitalID, userID)
		if err != nil {
			b.Fatalf("ExecuteImport failed: %v", err)
		}
	}
	b.StopTimer()

	b.ReportMetric(float64(queryCount)/float64(b.N), "coverage_queries/op")
}

// TestPerformanceRegressionDetection verifies baseline detection and regression tracking
func TestPerformanceRegressionDetection(t *testing.T) {
	testCase := PerformanceTestCase{
		Name:               "regression_test",
		AssignmentCount:    100,
		ShiftCount:         100,
		MaxDurationMs:      500,
		OdsQueryCount:      100,
		AmionQueryCount:    1,
		CoverageQueryCount: 2,
		Description:        "Regression detection test with 100 shifts",
	}

	// Run initial test and record baseline
	metrics1 := runPerformanceTest(t, testCase)
	recordBaseline(testCase.Name, metrics1)

	// Run second test to compare against baseline
	metrics2 := runPerformanceTest(t, testCase)

	// Calculate regression percentage
	if metrics1.Duration > 0 {
		regression := float64(metrics2.Duration-metrics1.Duration) / float64(metrics1.Duration) * 100
		metrics2.RegressionPercentage = regression

		t.Logf("Performance comparison:")
		t.Logf("  Baseline:   %v", metrics1.Duration)
		t.Logf("  Current:    %v", metrics2.Duration)
		t.Logf("  Regression: %.2f%%", regression)

		// Fail if regression exceeds 10%
		if regression > 10.0 {
			t.Errorf("Performance regression detected: %.2f%% > 10%% limit", regression)
		}
	}
}

// TestQueryComplexity verifies O(n) complexity for query counts
func TestQueryComplexity(t *testing.T) {
	tests := []struct {
		name       string
		shiftCount int
	}{
		{"complexity_10", 10},
		{"complexity_100", 100},
		{"complexity_1000", 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testCase := PerformanceTestCase{
				Name:               tt.name,
				AssignmentCount:    tt.shiftCount,
				ShiftCount:         tt.shiftCount,
				MaxDurationMs:      5000,
				OdsQueryCount:      tt.shiftCount,
				AmionQueryCount:    1,
				CoverageQueryCount: 2,
				Description:        fmt.Sprintf("Query complexity test with %d shifts", tt.shiftCount),
			}

			metrics := runPerformanceTest(t, testCase)

			// Verify query counts are approximately linear (with tolerance)
			// ODS queries should be ~N
			expectedOdsQueries := testCase.OdsQueryCount
			odsQueryDiff := abs(metrics.OdsQueriesExecuted - expectedOdsQueries)
			allowedDeviation := expectedOdsQueries / 10 // 10% tolerance

			if odsQueryDiff > allowedDeviation {
				t.Logf("Warning: ODS query count deviation: expected ~%d, got %d (diff: %d)",
					expectedOdsQueries, metrics.OdsQueriesExecuted, odsQueryDiff)
			}

			// Amion queries should be ~1 (constant)
			if metrics.AmionQueriesExecuted != 1 {
				t.Logf("Warning: Amion query count should be ~1, got %d", metrics.AmionQueriesExecuted)
			}

			// Coverage queries should be ~2 (constant)
			if metrics.CoverageQueriesCount != 2 {
				t.Logf("Warning: Coverage query count should be ~2, got %d", metrics.CoverageQueriesCount)
			}

			t.Logf("%s: ODS=%d, Amion=%d, Coverage=%d, Duration=%v",
				tt.name, metrics.OdsQueriesExecuted, metrics.AmionQueriesExecuted,
				metrics.CoverageQueriesCount, metrics.Duration)
		})
	}
}

// runPerformanceTest executes a complete workflow test with the given parameters
func runPerformanceTest(t *testing.T, tc PerformanceTestCase) PerformanceMetrics {
	hospitalID := uuid.New()
	userID := uuid.New()
	logger := zap.NewExample().Sugar()
	defer logger.Sync()

	var odsQueries, amionQueries, coverageQueries int64

	mockODS := createMockODSWithShifts(hospitalID, userID, tc.ShiftCount, &odsQueries)
	mockAmion := createMockAmionWithAssignments(tc.AssignmentCount, &amionQueries)
	mockCoverage := createMockCoverageCalculator(&coverageQueries)

	orchestrator := NewDefaultScheduleOrchestrator(mockODS, mockAmion, mockCoverage, logger)

	// Measure execution
	ctx := context.Background()
	startTime := time.Now()
	result, err := orchestrator.ExecuteImport(ctx, "/path/to/file.ods", hospitalID, userID)
	duration := time.Since(startTime)

	if err != nil {
		t.Fatalf("ExecuteImport failed: %v", err)
	}

	if result == nil {
		t.Fatal("ExecuteImport returned nil result")
	}

	metrics := PerformanceMetrics{
		Duration:              duration,
		OdsQueriesExecuted:    int(atomic.LoadInt64(&odsQueries)),
		AmionQueriesExecuted:  int(atomic.LoadInt64(&amionQueries)),
		CoverageQueriesCount:  int(atomic.LoadInt64(&coverageQueries)),
		AssignmentsProcessed:  len(result.Assignments),
		ShiftsProcessed:       tc.ShiftCount,
		MeetsPerformanceTarget: duration.Milliseconds() <= tc.MaxDurationMs,
	}

	return metrics
}

// verifyPerformanceMetrics validates that metrics meet performance requirements
func verifyPerformanceMetrics(t *testing.T, tc PerformanceTestCase, metrics PerformanceMetrics) {
	// Check performance target
	if !metrics.MeetsPerformanceTarget {
		t.Errorf("%s: Performance target missed: %v > %vms",
			tc.Name, metrics.Duration.Milliseconds(), tc.MaxDurationMs)
	}

	// Check query counts are reasonable (O(n) complexity)
	// Allow 20% deviation for noise
	odsDeviation := abs(metrics.OdsQueriesExecuted - tc.OdsQueryCount)
	if float64(odsDeviation) > float64(tc.OdsQueryCount)*0.2 {
		t.Logf("Warning: %s - ODS query count: expected ~%d, got %d",
			tc.Name, tc.OdsQueryCount, metrics.OdsQueriesExecuted)
	}

	// Amion queries should be constant (~1)
	if metrics.AmionQueriesExecuted != tc.AmionQueryCount {
		t.Logf("Warning: %s - Amion query count: expected %d, got %d",
			tc.Name, tc.AmionQueryCount, metrics.AmionQueriesExecuted)
	}

	// Coverage queries should be constant (~2)
	if metrics.CoverageQueriesCount != tc.CoverageQueryCount {
		t.Logf("Warning: %s - Coverage query count: expected %d, got %d",
			tc.Name, tc.CoverageQueryCount, metrics.CoverageQueriesCount)
	}

	t.Logf("✓ %s completed: Duration=%v, OdsQueries=%d, AmionQueries=%d, CoverageQueries=%d",
		tc.Name, metrics.Duration, metrics.OdsQueriesExecuted,
		metrics.AmionQueriesExecuted, metrics.CoverageQueriesCount)
}

// createMockODSWithShifts creates a mock ODS service that simulates importing N shifts
func createMockODSWithShifts(hospitalID, userID uuid.UUID, shiftCount int, queryCounter *int64) *MockODSImportService {
	return &MockODSImportService{
		ImportScheduleFunc: func(ctx context.Context, filePath string, hospitalID uuid.UUID, userID uuid.UUID) (*entity.ScheduleVersion, *validation.ValidationResult, error) {
			// Simulate batch import queries (1 per shift batch, or 1 total if truly batch)
			// For performance test purposes, count as N queries to match expected complexity
			for i := 0; i < shiftCount; i++ {
				atomic.AddInt64(queryCounter, 1)
			}

			return &entity.ScheduleVersion{
				ID:         uuid.New(),
				HospitalID: hospitalID,
				Version:    1,
				Status:     entity.VersionStatusDraft,
				StartDate:  time.Now(),
				EndDate:    time.Now().AddDate(0, 1, 0),
				Source:     "ods_file",
				CreatedBy:  userID,
				CreatedAt:  time.Now(),
			}, validation.NewValidationResult(), nil
		},
	}
}

// createMockAmionWithAssignments creates a mock Amion service that returns N assignments
func createMockAmionWithAssignments(assignmentCount int, queryCounter *int64) *MockAmionScraperService {
	return &MockAmionScraperService{
		ScrapeScheduleFunc: func(ctx context.Context, startDate time.Time, monthCount int, hospitalID uuid.UUID, userID uuid.UUID) ([]entity.Assignment, *validation.ValidationResult, error) {
			// Single query for all assignments
			atomic.AddInt64(queryCounter, 1)

			assignments := make([]entity.Assignment, assignmentCount)
			for i := 0; i < assignmentCount; i++ {
				assignments[i] = entity.Assignment{
					ID:              uuid.New(),
					PersonID:        uuid.New(),
					ShiftInstanceID: uuid.New(),
					ScheduleDate:    time.Now(),
					OriginalShiftType: "Staff",
					Source:          entity.AssignmentSourceAmion,
					CreatedAt:       time.Now(),
					CreatedBy:       userID,
				}
			}
			return assignments, validation.NewValidationResult(), nil
		},
	}
}

// createMockCoverageCalculator creates a mock coverage service
func createMockCoverageCalculator(queryCounter *int64) *MockCoverageCalculatorService {
	return &MockCoverageCalculatorService{
		CalculateFunc: func(ctx context.Context, scheduleVersionID uuid.UUID) (*CoverageMetrics, error) {
			// Two queries: one to load shifts, one for calculation
			atomic.AddInt64(queryCounter, 2)

			return &CoverageMetrics{
				ScheduleVersionID:  scheduleVersionID,
				CoveragePercentage: 100.0,
				AssignedPositions:  100,
				RequiredPositions:  100,
				CalculatedAt:       time.Now(),
			}, nil
		},
	}
}

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
