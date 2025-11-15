CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255),
    role VARCHAR(50) NOT NULL DEFAULT 'USER'
        CHECK (role IN ('ADMIN', 'SCHEDULER', 'VIEWER', 'USER')),
    hospital_id UUID REFERENCES hospitals(id) ON DELETE SET NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_login_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE,
    deleted_by UUID
);

-- Indexes for common queries
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_hospital ON users(hospital_id, active) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_active ON users(active) WHERE deleted_at IS NULL;

COMMENT ON TABLE users IS 'System users with authentication and authorization. Supports role-based access control.';
COMMENT ON COLUMN users.role IS 'ADMIN (all access), SCHEDULER (can modify schedules), VIEWER (read-only), USER (limited access)';
COMMENT ON COLUMN users.hospital_id IS 'Hospital this user belongs to (NULL for system admins)';
COMMENT ON COLUMN users.password_hash IS 'Bcrypt hash of password. NULL if using SSO/Vault tokens.';
