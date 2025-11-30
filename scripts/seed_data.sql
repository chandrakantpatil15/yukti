-- Seed data for testing

-- Insert test customers
INSERT INTO yt_customers (id, tenant_id, company_name, email, onboarding_status, onboarding_step, created_at, completed_at) VALUES
('cust-001', 'tenant-001', 'Acme Corp', 'admin@acme.com', 'completed', 'review_findings', '2024-01-15 10:00:00', '2024-01-15 11:30:00'),
('cust-002', 'tenant-002', 'TechStart Inc', 'cto@techstart.io', 'in_progress', 'initial_scan', '2024-01-20 14:00:00', NULL),
('cust-003', 'tenant-003', 'CloudScale LLC', 'ops@cloudscale.com', 'completed', 'review_findings', '2024-01-10 09:00:00', '2024-01-10 10:45:00');

-- Insert AWS connections
INSERT INTO yt_aws_connections (tenant_id, account_id, role_arn, external_id, regions, created_at) VALUES
('tenant-001', '123456789012', 'arn:aws:iam::123456789012:role/YuktiRole', 'ext-001', '{"us-east-1", "us-west-2"}', '2024-01-15 10:30:00'),
('tenant-002', '234567890123', 'arn:aws:iam::234567890123:role/YuktiRole', 'ext-002', '{"us-east-1"}', '2024-01-20 14:30:00'),
('tenant-003', '345678901234', 'arn:aws:iam::345678901234:role/YuktiRole', 'ext-003', '{"us-east-1", "eu-west-1"}', '2024-01-10 09:30:00');

-- Insert hidden cost findings
INSERT INTO yt_hidden_cost_findings (id, tenant_id, detector_name, category, severity, title, description, resource_arn, estimated_savings, confidence, created_at) VALUES
-- Acme Corp findings
('find-001', 'tenant-001', 'cross_az_data_transfer', 'Data Transfer Costs', 'High', 'RDS Multi-AZ cross-AZ data transfer', 'RDS Multi-AZ incurs cross-AZ data transfer charges', 'arn:aws:rds:us-east-1:123456789012:db:prod-db', 450.00, 0.95, NOW()),
('find-002', 'tenant-001', 'ebs_gp2_vs_gp3', 'Storage Lifecycle Waste', 'Medium', 'EBS gp2 volumes should migrate to gp3', 'gp3 is 20% cheaper with better performance', 'arn:aws:ec2:us-east-1:123456789012:volume/vol-123', 20.00, 0.98, NOW()),
('find-003', 'tenant-001', 'idle_load_balancer', 'Networking Costs', 'Medium', 'Idle load balancer with minimal traffic', 'Load balancer costs $16.20/month with <100 requests/hour', 'arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/test-lb', 16.20, 0.85, NOW()),
('find-004', 'tenant-002', 'nat_gateway_waste', 'Data Transfer Costs', 'High', 'NAT Gateway processing high data volume', 'Use VPC endpoints for S3/DynamoDB to save 84%', 'arn:aws:ec2:us-east-1:234567890123:natgateway/nat-456', 378.00, 0.90, NOW()),
('find-005', 'tenant-002', 'orphaned_snapshots', 'Storage Lifecycle Waste', 'Medium', 'Orphaned EBS snapshot older than 90 days', 'Snapshot has no associated volume or AMI', 'arn:aws:ec2:us-east-1:234567890123:snapshot/snap-789', 50.00, 0.98, NOW()),
('find-006', 'tenant-003', 'k8s_spot_opportunity', 'Kubernetes Optimization', 'High', 'Stateless workload on On-Demand nodes', 'Spot instances save 70% for fault-tolerant workloads', 'arn:aws:eks:us-east-1:345678901234:nodegroup/prod-cluster/ng-123', 3500.00, 0.85, NOW()),
('find-007', 'tenant-003', 'ec2_burstable_t2_vs_t3', 'Compute Waste', 'Medium', 'T2 instances should upgrade to T3', 'T3 is 10% cheaper with better performance', 'arn:aws:ec2:us-east-1:345678901234:instance/i-abc', 20.00, 0.92, NOW());

-- Insert budgets
INSERT INTO yt_budgets (id, tenant_id, name, amount, period, start_date, alert_threshold, current_spend, status, created_at) VALUES
('budget-001', 'tenant-001', 'Production AWS', 15000.00, 'monthly', '2024-01-01', 80.00, 12500.00, 'active', NOW()),
('budget-002', 'tenant-002', 'Development AWS', 5000.00, 'monthly', '2024-01-01', 80.00, 3200.00, 'active', NOW()),
('budget-003', 'tenant-003', 'Total AWS Spend', 50000.00, 'monthly', '2024-01-01', 80.00, 42000.00, 'active', NOW());

-- Insert cost data
INSERT INTO yt_cost_data (id, tenant_id, date, service, cost, created_at) VALUES
('cost-001', 'tenant-001', '2024-01-25', 'EC2', 5200.00, NOW()),
('cost-002', 'tenant-001', '2024-01-25', 'RDS', 3100.00, NOW()),
('cost-003', 'tenant-001', '2024-01-25', 'S3', 1800.00, NOW()),
('cost-004', 'tenant-002', '2024-01-25', 'EC2', 2100.00, NOW()),
('cost-005', 'tenant-002', '2024-01-25', 'Lambda', 800.00, NOW()),
('cost-006', 'tenant-003', '2024-01-25', 'EKS', 15000.00, NOW()),
('cost-007', 'tenant-003', '2024-01-25', 'EC2', 18000.00, NOW());

-- Insert RI recommendations
INSERT INTO yt_ri_recommendations (id, tenant_id, service, instance_type, region, term, payment_option, monthly_savings, upfront_cost, created_at) VALUES
('ri-001', 'tenant-001', 'EC2', 'm5.large', 'us-east-1', '1 Year', 'No Upfront', 450.00, 0.00, NOW()),
('ri-002', 'tenant-003', 'EC2', 'c5.xlarge', 'us-east-1', '1 Year', 'No Upfront', 2000.00, 0.00, NOW());

-- Insert SP recommendations
INSERT INTO yt_sp_recommendations (id, tenant_id, service, term, payment_option, hourly_commitment, monthly_savings, upfront_cost, created_at) VALUES
('sp-001', 'tenant-001', 'Compute', '1 Year', 'No Upfront', 2.50, 900.00, 0.00, NOW()),
('sp-002', 'tenant-003', 'Compute', '1 Year', 'No Upfront', 8.00, 5000.00, 0.00, NOW());
