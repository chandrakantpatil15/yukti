-- Fix and generate massive cost data
-- Clear existing cost data
TRUNCATE TABLE resource_costs CASCADE;

-- Generate massive cost data (2 years of daily costs)
INSERT INTO resource_costs (resource_id, date, cost_usd, usage_hours, data_source)
SELECT 
    r.id,
    date_series.date,
    CASE 
        WHEN r.status = 'running' THEN 
            CASE r.instance_type
                -- GPU instances (highest cost)
                WHEN 'p3.16xlarge' THEN 587.52 * date_series.daily_multiplier
                WHEN 'p3.8xlarge' THEN 293.76 * date_series.daily_multiplier
                WHEN 'p3.2xlarge' THEN 73.44 * date_series.daily_multiplier
                -- Large compute instances
                WHEN 'c5.24xlarge' THEN 97.92 * date_series.daily_multiplier
                WHEN 'c5.18xlarge' THEN 73.44 * date_series.daily_multiplier
                WHEN 'c5.12xlarge' THEN 48.96 * date_series.daily_multiplier
                WHEN 'c5.9xlarge' THEN 36.72 * date_series.daily_multiplier
                -- Large memory instances
                WHEN 'r5.24xlarge' THEN 145.15 * date_series.daily_multiplier
                WHEN 'r5.16xlarge' THEN 96.77 * date_series.daily_multiplier
                WHEN 'r5.12xlarge' THEN 72.58 * date_series.daily_multiplier
                WHEN 'r5.8xlarge' THEN 48.38 * date_series.daily_multiplier
                -- Large general purpose
                WHEN 'm5.24xlarge' THEN 110.59 * date_series.daily_multiplier
                WHEN 'm5.16xlarge' THEN 73.73 * date_series.daily_multiplier
                WHEN 'm5.12xlarge' THEN 55.30 * date_series.daily_multiplier
                WHEN 'm5.8xlarge' THEN 36.86 * date_series.daily_multiplier
                -- Storage optimized
                WHEN 'i3.16xlarge' THEN 119.81 * date_series.daily_multiplier
                WHEN 'i3.8xlarge' THEN 59.90 * date_series.daily_multiplier
                WHEN 'i3.4xlarge' THEN 29.95 * date_series.daily_multiplier
                -- Medium instances
                WHEN 'c5.4xlarge' THEN 16.32 * date_series.daily_multiplier
                WHEN 'c5.2xlarge' THEN 8.16 * date_series.daily_multiplier
                WHEN 'r5.4xlarge' THEN 24.19 * date_series.daily_multiplier
                WHEN 'r5.2xlarge' THEN 12.10 * date_series.daily_multiplier
                WHEN 'm5.4xlarge' THEN 18.43 * date_series.daily_multiplier
                WHEN 'm5.2xlarge' THEN 9.22 * date_series.daily_multiplier
                ELSE 4.61 * date_series.daily_multiplier
            END
        ELSE 0
    END as cost_usd,
    CASE WHEN r.status = 'running' THEN 24 ELSE 0 END as usage_hours,
    'aws_billing' as data_source
FROM resources r
CROSS JOIN (
    SELECT 
        (CURRENT_DATE - INTERVAL '1 day' * gs.day_offset) as date,
        CASE 
            WHEN EXTRACT(DOW FROM CURRENT_DATE - INTERVAL '1 day' * gs.day_offset) IN (0,6) THEN 0.7  -- Weekend reduction
            WHEN EXTRACT(MONTH FROM CURRENT_DATE - INTERVAL '1 day' * gs.day_offset) = 12 THEN 1.3    -- Holiday spike
            WHEN EXTRACT(MONTH FROM CURRENT_DATE - INTERVAL '1 day' * gs.day_offset) IN (6,7,8) THEN 1.2  -- Summer peak
            ELSE 1.0
        END as daily_multiplier
    FROM generate_series(0, 729) gs(day_offset)
) date_series
WHERE r.launch_time::date <= date_series.date;

-- Show final data counts
SELECT 
    'resources' as table_name, 
    COUNT(*) as count, 
    pg_size_pretty(pg_total_relation_size('resources')) as size
FROM resources
UNION ALL
SELECT 
    'resource_costs', 
    COUNT(*), 
    pg_size_pretty(pg_total_relation_size('resource_costs'))
FROM resource_costs
UNION ALL
SELECT 
    'resource_metrics', 
    COUNT(*), 
    pg_size_pretty(pg_total_relation_size('resource_metrics'))
FROM resource_metrics
UNION ALL
SELECT 
    'optimization_recommendations', 
    COUNT(*), 
    pg_size_pretty(pg_total_relation_size('optimization_recommendations'))
FROM optimization_recommendations
ORDER BY count DESC;