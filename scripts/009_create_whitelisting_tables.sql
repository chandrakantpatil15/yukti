-- Whitelisting tables
CREATE TABLE IF NOT EXISTS yt_whitelists (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES yt_tenants(id) ON DELETE CASCADE,
    
    whitelist_type VARCHAR(20) NOT NULL,
    resource_arn VARCHAR(500),
    tag_key VARCHAR(128),
    tag_value VARCHAR(256),
    service_name VARCHAR(50),
    recommendation_type VARCHAR(50),
    
    reason TEXT NOT NULL,
    business_justification TEXT,
    cost_impact_monthly DECIMAL(12,2),
    
    created_by VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    approved_by VARCHAR(100),
    approved_at TIMESTAMP,
    
    expires_at TIMESTAMP,
    expiry_reminder_sent BOOLEAN DEFAULT FALSE,
    
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    revoked_by VARCHAR(100),
    revoked_at TIMESTAMP,
    revoked_reason TEXT,
    
    CONSTRAINT valid_whitelist_type CHECK (whitelist_type IN ('resource', 'tag', 'service', 'recommendation_type')),
    CONSTRAINT valid_status CHECK (status IN ('active', 'expired', 'revoked', 'pending_approval'))
);

CREATE INDEX IF NOT EXISTS idx_whitelists_tenant ON yt_whitelists(tenant_id);
CREATE INDEX IF NOT EXISTS idx_whitelists_resource ON yt_whitelists(resource_arn);
CREATE INDEX IF NOT EXISTS idx_whitelists_status ON yt_whitelists(status);

-- Hidden cost findings
CREATE TABLE IF NOT EXISTS yt_hidden_cost_findings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES yt_tenants(id) ON DELETE CASCADE,
    
    detector_name VARCHAR(100) NOT NULL,
    category VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    resource_arn VARCHAR(500) NOT NULL,
    
    estimated_cost DECIMAL(12,2),
    estimated_savings DECIMAL(12,2),
    confidence DECIMAL(3,2),
    recommendation TEXT,
    
    detected_at TIMESTAMP NOT NULL DEFAULT NOW(),
    suppressed_until TIMESTAMP,
    
    UNIQUE(tenant_id, resource_arn, detector_name)
);

CREATE INDEX IF NOT EXISTS idx_hidden_costs_tenant ON yt_hidden_cost_findings(tenant_id);
CREATE INDEX IF NOT EXISTS idx_hidden_costs_category ON yt_hidden_cost_findings(category);
CREATE INDEX IF NOT EXISTS idx_hidden_costs_severity ON yt_hidden_cost_findings(severity);
