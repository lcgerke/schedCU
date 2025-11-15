-- Create scrape_batches table
-- Tracks atomic batch operations for schedule imports with full traceability
CREATE TABLE scrape_batches (
    id UUID PRIMARY KEY,
    hospital_id UUID NOT NULL REFERENCES hospitals(id),
    state VARCHAR(50) NOT NULL DEFAULT 'PENDING'
        CHECK (state IN ('PENDING', 'COMPLETE', 'FAILED')),
    window_start_date DATE NOT NULL,
    window_end_date DATE NOT NULL,
    scraped_at TIMESTAMP WITH TIME ZONE NOT NULL,
    completed_at TIMESTAMP WITH TIME ZONE,
    row_count INTEGER DEFAULT 0,
    ingest_checksum VARCHAR(255),
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,
    deleted_by UUID,
    archived_at TIMESTAMP WITH TIME ZONE,
    archived_by UUID
);

-- Indexes for common queries and integrity
CREATE INDEX idx_scrape_batches_hospital ON scrape_batches(hospital_id, state);
CREATE INDEX idx_scrape_batches_state ON scrape_batches(state);
CREATE INDEX idx_scrape_batches_window ON scrape_batches(window_start_date, window_end_date);
CREATE INDEX idx_scrape_batches_created ON scrape_batches(created_at DESC);
CREATE INDEX idx_scrape_batches_checksum ON scrape_batches(ingest_checksum);

-- Data integrity constraint
ALTER TABLE scrape_batches ADD CONSTRAINT check_date_range
    CHECK (window_start_date <= window_end_date);
