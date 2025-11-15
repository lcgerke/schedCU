// Package entity contains domain models for the schedCU application.
package entity

import (
	"time"

	"github.com/google/uuid"
)

// VersionStatus represents the status of a ScheduleVersion.
type VersionStatus string

const (
	VersionStatusDraft      VersionStatus = "draft"
	VersionStatusPublished  VersionStatus = "published"
	VersionStatusArchived   VersionStatus = "archived"
	VersionStatusDeprecated VersionStatus = "deprecated"
)

// ScheduleVersion represents a versioned snapshot of a hospital schedule.
// Each import creates a new version, allowing tracking of schedule changes over time.
type ScheduleVersion struct {
	ID          uuid.UUID
	HospitalID  uuid.UUID
	Version     int
	Status      VersionStatus
	StartDate   time.Time
	EndDate     time.Time
	Source      string // "ods_file", "manual", "amion", etc.
	Metadata    map[string]interface{}
	CreatedAt   time.Time
	CreatedBy   uuid.UUID
	UpdatedAt   time.Time
	UpdatedBy   uuid.UUID
	DeletedAt   *time.Time
}

// NewScheduleVersion creates a new ScheduleVersion with default values.
func NewScheduleVersion(
	hospitalID uuid.UUID,
	version int,
	startDate, endDate time.Time,
	source string,
	userID uuid.UUID,
) *ScheduleVersion {
	now := time.Now()
	return &ScheduleVersion{
		ID:         uuid.New(),
		HospitalID: hospitalID,
		Version:    version,
		Status:     VersionStatusDraft,
		StartDate:  startDate,
		EndDate:    endDate,
		Source:     source,
		Metadata:   make(map[string]interface{}),
		CreatedAt:  now,
		CreatedBy:  userID,
		UpdatedAt:  now,
		UpdatedBy:  userID,
	}
}
