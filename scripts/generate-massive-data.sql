-- Generate 2GB+ Netflix-scale FinOps data for performance testing
-- This will create millions of records across 2+ years

-- Clear existing data
TRUNCATE TABLE resource_costs, optimization_recommendations, resource_metrics, resources, aws_pricings CASCADE;

-- Comprehensive AWS pricing (all instance families)
INSERT INTO aws_pricings (instance_type, region, os, price_per_hour, ri_1yr_no_upfront, ri_1yr_partial_upfront, spot_price_avg) VALUES
-- General Purpose (M5 family)
('m5.large', 'us-east-1', 'Linux', 0.096, 0.0672, 0.0576, 0.0288),
('m5.xlarge', 'us-east-1', 'Linux', 0.192, 0.1344, 0.1152, 0.0576),
('m5.2xlarge', 'us-east-1', 'Linux', 0.384, 0.2688, 0.2304, 0.1152),
('m5.4xlarge', 'us-east-1', 'Linux', 0.768, 0.5376, 0.4608, 0.2304),
('m5.8xlarge', 'us-east-1', 'Linux', 1.536, 1.0752, 0.9216, 0.4608),
('m5.12xlarge', 'us-east-1', 'Linux', 2.304, 1.6128, 1.3824, 0.6912),
('m5.16xlarge', 'us-east-1', 'Linux', 3.072, 2.1504, 1.8432, 0.9216),
('m5.24xlarge', 'us-east-1', 'Linux', 4.608, 3.2256, 2.7648, 1.3824),

-- Compute Optimized (C5 family)
('c5.large', 'us-east-1', 'Linux', 0.085, 0.0595, 0.051, 0.0255),
('c5.xlarge', 'us-east-1', 'Linux', 0.17, 0.119, 0.102, 0.051),
('c5.2xlarge', 'us-east-1', 'Linux', 0.34, 0.238, 0.204, 0.102),
('c5.4xlarge', 'us-east-1', 'Linux', 0.68, 0.476, 0.408, 0.204),
('c5.9xlarge', 'us-east-1', 'Linux', 1.53, 1.071, 0.918, 0.459),
('c5.12xlarge', 'us-east-1', 'Linux', 2.04, 1.428, 1.224, 0.612),
('c5.18xlarge', 'us-east-1', 'Linux', 3.06, 2.142, 1.836, 0.918),
('c5.24xlarge', 'us-east-1', 'Linux', 4.08, 2.856, 2.448, 1.224),

-- Memory Optimized (R5 family)
('r5.large', 'us-east-1', 'Linux', 0.126, 0.0882, 0.0756, 0.0378),
('r5.xlarge', 'us-east-1', 'Linux', 0.252, 0.1764, 0.1512, 0.0756),
('r5.2xlarge', 'us-east-1', 'Linux', 0.504, 0.3528, 0.3024, 0.1512),
('r5.4xlarge', 'us-east-1', 'Linux', 1.008, 0.7056, 0.6048, 0.3024),
('r5.8xlarge', 'us-east-1', 'Linux', 2.016, 1.4112, 1.2096, 0.6048),
('r5.12xlarge', 'us-east-1', 'Linux', 3.024, 2.1168, 1.8144, 0.9072),
('r5.16xlarge', 'us-east-1', 'Linux', 4.032, 2.8224, 2.4192, 1.2096),
('r5.24xlarge', 'us-east-1', 'Linux', 6.048, 4.2336, 3.6288, 1.8144),

-- GPU Instances (P3 family)
('p3.2xlarge', 'us-east-1', 'Linux', 3.06, 2.142, 1.836, 0.918),
('p3.8xlarge', 'us-east-1', 'Linux', 12.24, 8.568, 7.344, 3.672),
('p3.16xlarge', 'us-east-1', 'Linux', 24.48, 17.136, 14.688, 7.344),

-- Storage Optimized (I3 family)
('i3.large', 'us-east-1', 'Linux', 0.156, 0.1092, 0.0936, 0.0468),
('i3.xlarge', 'us-east-1', 'Linux', 0.312, 0.2184, 0.1872, 0.0936),
('i3.2xlarge', 'us-east-1', 'Linux', 0.624, 0.4368, 0.3744, 0.1872),
('i3.4xlarge', 'us-east-1', 'Linux', 1.248, 0.8736, 0.7488, 0.3744),
('i3.8xlarge', 'us-east-1', 'Linux', 2.496, 1.7472, 1.4976, 0.7488),
('i3.16xlarge', 'us-east-1', 'Linux', 4.992, 3.4944, 2.9952, 1.4976);

-- Generate 10,000 Netflix-scale resources across multiple services
INSERT INTO resources (resource_id, resource_type, instance_type, region, status, project_id, environment, launch_time)
SELECT 
    'i-' || service_type || '-' || LPAD(resource_num::text, 6, '0'),
    'ec2',
    CASE 
        WHEN service_type = 'cdn' THEN (ARRAY['c5.4xlarge', 'c5.9xlarge', 'c5.12xlarge'])[1 + (resource_num % 3)]
        WHEN service_type = 'encode' THEN (ARRAY['p3.8xlarge', 'p3.16xlarge', 'c5.24xlarge'])[1 + (resource_num % 3)]
        WHEN service_type = 'ml' THEN (ARRAY['r5.8xlarge', 'r5.12xlarge', 'r5.16xlarge'])[1 + (resource_num % 3)]
        WHEN service_type = 'data' THEN (ARRAY['m5.8xlarge', 'm5.12xlarge', 'm5.16xlarge'])[1 + (resource_num % 3)]
        WHEN service_type = 'api' THEN (ARRAY['c5.2xlarge', 'c5.4xlarge', 'c5.9xlarge'])[1 + (resource_num % 3)]
        WHEN service_type = 'db' THEN (ARRAY['r5.4xlarge', 'r5.8xlarge', 'r5.12xlarge'])[1 + (resource_num % 3)]
        WHEN service_type = 'cache' THEN (ARRAY['r5.2xlarge', 'r5.4xlarge', 'r5.8xlarge'])[1 + (resource_num % 3)]
        WHEN service_type = 'search' THEN (ARRAY['i3.4xlarge', 'i3.8xlarge', 'i3.16xlarge'])[1 + (resource_num % 3)]
        WHEN service_type = 'stream' THEN (ARRAY['c5.12xlarge', 'c5.18xlarge', 'c5.24xlarge'])[1 + (resource_num % 3)]
        ELSE 'm5.large'
    END,
    'us-east-1',
    CASE WHEN RANDOM() < 0.95 THEN 'running' ELSE 'stopped' END,
    CASE 
        WHEN service_type IN ('cdn', 'encode', 'stream') THEN 1  -- Content delivery
        WHEN service_type IN ('ml', 'data', 'search') THEN 2     -- Data & ML
        ELSE 3  -- Infrastructure
    END,
    CASE 
        WHEN resource_num % 10 < 7 THEN 'prod'
        WHEN resource_num % 10 < 9 THEN 'staging'
        ELSE 'dev'
    END,
    NOW() - INTERVAL '1 day' * (RANDOM() * 730)::int  -- Random launch time in last 2 years
FROM 
    generate_series(1, 1000) as resource_num,
    unnest(ARRAY['cdn', 'encode', 'ml', 'data', 'api', 'db', 'cache', 'search', 'stream']) as service_type;

-- Generate massive cost data (2+ years of daily costs = ~2.7M records)
INSERT INTO resource_costs (resource_id, date, cost_usd, usage_hours, data_source)
SELECT 
    r.id,
    date_series.date,
    CASE 
        WHEN r.status = 'running' THEN 
            CASE r.instance_type
                -- GPU instances (highest cost)
                WHEN 'p3.16xlarge' THEN 587.52 * daily_multiplier
                WHEN 'p3.8xlarge' THEN 293.76 * daily_multiplier
                WHEN 'p3.2xlarge' THEN 73.44 * daily_multiplier
                -- Large compute instances
                WHEN 'c5.24xlarge' THEN 97.92 * daily_multiplier
                WHEN 'c5.18xlarge' THEN 73.44 * daily_multiplier
                WHEN 'c5.12xlarge' THEN 48.96 * daily_multiplier
                WHEN 'c5.9xlarge' THEN 36.72 * daily_multiplier
                -- Large memory instances
                WHEN 'r5.24xlarge' THEN 145.15 * daily_multiplier
                WHEN 'r5.16xlarge' THEN 96.77 * daily_multiplier
                WHEN 'r5.12xlarge' THEN 72.58 * daily_multiplier
                WHEN 'r5.8xlarge' THEN 48.38 * daily_multiplier
                -- Large general purpose
                WHEN 'm5.24xlarge' THEN 110.59 * daily_multiplier
                WHEN 'm5.16xlarge' THEN 73.73 * daily_multiplier
                WHEN 'm5.12xlarge' THEN 55.30 * daily_multiplier
                WHEN 'm5.8xlarge' THEN 36.86 * daily_multiplier
                -- Storage optimized
                WHEN 'i3.16xlarge' THEN 119.81 * daily_multiplier
                WHEN 'i3.8xlarge' THEN 59.90 * daily_multiplier
                WHEN 'i3.4xlarge' THEN 29.95 * daily_multiplier
                -- Medium instances
                WHEN 'c5.4xlarge' THEN 16.32 * daily_multiplier
                WHEN 'c5.2xlarge' THEN 8.16 * daily_multiplier
                WHEN 'r5.4xlarge' THEN 24.19 * daily_multiplier
                WHEN 'r5.2xlarge' THEN 12.10 * daily_multiplier
                WHEN 'm5.4xlarge' THEN 18.43 * daily_multiplier
                WHEN 'm5.2xlarge' THEN 9.22 * daily_multiplier
                ELSE 4.61 * daily_multiplier
            END
        ELSE 0
    END as cost_usd,
    CASE WHEN r.status = 'running' THEN 24 ELSE 0 END as usage_hours,
    'aws_billing' as data_source
FROM resources r
CROSS JOIN (
    SELECT 
        CURRENT_DATE - INTERVAL '1 day' * generate_series(0, 729) as date,
        CASE 
            WHEN EXTRACT(DOW FROM CURRENT_DATE - INTERVAL '1 day' * generate_series(0, 729)) IN (0,6) THEN 0.7  -- Weekend reduction
            WHEN EXTRACT(MONTH FROM CURRENT_DATE - INTERVAL '1 day' * generate_series(0, 729)) = 12 THEN 1.3    -- Holiday spike
            WHEN EXTRACT(MONTH FROM CURRENT_DATE - INTERVAL '1 day' * generate_series(0, 729)) IN (6,7,8) THEN 1.2  -- Summer peak
            ELSE 1.0
        END as daily_multiplier
) date_series
WHERE r.launch_time::date <= date_series.date;

-- Generate massive metrics data (hourly for last 30 days = ~7.2M records)
INSERT INTO resource_metrics (resource_id, timestamp, cpu_utilization, memory_utilization)
SELECT 
    r.id,
    timestamp_series.ts,
    CASE 
        WHEN r.resource_id LIKE '%-cdn-%' THEN 
            GREATEST(0, LEAST(100, 75 + 15 * SIN(EXTRACT(HOUR FROM timestamp_series.ts) * PI() / 12) + 10 * (RANDOM() - 0.5)))
        WHEN r.resource_id LIKE '%-encode-%' THEN 
            GREATEST(0, LEAST(100, 85 + 10 * (RANDOM() - 0.5)))
        WHEN r.resource_id LIKE '%-ml-%' THEN 
            GREATEST(0, LEAST(100, 70 + 20 * SIN(EXTRACT(HOUR FROM timestamp_series.ts) * PI() / 8) + 15 * (RANDOM() - 0.5)))
        WHEN r.resource_id LIKE '%-data-%' THEN 
            GREATEST(0, LEAST(100, 65 + 25 * SIN(EXTRACT(HOUR FROM timestamp_series.ts) * PI() / 6) + 10 * (RANDOM() - 0.5)))
        WHEN r.resource_id LIKE '%-api-%' THEN 
            GREATEST(0, LEAST(100, 60 + 30 * SIN(EXTRACT(HOUR FROM timestamp_series.ts) * PI() / 12) + 15 * (RANDOM() - 0.5)))
        WHEN r.resource_id LIKE '%-db-%' THEN 
            GREATEST(0, LEAST(100, 80 + 15 * (RANDOM() - 0.5)))
        WHEN r.resource_id LIKE '%-cache-%' THEN 
            GREATEST(0, LEAST(100, 55 + 35 * SIN(EXTRACT(HOUR FROM timestamp_series.ts) * PI() / 12) + 20 * (RANDOM() - 0.5)))
        WHEN r.resource_id LIKE '%-search-%' THEN 
            GREATEST(0, LEAST(100, 50 + 40 * SIN(EXTRACT(HOUR FROM timestamp_series.ts) * PI() / 12) + 25 * (RANDOM() - 0.5)))
        WHEN r.resource_id LIKE '%-stream-%' THEN 
            GREATEST(0, LEAST(100, 85 + 10 * SIN(EXTRACT(HOUR FROM timestamp_series.ts) * PI() / 12) + 5 * (RANDOM() - 0.5)))
        WHEN r.environment = 'dev' THEN 
            GREATEST(0, LEAST(100, 20 + 30 * (RANDOM() - 0.5)))
        WHEN r.environment = 'staging' THEN 
            GREATEST(0, LEAST(100, 35 + 25 * (RANDOM() - 0.5)))
        ELSE 
            GREATEST(0, LEAST(100, 50 + 40 * (RANDOM() - 0.5)))
    END as cpu_utilization,
    CASE 
        WHEN r.resource_id LIKE '%-ml-%' OR r.resource_id LIKE '%-db-%' THEN 
            GREATEST(0, LEAST(100, 80 + 15 * (RANDOM() - 0.5)))
        WHEN r.resource_id LIKE '%-cache-%' THEN 
            GREATEST(0, LEAST(100, 70 + 20 * (RANDOM() - 0.5)))
        WHEN r.resource_id LIKE '%-data-%' THEN 
            GREATEST(0, LEAST(100, 65 + 25 * (RANDOM() - 0.5)))
        WHEN r.resource_id LIKE '%-search-%' THEN 
            GREATEST(0, LEAST(100, 60 + 30 * (RANDOM() - 0.5)))
        ELSE 
            GREATEST(0, LEAST(100, 45 + 35 * (RANDOM() - 0.5)))
    END as memory_utilization
FROM resources r
CROSS JOIN (
    SELECT NOW() - INTERVAL '1 hour' * generate_series(0, 719) as ts  -- Last 30 days hourly
) timestamp_series
WHERE r.status = 'running'
AND r.launch_time <= timestamp_series.ts;

-- Generate optimization recommendations (thousands of them)
INSERT INTO optimization_recommendations (resource_id, recommendation_type, current_cost, optimized_cost, potential_savings, confidence, status)
SELECT 
    r.id,
    CASE 
        WHEN r.environment = 'prod' AND EXTRACT(DAYS FROM NOW() - r.launch_time) > 90 THEN 'reserved_instance'
        WHEN r.environment IN ('dev', 'staging') THEN 'spot_instance'
        WHEN r.resource_id LIKE '%-encode-%' AND RANDOM() < 0.3 THEN 'rightsizing'
        WHEN r.status = 'stopped' THEN 'termination'
        ELSE 'reserved_instance'
    END as recommendation_type,
    CASE r.instance_type
        WHEN 'p3.16xlarge' THEN 587.52
        WHEN 'p3.8xlarge' THEN 293.76
        WHEN 'c5.24xlarge' THEN 97.92
        WHEN 'r5.24xlarge' THEN 145.15
        WHEN 'm5.24xlarge' THEN 110.59
        ELSE 50.00
    END as current_cost,
    CASE 
        WHEN r.environment = 'prod' AND EXTRACT(DAYS FROM NOW() - r.launch_time) > 90 THEN 
            CASE r.instance_type
                WHEN 'p3.16xlarge' THEN 411.26  -- 30% RI savings
                WHEN 'p3.8xlarge' THEN 205.63
                WHEN 'c5.24xlarge' THEN 68.54
                ELSE 35.00
            END
        WHEN r.environment IN ('dev', 'staging') THEN 
            CASE r.instance_type
                WHEN 'p3.16xlarge' THEN 176.26  -- 70% spot savings
                WHEN 'p3.8xlarge' THEN 88.13
                ELSE 15.00
            END
        WHEN r.status = 'stopped' THEN 0.00
        ELSE 40.00
    END as optimized_cost,
    CASE 
        WHEN r.environment = 'prod' AND EXTRACT(DAYS FROM NOW() - r.launch_time) > 90 THEN 
            CASE r.instance_type
                WHEN 'p3.16xlarge' THEN 176.26
                WHEN 'p3.8xlarge' THEN 88.13
                WHEN 'c5.24xlarge' THEN 29.38
                ELSE 15.00
            END
        WHEN r.environment IN ('dev', 'staging') THEN 
            CASE r.instance_type
                WHEN 'p3.16xlarge' THEN 411.26
                WHEN 'p3.8xlarge' THEN 205.63
                ELSE 35.00
            END
        WHEN r.status = 'stopped' THEN 
            CASE r.instance_type
                WHEN 'p3.16xlarge' THEN 587.52
                WHEN 'p3.8xlarge' THEN 293.76
                ELSE 50.00
            END
        ELSE 10.00
    END as potential_savings,
    CASE 
        WHEN r.environment = 'prod' AND EXTRACT(DAYS FROM NOW() - r.launch_time) > 365 THEN 0.95
        WHEN r.environment = 'prod' AND EXTRACT(DAYS FROM NOW() - r.launch_time) > 90 THEN 0.85
        WHEN r.environment IN ('dev', 'staging') THEN 0.75
        WHEN r.status = 'stopped' THEN 0.90
        ELSE 0.70
    END as confidence,
    'active' as status
FROM resources r
WHERE RANDOM() < 0.4;  -- 40% of resources get recommendations

-- Create indexes for performance
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_resource_costs_date ON resource_costs(date);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_resource_costs_resource_date ON resource_costs(resource_id, date);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_resource_metrics_timestamp ON resource_metrics(timestamp);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_resource_metrics_resource_timestamp ON resource_metrics(resource_id, timestamp);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_resources_environment ON resources(environment);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_resources_instance_type ON resources(instance_type);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_resources_status ON resources(status);

-- Show data size
SELECT 
    schemaname,
    tablename,
    attname,
    n_distinct,
    correlation,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
FROM pg_stats 
WHERE schemaname = 'public' 
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;