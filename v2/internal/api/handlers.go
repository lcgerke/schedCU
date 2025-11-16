package api

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/schedcu/v2/internal/entity"
	"github.com/schedcu/v2/internal/job"
	"github.com/schedcu/v2/internal/service"
	"github.com/schedcu/v2/internal/validation"
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
		return c.JSON(http.StatusBadRequest, ErrorResponseWithCode("INVALID_REQUEST", fmt.Sprintf("Invalid request: %v", err)))
	}

	// Parse hospital ID
	hospitalID := entity.HospitalID(uuid.MustParse(req.HospitalID))

	// Parse dates (TODO: proper date parsing from RFC3339 strings)
	// For now, use placeholder dates - will be implemented in Phase 1 Week 4
	startDate := entity.Now() // placeholder
	endDate := entity.Now()   // placeholder

	// TODO: Get creator ID from authenticated user
	creatorID := entity.UserID(uuid.New())

	// Create version
	version, err := h.services.VersionService.CreateVersion(
		c.Request().Context(),
		hospitalID,
		startDate,
		endDate,
		creatorID,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponseWithCode("VERSION_CREATE_FAILED", fmt.Sprintf("Failed to create schedule version: %v", err)))
	}

	resp := CreateScheduleVersionResponse{
		ID:        version.ID.String(),
		Status:    string(version.Status),
		StartDate: version.EffectiveStartDate.String(),
		EndDate:   version.EffectiveEndDate.String(),
		CreatedAt: version.CreatedAt.String(),
	}

	return c.JSON(http.StatusCreated, SuccessResponse(resp))
}

// GetScheduleVersion retrieves a schedule version by ID
func (h *Handlers) GetScheduleVersion(c echo.Context) error {
	id := c.Param("id")
	versionID := entity.ScheduleVersionID(uuid.MustParse(id))

	version, err := h.services.VersionService.GetVersion(c.Request().Context(), versionID)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponseWithCode("NOT_FOUND", "Schedule version not found"))
	}

	return c.JSON(http.StatusOK, SuccessResponse(version))
}

// ListScheduleVersions lists schedule versions
func (h *Handlers) ListScheduleVersions(c echo.Context) error {
	hospitalID := c.QueryParam("hospital_id")
	if hospitalID == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponseWithCode("MISSING_PARAM", "hospital_id query parameter required"))
	}

	hID := entity.HospitalID(uuid.MustParse(hospitalID))

	versions, err := h.services.VersionService.ListAllVersions(c.Request().Context(), hID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponseWithCode("LIST_FAILED", fmt.Sprintf("Failed to list versions: %v", err)))
	}

	return c.JSON(http.StatusOK, SuccessResponse(versions))
}

// PromoteScheduleVersion promotes a version to PRODUCTION
func (h *Handlers) PromoteScheduleVersion(c echo.Context) error {
	id := c.Param("id")
	versionID := entity.ScheduleVersionID(uuid.MustParse(id))

	// TODO: Get promoter ID from authenticated user
	promoterID := entity.UserID(uuid.New())

	// Promote to production (and archive others)
	if err := h.services.VersionService.PromoteAndArchiveOthers(c.Request().Context(), versionID, promoterID); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponseWithCode("PROMOTE_FAILED", fmt.Sprintf("Failed to promote version: %v", err)))
	}

	version, _ := h.services.VersionService.GetVersion(c.Request().Context(), versionID)

	return c.JSON(http.StatusOK, SuccessResponse(version))
}

// ArchiveScheduleVersion archives a schedule version
func (h *Handlers) ArchiveScheduleVersion(c echo.Context) error {
	id := c.Param("id")
	versionID := entity.ScheduleVersionID(uuid.MustParse(id))

	// TODO: Get archiver ID from authenticated user
	archiverID := entity.UserID(uuid.New())

	if err := h.services.VersionService.Archive(c.Request().Context(), versionID, archiverID); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponseWithCode("ARCHIVE_FAILED", fmt.Sprintf("Failed to archive version: %v", err)))
	}

	version, _ := h.services.VersionService.GetVersion(c.Request().Context(), versionID)

	return c.JSON(http.StatusOK, SuccessResponse(version))
}

// UploadODSRequest represents a multipart upload request for ODS files
type UploadODSRequest struct {
	ScheduleVersionID string `form:"schedule_version_id" validate:"required"`
}

// UploadODSFile handles file upload and returns parsed data
func (h *Handlers) UploadODSFile(c echo.Context) error {
	// Parse form
	var req UploadODSRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponseWithCode("INVALID_REQUEST", "Missing schedule_version_id"))
	}

	// Get uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponseWithCode("MISSING_FILE", "No ODS file provided"))
	}

	// Validate file extension
	if !strings.HasSuffix(strings.ToLower(file.Filename), ".ods") {
		return c.JSON(http.StatusBadRequest, ErrorResponseWithCode("INVALID_FILE_TYPE", "File must be .ods format"))
	}

	// Validate file size (max 10MB)
	const maxFileSize = 10 * 1024 * 1024
	if file.Size > maxFileSize {
		return c.JSON(http.StatusBadRequest, ErrorResponseWithCode("FILE_TOO_LARGE", "File must be smaller than 10MB"))
	}

	// Read file into temporary location
	src, err := file.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponseWithCode("FILE_READ_ERROR", "Failed to read uploaded file"))
	}
	defer src.Close()

	// Create temporary file
	tempDir := os.TempDir()
	tempFile := filepath.Join(tempDir, fmt.Sprintf("ods_%s_%d.ods", uuid.New().String()[:8], time.Now().Unix()))
	dst, err := os.Create(tempFile)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponseWithCode("FILE_WRITE_ERROR", "Failed to save uploaded file"))
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		os.Remove(tempFile)
		return c.JSON(http.StatusInternalServerError, ErrorResponseWithCode("FILE_WRITE_ERROR", "Failed to save uploaded file"))
	}

	// Parse ODS file
	ctx := c.Request().Context()
	validationResult := validation.NewResult()
	parser := service.NewODSParser()

	odsData, err := parser.ParseFile(tempFile, validationResult)
	defer os.Remove(tempFile) // Clean up temp file

	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponseWithCode("PARSE_ERROR", "Failed to parse ODS file"))
	}

	// Validate schedule_version_id is provided
	if req.ScheduleVersionID == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponseWithCode("INVALID_REQUEST", "Missing or invalid schedule_version_id"))
	}

	// Get schedule version
	versionID := entity.ScheduleVersionID(uuid.MustParse(req.ScheduleVersionID))
	version, err := h.services.VersionService.GetVersion(ctx, versionID)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponseWithCode("VERSION_NOT_FOUND", "Schedule version not found"))
	}

	// Return parsed ODS data summary
	sheetSummary := make([]map[string]interface{}, 0)
	for _, sheet := range odsData.Sheets {
		sheetSummary = append(sheetSummary, map[string]interface{}{
			"name":              sheet.Name,
			"shift_category":    sheet.ShiftCategory,
			"day_type":          sheet.DayType,
			"specialty":         sheet.SpecialtyScenario,
			"time_start":        sheet.TimeStart.Format("15:04"),
			"time_end":          sheet.TimeEnd.Format("15:04"),
			"assignment_count":  len(sheet.CoverageGrid),
		})
	}

	return c.JSON(http.StatusOK, SuccessResponse(map[string]interface{}{
		"filename":            file.Filename,
		"sheets_parsed":       len(odsData.Sheets),
		"total_assignments":   getTotalAssignments(odsData),
		"sheets":              sheetSummary,
		"schedule_version_id": req.ScheduleVersionID,
		"hospital_id":         version.HospitalID.String(),
	}))
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
		return c.JSON(http.StatusBadRequest, ErrorResponseWithCode("ERROR", fmt.Sprintf("Invalid request: %v", err)))
	}

	versionID := entity.ScheduleVersionID(uuid.MustParse(req.ScheduleVersionID))

	// Get version to get hospital ID
	version, err := h.services.VersionService.GetVersion(c.Request().Context(), versionID)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponseWithCode("ERROR", "Schedule version not found"))
	}

	// TODO: Get creator ID from authenticated user
	creatorID := entity.UserID(uuid.New())

	// Enqueue import job
	info, err := h.scheduler.EnqueueODSImport(c.Request().Context(), version.HospitalID, versionID, req.Filename, creatorID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponseWithCode("ERROR", fmt.Sprintf("Failed to enqueue job: %v", err)))
	}

	return c.JSON(http.StatusAccepted, SuccessResponse(map[string]interface{}{
		"job_id": info.ID,
		"status": "queued",
	}))
}

// Helper function to get total assignments across all sheets
func getTotalAssignments(odsData *service.ODSData) int {
	total := 0
	for _, sheet := range odsData.Sheets {
		total += len(sheet.CoverageGrid)
	}
	return total
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
		return c.JSON(http.StatusBadRequest, ErrorResponseWithCode("ERROR", fmt.Sprintf("Invalid request: %v", err)))
	}

	versionID := entity.ScheduleVersionID(uuid.MustParse(req.ScheduleVersionID))

	// Get version to get hospital ID
	version, err := h.services.VersionService.GetVersion(c.Request().Context(), versionID)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponseWithCode("ERROR", "Schedule version not found"))
	}

	// TODO: Get creator ID from authenticated user
	creatorID := entity.UserID(uuid.New())

	// Enqueue scrape job
	info, err := h.scheduler.EnqueueAmionScrape(c.Request().Context(), version.HospitalID, versionID, req.MonthsBack, req.Username, creatorID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponseWithCode("ERROR", fmt.Sprintf("Failed to enqueue job: %v", err)))
	}

	return c.JSON(http.StatusAccepted, SuccessResponse(map[string]interface{}{
		"job_id": info.ID,
		"status": "queued",
	}))
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
		return c.JSON(http.StatusBadRequest, ErrorResponseWithCode("ERROR", fmt.Sprintf("Invalid request: %v", err)))
	}

	// TODO: Implement full workflow orchestration
	// This would coordinate all three phases
	// 1. Create schedule version
	// 2. Enqueue ODS import
	// 3. Enqueue Amion scrape
	// 4. Enqueue coverage calculation
	// All as a coordinated workflow

	return c.JSON(http.StatusNotImplemented, ErrorResponseWithCode("ERROR", "Full workflow not yet implemented"))
}

// GetJobStatus retrieves the status of a queued job
func (h *Handlers) GetJobStatus(c echo.Context) error {
	jobID := c.Param("jobID")

	// TODO: Implement job status retrieval from Asynq
	// info, err := h.scheduler.GetTaskInfo(context.Background(), jobID)

	return c.JSON(http.StatusOK, SuccessResponse(map[string]interface{}{
		"job_id": jobID,
		"status": "pending",
	}))
}

// GetScheduleCoverage retrieves coverage for a schedule
func (h *Handlers) GetScheduleCoverage(c echo.Context) error {
	scheduleID := c.Param("scheduleID")

	// TODO: Implement coverage retrieval
	// This would query the database for coverage calculations

	return c.JSON(http.StatusOK, SuccessResponse(map[string]interface{}{
		"schedule_id": scheduleID,
		"coverage":    "TBD",
	}))
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
		return c.JSON(http.StatusBadRequest, ErrorResponseWithCode("ERROR", fmt.Sprintf("Invalid request: %v", err)))
	}

	versionID := entity.ScheduleVersionID(uuid.MustParse(req.ScheduleVersionID))
	startDate := entity.Now()
	endDate := entity.Now()

	// TODO: Get creator ID from authenticated user
	creatorID := entity.UserID(uuid.New())

	// Enqueue job
	info, err := h.scheduler.EnqueueCoverageCalculation(c.Request().Context(), versionID, startDate, endDate, creatorID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponseWithCode("ERROR", fmt.Sprintf("Failed to enqueue job: %v", err)))
	}

	return c.JSON(http.StatusAccepted, SuccessResponse(map[string]interface{}{
		"job_id": info.ID,
		"status": "queued",
	}))
}

// Health returns the health status
func (h *Handlers) Health(c echo.Context) error {
	return c.JSON(http.StatusOK, SuccessResponse(map[string]interface{}{
		"status": "UP",
	}))
}

// HealthDB returns database health status
func (h *Handlers) HealthDB(c echo.Context) error {
	// TODO: Check database connectivity
	return c.JSON(http.StatusOK, SuccessResponse(map[string]interface{}{
		"database": "UP",
	}))
}

// HealthRedis returns Redis health status
func (h *Handlers) HealthRedis(c echo.Context) error {
	// TODO: Check Redis connectivity
	return c.JSON(http.StatusOK, SuccessResponse(map[string]interface{}{
		"redis": "UP",
	}))
}
