-- Insert tenants
INSERT INTO yt_tenants (tenant_id, tenant_name, aws_account_id, aws_role_arn, subscription_tier, max_resources, api_key_hash)
VALUES 
    ('tenant1', 'Enterprise Corp', '123456789012', 'arn:aws:iam::123456789012:role/YuktiFinOpsRole', 'enterprise', 5000, 'dummy_hash_1'),
    ('tenant2', 'Startup Inc', '234567890123', 'arn:aws:iam::234567890123:role/YuktiFinOpsRole', 'standard', 1000, 'dummy_hash_2'),
    ('tenant3', 'Tech LLC', '345678901234', 'arn:aws:iam::345678901234:role/YuktiFinOpsRole', 'premium', 2500, 'dummy_hash_3'),
    ('tenant4', 'Data Corp', '456789012345', 'arn:aws:iam::456789012345:role/YuktiFinOpsRole', 'enterprise', 5000, 'dummy_hash_4')
ON CONFLICT (tenant_id) DO UPDATE SET
    tenant_name = EXCLUDED.tenant_name,
    subscription_tier = EXCLUDED.subscription_tier,
    max_resources = EXCLUDED.max_resources;

-- Insert EC2 instances
INSERT INTO yt_aws_resources (
    tenant_id, instance_id, instance_type, region, availability_zone,
    state, platform, architecture, launch_time, environment,
    project_name, tags, sync_status
)
VALUES 
    -- Tenant 1 Resources
    ('tenant1', 'i-0abc123def456789a', 't3.xlarge', 'us-east-1', 'us-east-1a',
     'running', 'linux', 'x86_64', '2025-10-01 00:00:00+00', 'production',
     'e-commerce', '{"Name": "prod-web-01", "Environment": "production", "Project": "e-commerce", "CostCenter": "cc-123"}', 'active'),
    ('tenant1', 'i-0bcd234efg567890b', 'm5.2xlarge', 'us-west-2', 'us-west-2b',
     'running', 'linux', 'x86_64', '2025-09-15 00:00:00+00', 'production',
     'e-commerce', '{"Name": "prod-db-01", "Environment": "production", "Project": "e-commerce", "CostCenter": "cc-123"}', 'active'),
    ('tenant1', 'i-0ef1234ghi567890d', 'c5.2xlarge', 'us-east-1', 'us-east-1a',
     'running', 'linux', 'x86_64', '2025-10-01 00:00:00+00', 'production',
     'api-service', '{"Name": "prod-api-01", "Environment": "production", "Service": "api", "Team": "backend", "CostCenter": "cc-123"}', 'active'),
    
    -- Tenant 2 Resources
    ('tenant2', 'i-0cde345fgh678901c', 't3.medium', 'us-east-1', 'us-east-1b',
     'running', 'linux', 'x86_64', '2025-10-15 00:00:00+00', 'development',
     'mobile-app', '{"Name": "dev-app-01", "Environment": "development", "Project": "mobile-app", "CostCenter": "cc-456"}', 'active')
ON CONFLICT (instance_id) DO UPDATE SET
    instance_type = EXCLUDED.instance_type,
    state = EXCLUDED.state,
    tags = EXCLUDED.tags,
    sync_status = EXCLUDED.sync_status,
    last_synced = NOW();

-- Insert AWS Pricing data
INSERT INTO yt_aws_pricing (
    instance_type, region, os, vcpu, memory_gb,
    on_demand_price_usd, reserved_1yr_no_upfront, spot_price_avg
)
VALUES
    ('t3.medium', 'us-east-1', 'Linux', 2, 4, 0.0416, 0.0250, 0.0125),
    ('t3.xlarge', 'us-east-1', 'Linux', 4, 16, 0.1664, 0.0998, 0.0499),
    ('m5.2xlarge', 'us-west-2', 'Linux', 8, 32, 0.384, 0.230, 0.115),
    ('c5.2xlarge', 'us-east-1', 'Linux', 8, 16, 0.34, 0.204, 0.102),
    ('r5.4xlarge', 'us-east-1', 'Linux', 16, 128, 1.008, 0.605, 0.302),
    ('m5.xlarge', 'eu-west-1', 'Linux', 4, 16, 0.192, 0.115, 0.058)
ON CONFLICT (instance_type, region, os) DO UPDATE SET
    on_demand_price_usd = EXCLUDED.on_demand_price_usd,
    reserved_1yr_no_upfront = EXCLUDED.reserved_1yr_no_upfront,
    spot_price_avg = EXCLUDED.spot_price_avg,
    last_updated = NOW();

-- Insert Assessment Configurations
INSERT INTO yt_assessment_config (
    tenant_id, underutilized_cpu_threshold, underutilized_memory_threshold,
    overutilized_cpu_threshold, overutilized_memory_threshold
)
VALUES
    ('tenant1', 15.0, 20.0, 85.0, 85.0),
    ('tenant2', 20.0, 25.0, 80.0, 80.0),
    ('tenant3', 25.0, 30.0, 75.0, 75.0),
    ('tenant4', 15.0, 20.0, 85.0, 85.0)
ON CONFLICT (tenant_id) DO UPDATE SET
    underutilized_cpu_threshold = EXCLUDED.underutilized_cpu_threshold,
    underutilized_memory_threshold = EXCLUDED.underutilized_memory_threshold,
    overutilized_cpu_threshold = EXCLUDED.overutilized_cpu_threshold,
    overutilized_memory_threshold = EXCLUDED.overutilized_memory_threshold;

-- Insert Resource Assessments
INSERT INTO yt_resource_assessments (
    tenant_id, resource_arn, assessment_timestamp,
    assessment_window_hours, cpu_utilization_avg,
    memory_utilization_avg, optimization_score
)
VALUES
    ('tenant1', 'arn:aws:ec2:us-east-1:123456789012:instance/i-0abc123def456789a',
     NOW(), 24, 45.5, 62.3, 0.85),
    ('tenant1', 'arn:aws:ec2:us-west-2:123456789012:instance/i-0bcd234efg567890b',
     NOW(), 24, 78.2, 85.4, 0.65),
    ('tenant2', 'arn:aws:ec2:us-east-1:234567890123:instance/i-0cde345fgh678901c',
     NOW(), 24, 25.4, 35.6, 0.92);

-- Insert Resource Identifiers
INSERT INTO yt_resource_identifiers (
    tenant_id, instance_id, private_ip, hostname,
    application_name, service_name
)
VALUES
    ('tenant1', 'i-0abc123def456789a', '10.0.1.10', 'prod-web-01',
     'e-commerce', 'web-service'),
    ('tenant1', 'i-0bcd234efg567890b', '10.0.2.20', 'prod-db-01',
     'e-commerce', 'database'),
    ('tenant2', 'i-0cde345fgh678901c', '172.16.1.10', 'dev-app-01',
     'mobile-app', 'api-service')
ON CONFLICT (instance_id) DO UPDATE SET
    private_ip = EXCLUDED.private_ip,
    hostname = EXCLUDED.hostname,
    application_name = EXCLUDED.application_name,
    service_name = EXCLUDED.service_name;