CREATE TABLE assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    person_id UUID NOT NULL REFERENCES persons(id) ON DELETE CASCADE,
    shift_instance_id UUID NOT NULL REFERENCES shift_instances(id) ON DELETE CASCADE,
    schedule_date DATE NOT NULL,
    original_shift_type VARCHAR(255),
    source VARCHAR(50) NOT NULL DEFAULT 'MANUAL'
        CHECK (source IN ('AMION', 'MANUAL', 'OVERRIDE')),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,
    deleted_by UUID
);

-- Unique constraint on non-deleted assignments
CREATE UNIQUE INDEX idx_assignments_unique ON assignments(person_id, shift_instance_id, schedule_date)
    WHERE deleted_at IS NULL;

-- Regular indexes for filtering
CREATE INDEX idx_assignments_person ON assignments(person_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_assignments_shift ON assignments(shift_instance_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_assignments_date ON assignments(schedule_date) WHERE deleted_at IS NULL;
CREATE INDEX idx_assignments_source ON assignments(source);

COMMENT ON TABLE assignments IS 'Maps persons to shift instances. Shows who is assigned to work each shift.';
COMMENT ON COLUMN assignments.source IS 'Where the assignment came from: AMION (scraped), MANUAL (manually created), OVERRIDE (manual override of automatic assignment)';
COMMENT ON COLUMN assignments.original_shift_type IS 'Shift type from source system before any modifications';
