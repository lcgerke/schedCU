-- Create schedule_versions table
-- Temporal versioning enables time-travel queries and schedule promotion workflows
CREATE TABLE schedule_versions (
    id UUID PRIMARY KEY,
    hospital_id UUID NOT NULL REFERENCES hospitals(id),
    status VARCHAR(50) NOT NULL DEFAULT 'STAGING'
        CHECK (status IN ('STAGING', 'PRODUCTION', 'ARCHIVED')),
    effective_start_date DATE NOT NULL,
    effective_end_date DATE NOT NULL,
    scrape_batch_id UUID REFERENCES scrape_batches(id),
    validation_results JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_by UUID NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,
    deleted_by UUID
);

-- Indexes for common queries
CREATE INDEX idx_schedule_versions_hospital ON schedule_versions(hospital_id, status);
CREATE INDEX idx_schedule_versions_status ON schedule_versions(status);
CREATE INDEX idx_schedule_versions_effective_date ON schedule_versions(effective_start_date, effective_end_date);
CREATE INDEX idx_schedule_versions_batch ON schedule_versions(scrape_batch_id);
CREATE INDEX idx_schedule_versions_created ON schedule_versions(created_at DESC);

-- Data integrity constraint
ALTER TABLE schedule_versions ADD CONSTRAINT check_schedule_date_range
    CHECK (effective_start_date <= effective_end_date);
