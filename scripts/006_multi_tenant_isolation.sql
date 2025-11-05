-- Multi-Tenant Isolation Implementation for Enterprise Security

-- 1. Add tenant_id to all core tables
ALTER TABLE yt_aws_resources ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(50) DEFAULT 'default';
ALTER TABLE yt_resource_assessments ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(50) DEFAULT 'default';
ALTER TABLE yt_assessment_history ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(50) DEFAULT 'default';
ALTER TABLE yt_resource_identifiers ADD COLUMN IF NOT EXISTS tenant_id VARCHAR(50) DEFAULT 'default';

-- 2. Create tenant management table
CREATE TABLE IF NOT EXISTS yt_tenants (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(50) UNIQUE NOT NULL,
    tenant_name VARCHAR(255) NOT NULL,
    aws_account_id VARCHAR(12) NOT NULL,
    aws_role_arn VARCHAR(255) NOT NULL,
    
    -- Subscription details
    subscription_tier VARCHAR(20) DEFAULT 'standard', -- standard, premium, enterprise
    max_resources INTEGER DEFAULT 1000,
    
    -- Security settings
    api_key_hash VARCHAR(255) NOT NULL,
    allowed_ips JSONB DEFAULT '[]',
    
    -- Status
    status VARCHAR(20) DEFAULT 'active', -- active, suspended, trial
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 3. Create tenant-specific indexes for performance
CREATE INDEX IF NOT EXISTS idx_resources_tenant ON yt_aws_resources(tenant_id, sync_status);
CREATE INDEX IF NOT EXISTS idx_assessments_tenant ON yt_resource_assessments(tenant_id, assessment_timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_history_tenant ON yt_assessment_history(tenant_id, date DESC);

-- 4. Row Level Security (RLS) policies
ALTER TABLE yt_aws_resources ENABLE ROW LEVEL SECURITY;
ALTER TABLE yt_resource_assessments ENABLE ROW LEVEL SECURITY;
ALTER TABLE yt_assessment_history ENABLE ROW LEVEL SECURITY;

-- 5. Create tenant-specific database roles
DO $$
BEGIN
    -- Create tenant role template
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'tenant_role_template') THEN
        CREATE ROLE tenant_role_template;
        GRANT CONNECT ON DATABASE yukti TO tenant_role_template;
        GRANT USAGE ON SCHEMA public TO tenant_role_template;
        GRANT SELECT, INSERT, UPDATE ON yt_aws_resources TO tenant_role_template;
        GRANT SELECT, INSERT, UPDATE ON yt_resource_assessments TO tenant_role_template;
        GRANT SELECT ON yt_aws_pricing TO tenant_role_template;
    END IF;
END
$$;

-- 6. Tenant isolation functions
CREATE OR REPLACE FUNCTION get_current_tenant_id() RETURNS VARCHAR(50) AS $$
BEGIN
    -- Get tenant_id from session variable or default
    RETURN COALESCE(current_setting('app.current_tenant_id', true), 'default');
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- 7. RLS policies for tenant isolation
CREATE POLICY tenant_isolation_resources ON yt_aws_resources
    FOR ALL TO PUBLIC
    USING (tenant_id = get_current_tenant_id());

CREATE POLICY tenant_isolation_assessments ON yt_resource_assessments
    FOR ALL TO PUBLIC
    USING (tenant_id = get_current_tenant_id());

CREATE POLICY tenant_isolation_history ON yt_assessment_history
    FOR ALL TO PUBLIC
    USING (tenant_id = get_current_tenant_id());

-- 8. Tenant-aware views
CREATE OR REPLACE VIEW vw_tenant_resource_summary AS
SELECT 
    tenant_id,
    COUNT(*) as total_resources,
    COUNT(CASE WHEN state = 'running' THEN 1 END) as running_resources,
    SUM(CASE WHEN p.on_demand_price_usd IS NOT NULL 
        THEN p.on_demand_price_usd * 24 * 30 ELSE 0 END) as estimated_monthly_cost
FROM yt_aws_resources r
LEFT JOIN yt_aws_pricing p ON r.instance_type = p.instance_type AND r.region = p.region
WHERE r.sync_status = 'active'
GROUP BY tenant_id;

CREATE OR REPLACE VIEW vw_tenant_optimization_summary AS
SELECT 
    a.tenant_id,
    COUNT(*) as total_assessments,
    COUNT(CASE WHEN a.utilization_category = 'underutilized' THEN 1 END) as underutilized_count,
    COUNT(CASE WHEN a.utilization_category = 'overutilized' THEN 1 END) as overutilized_count,
    COUNT(CASE WHEN a.utilization_category = 'idle' THEN 1 END) as idle_count,
    AVG(a.optimization_score) as avg_optimization_score,
    SUM(a.potential_monthly_savings) as total_potential_savings
FROM yt_resource_assessments a
GROUP BY a.tenant_id;

-- 9. Tenant provisioning function
CREATE OR REPLACE FUNCTION provision_tenant(
    p_tenant_id VARCHAR(50),
    p_tenant_name VARCHAR(255),
    p_aws_account_id VARCHAR(12),
    p_aws_role_arn VARCHAR(255),
    p_api_key_hash VARCHAR(255)
) RETURNS BOOLEAN AS $$
BEGIN
    -- Insert tenant record
    INSERT INTO yt_tenants (tenant_id, tenant_name, aws_account_id, aws_role_arn, api_key_hash)
    VALUES (p_tenant_id, p_tenant_name, p_aws_account_id, p_aws_role_arn, p_api_key_hash);
    
    -- Create tenant-specific assessment config
    INSERT INTO yt_assessment_config (tenant_id)
    VALUES (p_tenant_id)
    ON CONFLICT (tenant_id) DO NOTHING;
    
    RETURN TRUE;
EXCEPTION
    WHEN OTHERS THEN
        RETURN FALSE;
END;
$$ LANGUAGE plpgsql;

-- 10. Tenant cleanup function
CREATE OR REPLACE FUNCTION cleanup_tenant_data(p_tenant_id VARCHAR(50)) RETURNS BOOLEAN AS $$
BEGIN
    -- Delete tenant data in correct order (respecting foreign keys)
    DELETE FROM yt_assessment_history WHERE tenant_id = p_tenant_id;
    DELETE FROM yt_resource_assessments WHERE tenant_id = p_tenant_id;
    DELETE FROM yt_resource_identifiers WHERE tenant_id = p_tenant_id;
    DELETE FROM yt_aws_resources WHERE tenant_id = p_tenant_id;
    DELETE FROM yt_assessment_config WHERE tenant_id = p_tenant_id;
    DELETE FROM yt_tenants WHERE tenant_id = p_tenant_id;
    
    RETURN TRUE;
EXCEPTION
    WHEN OTHERS THEN
        RETURN FALSE;
END;
$$ LANGUAGE plpgsql;

-- 11. Insert default tenant for existing data
INSERT INTO yt_tenants (tenant_id, tenant_name, aws_account_id, aws_role_arn, api_key_hash)
VALUES ('default', 'Default Tenant', '144403604430', 
        'arn:aws:iam::144403604430:role/YuktiFinOpsRole', 
        'default_api_key_hash')
ON CONFLICT (tenant_id) DO NOTHING;

-- 12. Update existing data with default tenant
UPDATE yt_aws_resources SET tenant_id = 'default' WHERE tenant_id IS NULL;
UPDATE yt_resource_assessments SET tenant_id = 'default' WHERE tenant_id IS NULL;
UPDATE yt_assessment_history SET tenant_id = 'default' WHERE tenant_id IS NULL;
UPDATE yt_resource_identifiers SET tenant_id = 'default' WHERE tenant_id IS NULL;

-- 13. Add NOT NULL constraints after data migration
ALTER TABLE yt_aws_resources ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE yt_resource_assessments ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE yt_assessment_history ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE yt_resource_identifiers ALTER COLUMN tenant_id SET NOT NULL;

SELECT 'Multi-tenant isolation implemented successfully' as status;