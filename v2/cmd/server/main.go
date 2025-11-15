package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/schedcu/v2/internal/api"
	"github.com/schedcu/v2/internal/job"
	"github.com/schedcu/v2/internal/repository/memory"
	"github.com/schedcu/v2/internal/service"
)

func main() {
	// Get server address from environment or use default
	addr := os.Getenv("SERVER_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	// For Phase 0, use in-memory repositories
	// In production (Phase 1 Week 4), these would be PostgreSQL
	// Note: Full repository implementations will be added in Phase 1 Week 4
	_ = memory.NewScheduleRepository() // Placeholder for now

	// Create service layer (placeholder for Phase 0)
	// Full implementations will be added in Phase 1
	// For now, we're testing the API routing structure
	coverageCalc := service.NewDynamicCoverageCalculator(nil, nil)
	var versionService *service.ScheduleVersionService
	// versionService will be properly initialized in Phase 1 Week 4

	// Create job scheduler (requires Redis)
	// For now, this will be initialized when Redis is available
	var scheduler *job.JobScheduler
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	var err error
	scheduler, err = job.NewJobScheduler(redisAddr)
	if err != nil {
		log.Printf("Warning: Failed to initialize job scheduler: %v (jobs will not be queued)", err)
	}

	// Create API router with all services
	serviceDeps := &api.ServiceDeps{
		OdsImporter:    nil, // TODO: Initialize in Phase 3
		AmionImporter:  nil, // TODO: Initialize in Phase 3
		Orchestrator:   nil, // TODO: Initialize in Phase 3
		CoverageCalc:   coverageCalc,
		VersionService: versionService,
	}

	router := api.NewRouter(scheduler, serviceDeps)

	// Graceful shutdown with timeout
	go func() {
		log.Printf("Starting server on %s...\n", addr)
		if err := router.Start(addr); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Keep server running for development
	// In production, this would use os.Signal and proper shutdown
	select {
	case <-time.After(24 * time.Hour): // Run for a day before stopping
		log.Println("Shutting down server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := router.Shutdown(); err != nil {
			log.Fatalf("Server shutdown error: %v", err)
		}
		_ = shutdownCtx // Suppress unused warning if not used in Shutdown()
	}
}
