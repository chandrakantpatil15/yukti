-- Seed sample data for tenant 18 with CORRECT schema

-- Sample findings for Hidden Costs page (matching actual schema)
INSERT INTO yt_hidden_cost_findings (id, tenant_id, detector_name, category, severity, title, description, resource_arn, estimated_savings, confidence, status, created_at)
VALUES 
('finding-001', '18', 'unused-ebs-volumes', 'Storage', 'High', 'Unused EBS Volume', 'EBS volume not attached to any instance for 30+ days', 'arn:aws:ec2:us-east-1:999888777666:volume/vol-abc123', 50.00, 0.95, 'active', NOW()),
('finding-002', '18', 'underutilized-ec2', 'Compute', 'Medium', 'Underutilized EC2 Instance', 'EC2 instance CPU utilization < 10% for 7 days', 'arn:aws:ec2:us-east-1:999888777666:instance/i-xyz789', 100.00, 0.88, 'active', NOW()),
('finding-003', '18', 'unoptimized-s3-storage', 'Storage', 'Low', 'Unoptimized S3 Storage Class', 'S3 bucket using Standard storage for infrequent access patterns', 'arn:aws:s3:::analytics-bucket-2024', 15.00, 0.75, 'active', NOW()),
('finding-004', '18', 'nat-gateway-idle', 'Networking', 'High', 'Idle NAT Gateway', 'NAT Gateway with no traffic for 14 days', 'arn:aws:ec2:us-east-1:999888777666:natgateway/nat-0abc123', 45.00, 0.92, 'active', NOW()),
('finding-005', '18', 'old-snapshot', 'Storage', 'Low', 'Old EBS Snapshot', 'EBS snapshot older than 90 days', 'arn:aws:ec2:us-east-1:999888777666:snapshot/snap-xyz456', 12.00, 0.80, 'active', NOW()),
('finding-006', '18', 'unattached-eip', 'Networking', 'Medium', 'Unattached Elastic IP', 'Elastic IP not associated with any instance', 'arn:aws:ec2:us-east-1:999888777666:eip/eipalloc-abc123', 3.60, 0.99, 'active', NOW()),
('finding-007', '18', 'oversized-rds', 'Database', 'High', 'Oversized RDS Instance', 'RDS instance CPU < 20% and connections < 10', 'arn:aws:rds:us-east-1:999888777666:db:prod-db', 200.00, 0.85, 'active', NOW())
ON CONFLICT (id) DO NOTHING;

-- Sample resources (check actual schema first)
-- Skipping if table structure is different

-- Sample cost data (check actual schema)
-- Skipping if table structure is different

-- Update customer onboarding status (check actual schema)
-- Skipping if column doesn't exist

-- Sample budget (check actual schema)
INSERT INTO yt_budgets (tenant_id, name, amount, period, alert_threshold, created_at)
VALUES ('18', 'Monthly AWS Budget', 1000.00, 'monthly', 80, NOW())
ON CONFLICT DO NOTHING;

COMMENT ON TABLE yt_hidden_cost_findings IS 'Sample data seeded for tenant 18';
