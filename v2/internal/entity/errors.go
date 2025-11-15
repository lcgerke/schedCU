package entity

import "errors"

// Domain-specific errors
var (
	ErrInvalidVersionStateTransition = errors.New("invalid version state transition")
	ErrCannotArchiveNonProduction    = errors.New("cannot archive non-production version")
	ErrInvalidDateRange              = errors.New("invalid date range: end date must be after start date")
	ErrEmptyValidationResult         = errors.New("validation result cannot be empty")
	ErrUnknownShiftType              = errors.New("unknown shift type")
	ErrUnknownSpecialty              = errors.New("unknown specialty type")
)

// ValidateVersionStatus validates a version status string
func ValidateVersionStatus(status string) bool {
	return status == string(VersionStatusStaging) ||
		status == string(VersionStatusProduction) ||
		status == string(VersionStatusArchived)
}

// ValidateBatchState validates a batch state string
func ValidateBatchState(state string) bool {
	return state == string(BatchStatePending) ||
		state == string(BatchStateComplete) ||
		state == string(BatchStateFailed)
}

// ValidateSpecialty validates a specialty type
func ValidateSpecialty(specialty string) bool {
	return specialty == string(SpecialtyBodyOnly) ||
		specialty == string(SpecialtyNeuroOnly) ||
		specialty == string(SpecialtyBoth)
}

// ValidateShiftType validates a shift type
func ValidateShiftType(shiftType string) bool {
	return shiftType == string(ShiftTypeON1) ||
		shiftType == string(ShiftTypeON2) ||
		shiftType == string(ShiftTypeMidC) ||
		shiftType == string(ShiftTypeMidL) ||
		shiftType == string(ShiftTypeDay)
}
