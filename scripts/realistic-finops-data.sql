-- Realistic FinOps data based on actual production patterns

-- Real AWS pricing (updated December 2024)
INSERT INTO aws_pricing (instance_type, region, os, price_per_hour, ri_1yr_no_upfront, ri_1yr_partial_upfront, spot_price_avg) VALUES
-- T3 instances (burstable)
('t3.nano', 'us-east-1', 'Linux', 0.0052, 0.0037, 0.0031, 0.0016),
('t3.micro', 'us-east-1', 'Linux', 0.0104, 0.0073, 0.0062, 0.0031),
('t3.small', 'us-east-1', 'Linux', 0.0208, 0.0146, 0.0125, 0.0062),
('t3.medium', 'us-east-1', 'Linux', 0.0416, 0.0292, 0.0250, 0.0125),
('t3.large', 'us-east-1', 'Linux', 0.0832, 0.0584, 0.0500, 0.0250),
('t3.xlarge', 'us-east-1', 'Linux', 0.1664, 0.1168, 0.1000, 0.0500),
-- M5 instances (general purpose)
('m5.large', 'us-east-1', 'Linux', 0.096, 0.0672, 0.0576, 0.0288),
('m5.xlarge', 'us-east-1', 'Linux', 0.192, 0.1344, 0.1152, 0.0576),
('m5.2xlarge', 'us-east-1', 'Linux', 0.384, 0.2688, 0.2304, 0.1152),
('m5.4xlarge', 'us-east-1', 'Linux', 0.768, 0.5376, 0.4608, 0.2304),
-- C5 instances (compute optimized)
('c5.large', 'us-east-1', 'Linux', 0.085, 0.0595, 0.0510, 0.0255),
('c5.xlarge', 'us-east-1', 'Linux', 0.17, 0.119, 0.102, 0.051),
('c5.2xlarge', 'us-east-1', 'Linux', 0.34, 0.238, 0.204, 0.102),
-- R5 instances (memory optimized)
('r5.large', 'us-east-1', 'Linux', 0.126, 0.0882, 0.0756, 0.0378),
('r5.xlarge', 'us-east-1', 'Linux', 0.252, 0.1764, 0.1512, 0.0756)
ON CONFLICT (instance_type, region, os) DO UPDATE SET
    price_per_hour = EXCLUDED.price_per_hour,
    ri_1yr_no_upfront = EXCLUDED.ri_1yr_no_upfront,
    ri_1yr_partial_upfront = EXCLUDED.ri_1yr_partial_upfront,
    spot_price_avg = EXCLUDED.spot_price_avg,
    updated_at = CURRENT_TIMESTAMP;

-- Realistic production resources with different optimization scenarios
INSERT INTO resources (resource_id, resource_type, instance_type, region, status, project_id, environment, launch_time) VALUES
-- PRODUCTION WORKLOADS (Good RI candidates)
('i-prod-web-01', 'ec2', 'm5.large', 'us-east-1', 'running', 1, 'prod', NOW() - INTERVAL '120 days'),
('i-prod-web-02', 'ec2', 'm5.large', 'us-east-1', 'running', 1, 'prod', NOW() - INTERVAL '90 days'),
('i-prod-api-01', 'ec2', 'c5.xlarge', 'us-east-1', 'running', 1, 'prod', NOW() - INTERVAL '180 days'),
('i-prod-db-01', 'ec2', 'r5.xlarge', 'us-east-1', 'running', 1, 'prod', NOW() - INTERVAL '365 days'),

-- OVERPROVISIONED RESOURCES (Rightsizing candidates)
('i-oversized-01', 'ec2', 'm5.2xlarge', 'us-east-1', 'running', 1, 'prod', NOW() - INTERVAL '45 days'),
('i-oversized-02', 'ec2', 'c5.2xlarge', 'us-east-1', 'running', 2, 'staging', NOW() - INTERVAL '30 days'),

-- DEV/TEST WORKLOADS (Spot candidates)
('i-dev-test-01', 'ec2', 't3.medium', 'us-east-1', 'running', 2, 'dev', NOW() - INTERVAL '10 days'),
('i-dev-test-02', 'ec2', 't3.large', 'us-east-1', 'running', 2, 'test', NOW() - INTERVAL '15 days'),
('i-dev-batch-01', 'ec2', 'c5.xlarge', 'us-east-1', 'running', 2, 'dev', NOW() - INTERVAL '5 days'),

-- IDLE/STOPPED RESOURCES (Termination candidates)
('i-idle-01', 'ec2', 't3.medium', 'us-east-1', 'stopped', 2, 'dev', NOW() - INTERVAL '20 days'),
('i-idle-02', 'ec2', 'm5.large', 'us-east-1', 'stopped', 3, 'staging', NOW() - INTERVAL '7 days'),

-- SEASONAL WORKLOADS
('i-seasonal-01', 'ec2', 'm5.xlarge', 'us-east-1', 'running', 2, 'prod', NOW() - INTERVAL '60 days');

-- Realistic cost data with patterns
INSERT INTO resource_costs (resource_id, date, cost_usd, usage_hours, data_source)
SELECT 
    r.id,
    date_series.date,
    CASE 
        WHEN r.status = 'running' THEN 
            ap.price_per_hour * 24 * 
            CASE 
                -- Weekend usage reduction
                WHEN EXTRACT(DOW FROM date_series.date) IN (0, 6) THEN 0.7
                -- Holiday reduction (simulate Dec 25, Jan 1)
                WHEN date_series.date IN ('2024-12-25', '2025-01-01') THEN 0.3
                -- Normal weekday
                ELSE 1.0
            END
        ELSE 0
    END,
    CASE 
        WHEN r.status = 'running' THEN 24
        ELSE 0
    END,
    'billing'
FROM resources r
CROSS JOIN (
    SELECT CURRENT_DATE - INTERVAL '1 day' * generate_series(0, 89) as date
) date_series
JOIN aws_pricing ap ON r.instance_type = ap.instance_type AND r.region = ap.region
WHERE r.launch_time <= date_series.date;

-- Realistic utilization patterns
INSERT INTO resource_metrics (resource_id, timestamp, cpu_utilization, memory_utilization)
SELECT 
    r.id,
    timestamp_series.ts,
    CASE r.resource_id
        -- Production web servers (steady load with peaks)
        WHEN 'i-prod-web-01' THEN 55 + 15 * SIN(EXTRACT(HOUR FROM timestamp_series.ts) * PI() / 12)
        WHEN 'i-prod-web-02' THEN 50 + 20 * SIN(EXTRACT(HOUR FROM timestamp_series.ts) * PI() / 12)
        
        -- API servers (higher utilization)
        WHEN 'i-prod-api-01' THEN 70 + 10 * SIN(EXTRACT(HOUR FROM timestamp_series.ts) * PI() / 12)
        
        -- Database (consistent high utilization)
        WHEN 'i-prod-db-01' THEN 75 + 5 * RANDOM()
        
        -- Overprovisioned (low utilization - rightsizing candidates)
        WHEN 'i-oversized-01' THEN 15 + 10 * RANDOM()
        WHEN 'i-oversized-02' THEN 20 + 8 * RANDOM()
        
        -- Dev/Test (variable, low utilization)
        WHEN 'i-dev-test-01' THEN 10 + 15 * RANDOM()
        WHEN 'i-dev-test-02' THEN 12 + 18 * RANDOM()
        WHEN 'i-dev-batch-01' THEN 
            CASE 
                WHEN EXTRACT(HOUR FROM timestamp_series.ts) BETWEEN 2 AND 6 THEN 85 + 10 * RANDOM()
                ELSE 5 + 10 * RANDOM()
            END
        
        -- Seasonal workload
        WHEN 'i-seasonal-01' THEN 
            CASE 
                WHEN EXTRACT(MONTH FROM timestamp_series.ts) IN (11, 12) THEN 80 + 15 * RANDOM()
                ELSE 25 + 20 * RANDOM()
            END
        
        ELSE 0
    END,
    CASE r.resource_id
        -- Memory utilization typically correlates with CPU but with different patterns
        WHEN 'i-prod-web-01' THEN 45 + 10 * SIN(EXTRACT(HOUR FROM timestamp_series.ts) * PI() / 12)
        WHEN 'i-prod-web-02' THEN 40 + 15 * SIN(EXTRACT(HOUR FROM timestamp_series.ts) * PI() / 12)
        WHEN 'i-prod-api-01' THEN 60 + 8 * SIN(EXTRACT(HOUR FROM timestamp_series.ts) * PI() / 12)
        WHEN 'i-prod-db-01' THEN 85 + 5 * RANDOM() -- Database memory intensive
        WHEN 'i-oversized-01' THEN 25 + 10 * RANDOM()
        WHEN 'i-oversized-02' THEN 30 + 8 * RANDOM()
        WHEN 'i-dev-test-01' THEN 20 + 15 * RANDOM()
        WHEN 'i-dev-test-02' THEN 22 + 18 * RANDOM()
        WHEN 'i-dev-batch-01' THEN 
            CASE 
                WHEN EXTRACT(HOUR FROM timestamp_series.ts) BETWEEN 2 AND 6 THEN 70 + 15 * RANDOM()
                ELSE 15 + 10 * RANDOM()
            END
        WHEN 'i-seasonal-01' THEN 
            CASE 
                WHEN EXTRACT(MONTH FROM timestamp_series.ts) IN (11, 12) THEN 70 + 20 * RANDOM()
                ELSE 35 + 15 * RANDOM()
            END
        ELSE 0
    END
FROM resources r
CROSS JOIN (
    SELECT NOW() - INTERVAL '1 hour' * generate_series(0, 167) as ts -- Last 7 days hourly
) timestamp_series
WHERE r.status = 'running';

-- Realistic resource tags
INSERT INTO resource_tags (resource_id, key, value)
SELECT r.id, 'Environment', r.environment FROM resources r
UNION ALL
SELECT r.id, 'Project', p.code FROM resources r JOIN projects p ON r.project_id = p.id
UNION ALL
SELECT r.id, 'Owner', p.owner_email FROM resources r JOIN projects p ON r.project_id = p.id
UNION ALL
SELECT r.id, 'CostCenter', cc.code FROM resources r JOIN projects p ON r.project_id = p.id JOIN cost_centers cc ON p.cost_center_id = cc.id
UNION ALL
-- Application-specific tags
SELECT r.id, 'Application', 
    CASE r.resource_id
        WHEN 'i-prod-web-01' THEN 'frontend'
        WHEN 'i-prod-web-02' THEN 'frontend'
        WHEN 'i-prod-api-01' THEN 'api-gateway'
        WHEN 'i-prod-db-01' THEN 'database'
        WHEN 'i-oversized-01' THEN 'legacy-app'
        WHEN 'i-oversized-02' THEN 'staging-env'
        WHEN 'i-dev-test-01' THEN 'development'
        WHEN 'i-dev-test-02' THEN 'testing'
        WHEN 'i-dev-batch-01' THEN 'batch-processing'
        WHEN 'i-seasonal-01' THEN 'holiday-campaign'
        ELSE 'unknown'
    END
FROM resources r
UNION ALL
-- Business criticality
SELECT r.id, 'Criticality',
    CASE r.environment
        WHEN 'prod' THEN 'high'
        WHEN 'staging' THEN 'medium'
        ELSE 'low'
    END
FROM resources r;