-- Migration 009: Admin Audit Logs
-- Purpose: Track all admin actions for security and compliance

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================================
-- 1. yt_admin_audit_logs: Track admin actions
-- ============================================================================
CREATE TABLE IF NOT EXISTS yt_admin_audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    admin_user_id UUID NOT NULL REFERENCES yt_users(id) ON DELETE CASCADE,
    action TEXT NOT NULL,
    resource_type TEXT NOT NULL,
    resource_id TEXT,
    tenant_id VARCHAR(50) REFERENCES yt_customers(id) ON DELETE SET NULL,
    target_user_id UUID REFERENCES yt_users(id) ON DELETE SET NULL,
    details JSONB,
    ip_address TEXT,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_admin_audit_admin ON yt_admin_audit_logs(admin_user_id);
CREATE INDEX idx_admin_audit_tenant ON yt_admin_audit_logs(tenant_id);
CREATE INDEX idx_admin_audit_action ON yt_admin_audit_logs(action);
CREATE INDEX idx_admin_audit_created ON yt_admin_audit_logs(created_at DESC);
CREATE INDEX idx_admin_audit_resource ON yt_admin_audit_logs(resource_type, resource_id);

-- ============================================================================
-- 2. yt_impersonation_sessions: Track admin impersonation
-- ============================================================================
CREATE TABLE IF NOT EXISTS yt_impersonation_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    admin_user_id UUID NOT NULL REFERENCES yt_users(id) ON DELETE CASCADE,
    target_user_id UUID NOT NULL REFERENCES yt_users(id) ON DELETE CASCADE,
    tenant_id VARCHAR(50) NOT NULL REFERENCES yt_customers(id) ON DELETE CASCADE,
    reason TEXT NOT NULL,
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    ended_at TIMESTAMP WITH TIME ZONE,
    ip_address TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true
);

CREATE INDEX idx_impersonation_admin ON yt_impersonation_sessions(admin_user_id);
CREATE INDEX idx_impersonation_target ON yt_impersonation_sessions(target_user_id);
CREATE INDEX idx_impersonation_active ON yt_impersonation_sessions(is_active) WHERE is_active = true;

SELECT 'Migration 009: Admin Audit Logs completed' as status;
