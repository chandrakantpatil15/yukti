-- ============================================
-- Yukti Platform - E2E Test Cleanup Script
-- ============================================
-- This script cleans up all user data for fresh end-to-end testing
-- Run this before starting a new onboarding test

-- 1. Delete all JWT refresh tokens
DELETE FROM yt_refresh_tokens;

-- 2. Delete all AWS connections
DELETE FROM yt_aws_connections;

-- 3. Delete all discovered resources
DELETE FROM yt_tenant_resources;

-- 4. Delete all AWS accounts
DELETE FROM yt_aws_accounts;

-- 5. Delete all cost optimization findings
DELETE FROM yt_hidden_cost_findings;

-- 6. Delete all whitelists
DELETE FROM yt_whitelists;

-- 7. Delete all budgets
DELETE FROM yt_budgets;

-- 8. Delete all cost data
DELETE FROM yt_cost_data;

-- 9. Delete all RI/SP recommendations
DELETE FROM yt_ri_recommendations;
DELETE FROM yt_sp_recommendations;

-- 10. Delete all metrics integrations
DELETE FROM yt_metrics_integrations;

-- 11. Delete all customers (tenants)
DELETE FROM yt_customers;

-- 12. Delete all users
DELETE FROM yt_users;

-- 13. Reset sequences (optional - for clean IDs)
ALTER SEQUENCE yt_customers_id_seq RESTART WITH 1;
ALTER SEQUENCE yt_aws_accounts_id_seq RESTART WITH 1;

-- ============================================
-- Verification Queries
-- ============================================
-- Run these to confirm cleanup

SELECT 'Users' as table_name, COUNT(*) as count FROM yt_users
UNION ALL
SELECT 'Customers', COUNT(*) FROM yt_customers
UNION ALL
SELECT 'AWS Connections', COUNT(*) FROM yt_aws_connections
UNION ALL
SELECT 'Resources', COUNT(*) FROM yt_tenant_resources
UNION ALL
SELECT 'Findings', COUNT(*) FROM yt_hidden_cost_findings
UNION ALL
SELECT 'Budgets', COUNT(*) FROM yt_budgets
UNION ALL
SELECT 'Whitelists', COUNT(*) FROM yt_whitelists;

-- ============================================
-- Expected Result: All counts should be 0
-- ============================================
