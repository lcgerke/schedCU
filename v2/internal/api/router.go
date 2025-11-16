package api

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/schedcu/v2/internal/job"
	"github.com/schedcu/v2/internal/service"
)

// Router creates and configures the Echo router
type Router struct {
	echo       *echo.Echo
	scheduler  *job.JobScheduler
	services   *ServiceDeps
	handlers   *Handlers
}

// ServiceDeps holds all business logic services
type ServiceDeps struct {
	OdsImporter    service.ODSImportService
	AmionImporter  service.AmionImportService
	Orchestrator   service.ScheduleOrchestrator
	CoverageCalc   service.CoverageCalculator
	VersionService service.ScheduleVersionService
}

// NewRouter creates a new Echo router with all routes
func NewRouter(
	scheduler *job.JobScheduler,
	services *ServiceDeps,
) *Router {

	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE, echo.PATCH},
		AllowHeaders: []string{echo.HeaderContentType, echo.HeaderAuthorization},
	}))

	r := &Router{
		echo:      e,
		scheduler: scheduler,
		handlers: &Handlers{
			scheduler:  scheduler,
			services:   services,
		},
	}

	// Register routes
	r.registerRoutes()

	return r
}

// registerRoutes configures all API routes
func (r *Router) registerRoutes() {
	// Health check
	r.echo.GET("/api/health", r.handlers.Health)

	// Schedules
	scheduleGroup := r.echo.Group("/api/schedules")
	scheduleGroup.POST("", r.handlers.CreateScheduleVersion)
	scheduleGroup.GET("/:id", r.handlers.GetScheduleVersion)
	scheduleGroup.GET("", r.handlers.ListScheduleVersions)
	scheduleGroup.POST("/:id/promote", r.handlers.PromoteScheduleVersion)
	scheduleGroup.POST("/:id/archive", r.handlers.ArchiveScheduleVersion)

	// Import operations
	importGroup := r.echo.Group("/api/imports")
	importGroup.POST("/ods/upload", r.handlers.UploadODSFile) // File upload handler
	importGroup.POST("/ods", r.handlers.StartODSImport)       // Start import job
	importGroup.POST("/amion", r.handlers.StartAmionImport)
	importGroup.POST("/full-workflow", r.handlers.StartFullWorkflow)
	importGroup.GET("/:jobID/status", r.handlers.GetJobStatus)

	// Coverage
	coverageGroup := r.echo.Group("/api/coverage")
	coverageGroup.GET("/schedule/:scheduleID", r.handlers.GetScheduleCoverage)
	coverageGroup.POST("/calculate", r.handlers.CalculateCoverage)

	// Health checks
	r.echo.GET("/api/health/db", r.handlers.HealthDB)
	r.echo.GET("/api/health/redis", r.handlers.HealthRedis)
}

// Start starts the HTTP server
func (r *Router) Start(addr string) error {
	return r.echo.Start(addr)
}

// Shutdown gracefully shuts down the server
func (r *Router) Shutdown() error {
	return r.echo.Close()
}

// Note: Response functions are defined in response.go
// - SuccessResponse(data interface{}) *APIResponse
// - ErrorResponseWithCode(code, message string) *APIResponse
// - ValidationErrorResponse(code, message string) *APIResponse
