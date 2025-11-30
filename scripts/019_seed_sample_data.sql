-- Seed sample data for tenant 18 (yourname123@example.com)
-- This makes the UI functional with realistic data

-- Sample findings for Hidden Costs page
INSERT INTO yt_hidden_cost_findings (tenant_id, detector_name, resource_arn, resource_type, severity, category, estimated_monthly_cost, estimated_savings, description, recommendation, status, created_at)
VALUES 
('18', 'unused-ebs-volumes', 'arn:aws:ec2:us-east-1:999888777666:volume/vol-abc123', 'ebs-volume', 'high', 'storage', 50.00, 50.00, 'EBS volume not attached to any instance for 30+ days', 'Delete unused volume or create snapshot and delete', 'open', NOW()),
('18', 'underutilized-ec2', 'arn:aws:ec2:us-east-1:999888777666:instance/i-xyz789', 'ec2-instance', 'medium', 'compute', 200.00, 100.00, 'EC2 instance CPU utilization < 10% for 7 days', 'Downsize from t3.large to t3.small', 'open', NOW()),
('18', 'unoptimized-s3-storage', 'arn:aws:s3:::analytics-bucket-2024', 's3-bucket', 'low', 'storage', 30.00, 15.00, 'S3 bucket using Standard storage for infrequent access patterns', 'Move to S3 Intelligent-Tiering or Glacier', 'open', NOW()),
('18', 'nat-gateway-idle', 'arn:aws:ec2:us-east-1:999888777666:natgateway/nat-0abc123', 'nat-gateway', 'high', 'networking', 45.00, 45.00, 'NAT Gateway with no traffic for 14 days', 'Delete NAT Gateway if not needed', 'open', NOW()),
('18', 'old-snapshot', 'arn:aws:ec2:us-east-1:999888777666:snapshot/snap-xyz456', 'ebs-snapshot', 'low', 'storage', 12.00, 12.00, 'EBS snapshot older than 90 days', 'Delete old snapshots or move to archive', 'open', NOW());

-- Sample resources for Resources page
INSERT INTO yt_tenant_resources (tenant_id, resource_id, resource_type, resource_name, region, tags, monthly_cost, created_at)
VALUES
('18', 'i-xyz789', 'ec2-instance', 'web-server-prod-1', 'us-east-1', '{"env":"production","team":"backend","app":"api"}', 200.00, NOW()),
('18', 'i-abc456', 'ec2-instance', 'worker-node-1', 'us-east-1', '{"env":"production","team":"data","app":"etl"}', 150.00, NOW()),
('18', 'vol-abc123', 'ebs-volume', 'unused-volume-old', 'us-east-1', '{"created":"2023-01-15"}', 50.00, NOW()),
('18', 'analytics-bucket-2024', 's3-bucket', 'analytics-bucket-2024', 'us-east-1', '{"project":"analytics","team":"data"}', 30.00, NOW()),
('18', 'logs-bucket', 's3-bucket', 'logs-bucket', 'us-east-1', '{"project":"logging","retention":"90days"}', 80.00, NOW()),
('18', 'nat-0abc123', 'nat-gateway', 'nat-gateway-public', 'us-east-1', '{"vpc":"vpc-main"}', 45.00, NOW()),
('18', 'db-prod-1', 'rds-instance', 'postgres-prod', 'us-east-1', '{"env":"production","engine":"postgres"}', 350.00, NOW());

-- Sample cost data for Dashboard
INSERT INTO yt_cost_data (tenant_id, date, service, cost, usage_type, region, created_at)
VALUES
('18', CURRENT_DATE - INTERVAL '1 day', 'EC2', 350.00, 'compute', 'us-east-1', NOW()),
('18', CURRENT_DATE - INTERVAL '1 day', 'S3', 110.00, 'storage', 'us-east-1', NOW()),
('18', CURRENT_DATE - INTERVAL '1 day', 'RDS', 350.00, 'database', 'us-east-1', NOW()),
('18', CURRENT_DATE - INTERVAL '1 day', 'NAT Gateway', 45.00, 'networking', 'us-east-1', NOW()),
('18', CURRENT_DATE - INTERVAL '2 days', 'EC2', 340.00, 'compute', 'us-east-1', NOW()),
('18', CURRENT_DATE - INTERVAL '2 days', 'S3', 105.00, 'storage', 'us-east-1', NOW()),
('18', CURRENT_DATE - INTERVAL '2 days', 'RDS', 350.00, 'database', 'us-east-1', NOW());

-- Update customer onboarding status
UPDATE yt_customers 
SET onboarding_status = 'completed', 
    completed_at = NOW(),
    current_step = 'review_findings'
WHERE tenant_id = '18';

-- Sample budget
INSERT INTO yt_budgets (tenant_id, name, amount, period, alert_threshold, created_at)
VALUES ('18', 'Monthly AWS Budget', 1000.00, 'monthly', 80, NOW())
ON CONFLICT (tenant_id, name) DO NOTHING;

-- Sample RI recommendations
INSERT INTO yt_ri_recommendations (tenant_id, instance_type, region, recommended_quantity, estimated_savings, upfront_cost, term_length, created_at)
VALUES 
('18', 't3.large', 'us-east-1', 2, 120.00, 500.00, '1year', NOW()),
('18', 'db.t3.large', 'us-east-1', 1, 200.00, 800.00, '1year', NOW())
ON CONFLICT DO NOTHING;

COMMENT ON TABLE yt_hidden_cost_findings IS 'Sample data seeded for tenant 18 - demo purposes';
