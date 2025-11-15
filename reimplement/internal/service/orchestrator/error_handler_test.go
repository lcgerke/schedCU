package orchestrator

import (
	"testing"

	"github.com/schedcu/reimplement/internal/validation"
)

func TestMergeValidationResultsEmpty(t *testing.T) {
	merger := NewErrorPropagator()
	result := merger.MergeValidationResults()

	if result == nil {
		t.Fatal("expected non-nil result, got nil")
	}
	if !result.IsValid() {
		t.Error("expected valid result for empty merge")
	}
}

func TestMergeValidationResultsSingle(t *testing.T) {
	merger := NewErrorPropagator()

	vr1 := validation.NewValidationResult()
	vr1.AddError("field1", "error message 1")
	vr1.AddWarning("field2", "warning message 1")

	result := merger.MergeValidationResults(vr1)

	if result.ErrorCount() != 1 {
		t.Errorf("expected 1 error, got %d", result.ErrorCount())
	}
	if result.WarningCount() != 1 {
		t.Errorf("expected 1 warning, got %d", result.WarningCount())
	}
}

func TestMergeValidationResultsMultiple(t *testing.T) {
	merger := NewErrorPropagator()

	vr1 := validation.NewValidationResult()
	vr1.AddError("phase1", "error1")
	vr2 := validation.NewValidationResult()
	vr2.AddError("phase2", "error2")
	vr3 := validation.NewValidationResult()
	vr3.AddWarning("phase3", "warning1")

	result := merger.MergeValidationResults(vr1, vr2, vr3)

	if result.ErrorCount() != 2 {
		t.Errorf("expected 2 errors, got %d", result.ErrorCount())
	}
	if result.WarningCount() != 1 {
		t.Errorf("expected 1 warning, got %d", result.WarningCount())
	}
}

func TestMergeValidationResultsWithPhaseContext(t *testing.T) {
	merger := NewErrorPropagator()

	vr1 := validation.NewValidationResult()
	vr1.AddError("file", "invalid format")
	vr2 := validation.NewValidationResult()
	vr2.AddError("conn", "timeout")
	vr3 := validation.NewValidationResult()
	vr3.AddError("algo", "failed")

	result := merger.MergeValidationResultsWithContext(map[int]*validation.ValidationResult{
		0: vr1,
		1: vr2,
		2: vr3,
	})

	if result.ErrorCount() != 3 {
		t.Errorf("expected 3 errors, got %d", result.ErrorCount())
	}
	if _, ok := result.GetContext("phases_with_errors"); !ok {
		t.Error("expected phases_with_errors in context")
	}
}

func TestShouldContinueOnWarning(t *testing.T) {
	merger := NewErrorPropagator()
	vr := validation.NewValidationResult()
	vr.AddWarning("field", "warning")

	if !merger.ShouldContinue(vr, PhaseODSImport) {
		t.Error("expected to continue on warnings")
	}
}

func TestShouldStopOnCriticalError(t *testing.T) {
	merger := NewErrorPropagator()
	vr := validation.NewValidationResult()
	vr.AddError("file", "parse error in file")

	if merger.ShouldContinue(vr, PhaseODSImport) {
		t.Error("expected to stop on critical error")
	}
}

func TestShouldStopOnConstraintViolation(t *testing.T) {
	merger := NewErrorPropagator()
	vr := validation.NewValidationResult()
	vr.AddError("db", "unique constraint violation")

	if merger.ShouldContinue(vr, PhaseAmionScrape) {
		t.Error("expected to stop on constraint violation")
	}
}

func TestShouldStopOnDiskFull(t *testing.T) {
	merger := NewErrorPropagator()
	vr := validation.NewValidationResult()
	vr.AddError("disk", "no space left on device")

	if merger.ShouldContinue(vr, PhaseCoverageCalculation) {
		t.Error("expected to stop on disk full")
	}
}

func TestContinueOnMajorErrorPhase2(t *testing.T) {
	merger := NewErrorPropagator()
	vr := validation.NewValidationResult()
	vr.AddError("shift", "invalid date format")

	if !merger.ShouldContinue(vr, PhaseAmionScrape) {
		t.Error("expected to continue with major error in Phase 2")
	}
}

func TestIsCriticalError(t *testing.T) {
	merger := NewErrorPropagator()
	vr := validation.NewValidationResult()
	vr.AddError("file", "invalid zip format")

	if !merger.IsCriticalError(vr, PhaseODSImport) {
		t.Error("expected invalid zip to be critical")
	}
}

func TestIsMajorError(t *testing.T) {
	merger := NewErrorPropagator()
	vr := validation.NewValidationResult()
	vr.AddError("shift", "invalid date format")

	if !merger.IsMajorError(vr) {
		t.Error("expected invalid date to be major error")
	}
}

func TestIsMinorError(t *testing.T) {
	merger := NewErrorPropagator()
	vr := validation.NewValidationResult()
	vr.AddWarning("field", "optional field missing")

	if !merger.IsMinorError(vr) {
		t.Error("expected warning to be minor")
	}
}

func TestMergeConflictingSeverities(t *testing.T) {
	merger := NewErrorPropagator()

	vr1 := validation.NewValidationResult()
	vr1.AddWarning("fmt", "deprecated format")

	vr2 := validation.NewValidationResult()
	vr2.AddError("conn", "connection timeout")

	result := merger.MergeValidationResults(vr1, vr2)

	if result.WarningCount() != 1 {
		t.Errorf("expected 1 warning, got %d", result.WarningCount())
	}
	if result.ErrorCount() != 1 {
		t.Errorf("expected 1 error, got %d", result.ErrorCount())
	}
	if result.IsValid() {
		t.Error("expected invalid result due to error")
	}
}

func TestNilValidationResultHandling(t *testing.T) {
	merger := NewErrorPropagator()

	vr1 := validation.NewValidationResult()
	vr1.AddError("field", "error")

	var vrNil *validation.ValidationResult

	result := merger.MergeValidationResults(vr1, vrNil)

	if result.ErrorCount() != 1 {
		t.Errorf("expected 1 error, got %d", result.ErrorCount())
	}
}

func TestInfoMessagesPreserved(t *testing.T) {
	merger := NewErrorPropagator()

	vr1 := validation.NewValidationResult()
	vr1.AddInfo("progress", "loaded 100 shifts")
	vr1.AddInfo("timing", "took 2.5s")

	vr2 := validation.NewValidationResult()
	vr2.AddInfo("progress", "created 50 assignments")

	result := merger.MergeValidationResults(vr1, vr2)

	if result.InfoCount() != 3 {
		t.Errorf("expected 3 infos, got %d", result.InfoCount())
	}
}

func TestCriticalErrorPatterns(t *testing.T) {
	tests := []struct {
		name    string
		msg     string
		want    bool
	}{
		{"parse_error", "parse error in file", true},
		{"invalid_zip", "invalid zip archive", true},
		{"corrupted", "corrupted file", true},
		{"constraint", "unique constraint violation", true},
		{"disk_full", "no space left on device", true},
		{"permission", "permission denied", true},
		{"timeout", "connection timeout", true},
		{"warning", "optional field missing", false},
		{"deprecated", "deprecated format", false},
	}

	for _, tt := range tests {
		vr := validation.NewValidationResult()
		vr.AddError("field", tt.msg)
		merger := NewErrorPropagator()
		got := merger.IsCriticalError(vr, PhaseODSImport)
		if got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestMajorErrorPatterns(t *testing.T) {
	tests := []struct {
		name string
		msg  string
		want bool
	}{
		{"invalid_date", "invalid date format", true},
		{"invalid_shift", "invalid shift type", true},
		{"unsupported", "unsupported shift type", true},
		{"missing_required", "missing required field", true},
		{"type_mismatch", "type mismatch in field", true},
		{"parse_error", "parse error in file", false},
		{"duplicate", "duplicate entry exists", false},
		{"optional", "optional field missing", false},
	}

	for _, tt := range tests {
		vr := validation.NewValidationResult()
		vr.AddError("field", tt.msg)
		merger := NewErrorPropagator()
		got := merger.IsMajorError(vr)
		if got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestPhaseContextTracking(t *testing.T) {
	merger := NewErrorPropagator()

	vr1 := validation.NewValidationResult()
	vr1.AddError("file", "parse error")

	vr2 := validation.NewValidationResult()
	vr2.AddError("conn", "timeout")

	result := merger.MergeValidationResultsWithContext(map[int]*validation.ValidationResult{
		0: vr1,
		1: vr2,
	})

	if len(result.Context) == 0 {
		t.Error("expected context to contain phase information")
	}
}

func TestMergeEmptyValidationResults(t *testing.T) {
	merger := NewErrorPropagator()

	vr1 := validation.NewValidationResult()
	vr2 := validation.NewValidationResult()
	vr3 := validation.NewValidationResult()

	result := merger.MergeValidationResults(vr1, vr2, vr3)

	if !result.IsValid() {
		t.Error("expected valid result when merging empty results")
	}
}

func TestPreservesContext(t *testing.T) {
	merger := NewErrorPropagator()

	vr1 := validation.NewValidationResult()
	vr1.SetContext("id", "123")

	vr2 := validation.NewValidationResult()
	vr2.SetContext("count", 42)

	result := merger.MergeValidationResults(vr1, vr2)

	if len(result.Context) < 2 {
		t.Error("expected context to be preserved")
	}
}
