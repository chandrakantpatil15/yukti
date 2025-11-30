-- Security tables for Week 8

-- API Keys with hashing
CREATE TABLE IF NOT EXISTS yt_api_keys (
    id SERIAL PRIMARY KEY,
    tenant_id INTEGER NOT NULL REFERENCES yt_tenants(id) ON DELETE CASCADE,
    key_name VARCHAR(100) NOT NULL,
    key_hash VARCHAR(64) NOT NULL UNIQUE,
    key_prefix VARCHAR(20) NOT NULL,
    scopes TEXT[] DEFAULT '{}',
    last_used TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    revoked BOOLEAN DEFAULT FALSE
);

CREATE INDEX idx_api_keys_tenant ON yt_api_keys(tenant_id);
CREATE INDEX idx_api_keys_hash ON yt_api_keys(key_hash);

-- Audit logs
CREATE TABLE IF NOT EXISTS yt_audit_logs (
    id SERIAL PRIMARY KEY,
    tenant_id INTEGER REFERENCES yt_tenants(id) ON DELETE SET NULL,
    user_id VARCHAR(100),
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50),
    resource_id VARCHAR(100),
    ip_address INET,
    user_agent TEXT,
    request_method VARCHAR(10),
    request_path TEXT,
    status_code INTEGER,
    error_message TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_audit_tenant ON yt_audit_logs(tenant_id);
CREATE INDEX idx_audit_action ON yt_audit_logs(action);
CREATE INDEX idx_audit_created ON yt_audit_logs(created_at);

-- Secrets (encrypted)
CREATE TABLE IF NOT EXISTS yt_secrets (
    id SERIAL PRIMARY KEY,
    tenant_id INTEGER NOT NULL REFERENCES yt_tenants(id) ON DELETE CASCADE,
    secret_key VARCHAR(100) NOT NULL,
    secret_value_encrypted TEXT NOT NULL,
    secret_type VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(tenant_id, secret_key)
);

CREATE INDEX idx_secrets_tenant ON yt_secrets(tenant_id);

SELECT 'Security tables created successfully' as status;
