-- Create yt_resource_metrics table for time series performance data
CREATE TABLE yt_resource_metrics (
    id SERIAL PRIMARY KEY,
    instance_id VARCHAR(50) NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    
    -- Core performance metrics from CloudWatch
    cpu_utilization DECIMAL(5,2), -- 0-100%
    memory_utilization DECIMAL(5,2), -- 0-100%
    disk_utilization DECIMAL(5,2), -- 0-100%
    network_in_mbps DECIMAL(10,2),
    network_out_mbps DECIMAL(10,2),
    disk_read_iops INTEGER,
    disk_write_iops INTEGER,
    
    -- Derived metrics for optimization
    cpu_credits_remaining INTEGER, -- for burstable instances
    memory_available_mb INTEGER,
    
    -- Workload pattern classification
    workload_pattern VARCHAR(20), -- steady, bursty, batch, idle
    optimization_score DECIMAL(3,2), -- 0-1 score for optimization potential
    
    -- Data source tracking
    metric_source VARCHAR(20) DEFAULT 'cloudwatch',
    
    FOREIGN KEY (instance_id) REFERENCES yt_aws_resources(instance_id) ON DELETE CASCADE
);

-- Partitioning by time for performance (monthly partitions)
CREATE TABLE yt_resource_metrics_2024_10 PARTITION OF yt_resource_metrics
    FOR VALUES FROM ('2024-10-01') TO ('2024-11-01');
CREATE TABLE yt_resource_metrics_2024_11 PARTITION OF yt_resource_metrics
    FOR VALUES FROM ('2024-11-01') TO ('2024-12-01');
CREATE TABLE yt_resource_metrics_2024_12 PARTITION OF yt_resource_metrics
    FOR VALUES FROM ('2024-12-01') TO ('2025-01-01');

-- Indexes for time series queries
CREATE INDEX idx_metrics_instance_time ON yt_resource_metrics(instance_id, timestamp DESC);
CREATE INDEX idx_metrics_timestamp ON yt_resource_metrics(timestamp DESC);
CREATE INDEX idx_metrics_workload_pattern ON yt_resource_metrics(workload_pattern);

-- Function to detect workload patterns
CREATE OR REPLACE FUNCTION detect_workload_pattern(
    avg_cpu DECIMAL,
    max_cpu DECIMAL,
    cpu_variance DECIMAL
) RETURNS VARCHAR(20) AS $$
BEGIN
    -- Batch processing: High CPU spikes then idle
    IF max_cpu > 80 AND avg_cpu < 30 AND cpu_variance > 25 THEN
        RETURN 'batch';
    -- Steady workload: Consistent CPU usage
    ELSIF avg_cpu > 20 AND cpu_variance < 15 THEN
        RETURN 'steady';
    -- Bursty workload: Variable CPU with spikes
    ELSIF cpu_variance > 20 THEN
        RETURN 'bursty';
    -- Idle: Low CPU usage
    ELSIF avg_cpu < 10 THEN
        RETURN 'idle';
    ELSE
        RETURN 'unknown';
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Function to calculate optimization score
CREATE OR REPLACE FUNCTION calculate_optimization_score(
    avg_cpu DECIMAL,
    avg_memory DECIMAL,
    workload_pattern VARCHAR
) RETURNS DECIMAL AS $$
BEGIN
    -- High optimization potential for underutilized resources
    IF avg_cpu < 20 AND avg_memory < 30 THEN
        RETURN 0.9; -- High potential for rightsizing
    ELSIF workload_pattern = 'batch' THEN
        RETURN 0.8; -- Good candidate for spot instances
    ELSIF workload_pattern = 'idle' THEN
        RETURN 0.95; -- Should be terminated or scheduled
    ELSIF avg_cpu < 50 AND avg_memory < 50 THEN
        RETURN 0.6; -- Moderate optimization potential
    ELSE
        RETURN 0.2; -- Well utilized, low optimization potential
    END IF;
END;
$$ LANGUAGE plpgsql;

-- View for latest metrics with optimization analysis
CREATE OR REPLACE VIEW vw_latest_metrics_with_optimization AS
SELECT 
    r.instance_id,
    r.instance_type,
    r.state,
    r.environment,
    m.timestamp as last_metric_time,
    m.cpu_utilization,
    m.memory_utilization,
    m.workload_pattern,
    m.optimization_score,
    p.on_demand_price_usd as hourly_cost,
    (p.on_demand_price_usd * 24 * 30) as monthly_cost,
    CASE 
        WHEN m.optimization_score > 0.8 THEN 'High'
        WHEN m.optimization_score > 0.6 THEN 'Medium'
        ELSE 'Low'
    END as optimization_priority
FROM yt_aws_resources r
LEFT JOIN LATERAL (
    SELECT * FROM yt_resource_metrics 
    WHERE instance_id = r.instance_id 
    ORDER BY timestamp DESC 
    LIMIT 1
) m ON true
LEFT JOIN yt_aws_pricing p ON r.instance_type = p.instance_type AND r.region = p.region
WHERE r.sync_status = 'active';

SELECT 'yt_resource_metrics table created successfully' as status;