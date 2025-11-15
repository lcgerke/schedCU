package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/schedcu/v2/internal/api"
	"github.com/schedcu/v2/internal/entity"
	"github.com/schedcu/v2/internal/repository"
	"github.com/schedcu/v2/internal/service"
)

// ScheduleHandler handles HTTP requests for schedule operations.
type ScheduleHandler struct {
	svc *service.ScheduleService
}

// NewScheduleHandler creates a new schedule handler.
func NewScheduleHandler(svc *service.ScheduleService) *ScheduleHandler {
	return &ScheduleHandler{svc: svc}
}

// CreateScheduleRequest contains the request body for creating a schedule.
type CreateScheduleRequest struct {
	HospitalID string  `json:"hospital_id" validate:"required,uuid"`
	StartDate  string  `json:"start_date" validate:"required"`
	EndDate    string  `json:"end_date" validate:"required"`
	Source     string  `json:"source" validate:"required,oneof=amion ods_file manual"`
	SourceID   *string `json:"source_id,omitempty"`
}

// UpdateScheduleRequest contains the request body for updating a schedule.
type UpdateScheduleRequest struct {
	EndDate  *string `json:"end_date,omitempty"`
	Source   *string `json:"source,omitempty"`
	SourceID *string `json:"source_id,omitempty"`
}

// CreateSchedule handles POST /api/schedules
func (h *ScheduleHandler) CreateSchedule(c echo.Context) error {
	var req CreateScheduleRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, api.ErrorResponseWithCode(
			"INVALID_REQUEST",
			"Invalid request body: "+err.Error(),
		))
	}

	// Parse hospital ID
	hospitalID, err := uuid.Parse(req.HospitalID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, api.ErrorResponseWithCode(
			"INVALID_HOSPITAL_ID",
			"Invalid hospital_id format",
		))
	}

	// Get user ID from context (would come from auth middleware in production)
	userID := uuid.New()

	// Create schedule via service
	svcReq := &service.CreateScheduleRequest{
		HospitalID: hospitalID,
		StartDate:  req.StartDate,
		EndDate:    req.EndDate,
		Source:     req.Source,
		SourceID:   req.SourceID,
		UserID:     userID,
	}

	schedule, err := h.svc.CreateSchedule(c.Request().Context(), svcReq)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, api.ErrorResponseWithCode(
			"CREATION_FAILED",
			"Failed to create schedule: "+err.Error(),
		))
	}

	return c.JSON(http.StatusCreated, api.SuccessResponse(schedule))
}

// GetSchedule handles GET /api/schedules/:id
func (h *ScheduleHandler) GetSchedule(c echo.Context) error {
	id := c.Param("id")

	scheduleID, err := uuid.Parse(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, api.ErrorResponseWithCode(
			"INVALID_ID",
			"Invalid schedule ID format",
		))
	}

	schedule, err := h.svc.GetSchedule(c.Request().Context(), scheduleID)
	if err != nil {
		if repository.IsNotFound(err) {
			return c.JSON(http.StatusNotFound, api.ErrorResponseWithCode(
				"NOT_FOUND",
				"Schedule not found",
			))
		}

		return c.JSON(http.StatusInternalServerError, api.ErrorResponseWithCode(
			"RETRIEVAL_FAILED",
			"Failed to retrieve schedule: "+err.Error(),
		))
	}

	return c.JSON(http.StatusOK, api.SuccessResponse(schedule))
}

// UpdateSchedule handles PUT /api/schedules/:id
func (h *ScheduleHandler) UpdateSchedule(c echo.Context) error {
	id := c.Param("id")

	scheduleID, err := uuid.Parse(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, api.ErrorResponseWithCode(
			"INVALID_ID",
			"Invalid schedule ID format",
		))
	}

	var req UpdateScheduleRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, api.ErrorResponseWithCode(
			"INVALID_REQUEST",
			"Invalid request body: "+err.Error(),
		))
	}

	// Get user ID from context (would come from auth middleware in production)
	userID := uuid.New()

	// Update schedule via service
	svcReq := &service.UpdateScheduleRequest{
		ID:       scheduleID,
		EndDate:  req.EndDate,
		Source:   req.Source,
		SourceID: req.SourceID,
		UserID:   userID,
	}

	schedule, err := h.svc.UpdateSchedule(c.Request().Context(), svcReq)
	if err != nil {
		if repository.IsNotFound(err) {
			return c.JSON(http.StatusNotFound, api.ErrorResponseWithCode(
				"NOT_FOUND",
				"Schedule not found",
			))
		}

		return c.JSON(http.StatusInternalServerError, api.ErrorResponseWithCode(
			"UPDATE_FAILED",
			"Failed to update schedule: "+err.Error(),
		))
	}

	return c.JSON(http.StatusOK, api.SuccessResponse(schedule))
}

// DeleteSchedule handles DELETE /api/schedules/:id
func (h *ScheduleHandler) DeleteSchedule(c echo.Context) error {
	id := c.Param("id")

	scheduleID, err := uuid.Parse(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, api.ErrorResponseWithCode(
			"INVALID_ID",
			"Invalid schedule ID format",
		))
	}

	// Get user ID from context (would come from auth middleware in production)
	deleterID := uuid.New()

	if err := h.svc.DeleteSchedule(c.Request().Context(), scheduleID, deleterID); err != nil {
		if repository.IsNotFound(err) {
			return c.JSON(http.StatusNotFound, api.ErrorResponseWithCode(
				"NOT_FOUND",
				"Schedule not found",
			))
		}

		return c.JSON(http.StatusInternalServerError, api.ErrorResponseWithCode(
			"DELETE_FAILED",
			"Failed to delete schedule: "+err.Error(),
		))
	}

	return c.JSON(http.StatusNoContent, nil)
}

// GetSchedulesForHospital handles GET /api/hospitals/:hospital_id/schedules
func (h *ScheduleHandler) GetSchedulesForHospital(c echo.Context) error {
	hospitalID := c.Param("hospital_id")

	hID, err := uuid.Parse(hospitalID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, api.ErrorResponseWithCode(
			"INVALID_HOSPITAL_ID",
			"Invalid hospital ID format",
		))
	}

	schedules, err := h.svc.GetSchedulesForHospital(c.Request().Context(), hID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, api.ErrorResponseWithCode(
			"RETRIEVAL_FAILED",
			"Failed to retrieve schedules: "+err.Error(),
		))
	}

	return c.JSON(http.StatusOK, api.SuccessResponse(map[string]interface{}{
		"schedules": schedules,
		"count":     len(schedules),
	}))
}

// AddShiftRequest contains the request body for adding a shift.
type AddShiftRequest struct {
	Position    string `json:"position" validate:"required"`
	StartTime   string `json:"start_time" validate:"required"`
	EndTime     string `json:"end_time" validate:"required"`
	StaffMember string `json:"staff_member" validate:"required"`
	Location    string `json:"location" validate:"required"`
}

// AddShift handles POST /api/schedules/:id/shifts
func (h *ScheduleHandler) AddShift(c echo.Context) error {
	id := c.Param("id")

	scheduleID, err := uuid.Parse(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, api.ErrorResponseWithCode(
			"INVALID_ID",
			"Invalid schedule ID format",
		))
	}

	var req AddShiftRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, api.ErrorResponseWithCode(
			"INVALID_REQUEST",
			"Invalid request body: "+err.Error(),
		))
	}

	shift := &entity.ShiftInstance{
		Position:    req.Position,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		StaffMember: req.StaffMember,
		Location:    req.Location,
	}

	if err := h.svc.AddShiftToSchedule(c.Request().Context(), scheduleID, shift); err != nil {
		if repository.IsNotFound(err) {
			return c.JSON(http.StatusNotFound, api.ErrorResponseWithCode(
				"NOT_FOUND",
				"Schedule not found",
			))
		}

		return c.JSON(http.StatusInternalServerError, api.ErrorResponseWithCode(
			"SHIFT_CREATION_FAILED",
			"Failed to add shift: "+err.Error(),
		))
	}

	return c.JSON(http.StatusCreated, api.SuccessResponse(shift))
}

// HealthCheck handles GET /api/health
func HealthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status": "UP",
	})
}
