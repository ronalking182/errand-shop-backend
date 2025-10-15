-- Migration: Create Audit Logs Table
-- This migration creates the audit_logs table for system logging

CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID,
    action VARCHAR(100) NOT NULL,
    resource VARCHAR(100) NOT NULL,
    resource_id VARCHAR(100),
    description TEXT,
    ip_address VARCHAR(45),
    user_agent TEXT,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource ON audit_logs(resource);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource_id ON audit_logs(resource_id);

-- Add comments
COMMENT ON TABLE audit_logs IS 'System audit logs for tracking user and system actions';
COMMENT ON COLUMN audit_logs.user_id IS 'ID of the user who performed the action (null for system actions)';
COMMENT ON COLUMN audit_logs.action IS 'Action performed (e.g., CREATE, UPDATE, DELETE)';
COMMENT ON COLUMN audit_logs.resource IS 'Resource type affected (e.g., user, product, order)';
COMMENT ON COLUMN audit_logs.resource_id IS 'ID of the specific resource affected';
COMMENT ON COLUMN audit_logs.metadata IS 'Additional metadata about the action in JSON format';