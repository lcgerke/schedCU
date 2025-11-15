package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/schedcu/v2/internal/entity"
	"github.com/schedcu/v2/internal/repository"
)

// ScheduleVersionService manages schedule versions (creation, promotion, archival)
type ScheduleVersionService struct {
	repo repository.ScheduleVersionRepository
}

// NewScheduleVersionService creates a new schedule version service
func NewScheduleVersionService(repo repository.ScheduleVersionRepository) *ScheduleVersionService {
	return &ScheduleVersionService{repo: repo}
}

// CreateVersion creates a new schedule version in STAGING status
func (s *ScheduleVersionService) CreateVersion(
	ctx context.Context,
	hospitalID entity.HospitalID,
	startDate, endDate entity.Date,
	creatorID entity.UserID,
) (*entity.ScheduleVersion, error) {

	version := &entity.ScheduleVersion{
		ID:                  entity.ScheduleVersionID(uuid.New()),
		HospitalID:          hospitalID,
		EffectiveStartDate:  startDate,
		EffectiveEndDate:    endDate,
		Status:              entity.VersionStatusStaging,
		CreatedAt:           entity.Now(),
		CreatedBy:           creatorID,
	}

	if err := s.repo.Create(ctx, version); err != nil {
		return nil, fmt.Errorf("failed to create schedule version: %w", err)
	}

	return version, nil
}

// GetVersion retrieves a schedule version by ID
func (s *ScheduleVersionService) GetVersion(
	ctx context.Context,
	id entity.ScheduleVersionID,
) (*entity.ScheduleVersion, error) {

	version, err := s.repo.GetByID(ctx, uuid.UUID(id))
	if err != nil {
		return nil, err
	}

	return version, nil
}

// GetActiveVersion retrieves the active (PRODUCTION) version for a hospital on a given date
func (s *ScheduleVersionService) GetActiveVersion(
	ctx context.Context,
	hospitalID entity.HospitalID,
	date entity.Date,
) (*entity.ScheduleVersion, error) {

	version, err := s.repo.GetActiveVersion(ctx, uuid.UUID(hospitalID), date)
	if err != nil {
		return nil, err
	}

	return version, nil
}

// ListVersionsByStatus lists all versions for a hospital with a specific status
func (s *ScheduleVersionService) ListVersionsByStatus(
	ctx context.Context,
	hospitalID entity.HospitalID,
	status entity.VersionStatus,
) ([]*entity.ScheduleVersion, error) {

	versions, err := s.repo.GetByHospitalAndStatus(ctx, uuid.UUID(hospitalID), status)
	if err != nil {
		return nil, err
	}

	return versions, nil
}

// ListAllVersions lists all versions for a hospital
func (s *ScheduleVersionService) ListAllVersions(
	ctx context.Context,
	hospitalID entity.HospitalID,
) ([]*entity.ScheduleVersion, error) {

	versions, err := s.repo.ListByHospital(ctx, uuid.UUID(hospitalID))
	if err != nil {
		return nil, err
	}

	return versions, nil
}

// PromoteToProduction transitions a version from STAGING to PRODUCTION
// This makes it the active schedule that staff see
func (s *ScheduleVersionService) PromoteToProduction(
	ctx context.Context,
	id entity.ScheduleVersionID,
	promoterID entity.UserID,
) error {

	version, err := s.repo.GetByID(ctx, uuid.UUID(id))
	if err != nil {
		return err
	}

	// Validate state transition
	if version.Status != entity.VersionStatusStaging {
		return fmt.Errorf("can only promote STAGING versions, current status: %s", version.Status)
	}

	// Promote to PRODUCTION
	version.Status = entity.VersionStatusProduction
	version.UpdatedAt = entity.Now()
	version.UpdatedBy = promoterID

	if err := s.repo.Update(ctx, version); err != nil {
		return fmt.Errorf("failed to promote schedule version: %w", err)
	}

	return nil
}

// Archive transitions a version from PRODUCTION to ARCHIVED
// This removes it as the active schedule
func (s *ScheduleVersionService) Archive(
	ctx context.Context,
	id entity.ScheduleVersionID,
	archiverID entity.UserID,
) error {

	version, err := s.repo.GetByID(ctx, uuid.UUID(id))
	if err != nil {
		return err
	}

	// Validate state transition (can only archive PRODUCTION versions)
	if version.Status != entity.VersionStatusProduction {
		return fmt.Errorf("can only archive PRODUCTION versions, current status: %s", version.Status)
	}

	// Archive
	version.Status = entity.VersionStatusArchived
	version.UpdatedAt = entity.Now()
	version.UpdatedBy = archiverID

	if err := s.repo.Update(ctx, version); err != nil {
		return fmt.Errorf("failed to archive schedule version: %w", err)
	}

	return nil
}

// Delete performs a soft delete on a schedule version
// This prevents accidental deletion of historical data
func (s *ScheduleVersionService) Delete(
	ctx context.Context,
	id entity.ScheduleVersionID,
	deleterID entity.UserID,
) error {

	if err := s.repo.Delete(ctx, uuid.UUID(id), uuid.UUID(deleterID)); err != nil {
		return fmt.Errorf("failed to delete schedule version: %w", err)
	}

	return nil
}

// PromoteAndArchiveOthers promotes a version to PRODUCTION and archives any other PRODUCTION versions
// This ensures only one PRODUCTION version exists at a time
func (s *ScheduleVersionService) PromoteAndArchiveOthers(
	ctx context.Context,
	id entity.ScheduleVersionID,
	promoterID entity.UserID,
) error {

	version, err := s.repo.GetByID(ctx, uuid.UUID(id))
	if err != nil {
		return err
	}

	// Find and archive other PRODUCTION versions for the same hospital
	others, err := s.repo.GetByHospitalAndStatus(ctx, uuid.UUID(version.HospitalID), entity.VersionStatusProduction)
	if err != nil {
		return err
	}

	// Archive each existing PRODUCTION version
	for _, other := range others {
		if other.ID != version.ID {
			if err := s.Archive(ctx, entity.ScheduleVersionID(other.ID), promoterID); err != nil {
				return fmt.Errorf("failed to archive existing production version: %w", err)
			}
		}
	}

	// Promote the new version
	return s.PromoteToProduction(ctx, id, promoterID)
}
