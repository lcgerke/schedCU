CREATE TABLE job_queue (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_type VARCHAR(255) NOT NULL,
    payload JSONB NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING'
        CHECK (status IN ('PENDING', 'PROCESSING', 'COMPLETE', 'FAILED', 'RETRY')),
    result JSONB,
    error_message TEXT,
    retry_count INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE
);

-- Indexes for common queries
CREATE INDEX idx_job_queue_status ON job_queue(status);
CREATE INDEX idx_job_queue_type ON job_queue(job_type, status);
CREATE INDEX idx_job_queue_created ON job_queue(created_at DESC);
CREATE INDEX idx_job_queue_pending ON job_queue(status) WHERE status IN ('PENDING', 'PROCESSING');

COMMENT ON TABLE job_queue IS 'Fallback job tracking table for Asynq or alternative job queues. Provides visibility into async task execution.';
COMMENT ON COLUMN job_queue.job_type IS 'Type of job: ODS_IMPORT, AMION_SCRAPE, COVERAGE_CALCULATION, etc.';
COMMENT ON COLUMN job_queue.payload IS 'JSON payload containing job parameters (hospital_id, dates, etc.)';
COMMENT ON COLUMN job_queue.status IS 'PENDING (waiting), PROCESSING (running), COMPLETE (success), FAILED (gave up), RETRY (retrying)';
COMMENT ON COLUMN job_queue.result IS 'JSON result data on COMPLETE (e.g., coverage calculation results)';
COMMENT ON COLUMN job_queue.error_message IS 'Human-readable error message on FAILED or final retry';
