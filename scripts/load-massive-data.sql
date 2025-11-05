-- Load massive Netflix-scale data (simplified for current schema)
TRUNCATE TABLE resource_costs, optimization_recommendations, resource_metrics, resources, aws_pricings CASCADE;

-- Load AWS pricing data
INSERT INTO aws_pricings (instance_type, region, os, price_per_hour, ri_1yr_no_upfront, ri_1yr_partial_upfront, spot_price_avg) VALUES
('c5.large', 'us-east-1', 'Linux', 0.085, 0.0595, 0.051, 0.0255),
('c5.xlarge', 'us-east-1', 'Linux', 0.17, 0.119, 0.102, 0.051),
('c5.2xlarge', 'us-east-1', 'Linux', 0.34, 0.238, 0.204, 0.102),
('c5.4xlarge', 'us-east-1', 'Linux', 0.68, 0.476, 0.408, 0.204),
('c5.9xlarge', 'us-east-1', 'Linux', 1.53, 1.071, 0.918, 0.459),
('c5.18xlarge', 'us-east-1', 'Linux', 3.06, 2.142, 1.836, 0.918),
('r5.large', 'us-east-1', 'Linux', 0.126, 0.0882, 0.0756, 0.0378),
('r5.xlarge', 'us-east-1', 'Linux', 0.252, 0.1764, 0.1512, 0.0756),
('r5.2xlarge', 'us-east-1', 'Linux', 0.504, 0.3528, 0.3024, 0.1512),
('r5.4xlarge', 'us-east-1', 'Linux', 1.008, 0.7056, 0.6048, 0.3024),
('r5.8xlarge', 'us-east-1', 'Linux', 2.016, 1.4112, 1.2096, 0.6048),
('m5.large', 'us-east-1', 'Linux', 0.096, 0.0672, 0.0576, 0.0288),
('m5.xlarge', 'us-east-1', 'Linux', 0.192, 0.1344, 0.1152, 0.0576),
('m5.2xlarge', 'us-east-1', 'Linux', 0.384, 0.2688, 0.2304, 0.1152),
('m5.4xlarge', 'us-east-1', 'Linux', 0.768, 0.5376, 0.4608, 0.2304),
('m5.8xlarge', 'us-east-1', 'Linux', 1.536, 1.0752, 0.9216, 0.4608),
('p3.2xlarge', 'us-east-1', 'Linux', 3.06, 2.142, 1.836, 0.918),
('p3.8xlarge', 'us-east-1', 'Linux', 12.24, 8.568, 7.344, 3.672);

-- Generate 5000 Netflix-scale resources
INSERT INTO resources (resource_id, resource_type, instance_type, region, status, project_id, environment, launch_time)
SELECT 
    'i-' || service_type || '-' || LPAD(resource_num::text, 6, '0'),
    'ec2',
    CASE 
        WHEN service_type = 'cdn' THEN (ARRAY['c5.4xlarge', 'c5.9xlarge'])[1 + (resource_num % 2)]
        WHEN service_type = 'encode' THEN (ARRAY['p3.8xlarge', 'p3.2xlarge'])[1 + (resource_num % 2)]
        WHEN service_type = 'ml' THEN (ARRAY['r5.8xlarge', 'r5.4xlarge'])[1 + (resource_num % 2)]
        WHEN service_type = 'data' THEN (ARRAY['m5.8xlarge', 'm5.4xlarge'])[1 + (resource_num % 2)]
        WHEN service_type = 'api' THEN (ARRAY['c5.2xlarge', 'c5.4xlarge'])[1 + (resource_num % 2)]
        ELSE 'm5.large'
    END,
    'us-east-1',
    CASE WHEN RANDOM() < 0.95 THEN 'running' ELSE 'stopped' END,
    CASE 
        WHEN service_type IN ('cdn', 'encode') THEN 1
        WHEN service_type IN ('ml', 'data') THEN 2
        ELSE 3
    END,
    CASE 
        WHEN resource_num % 10 < 7 THEN 'prod'
        WHEN resource_num % 10 < 9 THEN 'staging'
        ELSE 'dev'
    END,
    NOW() - INTERVAL '1 day' * (RANDOM() * 365)::int
FROM 
    generate_series(1, 1000) as resource_num,
    unnest(ARRAY['cdn', 'encode', 'ml', 'data', 'api']) as service_type;

-- Generate cost data for last 365 days (1.8M+ records)
INSERT INTO resource_costs (resource_id, date, cost_usd, usage_hours, data_source)
SELECT 
    r.id,
    CURRENT_DATE - INTERVAL '1 day' * day_offset,
    CASE 
        WHEN r.status = 'running' THEN 
            CASE r.instance_type
                WHEN 'p3.8xlarge' THEN 293.76
                WHEN 'p3.2xlarge' THEN 73.44
                WHEN 'r5.8xlarge' THEN 48.38
                WHEN 'r5.4xlarge' THEN 24.19
                WHEN 'm5.8xlarge' THEN 36.86
                WHEN 'm5.4xlarge' THEN 18.43
                WHEN 'c5.9xlarge' THEN 36.72
                WHEN 'c5.4xlarge' THEN 16.32
                WHEN 'c5.2xlarge' THEN 8.16
                ELSE 4.61
            END * 
            CASE 
                WHEN EXTRACT(DOW FROM CURRENT_DATE - INTERVAL '1 day' * day_offset) IN (0,6) THEN 0.7
                ELSE 1.0
            END
        ELSE 0
    END as cost_usd,
    CASE WHEN r.status = 'running' THEN 24 ELSE 0 END as usage_hours,
    'aws_billing' as data_source
FROM resources r
CROSS JOIN generate_series(0, 364) as day_offset
WHERE r.launch_time::date <= (CURRENT_DATE - INTERVAL '1 day' * day_offset);

-- Generate metrics data for last 30 days (3.6M+ records)
INSERT INTO resource_metrics (resource_id, timestamp, cpu_utilization, memory_utilization)
SELECT 
    r.id,
    NOW() - INTERVAL '1 hour' * hour_offset,
    CASE 
        WHEN r.resource_id LIKE '%-cdn-%' THEN 75 + 15 * RANDOM()
        WHEN r.resource_id LIKE '%-encode-%' THEN 85 + 10 * RANDOM()
        WHEN r.resource_id LIKE '%-ml-%' THEN 70 + 20 * RANDOM()
        WHEN r.resource_id LIKE '%-data-%' THEN 65 + 25 * RANDOM()
        WHEN r.resource_id LIKE '%-api-%' THEN 60 + 30 * RANDOM()
        WHEN r.environment = 'dev' THEN 20 + 30 * RANDOM()
        ELSE 50 + 40 * RANDOM()
    END as cpu_utilization,
    CASE 
        WHEN r.resource_id LIKE '%-ml-%' THEN 80 + 15 * RANDOM()
        WHEN r.resource_id LIKE '%-data-%' THEN 65 + 25 * RANDOM()
        ELSE 45 + 35 * RANDOM()
    END as memory_utilization
FROM resources r
CROSS JOIN generate_series(0, 719) as hour_offset  -- 30 days * 24 hours
WHERE r.status = 'running'
AND r.launch_time <= (NOW() - INTERVAL '1 hour' * hour_offset);

-- Generate optimization recommendations
INSERT INTO optimization_recommendations (resource_id, recommendation_type, current_cost, optimized_cost, potential_savings, confidence, status)
SELECT 
    r.id,
    CASE 
        WHEN r.environment = 'prod' THEN 'reserved_instance'
        WHEN r.environment IN ('dev', 'staging') THEN 'spot_instance'
        WHEN r.status = 'stopped' THEN 'termination'
        ELSE 'rightsizing'
    END as recommendation_type,
    CASE r.instance_type
        WHEN 'p3.8xlarge' THEN 293.76
        WHEN 'p3.2xlarge' THEN 73.44
        WHEN 'r5.8xlarge' THEN 48.38
        WHEN 'r5.4xlarge' THEN 24.19
        WHEN 'm5.8xlarge' THEN 36.86
        WHEN 'm5.4xlarge' THEN 18.43
        WHEN 'c5.9xlarge' THEN 36.72
        WHEN 'c5.4xlarge' THEN 16.32
        WHEN 'c5.2xlarge' THEN 8.16
        ELSE 4.61
    END as current_cost,
    CASE 
        WHEN r.environment = 'prod' THEN 
            CASE r.instance_type
                WHEN 'p3.8xlarge' THEN 205.63  -- 30% RI savings
                WHEN 'p3.2xlarge' THEN 51.41
                WHEN 'r5.8xlarge' THEN 33.87
                WHEN 'r5.4xlarge' THEN 16.93
                ELSE 3.23
            END
        WHEN r.environment IN ('dev', 'staging') THEN 
            CASE r.instance_type
                WHEN 'p3.8xlarge' THEN 88.13   -- 70% spot savings
                WHEN 'p3.2xlarge' THEN 22.03
                ELSE 1.38
            END
        WHEN r.status = 'stopped' THEN 0.00
        ELSE 
            CASE r.instance_type
                WHEN 'p3.8xlarge' THEN 146.88  -- 50% rightsizing
                WHEN 'p3.2xlarge' THEN 36.72
                ELSE 2.31
            END
    END as optimized_cost,
    CASE 
        WHEN r.environment = 'prod' THEN 
            CASE r.instance_type
                WHEN 'p3.8xlarge' THEN 88.13
                WHEN 'p3.2xlarge' THEN 22.03
                WHEN 'r5.8xlarge' THEN 14.51
                WHEN 'r5.4xlarge' THEN 7.26
                ELSE 1.38
            END
        WHEN r.environment IN ('dev', 'staging') THEN 
            CASE r.instance_type
                WHEN 'p3.8xlarge' THEN 205.63
                WHEN 'p3.2xlarge' THEN 51.41
                ELSE 3.23
            END
        WHEN r.status = 'stopped' THEN 
            CASE r.instance_type
                WHEN 'p3.8xlarge' THEN 293.76
                WHEN 'p3.2xlarge' THEN 73.44
                ELSE 4.61
            END
        ELSE 
            CASE r.instance_type
                WHEN 'p3.8xlarge' THEN 146.88
                WHEN 'p3.2xlarge' THEN 36.72
                ELSE 2.30
            END
    END as potential_savings,
    CASE 
        WHEN r.environment = 'prod' THEN 0.85
        WHEN r.environment IN ('dev', 'staging') THEN 0.75
        WHEN r.status = 'stopped' THEN 0.95
        ELSE 0.70
    END as confidence,
    'active' as status
FROM resources r
WHERE RANDOM() < 0.3;  -- 30% of resources get recommendations

-- Show final data size
SELECT 
    'resources' as table_name, 
    COUNT(*) as record_count,
    pg_size_pretty(pg_total_relation_size('resources')) as size
FROM resources
UNION ALL
SELECT 
    'resource_costs' as table_name, 
    COUNT(*) as record_count,
    pg_size_pretty(pg_total_relation_size('resource_costs')) as size
FROM resource_costs
UNION ALL
SELECT 
    'resource_metrics' as table_name, 
    COUNT(*) as record_count,
    pg_size_pretty(pg_total_relation_size('resource_metrics')) as size
FROM resource_metrics
UNION ALL
SELECT 
    'optimization_recommendations' as table_name, 
    COUNT(*) as record_count,
    pg_size_pretty(pg_total_relation_size('optimization_recommendations')) as size
FROM optimization_recommendations;