-- Optimized Assessment-Based Schema for Yukti FinOps Platform
-- Stores only ratings/assessments, not raw time-series data

-- 1. Updated resources table with ARN as primary identifier
ALTER TABLE yt_aws_resources ADD COLUMN IF NOT EXISTS resource_arn VARCHAR(255) UNIQUE;
UPDATE yt_aws_resources SET resource_arn = 
    'arn:aws:ec2:' || region || ':' || '144403604430' || ':instance/' || instance_id 
    WHERE resource_arn IS NULL;

-- 2. Assessment configuration table (user-configurable thresholds)
CREATE TABLE IF NOT EXISTS yt_assessment_config (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(50) DEFAULT 'default',
    
    -- Underutilized thresholds
    underutilized_cpu_threshold DECIMAL(5,2) DEFAULT 20.0,
    underutilized_memory_threshold DECIMAL(5,2) DEFAULT 25.0,
    underutilized_window_days INTEGER DEFAULT 7,
    
    -- Overutilized thresholds  
    overutilized_cpu_threshold DECIMAL(5,2) DEFAULT 80.0,
    overutilized_memory_threshold DECIMAL(5,2) DEFAULT 80.0,
    
    -- Batch workload detection
    batch_high_threshold DECIMAL(5,2) DEFAULT 80.0,
    batch_low_threshold DECIMAL(5,2) DEFAULT 30.0,
    batch_variance_threshold DECIMAL(5,2) DEFAULT 25.0,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Insert default configuration
INSERT INTO yt_assessment_config (tenant_id) VALUES ('default') 
ON CONFLICT DO NOTHING;

-- 3. Lightweight resource assessments table (replaces heavy metrics table)
CREATE TABLE IF NOT EXISTS yt_resource_assessments (
    id SERIAL PRIMARY KEY,
    resource_arn VARCHAR(255) NOT NULL,
    assessment_timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    
    -- Assessment period
    assessment_window_hours INTEGER DEFAULT 24,
    assessment_window_start TIMESTAMP WITH TIME ZONE,
    assessment_window_end TIMESTAMP WITH TIME ZONE,
    
    -- Utilization category (main classification)
    utilization_category VARCHAR(20) NOT NULL, -- underutilized, overutilized, intermittent, batch, idle
    
    -- Summary metrics (aggregated from time-series)
    avg_cpu_utilization DECIMAL(5,2),
    max_cpu_utilization DECIMAL(5,2),
    avg_memory_utilization DECIMAL(5,2),
    max_memory_utilization DECIMAL(5,2),
    cpu_variance DECIMAL(5,2),
    
    -- Pattern analysis
    usage_pattern VARCHAR(20), -- steady, bursty, scheduled, batch, idle
    peak_hours VARCHAR(50), -- e.g., "09:00-17:00" or "22:00-02:00"
    idle_percentage DECIMAL(5,2), -- % of time idle
    
    -- Optimization recommendations
    optimization_score DECIMAL(3,2), -- 0-1 score
    recommended_action VARCHAR(50), -- downsize, upsize, spot, schedule, terminate
    recommended_instance_type VARCHAR(50),
    potential_monthly_savings DECIMAL(10,2),
    
    -- Cost attribution
    current_hourly_cost DECIMAL(8,4),
    projected_hourly_cost DECIMAL(8,4),
    
    -- Assessment metadata
    confidence_score INTEGER DEFAULT 85, -- 0-100
    data_points_analyzed INTEGER,
    assessment_engine_version VARCHAR(10) DEFAULT '1.0',
    
    FOREIGN KEY (resource_arn) REFERENCES yt_aws_resources(resource_arn) ON DELETE CASCADE
);

-- 4. Assessment history for trend analysis (lightweight)
CREATE TABLE IF NOT EXISTS yt_assessment_history (
    id SERIAL PRIMARY KEY,
    resource_arn VARCHAR(255) NOT NULL,
    date DATE NOT NULL,
    
    -- Daily summary
    utilization_category VARCHAR(20),
    optimization_score DECIMAL(3,2),
    monthly_cost DECIMAL(10,2),
    potential_savings DECIMAL(10,2),
    
    -- Change tracking
    category_changed BOOLEAN DEFAULT FALSE,
    score_change DECIMAL(3,2) DEFAULT 0,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(resource_arn, date),
    FOREIGN KEY (resource_arn) REFERENCES yt_aws_resources(resource_arn) ON DELETE CASCADE
);

-- 5. Indexes for performance
CREATE INDEX IF NOT EXISTS idx_assessments_arn_timestamp ON yt_resource_assessments(resource_arn, assessment_timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_assessments_category ON yt_resource_assessments(utilization_category);
CREATE INDEX IF NOT EXISTS idx_assessments_score ON yt_resource_assessments(optimization_score DESC);
CREATE INDEX IF NOT EXISTS idx_history_arn_date ON yt_assessment_history(resource_arn, date DESC);

-- 6. Assessment functions with configurable thresholds
CREATE OR REPLACE FUNCTION classify_utilization(
    avg_cpu DECIMAL,
    max_cpu DECIMAL,
    avg_memory DECIMAL,
    max_memory DECIMAL,
    cpu_variance DECIMAL,
    tenant_id VARCHAR DEFAULT 'default'
) RETURNS VARCHAR(20) AS $$
DECLARE
    config RECORD;
BEGIN
    -- Get tenant configuration
    SELECT * INTO config FROM yt_assessment_config WHERE yt_assessment_config.tenant_id = classify_utilization.tenant_id;
    
    -- Batch workload pattern
    IF max_cpu > config.batch_high_threshold 
       AND avg_cpu < config.batch_low_threshold 
       AND cpu_variance > config.batch_variance_threshold THEN
        RETURN 'batch';
    END IF;
    
    -- Overutilized
    IF avg_cpu > config.overutilized_cpu_threshold 
       OR avg_memory > config.overutilized_memory_threshold THEN
        RETURN 'overutilized';
    END IF;
    
    -- Underutilized
    IF avg_cpu < config.underutilized_cpu_threshold 
       AND avg_memory < config.underutilized_memory_threshold THEN
        RETURN 'underutilized';
    END IF;
    
    -- Intermittent/Bursty
    IF cpu_variance > 20 THEN
        RETURN 'intermittent';
    END IF;
    
    -- Idle
    IF avg_cpu < 5 AND avg_memory < 10 THEN
        RETURN 'idle';
    END IF;
    
    RETURN 'normal';
END;
$$ LANGUAGE plpgsql;

-- 7. View for current resource status (user-facing)
CREATE OR REPLACE VIEW vw_current_resource_status AS
SELECT 
    r.resource_arn,
    r.instance_id,
    r.instance_type,
    r.region,
    r.environment,
    
    -- Latest assessment
    a.utilization_category,
    a.usage_pattern,
    a.optimization_score,
    a.recommended_action,
    a.recommended_instance_type,
    
    -- Cost information
    a.current_hourly_cost,
    a.projected_hourly_cost,
    a.potential_monthly_savings,
    
    -- Utilization summary
    a.avg_cpu_utilization,
    a.avg_memory_utilization,
    a.idle_percentage,
    a.peak_hours,
    
    -- Assessment metadata
    a.assessment_timestamp,
    a.confidence_score,
    
    -- Trend (compared to previous assessment)
    CASE 
        WHEN h.score_change > 0.1 THEN 'Improving'
        WHEN h.score_change < -0.1 THEN 'Degrading'
        ELSE 'Stable'
    END as trend
    
FROM yt_aws_resources r
LEFT JOIN LATERAL (
    SELECT * FROM yt_resource_assessments 
    WHERE resource_arn = r.resource_arn 
    ORDER BY assessment_timestamp DESC 
    LIMIT 1
) a ON true
LEFT JOIN LATERAL (
    SELECT * FROM yt_assessment_history 
    WHERE resource_arn = r.resource_arn 
    ORDER BY date DESC 
    LIMIT 1
) h ON true
WHERE r.sync_status = 'active';

-- 8. Function for user-configurable assessment timeline queries
CREATE OR REPLACE FUNCTION get_assessments_by_timeline(
    p_resource_arn VARCHAR DEFAULT NULL,
    p_start_date DATE DEFAULT CURRENT_DATE - INTERVAL '30 days',
    p_end_date DATE DEFAULT CURRENT_DATE,
    p_category VARCHAR DEFAULT NULL
) RETURNS TABLE(
    resource_arn VARCHAR,
    instance_id VARCHAR,
    date DATE,
    utilization_category VARCHAR,
    optimization_score DECIMAL,
    potential_savings DECIMAL
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        h.resource_arn,
        r.instance_id,
        h.date,
        h.utilization_category,
        h.optimization_score,
        h.potential_savings
    FROM yt_assessment_history h
    JOIN yt_aws_resources r ON h.resource_arn = r.resource_arn
    WHERE h.date BETWEEN p_start_date AND p_end_date
    AND (p_resource_arn IS NULL OR h.resource_arn = p_resource_arn)
    AND (p_category IS NULL OR h.utilization_category = p_category)
    ORDER BY h.date DESC, h.resource_arn;
END;
$$ LANGUAGE plpgsql;

-- 9. Caching mechanism for live data
CREATE TABLE IF NOT EXISTS yt_assessment_cache (
    cache_key VARCHAR(255) PRIMARY KEY,
    cache_data JSONB NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_cache_expires ON yt_assessment_cache(expires_at);

SELECT 'Optimized assessment-based schema created successfully' as status;