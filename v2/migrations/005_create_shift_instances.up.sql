CREATE TABLE shift_instances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    schedule_version_id UUID NOT NULL REFERENCES schedule_versions(id) ON DELETE CASCADE,
    shift_type VARCHAR(50) NOT NULL
        CHECK (shift_type IN ('ON1', 'ON2', 'MidC', 'MidL', 'DAY')),
    schedule_date DATE NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    hospital_id UUID NOT NULL REFERENCES hospitals(id) ON DELETE CASCADE,
    study_type VARCHAR(50) NOT NULL
        CHECK (study_type IN ('GENERAL', 'BODY', 'NEURO')),
    specialty_constraint VARCHAR(50) NOT NULL
        CHECK (specialty_constraint IN ('BODY_ONLY', 'NEURO_ONLY', 'BOTH')),
    desired_coverage INTEGER NOT NULL DEFAULT 1,
    is_mandatory BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,
    deleted_by UUID
);

-- Indexes for common queries
CREATE INDEX idx_shift_instances_schedule_version ON shift_instances(schedule_version_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_shift_instances_date ON shift_instances(schedule_date) WHERE deleted_at IS NULL;
CREATE INDEX idx_shift_instances_hospital_date ON shift_instances(hospital_id, schedule_date) WHERE deleted_at IS NULL;
CREATE INDEX idx_shift_instances_shift_type ON shift_instances(shift_type);

COMMENT ON TABLE shift_instances IS 'Individual shifts within a schedule version. Each represents a single shift slot for a specific date.';
COMMENT ON COLUMN shift_instances.shift_type IS 'Type of shift: ON1 (overnight 1), ON2 (overnight 2), MidC (midday call), MidL (midday long), DAY (daytime)';
COMMENT ON COLUMN shift_instances.study_type IS 'Types of studies this shift handles: GENERAL, BODY, NEURO';
COMMENT ON COLUMN shift_instances.specialty_constraint IS 'Which specialties can fill this shift: BODY_ONLY, NEURO_ONLY, or BOTH';
