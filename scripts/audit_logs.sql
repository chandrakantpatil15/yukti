-- Audit logs table for security monitoring
CREATE TABLE IF NOT EXISTS yt_audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_user VARCHAR(255) NOT NULL,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    resource_id VARCHAR(255),
    tenant_id VARCHAR(255),
    ip_address VARCHAR(45),
    user_agent TEXT,
    details JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_admin ON yt_audit_logs(admin_user);
CREATE INDEX idx_audit_logs_tenant ON yt_audit_logs(tenant_id);
CREATE INDEX idx_audit_logs_created ON yt_audit_logs(created_at DESC);
