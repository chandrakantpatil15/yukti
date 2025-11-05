-- Netflix-scale FinOps data for realistic enterprise operations
-- Clear existing data
TRUNCATE TABLE resource_costs, optimization_recommendations, resource_metrics, resources, aws_pricings CASCADE;

-- Netflix-scale AWS pricing (all regions)
INSERT INTO aws_pricings (instance_type, region, os, price_per_hour, ri_1yr_no_upfront, ri_1yr_partial_upfront, spot_price_avg) VALUES
-- Compute instances (Netflix uses heavily)
('c5.large', 'us-east-1', 'Linux', 0.085, 0.0595, 0.051, 0.0255),
('c5.xlarge', 'us-east-1', 'Linux', 0.17, 0.119, 0.102, 0.051),
('c5.2xlarge', 'us-east-1', 'Linux', 0.34, 0.238, 0.204, 0.102),
('c5.4xlarge', 'us-east-1', 'Linux', 0.68, 0.476, 0.408, 0.204),
('c5.9xlarge', 'us-east-1', 'Linux', 1.53, 1.071, 0.918, 0.459),
('c5.18xlarge', 'us-east-1', 'Linux', 3.06, 2.142, 1.836, 0.918),

-- Memory optimized (for data processing)
('r5.large', 'us-east-1', 'Linux', 0.126, 0.0882, 0.0756, 0.0378),
('r5.xlarge', 'us-east-1', 'Linux', 0.252, 0.1764, 0.1512, 0.0756),
('r5.2xlarge', 'us-east-1', 'Linux', 0.504, 0.3528, 0.3024, 0.1512),
('r5.4xlarge', 'us-east-1', 'Linux', 1.008, 0.7056, 0.6048, 0.3024),
('r5.8xlarge', 'us-east-1', 'Linux', 2.016, 1.4112, 1.2096, 0.6048),

-- General purpose (mixed workloads)
('m5.large', 'us-east-1', 'Linux', 0.096, 0.0672, 0.0576, 0.0288),
('m5.xlarge', 'us-east-1', 'Linux', 0.192, 0.1344, 0.1152, 0.0576),
('m5.2xlarge', 'us-east-1', 'Linux', 0.384, 0.2688, 0.2304, 0.1152),
('m5.4xlarge', 'us-east-1', 'Linux', 0.768, 0.5376, 0.4608, 0.2304),
('m5.8xlarge', 'us-east-1', 'Linux', 1.536, 1.0752, 0.9216, 0.4608),

-- GPU instances (ML/encoding)
('p3.2xlarge', 'us-east-1', 'Linux', 3.06, 2.142, 1.836, 0.918),
('p3.8xlarge', 'us-east-1', 'Linux', 12.24, 8.568, 7.344, 3.672);

-- Netflix-scale resources (streaming infrastructure)
INSERT INTO resources (resource_id, resource_type, instance_type, region, status, project_id, environment, launch_time) VALUES
-- Content Delivery Network
('i-cdn-edge-001', 'ec2', 'c5.4xlarge', 'us-east-1', 'running', 1, 'prod', NOW() - INTERVAL '365 days'),
('i-cdn-edge-002', 'ec2', 'c5.4xlarge', 'us-east-1', 'running', 1, 'prod', NOW() - INTERVAL '350 days'),
('i-cdn-edge-003', 'ec2', 'c5.4xlarge', 'us-east-1', 'running', 1, 'prod', NOW() - INTERVAL '340 days'),
('i-cdn-edge-004', 'ec2', 'c5.4xlarge', 'us-east-1', 'running', 1, 'prod', NOW() - INTERVAL '330 days'),
('i-cdn-edge-005', 'ec2', 'c5.4xlarge', 'us-east-1', 'running', 1, 'prod', NOW() - INTERVAL '320 days'),

-- Video Encoding Pipeline
('i-encode-001', 'ec2', 'p3.8xlarge', 'us-east-1', 'running', 1, 'prod', NOW() - INTERVAL '200 days'),
('i-encode-002', 'ec2', 'p3.8xlarge', 'us-east-1', 'running', 1, 'prod', NOW() - INTERVAL '190 days'),
('i-encode-003', 'ec2', 'p3.2xlarge', 'us-east-1', 'running', 1, 'prod', NOW() - INTERVAL '180 days'),

-- Recommendation Engine
('i-ml-rec-001', 'ec2', 'r5.8xlarge', 'us-east-1', 'running', 2, 'prod', NOW() - INTERVAL '400 days'),
('i-ml-rec-002', 'ec2', 'r5.8xlarge', 'us-east-1', 'running', 2, 'prod', NOW() - INTERVAL '390 days'),
('i-ml-rec-003', 'ec2', 'r5.4xlarge', 'us-east-1', 'running', 2, 'prod', NOW() - INTERVAL '380 days'),

-- User Data Processing
('i-data-proc-001', 'ec2', 'm5.8xlarge', 'us-east-1', 'running', 2, 'prod', NOW() - INTERVAL '300 days'),
('i-data-proc-002', 'ec2', 'm5.8xlarge', 'us-east-1', 'running', 2, 'prod', NOW() - INTERVAL '290 days'),
('i-data-proc-003', 'ec2', 'm5.4xlarge', 'us-east-1', 'running', 2, 'prod', NOW() - INTERVAL '280 days'),

-- API Gateway Cluster
('i-api-gw-001', 'ec2', 'c5.2xlarge', 'us-east-1', 'running', 1, 'prod', NOW() - INTERVAL '250 days'),
('i-api-gw-002', 'ec2', 'c5.2xlarge', 'us-east-1', 'running', 1, 'prod', NOW() - INTERVAL '240 days'),
('i-api-gw-003', 'ec2', 'c5.2xlarge', 'us-east-1', 'running', 1, 'prod', NOW() - INTERVAL '230 days'),

-- Database Cluster
('i-db-primary-001', 'ec2', 'r5.4xlarge', 'us-east-1', 'running', 2, 'prod', NOW() - INTERVAL '500 days'),
('i-db-replica-001', 'ec2', 'r5.4xlarge', 'us-east-1', 'running', 2, 'prod', NOW() - INTERVAL '490 days'),
('i-db-replica-002', 'ec2', 'r5.2xlarge', 'us-east-1', 'running', 2, 'prod', NOW() - INTERVAL '480 days'),

-- Development/Testing (Optimization candidates)
('i-dev-test-001', 'ec2', 'm5.2xlarge', 'us-east-1', 'running', 1, 'dev', NOW() - INTERVAL '30 days'),
('i-dev-test-002', 'ec2', 'm5.xlarge', 'us-east-1', 'running', 1, 'dev', NOW() - INTERVAL '25 days'),
('i-staging-001', 'ec2', 'c5.xlarge', 'us-east-1', 'running', 2, 'staging', NOW() - INTERVAL '45 days'),
('i-staging-002', 'ec2', 'r5.xlarge', 'us-east-1', 'running', 2, 'staging', NOW() - INTERVAL '40 days'),

-- Idle/Overprovisioned (Cost optimization targets)
('i-legacy-001', 'ec2', 'm5.4xlarge', 'us-east-1', 'running', 3, 'prod', NOW() - INTERVAL '600 days'),
('i-legacy-002', 'ec2', 'c5.4xlarge', 'us-east-1', 'stopped', 3, 'prod', NOW() - INTERVAL '590 days'),
('i-unused-001', 'ec2', 'r5.2xlarge', 'us-east-1', 'stopped', 3, 'dev', NOW() - INTERVAL '60 days');

-- Generate Netflix-scale cost data (last 90 days)
INSERT INTO resource_costs (resource_id, date, cost_usd, usage_hours, data_source)
SELECT 
    r.id,
    date_series.date,
    CASE 
        WHEN r.status = 'running' THEN 
            CASE r.instance_type
                -- High-cost GPU instances
                WHEN 'p3.8xlarge' THEN 293.76 * CASE WHEN EXTRACT(DOW FROM date_series.date) IN (0,6) THEN 0.8 ELSE 1.0 END
                WHEN 'p3.2xlarge' THEN 73.44 * CASE WHEN EXTRACT(DOW FROM date_series.date) IN (0,6) THEN 0.8 ELSE 1.0 END
                -- Memory optimized
                WHEN 'r5.8xlarge' THEN 48.38 * CASE WHEN EXTRACT(DOW FROM date_series.date) IN (0,6) THEN 0.9 ELSE 1.0 END
                WHEN 'r5.4xlarge' THEN 24.19 * CASE WHEN EXTRACT(DOW FROM date_series.date) IN (0,6) THEN 0.9 ELSE 1.0 END
                WHEN 'r5.2xlarge' THEN 12.10 * CASE WHEN EXTRACT(DOW FROM date_series.date) IN (0,6) THEN 0.9 ELSE 1.0 END
                WHEN 'r5.xlarge' THEN 6.05 * CASE WHEN EXTRACT(DOW FROM date_series.date) IN (0,6) THEN 0.9 ELSE 1.0 END
                -- Compute optimized
                WHEN 'c5.4xlarge' THEN 16.32 * CASE WHEN EXTRACT(DOW FROM date_series.date) IN (0,6) THEN 0.7 ELSE 1.0 END
                WHEN 'c5.2xlarge' THEN 8.16 * CASE WHEN EXTRACT(DOW FROM date_series.date) IN (0,6) THEN 0.7 ELSE 1.0 END
                WHEN 'c5.xlarge' THEN 4.08 * CASE WHEN EXTRACT(DOW FROM date_series.date) IN (0,6) THEN 0.7 ELSE 1.0 END
                -- General purpose
                WHEN 'm5.8xlarge' THEN 36.86 * CASE WHEN EXTRACT(DOW FROM date_series.date) IN (0,6) THEN 0.8 ELSE 1.0 END
                WHEN 'm5.4xlarge' THEN 18.43 * CASE WHEN EXTRACT(DOW FROM date_series.date) IN (0,6) THEN 0.8 ELSE 1.0 END
                WHEN 'm5.2xlarge' THEN 9.22 * CASE WHEN EXTRACT(DOW FROM date_series.date) IN (0,6) THEN 0.8 ELSE 1.0 END
                WHEN 'm5.xlarge' THEN 4.61 * CASE WHEN EXTRACT(DOW FROM date_series.date) IN (0,6) THEN 0.8 ELSE 1.0 END
                ELSE 2.30
            END
        ELSE 0
    END as cost_usd,
    CASE WHEN r.status = 'running' THEN 24 ELSE 0 END as usage_hours,
    'aws_billing' as data_source
FROM resources r
CROSS JOIN (
    SELECT CURRENT_DATE - INTERVAL '1 day' * generate_series(0, 89) as date
) date_series
WHERE r.launch_time <= date_series.date;

-- Netflix-scale optimization recommendations
INSERT INTO optimization_recommendations (resource_id, recommendation_type, current_cost, optimized_cost, potential_savings, confidence, status) VALUES
-- Reserved Instance recommendations (production workloads)
((SELECT id FROM resources WHERE resource_id = 'i-cdn-edge-001'), 'reserved_instance', 16.32, 11.42, 4.90, 0.95, 'active'),
((SELECT id FROM resources WHERE resource_id = 'i-cdn-edge-002'), 'reserved_instance', 16.32, 11.42, 4.90, 0.95, 'active'),
((SELECT id FROM resources WHERE resource_id = 'i-ml-rec-001'), 'reserved_instance', 48.38, 33.87, 14.51, 0.92, 'active'),
((SELECT id FROM resources WHERE resource_id = 'i-ml-rec-002'), 'reserved_instance', 48.38, 33.87, 14.51, 0.92, 'active'),
((SELECT id FROM resources WHERE resource_id = 'i-data-proc-001'), 'reserved_instance', 36.86, 25.80, 11.06, 0.90, 'active'),
((SELECT id FROM resources WHERE resource_id = 'i-db-primary-001'), 'reserved_instance', 24.19, 16.93, 7.26, 0.88, 'active'),

-- Spot instance recommendations (dev/test)
((SELECT id FROM resources WHERE resource_id = 'i-dev-test-001'), 'spot_instance', 9.22, 2.76, 6.46, 0.85, 'active'),
((SELECT id FROM resources WHERE resource_id = 'i-dev-test-002'), 'spot_instance', 4.61, 1.38, 3.23, 0.85, 'active'),
((SELECT id FROM resources WHERE resource_id = 'i-staging-001'), 'spot_instance', 4.08, 1.22, 2.86, 0.80, 'active'),
((SELECT id FROM resources WHERE resource_id = 'i-staging-002'), 'spot_instance', 6.05, 1.81, 4.24, 0.80, 'active'),

-- Rightsizing recommendations (overprovisioned)
((SELECT id FROM resources WHERE resource_id = 'i-legacy-001'), 'rightsizing', 18.43, 9.22, 9.21, 0.75, 'active'),
((SELECT id FROM resources WHERE resource_id = 'i-encode-003'), 'rightsizing', 73.44, 36.72, 36.72, 0.70, 'active'),

-- Termination recommendations (idle resources)
((SELECT id FROM resources WHERE resource_id = 'i-legacy-002'), 'termination', 16.32, 0.00, 16.32, 0.95, 'active'),
((SELECT id FROM resources WHERE resource_id = 'i-unused-001'), 'termination', 12.10, 0.00, 12.10, 0.90, 'active');

-- Generate realistic utilization metrics for last 7 days
INSERT INTO resource_metrics (resource_id, timestamp, cpu_utilization, memory_utilization)
SELECT 
    r.id,
    timestamp_series.ts,
    CASE 
        -- CDN edge servers (high CPU, moderate memory)
        WHEN r.resource_id LIKE 'i-cdn-edge-%' THEN 75 + 15 * SIN(EXTRACT(HOUR FROM timestamp_series.ts) * PI() / 12) + 5 * RANDOM()
        -- Video encoding (very high GPU/CPU)
        WHEN r.resource_id LIKE 'i-encode-%' THEN 85 + 10 * RANDOM()
        -- ML recommendation (high memory, moderate CPU)
        WHEN r.resource_id LIKE 'i-ml-rec-%' THEN 60 + 20 * SIN(EXTRACT(HOUR FROM timestamp_series.ts) * PI() / 12) + 10 * RANDOM()
        -- Data processing (variable load)
        WHEN r.resource_id LIKE 'i-data-proc-%' THEN 70 + 25 * SIN(EXTRACT(HOUR FROM timestamp_series.ts) * PI() / 8) + 5 * RANDOM()
        -- API Gateway (steady high load)
        WHEN r.resource_id LIKE 'i-api-gw-%' THEN 65 + 20 * SIN(EXTRACT(HOUR FROM timestamp_series.ts) * PI() / 12) + 10 * RANDOM()
        -- Database (consistent high)
        WHEN r.resource_id LIKE 'i-db-%' THEN 80 + 10 * RANDOM()
        -- Dev/Test (low utilization - optimization target)
        WHEN r.environment IN ('dev', 'staging') THEN 15 + 20 * RANDOM()
        -- Legacy/unused (very low - termination candidates)
        WHEN r.resource_id LIKE 'i-legacy-%' OR r.resource_id LIKE 'i-unused-%' THEN 5 + 10 * RANDOM()
        ELSE 50 + 30 * RANDOM()
    END as cpu_utilization,
    CASE 
        -- Memory utilization patterns
        WHEN r.resource_id LIKE 'i-ml-rec-%' THEN 85 + 10 * RANDOM()
        WHEN r.resource_id LIKE 'i-db-%' THEN 75 + 15 * RANDOM()
        WHEN r.resource_id LIKE 'i-data-proc-%' THEN 70 + 20 * RANDOM()
        WHEN r.resource_id LIKE 'i-encode-%' THEN 60 + 15 * RANDOM()
        WHEN r.resource_id LIKE 'i-cdn-edge-%' THEN 45 + 20 * RANDOM()
        WHEN r.resource_id LIKE 'i-api-gw-%' THEN 55 + 25 * RANDOM()
        WHEN r.environment IN ('dev', 'staging') THEN 20 + 15 * RANDOM()
        WHEN r.resource_id LIKE 'i-legacy-%' OR r.resource_id LIKE 'i-unused-%' THEN 10 + 10 * RANDOM()
        ELSE 40 + 30 * RANDOM()
    END as memory_utilization
FROM resources r
CROSS JOIN (
    SELECT NOW() - INTERVAL '1 hour' * generate_series(0, 167) as ts -- Last 7 days hourly
) timestamp_series
WHERE r.status = 'running';