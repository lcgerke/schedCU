package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/schedcu/v2/internal/entity"
	"github.com/schedcu/v2/internal/job"
	"github.com/schedcu/v2/internal/service"
)

// Handlers contains all HTTP request handlers
type Handlers struct {
	scheduler *job.JobScheduler
	services  *ServiceDeps
}

// CreateScheduleVersionRequest represents a request to create a new schedule version
type CreateScheduleVersionRequest struct {
	HospitalID string `json:"hospital_id" validate:"required"`
	StartDate  string `json:"start_date" validate:"required"`
	EndDate    string `json:"end_date" validate:"required"`
}

// CreateScheduleVersionResponse represents the response from creating a schedule version
type CreateScheduleVersionResponse struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	CreatedAt string `json:"created_at"`
}

// CreateScheduleVersion creates a new schedule version
func (h *Handlers) CreateScheduleVersion(c echo.Context) error {
	var req CreateScheduleVersionRequest

	if err := c.Bind(&req); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
	}

	// Parse hospital ID
	hospitalID := entity.HospitalID(uuid.MustParse(req.HospitalID))

	// Parse dates
	startDate := entity.Date(req.StartDate)
	endDate := entity.Date(req.EndDate)

	// TODO: Get creator ID from authenticated user
	creatorID := entity.UserID(uuid.New())

	// Create version
	version, err := h.services.VersionService.CreateVersion(
		context.Background(),
		hospitalID,
		startDate,
		endDate,
		creatorID,
	)
	if err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to create schedule version: %v", err))
	}

	resp := CreateScheduleVersionResponse{
		ID:        string(version.ID),
		Status:    string(version.Status),
		StartDate: string(version.StartDate),
		EndDate:   string(version.EndDate),
		CreatedAt: version.CreatedAt.String(),
	}

	return SuccessResponse(c, http.StatusCreated, resp)
}

// GetScheduleVersion retrieves a schedule version by ID
func (h *Handlers) GetScheduleVersion(c echo.Context) error {
	id := c.Param("id")
	versionID := entity.ScheduleVersionID(uuid.MustParse(id))

	version, err := h.services.VersionService.GetVersion(context.Background(), versionID)
	if err != nil {
		return ErrorResponse(c, http.StatusNotFound, "Schedule version not found")
	}

	return SuccessResponse(c, http.StatusOK, version)
}

// ListScheduleVersions lists schedule versions
func (h *Handlers) ListScheduleVersions(c echo.Context) error {
	hospitalID := c.QueryParam("hospital_id")
	if hospitalID == "" {
		return ErrorResponse(c, http.StatusBadRequest, "hospital_id query parameter required")
	}

	hID := entity.HospitalID(uuid.MustParse(hospitalID))

	versions, err := h.services.VersionService.ListAllVersions(context.Background(), hID)
	if err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to list versions: %v", err))
	}

	return SuccessResponse(c, http.StatusOK, versions)
}

// PromoteScheduleVersion promotes a version to PRODUCTION
func (h *Handlers) PromoteScheduleVersion(c echo.Context) error {
	id := c.Param("id")
	versionID := entity.ScheduleVersionID(uuid.MustParse(id))

	// TODO: Get promoter ID from authenticated user
	promoterID := entity.UserID(uuid.New())

	// Promote to production (and archive others)
	if err := h.services.VersionService.PromoteAndArchiveOthers(context.Background(), versionID, promoterID); err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to promote version: %v", err))
	}

	version, _ := h.services.VersionService.GetVersion(context.Background(), versionID)

	return SuccessResponse(c, http.StatusOK, version)
}

// ArchiveScheduleVersion archives a schedule version
func (h *Handlers) ArchiveScheduleVersion(c echo.Context) error {
	id := c.Param("id")
	versionID := entity.ScheduleVersionID(uuid.MustParse(id))

	// TODO: Get archiver ID from authenticated user
	archiverID := entity.UserID(uuid.New())

	if err := h.services.VersionService.Archive(context.Background(), versionID, archiverID); err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to archive version: %v", err))
	}

	version, _ := h.services.VersionService.GetVersion(context.Background(), versionID)

	return SuccessResponse(c, http.StatusOK, version)
}

// StartODSImportRequest represents a request to start ODS import
type StartODSImportRequest struct {
	ScheduleVersionID string `json:"schedule_version_id" validate:"required"`
	Filename          string `json:"filename" validate:"required"`
}

// StartODSImport enqueues an ODS import job
func (h *Handlers) StartODSImport(c echo.Context) error {
	var req StartODSImportRequest

	if err := c.Bind(&req); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
	}

	versionID := entity.ScheduleVersionID(uuid.MustParse(req.ScheduleVersionID))

	// Get version to get hospital ID
	version, err := h.services.VersionService.GetVersion(context.Background(), versionID)
	if err != nil {
		return ErrorResponse(c, http.StatusNotFound, "Schedule version not found")
	}

	// TODO: Get creator ID from authenticated user
	creatorID := entity.UserID(uuid.New())

	// Enqueue import job
	info, err := h.scheduler.EnqueueODSImport(context.Background(), version.HospitalID, versionID, req.Filename, creatorID)
	if err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to enqueue job: %v", err))
	}

	return SuccessResponse(c, http.StatusAccepted, map[string]interface{}{
		"job_id": info.ID,
		"status": "queued",
	})
}

// StartAmionImportRequest represents a request to start Amion import
type StartAmionImportRequest struct {
	ScheduleVersionID string `json:"schedule_version_id" validate:"required"`
	MonthsBack        int    `json:"months_back" validate:"required,min=1,max=24"`
	Username          string `json:"username" validate:"required"`
}

// StartAmionImport enqueues an Amion scraping job
func (h *Handlers) StartAmionImport(c echo.Context) error {
	var req StartAmionImportRequest

	if err := c.Bind(&req); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
	}

	versionID := entity.ScheduleVersionID(uuid.MustParse(req.ScheduleVersionID))

	// Get version to get hospital ID
	version, err := h.services.VersionService.GetVersion(context.Background(), versionID)
	if err != nil {
		return ErrorResponse(c, http.StatusNotFound, "Schedule version not found")
	}

	// TODO: Get creator ID from authenticated user
	creatorID := entity.UserID(uuid.New())

	// Enqueue scrape job
	info, err := h.scheduler.EnqueueAmionScrape(context.Background(), version.HospitalID, versionID, req.MonthsBack, req.Username, creatorID)
	if err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to enqueue job: %v", err))
	}

	return SuccessResponse(c, http.StatusAccepted, map[string]interface{}{
		"job_id": info.ID,
		"status": "queued",
	})
}

// StartFullWorkflowRequest represents a request to start the full import workflow
type StartFullWorkflowRequest struct {
	HospitalID string `json:"hospital_id" validate:"required"`
	StartDate  string `json:"start_date" validate:"required"`
	EndDate    string `json:"end_date" validate:"required"`
	Filename   string `json:"filename" validate:"required"`
	MonthsBack int    `json:"months_back" validate:"min=1,max=24"`
	Username   string `json:"username"`
}

// StartFullWorkflow starts the full 3-phase workflow
func (h *Handlers) StartFullWorkflow(c echo.Context) error {
	var req StartFullWorkflowRequest

	if err := c.Bind(&req); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
	}

	// TODO: Implement full workflow orchestration
	// This would coordinate all three phases
	// 1. Create schedule version
	// 2. Enqueue ODS import
	// 3. Enqueue Amion scrape
	// 4. Enqueue coverage calculation
	// All as a coordinated workflow

	return ErrorResponse(c, http.StatusNotImplemented, "Full workflow not yet implemented")
}

// GetJobStatus retrieves the status of a queued job
func (h *Handlers) GetJobStatus(c echo.Context) error {
	jobID := c.Param("jobID")

	// TODO: Implement job status retrieval from Asynq
	// info, err := h.scheduler.GetTaskInfo(context.Background(), jobID)

	return SuccessResponse(c, http.StatusOK, map[string]interface{}{
		"job_id": jobID,
		"status": "pending",
	})
}

// GetScheduleCoverage retrieves coverage for a schedule
func (h *Handlers) GetScheduleCoverage(c echo.Context) error {
	scheduleID := c.Param("scheduleID")

	// TODO: Implement coverage retrieval
	// This would query the database for coverage calculations

	return SuccessResponse(c, http.StatusOK, map[string]interface{}{
		"schedule_id": scheduleID,
		"coverage":    "TBD",
	})
}

// CalculateCoverageRequest represents a request to calculate coverage
type CalculateCoverageRequest struct {
	ScheduleVersionID string `json:"schedule_version_id" validate:"required"`
	StartDate         string `json:"start_date" validate:"required"`
	EndDate           string `json:"end_date" validate:"required"`
}

// CalculateCoverage enqueues a coverage calculation job
func (h *Handlers) CalculateCoverage(c echo.Context) error {
	var req CalculateCoverageRequest

	if err := c.Bind(&req); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
	}

	versionID := entity.ScheduleVersionID(uuid.MustParse(req.ScheduleVersionID))
	startDate := entity.Date(req.StartDate)
	endDate := entity.Date(req.EndDate)

	// TODO: Get creator ID from authenticated user
	creatorID := entity.UserID(uuid.New())

	// Enqueue job
	info, err := h.scheduler.EnqueueCoverageCalculation(context.Background(), versionID, startDate, endDate, creatorID)
	if err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to enqueue job: %v", err))
	}

	return SuccessResponse(c, http.StatusAccepted, map[string]interface{}{
		"job_id": info.ID,
		"status": "queued",
	})
}

// Health returns the health status
func (h *Handlers) Health(c echo.Context) error {
	return SuccessResponse(c, http.StatusOK, map[string]interface{}{
		"status": "UP",
	})
}

// HealthDB returns database health status
func (h *Handlers) HealthDB(c echo.Context) error {
	// TODO: Check database connectivity
	return SuccessResponse(c, http.StatusOK, map[string]interface{}{
		"database": "UP",
	})
}

// HealthRedis returns Redis health status
func (h *Handlers) HealthRedis(c echo.Context) error {
	// TODO: Check Redis connectivity
	return SuccessResponse(c, http.StatusOK, map[string]interface{}{
		"redis": "UP",
	})
}
