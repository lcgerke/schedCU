-- Create persons table (staff registry)
-- Stores radiologists and other hospital staff
CREATE TABLE persons (
    id UUID PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    specialty VARCHAR(50) NOT NULL CHECK (specialty IN ('BODY_ONLY', 'NEURO_ONLY', 'BOTH')),
    active BOOLEAN NOT NULL DEFAULT TRUE,
    aliases TEXT[] DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Indexes for common queries
CREATE INDEX idx_persons_email ON persons(email);
CREATE INDEX idx_persons_specialty ON persons(specialty);
CREATE INDEX idx_persons_active ON persons(active, deleted_at);
CREATE INDEX idx_persons_created_at ON persons(created_at DESC);
