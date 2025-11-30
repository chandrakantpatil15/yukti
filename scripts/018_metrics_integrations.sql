-- Create metrics integrations table for onboarding
CREATE TABLE IF NOT EXISTS yt_metrics_integrations (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    source VARCHAR(50) NOT NULL,
    endpoint TEXT,
    verified BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(tenant_id)
);

CREATE INDEX idx_metrics_integrations_tenant ON yt_metrics_integrations(tenant_id);

COMMENT ON TABLE yt_metrics_integrations IS 'Stores metrics integration configurations for customer onboarding';
