-- Realistic FinOps Simulation with Real-World Use Cases
-- This simulates actual enterprise scenarios with proper cost patterns

-- Clear existing data
TRUNCATE TABLE resource_costs, optimization_recommendations, resource_metrics, resources CASCADE;

-- Real-world resource scenarios
INSERT INTO resources (resource_id, resource_type, instance_type, region, status, project_id, environment, launch_time) VALUES
-- Scenario 1: Over-provisioned development environments (common waste)
('i-dev-oversized-001', 'ec2', 'r5.8xlarge', 'us-east-1', 'running', 3, 'dev', NOW() - INTERVAL '45 days'),
('i-dev-oversized-002', 'ec2', 'c5.12xlarge', 'us-east-1', 'running', 3, 'dev', NOW() - INTERVAL '30 days'),
('i-dev-oversized-003', 'ec2', 'm5.16xlarge', 'us-east-1', 'running', 3, 'dev', NOW() - INTERVAL '60 days'),

-- Scenario 2: Zombie resources (stopped but still incurring storage costs)
('i-zombie-001', 'ec2', 'm5.2xlarge', 'us-east-1', 'stopped', 2, 'staging', NOW() - INTERVAL '90 days'),
('i-zombie-002', 'ec2', 'c5.4xlarge', 'us-east-1', 'stopped', 2, 'staging', NOW() - INTERVAL '120 days'),
('i-zombie-003', 'ec2', 'r5.4xlarge', 'us-east-1', 'stopped', 1, 'prod', NOW() - INTERVAL '180 days'),

-- Scenario 3: Long-running production workloads (RI candidates)
('i-prod-longrun-001', 'ec2', 'c5.9xlarge', 'us-east-1', 'running', 1, 'prod', NOW() - INTERVAL '400 days'),
('i-prod-longrun-002', 'ec2', 'r5.12xlarge', 'us-east-1', 'running', 1, 'prod', NOW() - INTERVAL '365 days'),
('i-prod-longrun-003', 'ec2', 'm5.8xlarge', 'us-east-1', 'running', 1, 'prod', NOW() - INTERVAL '500 days'),

-- Scenario 4: Batch processing workloads (spot candidates)
('i-batch-001', 'ec2', 'c5.18xlarge', 'us-east-1', 'running', 2, 'prod', NOW() - INTERVAL '10 days'),
('i-batch-002', 'ec2', 'c5.24xlarge', 'us-east-1', 'running', 2, 'prod', NOW() - INTERVAL '15 days'),
('i-batch-003', 'ec2', 'm5.24xlarge', 'us-east-1', 'running', 2, 'prod', NOW() - INTERVAL '5 days'),

-- Scenario 5: GPU workloads for ML (expensive, need optimization)
('i-ml-gpu-001', 'ec2', 'p3.8xlarge', 'us-east-1', 'running', 2, 'prod', NOW() - INTERVAL '20 days'),
('i-ml-gpu-002', 'ec2', 'p3.16xlarge', 'us-east-1', 'running', 2, 'prod', NOW() - INTERVAL '25 days'),

-- Scenario 6: Weekend-only development resources (schedule candidates)
('i-weekend-dev-001', 'ec2', 'm5.4xlarge', 'us-east-1', 'running', 3, 'dev', NOW() - INTERVAL '14 days'),
('i-weekend-dev-002', 'ec2', 'c5.4xlarge', 'us-east-1', 'running', 3, 'dev', NOW() - INTERVAL '21 days'),

-- Scenario 7: Right-sizing candidates (low utilization)
('i-underutil-001', 'ec2', 'r5.16xlarge', 'us-east-1', 'running', 1, 'prod', NOW() - INTERVAL '100 days'),
('i-underutil-002', 'ec2', 'c5.12xlarge', 'us-east-1', 'running', 1, 'prod', NOW() - INTERVAL '80 days');

-- Generate realistic cost patterns for each scenario
INSERT INTO resource_costs (resource_id, date, cost_usd, usage_hours, data_source)
SELECT 
    r.id,
    date_series.date,
    CASE 
        -- Over-provisioned dev environments (high cost, low value)
        WHEN r.resource_id LIKE 'i-dev-oversized%' THEN
            CASE r.instance_type
                WHEN 'r5.8xlarge' THEN 48.38 * 1.0  -- Full cost, should be much smaller
                WHEN 'c5.12xlarge' THEN 48.96 * 1.0
                WHEN 'm5.16xlarge' THEN 73.73 * 1.0
            END
        
        -- Zombie resources (storage costs only)
        WHEN r.resource_id LIKE 'i-zombie%' AND r.status = 'stopped' THEN
            CASE r.instance_type
                WHEN 'm5.2xlarge' THEN 2.50  -- EBS storage cost only
                WHEN 'c5.4xlarge' THEN 3.20
                WHEN 'r5.4xlarge' THEN 4.80
            END
        
        -- Long-running production (RI candidates)
        WHEN r.resource_id LIKE 'i-prod-longrun%' THEN
            CASE r.instance_type
                WHEN 'c5.9xlarge' THEN 36.72 * 1.0  -- On-demand pricing
                WHEN 'r5.12xlarge' THEN 72.58 * 1.0
                WHEN 'm5.8xlarge' THEN 36.86 * 1.0
            END
        
        -- Batch processing (high cost, good spot candidates)
        WHEN r.resource_id LIKE 'i-batch%' THEN
            CASE r.instance_type
                WHEN 'c5.18xlarge' THEN 73.44 * 1.0
                WHEN 'c5.24xlarge' THEN 97.92 * 1.0
                WHEN 'm5.24xlarge' THEN 110.59 * 1.0
            END
        
        -- GPU workloads (very expensive)
        WHEN r.resource_id LIKE 'i-ml-gpu%' THEN
            CASE r.instance_type
                WHEN 'p3.8xlarge' THEN 293.76 * 1.0
                WHEN 'p3.16xlarge' THEN 587.52 * 1.0
            END
        
        -- Weekend-only dev (should be scheduled)
        WHEN r.resource_id LIKE 'i-weekend-dev%' THEN
            CASE 
                WHEN EXTRACT(DOW FROM date_series.date) IN (0,6) THEN 0  -- Should be off on weekends
                ELSE 
                    CASE r.instance_type
                        WHEN 'm5.4xlarge' THEN 18.43
                        WHEN 'c5.4xlarge' THEN 16.32
                    END
            END
        
        -- Under-utilized resources
        WHEN r.resource_id LIKE 'i-underutil%' THEN
            CASE r.instance_type
                WHEN 'r5.16xlarge' THEN 96.77 * 1.0  -- High cost but low utilization
                WHEN 'c5.12xlarge' THEN 48.96 * 1.0
            END
        
        ELSE 10.00
    END as cost_usd,
    
    CASE 
        WHEN r.status = 'stopped' THEN 0
        WHEN r.resource_id LIKE 'i-weekend-dev%' AND EXTRACT(DOW FROM date_series.date) IN (0,6) THEN 0
        ELSE 24
    END as usage_hours,
    
    'aws_billing' as data_source
FROM resources r
CROSS JOIN (
    SELECT (CURRENT_DATE - INTERVAL '1 day' * generate_series(0, 89)) as date  -- Last 90 days
) date_series
WHERE r.launch_time::date <= date_series.date;

-- Generate realistic utilization metrics showing the problems
INSERT INTO resource_metrics (resource_id, timestamp, cpu_utilization, memory_utilization)
SELECT 
    r.id,
    timestamp_series.ts,
    CASE 
        -- Over-provisioned dev: Very low utilization (5-15%)
        WHEN r.resource_id LIKE 'i-dev-oversized%' THEN 
            GREATEST(0, LEAST(100, 8 + 7 * (RANDOM() - 0.5)))
        
        -- Long-running prod: Steady moderate utilization (good RI candidates)
        WHEN r.resource_id LIKE 'i-prod-longrun%' THEN 
            GREATEST(0, LEAST(100, 65 + 15 * (RANDOM() - 0.5)))
        
        -- Batch processing: High utilization during business hours
        WHEN r.resource_id LIKE 'i-batch%' THEN 
            CASE 
                WHEN EXTRACT(HOUR FROM timestamp_series.ts) BETWEEN 9 AND 17 THEN
                    GREATEST(0, LEAST(100, 85 + 10 * (RANDOM() - 0.5)))
                ELSE
                    GREATEST(0, LEAST(100, 20 + 15 * (RANDOM() - 0.5)))
            END
        
        -- GPU workloads: High utilization but expensive
        WHEN r.resource_id LIKE 'i-ml-gpu%' THEN 
            GREATEST(0, LEAST(100, 80 + 15 * (RANDOM() - 0.5)))
        
        -- Weekend dev: Should be zero on weekends
        WHEN r.resource_id LIKE 'i-weekend-dev%' THEN 
            CASE 
                WHEN EXTRACT(DOW FROM timestamp_series.ts) IN (0,6) THEN 0
                ELSE GREATEST(0, LEAST(100, 45 + 25 * (RANDOM() - 0.5)))
            END
        
        -- Under-utilized: Very low CPU despite large instance
        WHEN r.resource_id LIKE 'i-underutil%' THEN 
            GREATEST(0, LEAST(100, 12 + 8 * (RANDOM() - 0.5)))
        
        ELSE 50 + 30 * (RANDOM() - 0.5)
    END as cpu_utilization,
    
    CASE 
        -- Memory patterns similar to CPU but different ratios
        WHEN r.resource_id LIKE 'i-dev-oversized%' THEN 
            GREATEST(0, LEAST(100, 15 + 10 * (RANDOM() - 0.5)))
        
        WHEN r.resource_id LIKE 'i-prod-longrun%' THEN 
            GREATEST(0, LEAST(100, 70 + 20 * (RANDOM() - 0.5)))
        
        WHEN r.resource_id LIKE 'i-batch%' THEN 
            GREATEST(0, LEAST(100, 75 + 15 * (RANDOM() - 0.5)))
        
        WHEN r.resource_id LIKE 'i-ml-gpu%' THEN 
            GREATEST(0, LEAST(100, 85 + 10 * (RANDOM() - 0.5)))
        
        WHEN r.resource_id LIKE 'i-weekend-dev%' THEN 
            CASE 
                WHEN EXTRACT(DOW FROM timestamp_series.ts) IN (0,6) THEN 0
                ELSE GREATEST(0, LEAST(100, 40 + 20 * (RANDOM() - 0.5)))
            END
        
        WHEN r.resource_id LIKE 'i-underutil%' THEN 
            GREATEST(0, LEAST(100, 25 + 15 * (RANDOM() - 0.5)))
        
        ELSE 45 + 35 * (RANDOM() - 0.5)
    END as memory_utilization
FROM resources r
CROSS JOIN (
    SELECT NOW() - INTERVAL '1 hour' * generate_series(0, 167) as ts  -- Last 7 days hourly
) timestamp_series
WHERE r.status = 'running';

-- Generate realistic optimization recommendations based on actual patterns
INSERT INTO optimization_recommendations (resource_id, recommendation_type, current_cost, optimized_cost, potential_savings, confidence, status)
SELECT 
    r.id,
    CASE 
        WHEN r.resource_id LIKE 'i-dev-oversized%' THEN 'rightsizing'
        WHEN r.resource_id LIKE 'i-zombie%' THEN 'termination'
        WHEN r.resource_id LIKE 'i-prod-longrun%' THEN 'reserved_instance'
        WHEN r.resource_id LIKE 'i-batch%' THEN 'spot_instance'
        WHEN r.resource_id LIKE 'i-ml-gpu%' THEN 'scheduling'
        WHEN r.resource_id LIKE 'i-weekend-dev%' THEN 'scheduling'
        WHEN r.resource_id LIKE 'i-underutil%' THEN 'rightsizing'
        ELSE 'reserved_instance'
    END as recommendation_type,
    
    -- Current monthly costs
    CASE 
        WHEN r.resource_id LIKE 'i-dev-oversized-001' THEN 1451.40  -- r5.8xlarge
        WHEN r.resource_id LIKE 'i-dev-oversized-002' THEN 1468.80  -- c5.12xlarge
        WHEN r.resource_id LIKE 'i-dev-oversized-003' THEN 2211.90  -- m5.16xlarge
        WHEN r.resource_id LIKE 'i-zombie%' THEN 
            CASE r.instance_type
                WHEN 'm5.2xlarge' THEN 75.00
                WHEN 'c5.4xlarge' THEN 96.00
                WHEN 'r5.4xlarge' THEN 144.00
            END
        WHEN r.resource_id LIKE 'i-prod-longrun-001' THEN 1101.60  -- c5.9xlarge
        WHEN r.resource_id LIKE 'i-prod-longrun-002' THEN 2177.40  -- r5.12xlarge
        WHEN r.resource_id LIKE 'i-prod-longrun-003' THEN 1105.80  -- m5.8xlarge
        WHEN r.resource_id LIKE 'i-batch-001' THEN 2203.20  -- c5.18xlarge
        WHEN r.resource_id LIKE 'i-batch-002' THEN 2937.60  -- c5.24xlarge
        WHEN r.resource_id LIKE 'i-batch-003' THEN 3317.70  -- m5.24xlarge
        WHEN r.resource_id LIKE 'i-ml-gpu-001' THEN 8812.80  -- p3.8xlarge
        WHEN r.resource_id LIKE 'i-ml-gpu-002' THEN 17625.60  -- p3.16xlarge
        WHEN r.resource_id LIKE 'i-weekend-dev%' THEN 
            CASE r.instance_type
                WHEN 'm5.4xlarge' THEN 552.90
                WHEN 'c5.4xlarge' THEN 489.60
            END
        WHEN r.resource_id LIKE 'i-underutil-001' THEN 2903.10  -- r5.16xlarge
        WHEN r.resource_id LIKE 'i-underutil-002' THEN 1468.80  -- c5.12xlarge
        ELSE 300.00
    END::decimal as current_cost,
    
    -- Optimized costs after recommendations
    CASE 
        -- Rightsizing: Move to smaller instances (60% savings)
        WHEN r.resource_id LIKE 'i-dev-oversized-001' THEN 580.56   -- r5.8xlarge -> r5.2xlarge
        WHEN r.resource_id LIKE 'i-dev-oversized-002' THEN 587.52   -- c5.12xlarge -> c5.4xlarge
        WHEN r.resource_id LIKE 'i-dev-oversized-003' THEN 884.76   -- m5.16xlarge -> m5.4xlarge
        
        -- Termination: Zero cost
        WHEN r.resource_id LIKE 'i-zombie%' THEN 0.00
        
        -- Reserved Instances: 30% savings
        WHEN r.resource_id LIKE 'i-prod-longrun-001' THEN 771.12   -- 30% RI savings
        WHEN r.resource_id LIKE 'i-prod-longrun-002' THEN 1524.18
        WHEN r.resource_id LIKE 'i-prod-longrun-003' THEN 774.06
        
        -- Spot Instances: 70% savings
        WHEN r.resource_id LIKE 'i-batch-001' THEN 660.96
        WHEN r.resource_id LIKE 'i-batch-002' THEN 881.28
        WHEN r.resource_id LIKE 'i-batch-003' THEN 995.31
        
        -- Scheduling: 60% savings (run only when needed)
        WHEN r.resource_id LIKE 'i-ml-gpu-001' THEN 3525.12
        WHEN r.resource_id LIKE 'i-ml-gpu-002' THEN 7050.24
        WHEN r.resource_id LIKE 'i-weekend-dev%' THEN 
            CASE r.instance_type
                WHEN 'm5.4xlarge' THEN 165.87  -- 70% savings with scheduling
                WHEN 'c5.4xlarge' THEN 146.88
            END
        
        -- Rightsizing for under-utilized
        WHEN r.resource_id LIKE 'i-underutil-001' THEN 1161.24  -- r5.16xlarge -> r5.4xlarge
        WHEN r.resource_id LIKE 'i-underutil-002' THEN 587.52   -- c5.12xlarge -> c5.4xlarge
        
        ELSE 210.00
    END::decimal as optimized_cost,
    
    -- Calculate savings
    CASE 
        WHEN r.resource_id LIKE 'i-dev-oversized-001' THEN 870.84
        WHEN r.resource_id LIKE 'i-dev-oversized-002' THEN 881.28
        WHEN r.resource_id LIKE 'i-dev-oversized-003' THEN 1327.14
        WHEN r.resource_id LIKE 'i-zombie%' THEN 
            CASE r.instance_type
                WHEN 'm5.2xlarge' THEN 75.00
                WHEN 'c5.4xlarge' THEN 96.00
                WHEN 'r5.4xlarge' THEN 144.00
            END
        WHEN r.resource_id LIKE 'i-prod-longrun-001' THEN 330.48
        WHEN r.resource_id LIKE 'i-prod-longrun-002' THEN 653.22
        WHEN r.resource_id LIKE 'i-prod-longrun-003' THEN 331.74
        WHEN r.resource_id LIKE 'i-batch-001' THEN 1542.24
        WHEN r.resource_id LIKE 'i-batch-002' THEN 2056.32
        WHEN r.resource_id LIKE 'i-batch-003' THEN 2322.39
        WHEN r.resource_id LIKE 'i-ml-gpu-001' THEN 5287.68
        WHEN r.resource_id LIKE 'i-ml-gpu-002' THEN 10575.36
        WHEN r.resource_id LIKE 'i-weekend-dev%' THEN 
            CASE r.instance_type
                WHEN 'm5.4xlarge' THEN 387.03
                WHEN 'c5.4xlarge' THEN 342.72
            END
        WHEN r.resource_id LIKE 'i-underutil-001' THEN 1741.86
        WHEN r.resource_id LIKE 'i-underutil-002' THEN 881.28
        ELSE 90.00
    END::decimal as potential_savings,
    
    -- Confidence based on data quality and patterns
    CASE 
        WHEN r.resource_id LIKE 'i-dev-oversized%' THEN 0.95  -- Clear over-provisioning
        WHEN r.resource_id LIKE 'i-zombie%' THEN 0.99        -- Obviously should be terminated
        WHEN r.resource_id LIKE 'i-prod-longrun%' THEN 0.90  -- Good RI candidates
        WHEN r.resource_id LIKE 'i-batch%' THEN 0.85         -- Good spot candidates
        WHEN r.resource_id LIKE 'i-ml-gpu%' THEN 0.80        -- Scheduling needs validation
        WHEN r.resource_id LIKE 'i-weekend-dev%' THEN 0.92   -- Clear scheduling opportunity
        WHEN r.resource_id LIKE 'i-underutil%' THEN 0.88     -- Clear rightsizing opportunity
        ELSE 0.70
    END as confidence,
    
    'active' as status
FROM resources r;

-- Show realistic summary
SELECT 
    'REALISTIC FINOPS SIMULATION SUMMARY' as title,
    '' as separator;

SELECT 
    'Total Monthly Cost' as metric,
    '$' || to_char(SUM(rc.cost_usd * 30), '999,999.99') as value
FROM resource_costs rc
WHERE rc.date >= CURRENT_DATE - INTERVAL '1 day';

SELECT 
    'Total Potential Monthly Savings' as metric,
    '$' || to_char(SUM(potential_savings), '999,999.99') as value
FROM optimization_recommendations
WHERE status = 'active';

SELECT 
    recommendation_type,
    COUNT(*) as resources,
    '$' || to_char(SUM(potential_savings), '999,999.99') as monthly_savings
FROM optimization_recommendations
WHERE status = 'active'
GROUP BY recommendation_type
ORDER BY SUM(potential_savings) DESC;