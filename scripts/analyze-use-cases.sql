-- Detailed Analysis of Realistic FinOps Use Cases
-- This shows how each scenario would be identified and optimized

SELECT '🎯 FINOPS USE CASE ANALYSIS' as title;

-- Use Case 1: Over-provisioned Development Environments
SELECT 
    '1. OVER-PROVISIONED DEV ENVIRONMENTS' as use_case,
    '' as separator;

SELECT 
    r.resource_id,
    r.instance_type,
    r.environment,
    ROUND(AVG(rm.cpu_utilization), 1) as avg_cpu_util,
    ROUND(AVG(rm.memory_utilization), 1) as avg_memory_util,
    '$' || ROUND(AVG(rc.cost_usd), 2) as daily_cost,
    '$' || ROUND(AVG(rc.cost_usd) * 30, 2) as monthly_cost,
    'Rightsize to smaller instance' as recommendation
FROM resources r
JOIN resource_costs rc ON r.id = rc.resource_id
JOIN resource_metrics rm ON r.id = rm.resource_id
WHERE r.resource_id LIKE 'i-dev-oversized%'
GROUP BY r.id, r.resource_id, r.instance_type, r.environment
ORDER BY AVG(rc.cost_usd) DESC;

-- Use Case 2: Zombie Resources (Stopped but Costing Money)
SELECT 
    '2. ZOMBIE RESOURCES (STOPPED BUT COSTING)' as use_case,
    '' as separator;

SELECT 
    r.resource_id,
    r.instance_type,
    r.status,
    EXTRACT(DAYS FROM NOW() - r.launch_time) as days_since_launch,
    '$' || ROUND(AVG(rc.cost_usd), 2) as daily_storage_cost,
    '$' || ROUND(AVG(rc.cost_usd) * 30, 2) as monthly_waste,
    'Terminate to eliminate costs' as recommendation
FROM resources r
JOIN resource_costs rc ON r.id = rc.resource_id
WHERE r.resource_id LIKE 'i-zombie%'
GROUP BY r.id, r.resource_id, r.instance_type, r.status, r.launch_time
ORDER BY AVG(rc.cost_usd) DESC;

-- Use Case 3: Reserved Instance Candidates
SELECT 
    '3. RESERVED INSTANCE CANDIDATES' as use_case,
    '' as separator;

SELECT 
    r.resource_id,
    r.instance_type,
    r.environment,
    EXTRACT(DAYS FROM NOW() - r.launch_time) as days_running,
    ROUND(AVG(rm.cpu_utilization), 1) as avg_cpu_util,
    '$' || ROUND(AVG(rc.cost_usd), 2) as daily_cost,
    '$' || ROUND(AVG(rc.cost_usd) * 30, 2) as monthly_cost,
    '$' || ROUND(AVG(rc.cost_usd) * 30 * 0.7, 2) as ri_monthly_cost,
    '$' || ROUND(AVG(rc.cost_usd) * 30 * 0.3, 2) as monthly_savings
FROM resources r
JOIN resource_costs rc ON r.id = rc.resource_id
JOIN resource_metrics rm ON r.id = rm.resource_id
WHERE r.resource_id LIKE 'i-prod-longrun%'
GROUP BY r.id, r.resource_id, r.instance_type, r.environment, r.launch_time
ORDER BY AVG(rc.cost_usd) DESC;

-- Use Case 4: Spot Instance Candidates (Batch Processing)
SELECT 
    '4. SPOT INSTANCE CANDIDATES (BATCH WORKLOADS)' as use_case,
    '' as separator;

SELECT 
    r.resource_id,
    r.instance_type,
    '$' || ROUND(AVG(rc.cost_usd), 2) as daily_cost,
    '$' || ROUND(AVG(rc.cost_usd) * 30, 2) as monthly_cost,
    '$' || ROUND(AVG(rc.cost_usd) * 30 * 0.3, 2) as spot_monthly_cost,
    '$' || ROUND(AVG(rc.cost_usd) * 30 * 0.7, 2) as monthly_savings,
    '70% savings with spot instances' as benefit
FROM resources r
JOIN resource_costs rc ON r.id = rc.resource_id
WHERE r.resource_id LIKE 'i-batch%'
GROUP BY r.id, r.resource_id, r.instance_type
ORDER BY AVG(rc.cost_usd) DESC;

-- Use Case 5: GPU Workloads Needing Scheduling
SELECT 
    '5. EXPENSIVE GPU WORKLOADS' as use_case,
    '' as separator;

SELECT 
    r.resource_id,
    r.instance_type,
    '$' || ROUND(AVG(rc.cost_usd), 2) as daily_cost,
    '$' || ROUND(AVG(rc.cost_usd) * 30, 2) as monthly_cost,
    '$' || ROUND(AVG(rc.cost_usd) * 30 * 0.4, 2) as scheduled_cost,
    '$' || ROUND(AVG(rc.cost_usd) * 30 * 0.6, 2) as monthly_savings,
    'Schedule for business hours only' as recommendation
FROM resources r
JOIN resource_costs rc ON r.id = rc.resource_id
WHERE r.resource_id LIKE 'i-ml-gpu%'
GROUP BY r.id, r.resource_id, r.instance_type
ORDER BY AVG(rc.cost_usd) DESC;

-- Use Case 6: Under-utilized Large Instances
SELECT 
    '6. UNDER-UTILIZED LARGE INSTANCES' as use_case,
    '' as separator;

SELECT 
    r.resource_id,
    r.instance_type,
    ROUND(AVG(rm.cpu_utilization), 1) as avg_cpu_util,
    ROUND(AVG(rm.memory_utilization), 1) as avg_memory_util,
    '$' || ROUND(AVG(rc.cost_usd), 2) as daily_cost,
    '$' || ROUND(AVG(rc.cost_usd) * 30, 2) as monthly_cost,
    CASE 
        WHEN r.instance_type = 'r5.16xlarge' THEN 'Rightsize to r5.4xlarge'
        WHEN r.instance_type = 'c5.12xlarge' THEN 'Rightsize to c5.4xlarge'
        ELSE 'Rightsize to smaller instance'
    END as recommendation
FROM resources r
JOIN resource_costs rc ON r.id = rc.resource_id
JOIN resource_metrics rm ON r.id = rm.resource_id
WHERE r.resource_id LIKE 'i-underutil%'
GROUP BY r.id, r.resource_id, r.instance_type
ORDER BY AVG(rc.cost_usd) DESC;

-- Summary of All Optimization Opportunities
SELECT 
    'OPTIMIZATION SUMMARY BY TYPE' as summary,
    '' as separator;

SELECT 
    or_rec.recommendation_type,
    COUNT(*) as resource_count,
    '$' || ROUND(SUM(or_rec.current_cost), 2) as total_current_cost,
    '$' || ROUND(SUM(or_rec.optimized_cost), 2) as total_optimized_cost,
    '$' || ROUND(SUM(or_rec.potential_savings), 2) as total_monthly_savings,
    ROUND(AVG(or_rec.confidence) * 100, 1) || '%' as avg_confidence
FROM optimization_recommendations or_rec
WHERE or_rec.status = 'active'
GROUP BY or_rec.recommendation_type
ORDER BY SUM(or_rec.potential_savings) DESC;

-- ROI Analysis
SELECT 
    'ROI ANALYSIS' as analysis,
    '' as separator;

SELECT 
    '$' || ROUND(SUM(or_rec.current_cost), 2) as current_monthly_spend,
    '$' || ROUND(SUM(or_rec.potential_savings), 2) as potential_monthly_savings,
    ROUND((SUM(or_rec.potential_savings) / SUM(or_rec.current_cost)) * 100, 1) || '%' as cost_reduction_percentage,
    '$' || ROUND(SUM(or_rec.potential_savings) * 12, 2) as annual_savings_potential
FROM optimization_recommendations or_rec
WHERE or_rec.status = 'active';