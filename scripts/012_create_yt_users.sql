-- User accounts and RBAC migration
-- Note: tenant_id uses INTEGER to match yt_tenants.id (SERIAL)
-- If migrating to UUID, update yt_tenants.id first, then this table

-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table with tenant isolation and role-based access
CREATE TABLE IF NOT EXISTS yt_users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id INTEGER NOT NULL REFERENCES yt_tenants(id) ON DELETE CASCADE,
    email TEXT NOT NULL,
    password_hash TEXT NOT NULL, -- bcrypt hashed password
    role TEXT NOT NULL CHECK (role IN ('admin', 'editor', 'viewer')),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Unique constraint: email must be unique per tenant
    CONSTRAINT unique_email_per_tenant UNIQUE (tenant_id, email)
);

-- Indexes for performance
CREATE INDEX idx_users_tenant_email ON yt_users(tenant_id, email);
CREATE INDEX idx_users_tenant_role ON yt_users(tenant_id, role);
CREATE INDEX idx_users_email ON yt_users(email) WHERE is_active = true;

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_yt_users_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to auto-update updated_at
CREATE TRIGGER trigger_update_yt_users_updated_at
    BEFORE UPDATE ON yt_users
    FOR EACH ROW
    EXECUTE FUNCTION update_yt_users_updated_at();

-- Add user_id to audit logs for better tracking
ALTER TABLE yt_audit_logs 
    ADD COLUMN IF NOT EXISTS user_id UUID REFERENCES yt_users(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_audit_logs_user ON yt_audit_logs(user_id);

SELECT 'yt_users table created successfully' as status;

