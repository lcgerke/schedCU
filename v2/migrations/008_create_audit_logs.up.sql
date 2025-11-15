CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    action VARCHAR(255) NOT NULL,
    resource VARCHAR(255) NOT NULL,
    resource_id UUID,
    old_values TEXT,
    new_values TEXT,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    ip_address INET
);

-- Indexes for common queries
CREATE INDEX idx_audit_logs_user ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource);
CREATE INDEX idx_audit_logs_resource_id ON audit_logs(resource_id);
CREATE INDEX idx_audit_logs_timestamp ON audit_logs(timestamp DESC);

COMMENT ON TABLE audit_logs IS 'Comprehensive audit trail for HIPAA compliance. All administrative actions are logged here.';
COMMENT ON COLUMN audit_logs.action IS 'Action type: CREATE, UPDATE, DELETE, PROMOTE, ARCHIVE, IMPORT, etc.';
COMMENT ON COLUMN audit_logs.resource IS 'Resource type: schedule_version, shift_instance, assignment, etc.';
COMMENT ON COLUMN audit_logs.old_values IS 'Previous values as JSON string (for UPDATE/DELETE)';
COMMENT ON COLUMN audit_logs.new_values IS 'New values as JSON string (for CREATE/UPDATE)';
COMMENT ON COLUMN audit_logs.ip_address IS 'IP address of user making the change (for security audits)';
