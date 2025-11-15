CREATE TABLE coverage_calculations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    schedule_version_id UUID NOT NULL REFERENCES schedule_versions(id) ON DELETE CASCADE,
    hospital_id UUID NOT NULL REFERENCES hospitals(id) ON DELETE CASCADE,
    calculation_date DATE NOT NULL,
    calculation_period_start_date DATE NOT NULL,
    calculation_period_end_date DATE NOT NULL,
    coverage_by_position JSONB NOT NULL,
    coverage_summary JSONB,
    validation_errors JSONB,
    query_count INTEGER,
    calculated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    calculated_by UUID NOT NULL
);

-- Indexes for common queries
CREATE INDEX idx_coverage_calculations_schedule_version ON coverage_calculations(schedule_version_id);
CREATE INDEX idx_coverage_calculations_hospital_date ON coverage_calculations(hospital_id, calculation_date DESC);
CREATE INDEX idx_coverage_calculations_period ON coverage_calculations(calculation_period_start_date, calculation_period_end_date);

COMMENT ON TABLE coverage_calculations IS 'Results of DynamicCoverageCalculator service. Shows coverage percentage by position and shift type.';
COMMENT ON COLUMN coverage_calculations.coverage_by_position IS 'JSON object mapping positions to {required: N, assigned: N, coverage: 0.0-1.0}';
COMMENT ON COLUMN coverage_calculations.coverage_summary IS 'JSON object with aggregate coverage stats: {average_coverage: 0.0-1.0, critical_gaps: [...]}';
COMMENT ON COLUMN coverage_calculations.validation_errors IS 'JSON array of validation issues found during calculation';
COMMENT ON COLUMN coverage_calculations.query_count IS 'For testing/monitoring: number of queries used to calculate coverage (should be constant, never N+1)';
