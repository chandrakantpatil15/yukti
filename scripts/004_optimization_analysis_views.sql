-- Advanced views for resource utilization vs pricing analysis
-- This shows how utilization compares with cost for optimization decisions

-- 1. Resource Efficiency Analysis
CREATE OR REPLACE VIEW vw_resource_efficiency_analysis AS
SELECT 
    r.instance_id,
    r.instance_type,
    r.state,
    r.environment,
    
    -- Pricing data
    p.on_demand_price_usd as hourly_cost,
    p.reserved_1yr_no_upfront as reserved_hourly_cost,
    p.spot_price_avg as spot_hourly_cost,
    (p.on_demand_price_usd * 24 * 30) as monthly_cost,
    
    -- Resource specifications
    p.vcpu,
    p.memory_gb,
    
    -- Utilization metrics (last 7 days average)
    COALESCE(AVG(m.cpu_utilization), 0) as avg_cpu_7d,
    COALESCE(AVG(m.memory_utilization), 0) as avg_memory_7d,
    COALESCE(MAX(m.cpu_utilization), 0) as max_cpu_7d,
    COALESCE(MIN(m.cpu_utilization), 0) as min_cpu_7d,
    
    -- Cost efficiency calculations
    CASE 
        WHEN AVG(m.cpu_utilization) > 0 THEN 
            p.on_demand_price_usd / (AVG(m.cpu_utilization) / 100)
        ELSE p.on_demand_price_usd * 100 
    END as cost_per_cpu_utilization,
    
    CASE 
        WHEN AVG(m.memory_utilization) > 0 THEN 
            p.on_demand_price_usd / (AVG(m.memory_utilization) / 100)
        ELSE p.on_demand_price_usd * 100 
    END as cost_per_memory_utilization,
    
    -- Optimization potential
    CASE 
        WHEN AVG(m.cpu_utilization) < 20 AND AVG(m.memory_utilization) < 30 THEN 'HIGH'
        WHEN AVG(m.cpu_utilization) < 50 AND AVG(m.memory_utilization) < 50 THEN 'MEDIUM'
        ELSE 'LOW'
    END as optimization_potential,
    
    -- Recommended actions
    CASE 
        WHEN r.state = 'stopped' THEN 'TERMINATE'
        WHEN AVG(m.cpu_utilization) < 10 THEN 'SCHEDULE_OR_TERMINATE'
        WHEN AVG(m.cpu_utilization) < 20 AND AVG(m.memory_utilization) < 30 THEN 'RIGHTSIZE_DOWN'
        WHEN MAX(m.cpu_utilization) > 80 AND AVG(m.cpu_utilization) < 40 THEN 'SPOT_INSTANCE'
        WHEN r.environment = 'prod' AND EXTRACT(DAYS FROM NOW() - r.launch_time) > 90 THEN 'RESERVED_INSTANCE'
        ELSE 'MONITOR'
    END as recommended_action,
    
    -- Potential savings calculations
    CASE 
        WHEN AVG(m.cpu_utilization) < 20 AND AVG(m.memory_utilization) < 30 THEN 
            (p.on_demand_price_usd * 24 * 30) * 0.6  -- 60% savings from rightsizing
        WHEN r.environment != 'prod' THEN 
            (p.on_demand_price_usd * 24 * 30) * 0.7  -- 70% savings from spot
        WHEN r.environment = 'prod' AND EXTRACT(DAYS FROM NOW() - r.launch_time) > 90 THEN 
            (p.on_demand_price_usd * 24 * 30) * 0.3  -- 30% savings from RI
        ELSE 0
    END as potential_monthly_savings

FROM yt_aws_resources r
LEFT JOIN yt_aws_pricing p ON r.instance_type = p.instance_type AND r.region = p.region
LEFT JOIN yt_resource_metrics m ON r.instance_id = m.instance_id 
    AND m.timestamp > NOW() - INTERVAL '7 days'
WHERE r.sync_status = 'active'
GROUP BY r.instance_id, r.instance_type, r.state, r.environment, r.launch_time,
         p.on_demand_price_usd, p.reserved_1yr_no_upfront, p.spot_price_avg, 
         p.vcpu, p.memory_gb;

-- 2. Rightsizing Recommendations with Cost Impact
CREATE OR REPLACE VIEW vw_rightsizing_recommendations AS
SELECT 
    r.instance_id,
    r.instance_type as current_type,
    r.state,
    
    -- Current specs and cost
    p_current.vcpu as current_vcpu,
    p_current.memory_gb as current_memory,
    p_current.on_demand_price_usd as current_hourly_cost,
    (p_current.on_demand_price_usd * 24 * 30) as current_monthly_cost,
    
    -- Utilization
    AVG(m.cpu_utilization) as avg_cpu_utilization,
    AVG(m.memory_utilization) as avg_memory_utilization,
    
    -- Recommended instance type based on utilization
    CASE 
        WHEN AVG(m.cpu_utilization) < 20 AND AVG(m.memory_utilization) < 30 THEN
            CASE 
                WHEN r.instance_type LIKE 'm5.%' THEN 
                    CASE r.instance_type
                        WHEN 'm5.2xlarge' THEN 'm5.large'
                        WHEN 'm5.xlarge' THEN 'm5.large'
                        WHEN 'm5.4xlarge' THEN 'm5.xlarge'
                        WHEN 'm5.8xlarge' THEN 'm5.2xlarge'
                        ELSE r.instance_type
                    END
                WHEN r.instance_type LIKE 'c5.%' THEN 
                    CASE r.instance_type
                        WHEN 'c5.2xlarge' THEN 'c5.large'
                        WHEN 'c5.xlarge' THEN 'c5.large'
                        WHEN 'c5.4xlarge' THEN 'c5.xlarge'
                        WHEN 'c5.9xlarge' THEN 'c5.2xlarge'
                        ELSE r.instance_type
                    END
                ELSE r.instance_type
            END
        ELSE r.instance_type
    END as recommended_type,
    
    -- Cost comparison will be added via JOIN with pricing table
    'RIGHTSIZE' as recommendation_type,
    
    -- Confidence based on data quality
    CASE 
        WHEN COUNT(m.id) > 100 THEN 0.95  -- Lots of data points
        WHEN COUNT(m.id) > 50 THEN 0.85   -- Good data
        WHEN COUNT(m.id) > 10 THEN 0.70   -- Some data
        ELSE 0.50                         -- Limited data
    END as confidence_score

FROM yt_aws_resources r
LEFT JOIN yt_aws_pricing p_current ON r.instance_type = p_current.instance_type AND r.region = p_current.region
LEFT JOIN yt_resource_metrics m ON r.instance_id = m.instance_id 
    AND m.timestamp > NOW() - INTERVAL '7 days'
WHERE r.sync_status = 'active' 
    AND r.state = 'running'
    AND AVG(m.cpu_utilization) < 50  -- Only underutilized resources
GROUP BY r.instance_id, r.instance_type, r.state, r.region,
         p_current.vcpu, p_current.memory_gb, p_current.on_demand_price_usd
HAVING AVG(m.cpu_utilization) < 50 AND COUNT(m.id) > 5;

-- 3. Cost vs Performance Dashboard
CREATE OR REPLACE VIEW vw_cost_performance_dashboard AS
SELECT 
    r.environment,
    COUNT(*) as total_instances,
    SUM(p.on_demand_price_usd * 24 * 30) as total_monthly_cost,
    AVG(m.cpu_utilization) as avg_cpu_utilization,
    AVG(m.memory_utilization) as avg_memory_utilization,
    
    -- Efficiency metrics
    SUM(CASE WHEN m.cpu_utilization < 20 THEN 1 ELSE 0 END) as underutilized_instances,
    SUM(CASE WHEN m.cpu_utilization > 80 THEN 1 ELSE 0 END) as overutilized_instances,
    
    -- Cost efficiency
    SUM(p.on_demand_price_usd * 24 * 30) / NULLIF(AVG(m.cpu_utilization), 0) as cost_per_cpu_percent,
    
    -- Potential savings
    SUM(CASE 
        WHEN m.cpu_utilization < 20 THEN (p.on_demand_price_usd * 24 * 30) * 0.6
        ELSE 0 
    END) as potential_rightsizing_savings,
    
    SUM(CASE 
        WHEN r.environment != 'prod' THEN (p.on_demand_price_usd * 24 * 30) * 0.7
        ELSE 0 
    END) as potential_spot_savings

FROM yt_aws_resources r
LEFT JOIN yt_aws_pricing p ON r.instance_type = p.instance_type AND r.region = p.region
LEFT JOIN yt_resource_metrics m ON r.instance_id = m.instance_id 
    AND m.timestamp > NOW() - INTERVAL '7 days'
WHERE r.sync_status = 'active' AND r.state = 'running'
GROUP BY r.environment
ORDER BY total_monthly_cost DESC;

SELECT 'Optimization analysis views created successfully' as status;