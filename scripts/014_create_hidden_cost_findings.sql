-- Hidden cost findings table
-- Migration: 014_create_hidden_cost_findings.sql

CREATE TABLE IF NOT EXISTS yt_hidden_cost_findings (
    id VARCHAR(50) PRIMARY KEY,
    tenant_id VARCHAR(50) NOT NULL,
    detector_name VARCHAR(100) NOT NULL,
    category VARCHAR(100) NOT NULL,
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('Low', 'Medium', 'High', 'Critical')),
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    resource_arn TEXT NOT NULL,
    estimated_savings DECIMAL(12,2) NOT NULL DEFAULT 0,
    confidence DECIMAL(3,2) NOT NULL DEFAULT 0 CHECK (confidence >= 0 AND confidence <= 1),
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'resolved', 'whitelisted', 'ignored')),
    resolved_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_findings_tenant ON yt_hidden_cost_findings(tenant_id);
CREATE INDEX IF NOT EXISTS idx_findings_category ON yt_hidden_cost_findings(category);
CREATE INDEX IF NOT EXISTS idx_findings_severity ON yt_hidden_cost_findings(severity);
CREATE INDEX IF NOT EXISTS idx_findings_status ON yt_hidden_cost_findings(status);
CREATE INDEX IF NOT EXISTS idx_findings_created_at ON yt_hidden_cost_findings(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_findings_savings ON yt_hidden_cost_findings(estimated_savings DESC);

-- Composite index for common queries
CREATE INDEX IF NOT EXISTS idx_findings_tenant_status_created ON yt_hidden_cost_findings(tenant_id, status, created_at DESC);

-- Update trigger for updated_at
CREATE OR REPLACE FUNCTION update_findings_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_findings_updated_at
    BEFORE UPDATE ON yt_hidden_cost_findings
    FOR EACH ROW
    EXECUTE FUNCTION update_findings_updated_at();

SELECT 'yt_hidden_cost_findings table created successfully' as status;
