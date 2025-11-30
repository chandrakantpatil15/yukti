-- Performance indexes for time-range queries
-- Migration: 015_add_performance_indexes.sql

-- Cost data indexes
CREATE INDEX IF NOT EXISTS idx_cost_data_tenant_date ON yt_cost_data(tenant_id, date DESC);
CREATE INDEX IF NOT EXISTS idx_cost_data_service_date ON yt_cost_data(service, date DESC);
CREATE INDEX IF NOT EXISTS idx_cost_data_date ON yt_cost_data(date DESC);

-- Audit logs indexes
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON yt_audit_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_created ON yt_audit_logs(user_id, created_at DESC) WHERE user_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_audit_logs_action_created ON yt_audit_logs(action, created_at DESC);

-- Tenant resources indexes
CREATE INDEX IF NOT EXISTS idx_tenant_resources_synced ON yt_tenant_resources(last_synced DESC);
CREATE INDEX IF NOT EXISTS idx_tenant_resources_tenant_type ON yt_tenant_resources(tenant_id, resource_type);

-- Recommendations indexes
CREATE INDEX IF NOT EXISTS idx_tenant_recs_created ON yt_tenant_recommendations(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_tenant_recs_tenant_status ON yt_tenant_recommendations(tenant_id, status);

-- Budget indexes
CREATE INDEX IF NOT EXISTS idx_budgets_tenant_status ON yt_budgets(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_budgets_start_date ON yt_budgets(start_date DESC);

-- RI/SP recommendations indexes
CREATE INDEX IF NOT EXISTS idx_ri_recs_tenant_created ON yt_ri_recommendations(tenant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_sp_recs_tenant_created ON yt_sp_recommendations(tenant_id, created_at DESC);

SELECT 'Performance indexes created successfully' as status;
