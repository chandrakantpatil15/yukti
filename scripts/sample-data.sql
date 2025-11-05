-- Sample FinOps data for demo
INSERT INTO aws_pricings (instance_type, region, os, price_per_hour, ri_1yr_no_upfront, ri_1yr_partial_upfront, spot_price_avg) VALUES
('t3.micro', 'us-east-1', 'Linux', 0.0104, 0.0073, 0.0062, 0.0031),
('t3.small', 'us-east-1', 'Linux', 0.0208, 0.0146, 0.0125, 0.0062),
('m5.large', 'us-east-1', 'Linux', 0.096, 0.0672, 0.0576, 0.0288),
('c5.xlarge', 'us-east-1', 'Linux', 0.17, 0.119, 0.102, 0.051),
('r5.xlarge', 'us-east-1', 'Linux', 0.252, 0.1764, 0.1512, 0.0756)
ON CONFLICT (instance_type, region, os) DO UPDATE SET
    price_per_hour = EXCLUDED.price_per_hour,
    updated_at = CURRENT_TIMESTAMP;

INSERT INTO resources (resource_id, resource_type, instance_type, region, status, project_id, environment, launch_time) VALUES
('i-prod-web-01', 'ec2', 'm5.large', 'us-east-1', 'running', 1, 'prod', NOW() - INTERVAL '120 days'),
('i-prod-api-01', 'ec2', 'c5.xlarge', 'us-east-1', 'running', 2, 'prod', NOW() - INTERVAL '180 days'),
('i-dev-test-01', 'ec2', 't3.medium', 'us-east-1', 'running', 1, 'dev', NOW() - INTERVAL '10 days'),
('i-staging-01', 'ec2', 't3.small', 'us-east-1', 'running', 2, 'staging', NOW() - INTERVAL '30 days')
ON CONFLICT (resource_id) DO NOTHING;

-- Generate cost data for last 30 days
INSERT INTO resource_costs (resource_id, date, cost_usd, usage_hours, data_source)
SELECT 
    r.id,
    CURRENT_DATE - INTERVAL '1 day' * generate_series(0, 29) as date,
    CASE r.instance_type
        WHEN 'm5.large' THEN 2.30
        WHEN 'c5.xlarge' THEN 4.08
        WHEN 't3.medium' THEN 1.00
        WHEN 't3.small' THEN 0.50
        ELSE 1.00
    END as cost_usd,
    24 as usage_hours,
    'billing' as data_source
FROM resources r;

-- Sample optimization recommendations
INSERT INTO optimization_recommendations (resource_id, recommendation_type, current_cost, optimized_cost, potential_savings, confidence) 
SELECT 
    r.id,
    'reserved_instance' as recommendation_type,
    100.00 as current_cost,
    70.00 as optimized_cost,
    30.00 as potential_savings,
    0.85 as confidence
FROM resources r 
WHERE r.environment = 'prod'
LIMIT 2;

INSERT INTO optimization_recommendations (resource_id, recommendation_type, current_cost, optimized_cost, potential_savings, confidence)
SELECT 
    r.id,
    'spot_instance' as recommendation_type,
    50.00 as current_cost,
    15.00 as optimized_cost,
    35.00 as potential_savings,
    0.75 as confidence
FROM resources r 
WHERE r.environment IN ('dev', 'staging')
LIMIT 2;