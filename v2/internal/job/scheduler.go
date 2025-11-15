package job

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"github.com/schedcu/v2/internal/entity"
)

// JobScheduler manages job enqueueing to Asynq
type JobScheduler struct {
	client *asynq.Client
}

// NewJobScheduler creates a new job scheduler
func NewJobScheduler(redisAddr string) (*JobScheduler, error) {
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})

	// Test connection
	if err := client.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &JobScheduler{client: client}, nil
}

// Job types
const (
	TypeODSImport    = "ods:import"
	TypeAmionScrape  = "amion:scrape"
	TypeCoverageCalc = "coverage:calculate"
)

// ODSImportPayload represents the payload for ODS import job
type ODSImportPayload struct {
	HospitalID entity.HospitalID `json:"hospital_id"`
	VersionID  entity.ScheduleVersionID `json:"version_id"`
	Filename   string `json:"filename"`
	CreatorID  entity.UserID `json:"creator_id"`
}

// EnqueueODSImport enqueues an ODS import job
func (s *JobScheduler) EnqueueODSImport(
	ctx context.Context,
	hospitalID entity.HospitalID,
	versionID entity.ScheduleVersionID,
	filename string,
	creatorID entity.UserID,
) (*asynq.TaskInfo, error) {

	payload := ODSImportPayload{
		HospitalID: hospitalID,
		VersionID:  versionID,
		Filename:   filename,
		CreatorID:  creatorID,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	task := asynq.NewTask(TypeODSImport, payloadBytes)

	info, err := s.client.EnqueueContext(ctx, task, asynq.MaxRetry(3), asynq.Timeout(10*time.Minute))
	if err != nil {
		return nil, fmt.Errorf("failed to enqueue ODS import job: %w", err)
	}

	return info, nil
}

// AmionScrapePayload represents the payload for Amion scrape job
type AmionScrapePayload struct {
	HospitalID  entity.HospitalID `json:"hospital_id"`
	VersionID   entity.ScheduleVersionID `json:"version_id"`
	MonthsBack  int `json:"months_back"`
	Username    string `json:"username"`
	CreatorID   entity.UserID `json:"creator_id"`
}

// EnqueueAmionScrape enqueues an Amion scraping job
func (s *JobScheduler) EnqueueAmionScrape(
	ctx context.Context,
	hospitalID entity.HospitalID,
	versionID entity.ScheduleVersionID,
	monthsBack int,
	username string,
	creatorID entity.UserID,
) (*asynq.TaskInfo, error) {

	payload := AmionScrapePayload{
		HospitalID: hospitalID,
		VersionID:  versionID,
		MonthsBack: monthsBack,
		Username:   username,
		CreatorID:  creatorID,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	task := asynq.NewTask(TypeAmionScrape, payloadBytes)

	// Amion scraping can take longer (depends on months to scrape)
	// Estimate: 30s base + 10s per month
	timeout := time.Duration(30 + monthsBack*10) * time.Second
	if timeout < 2*time.Minute {
		timeout = 2 * time.Minute
	}

	info, err := s.client.EnqueueContext(
		ctx,
		task,
		asynq.MaxRetry(2),
		asynq.Timeout(timeout),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to enqueue Amion scrape job: %w", err)
	}

	return info, nil
}

// CoverageCalcPayload represents the payload for coverage calculation job
type CoverageCalcPayload struct {
	ScheduleVersionID entity.ScheduleVersionID `json:"schedule_version_id"`
	StartDate         entity.Date `json:"start_date"`
	EndDate           entity.Date `json:"end_date"`
	CreatorID         entity.UserID `json:"creator_id"`
}

// EnqueueCoverageCalculation enqueues a coverage calculation job
func (s *JobScheduler) EnqueueCoverageCalculation(
	ctx context.Context,
	versionID entity.ScheduleVersionID,
	startDate, endDate entity.Date,
	creatorID entity.UserID,
) (*asynq.TaskInfo, error) {

	payload := CoverageCalcPayload{
		ScheduleVersionID: versionID,
		StartDate:         startDate,
		EndDate:           endDate,
		CreatorID:         creatorID,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	task := asynq.NewTask(TypeCoverageCalc, payloadBytes)

	info, err := s.client.EnqueueContext(
		ctx,
		task,
		asynq.MaxRetry(1),
		asynq.Timeout(2*time.Minute),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to enqueue coverage calculation job: %w", err)
	}

	return info, nil
}

// Close closes the job scheduler and releases resources
func (s *JobScheduler) Close() error {
	return s.client.Close()
}

// GetTaskInfo retrieves information about a task
func (s *JobScheduler) GetTaskInfo(ctx context.Context, taskID string) (*asynq.TaskInfo, error) {
	inspector := asynq.NewInspector(asynq.RedisClientOpt{Addr: s.client.String()})
	defer inspector.Close()

	return inspector.GetTaskInfo(ctx, "default", taskID)
}
