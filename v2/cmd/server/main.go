package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/schedcu/v2/internal/api/handlers"
	"github.com/schedcu/v2/internal/repository/memory"
	"github.com/schedcu/v2/internal/service"
)

func main() {
	// Create echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// For Phase 0, use in-memory repository
	// In production, this would be PostgreSQL
	repo := memory.NewScheduleRepository()

	// Create service layer
	scheduleSvc := service.NewScheduleService(repo)

	// Create handlers
	scheduleHandler := handlers.NewScheduleHandler(scheduleSvc)

	// Health check endpoint
	e.GET("/api/health", handlers.HealthCheck)

	// Schedule routes
	e.POST("/api/schedules", scheduleHandler.CreateSchedule)
	e.GET("/api/schedules/:id", scheduleHandler.GetSchedule)
	e.PUT("/api/schedules/:id", scheduleHandler.UpdateSchedule)
	e.DELETE("/api/schedules/:id", scheduleHandler.DeleteSchedule)

	// Hospital-scoped routes
	e.GET("/api/hospitals/:hospital_id/schedules", scheduleHandler.GetSchedulesForHospital)

	// Shift routes
	e.POST("/api/schedules/:id/shifts", scheduleHandler.AddShift)

	// Get server address from environment or use default
	addr := os.Getenv("SERVER_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	// Graceful shutdown
	go func() {
		log.Printf("Starting server on %s...\n", addr)
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Create a channel to handle shutdown signal
	// In production, this would use os.Signal
	select {
	case <-time.After(1 * time.Hour): // Just keep running for development
		log.Println("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := e.Shutdown(ctx); err != nil {
			log.Fatalf("Server shutdown error: %v", err)
		}
	}
}
