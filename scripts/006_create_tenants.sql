-- Multi-tenant architecture tables

-- Tenants (customers)
CREATE TABLE IF NOT EXISTS yt_tenants (
    id SERIAL PRIMARY KEY,
    tenant_code VARCHAR(50) NOT NULL UNIQUE,
    company_name VARCHAR(200) NOT NULL,
    subscription_tier VARCHAR(20) NOT NULL DEFAULT 'FREE', -- FREE, PROFESSIONAL, ENTERPRISE
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- active, suspended, cancelled
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    trial_ends_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_tenant_code ON yt_tenants(tenant_code);
CREATE INDEX idx_tenant_status ON yt_tenants(status);

-- AWS accounts linked to tenants
CREATE TABLE IF NOT EXISTS yt_aws_accounts (
    id SERIAL PRIMARY KEY,
    tenant_id INTEGER NOT NULL REFERENCES yt_tenants(id) ON DELETE CASCADE,
    account_id VARCHAR(12) NOT NULL,
    account_name VARCHAR(200),
    role_arn VARCHAR(500) NOT NULL, -- arn:aws:iam::123456789012:role/YuktiReadOnlyRole
    external_id VARCHAR(100) NOT NULL, -- Security token for AssumeRole
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, active, error
    last_sync TIMESTAMP WITH TIME ZONE,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(tenant_id, account_id)
);

CREATE INDEX idx_tenant_accounts ON yt_aws_accounts(tenant_id);
CREATE INDEX idx_account_status ON yt_aws_accounts(status);

-- Tenant-specific resources (replaces yt_aws_resources)
CREATE TABLE IF NOT EXISTS yt_tenant_resources (
    id SERIAL PRIMARY KEY,
    tenant_id INTEGER NOT NULL REFERENCES yt_tenants(id) ON DELETE CASCADE,
    aws_account_id INTEGER NOT NULL REFERENCES yt_aws_accounts(id) ON DELETE CASCADE,
    resource_id VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50) NOT NULL, -- ec2, rds, s3, lambda, etc
    region VARCHAR(50) NOT NULL,
    instance_type VARCHAR(50),
    state VARCHAR(20),
    tags JSONB DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    monthly_cost DECIMAL(12,2),
    last_synced TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(tenant_id, aws_account_id, resource_id)
);

CREATE INDEX idx_tenant_resources ON yt_tenant_resources(tenant_id);
CREATE INDEX idx_resource_type ON yt_tenant_resources(resource_type);
CREATE INDEX idx_tags_gin ON yt_tenant_resources USING GIN(tags);

-- Tenant-specific recommendations
CREATE TABLE IF NOT EXISTS yt_tenant_recommendations (
    id SERIAL PRIMARY KEY,
    tenant_id INTEGER NOT NULL REFERENCES yt_tenants(id) ON DELETE CASCADE,
    resource_id INTEGER REFERENCES yt_tenant_resources(id) ON DELETE CASCADE,
    recommendation_type VARCHAR(50) NOT NULL,
    current_cost DECIMAL(12,2),
    optimized_cost DECIMAL(12,2),
    monthly_savings DECIMAL(12,2),
    confidence_score DECIMAL(3,2),
    status VARCHAR(20) DEFAULT 'pending', -- pending, accepted, rejected, applied
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_tenant_recs ON yt_tenant_recommendations(tenant_id);
CREATE INDEX idx_rec_status ON yt_tenant_recommendations(status);

SELECT 'Multi-tenant tables created successfully' as status;
