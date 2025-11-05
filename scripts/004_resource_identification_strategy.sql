-- Resource Identification Strategy for Log Correlation and Pricing

-- 1. Add resource identification fields to yt_aws_resources
ALTER TABLE yt_aws_resources ADD COLUMN IF NOT EXISTS private_ip VARCHAR(15);
ALTER TABLE yt_aws_resources ADD COLUMN IF NOT EXISTS public_ip VARCHAR(15);
ALTER TABLE yt_aws_resources ADD COLUMN IF NOT EXISTS private_dns VARCHAR(255);
ALTER TABLE yt_aws_resources ADD COLUMN IF NOT EXISTS public_dns VARCHAR(255);

-- 2. Create resource identification mapping table
CREATE TABLE IF NOT EXISTS yt_resource_identifiers (
    id SERIAL PRIMARY KEY,
    instance_id VARCHAR(50) NOT NULL,
    
    -- Network identifiers for log correlation
    private_ip VARCHAR(15),
    public_ip VARCHAR(15),
    hostname VARCHAR(255),
    
    -- Application identifiers
    application_name VARCHAR(100),
    service_name VARCHAR(100),
    container_id VARCHAR(64),
    
    -- Log source identifiers
    log_group_name VARCHAR(255),
    log_stream_prefix VARCHAR(255),
    
    -- Custom identifiers from tags
    custom_identifier VARCHAR(255),
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    FOREIGN KEY (instance_id) REFERENCES yt_aws_resources(instance_id) ON DELETE CASCADE
);

-- 3. Create pricing lookup view with resource correlation
CREATE OR REPLACE VIEW vw_resource_pricing_lookup AS
SELECT 
    r.instance_id,
    r.instance_type,
    r.region,
    r.state,
    r.private_ip,
    r.public_ip,
    r.environment,
    r.tags,
    
    -- Pricing information
    p.on_demand_price_usd as hourly_cost,
    p.spot_price_usd,
    p.reserved_1yr_price_usd,
    p.vcpus,
    p.memory_gb,
    p.storage_type,
    
    -- Cost calculations
    (p.on_demand_price_usd * 24) as daily_cost,
    (p.on_demand_price_usd * 24 * 30) as monthly_cost,
    
    -- Identifiers for log correlation
    ri.hostname,
    ri.application_name,
    ri.service_name,
    ri.log_group_name
    
FROM yt_aws_resources r
LEFT JOIN yt_aws_pricing p ON r.instance_type = p.instance_type AND r.region = p.region
LEFT JOIN yt_resource_identifiers ri ON r.instance_id = ri.instance_id
WHERE r.sync_status = 'active';

-- 4. Function to extract resource identifier from log entry
CREATE OR REPLACE FUNCTION extract_resource_from_log(
    log_message TEXT,
    source_ip TEXT DEFAULT NULL,
    hostname TEXT DEFAULT NULL
) RETURNS TABLE(instance_id VARCHAR, confidence_score INTEGER) AS $$
BEGIN
    -- Method 1: Direct instance ID in log
    IF log_message ~* 'i-[0-9a-f]{8,17}' THEN
        RETURN QUERY
        SELECT 
            (regexp_matches(log_message, '(i-[0-9a-f]{8,17})', 'i'))[1]::VARCHAR,
            95 as confidence_score;
        RETURN;
    END IF;
    
    -- Method 2: Match by IP address
    IF source_ip IS NOT NULL THEN
        RETURN QUERY
        SELECT r.instance_id, 85 as confidence_score
        FROM yt_aws_resources r 
        WHERE r.private_ip = source_ip OR r.public_ip = source_ip
        LIMIT 1;
        IF FOUND THEN RETURN; END IF;
    END IF;
    
    -- Method 3: Match by hostname
    IF hostname IS NOT NULL THEN
        RETURN QUERY
        SELECT ri.instance_id, 75 as confidence_score
        FROM yt_resource_identifiers ri
        WHERE ri.hostname = hostname
        LIMIT 1;
        IF FOUND THEN RETURN; END IF;
    END IF;
    
    -- Method 4: Match by application/service name in log
    RETURN QUERY
    SELECT ri.instance_id, 60 as confidence_score
    FROM yt_resource_identifiers ri
    WHERE ri.application_name IS NOT NULL 
    AND log_message ~* ri.application_name
    LIMIT 1;
END;
$$ LANGUAGE plpgsql;

-- 5. Create log entries table for cost correlation
CREATE TABLE IF NOT EXISTS yt_log_entries (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    instance_id VARCHAR(50),
    
    -- Log identification
    log_source VARCHAR(50), -- cloudwatch, application, system
    log_group VARCHAR(255),
    log_stream VARCHAR(255),
    
    -- Log content
    message TEXT,
    level VARCHAR(10), -- ERROR, WARN, INFO, DEBUG
    
    -- Resource correlation
    source_ip VARCHAR(15),
    hostname VARCHAR(255),
    confidence_score INTEGER DEFAULT 0,
    
    -- Cost attribution
    attributed_cost DECIMAL(10,4), -- Cost attributed to this log entry
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    FOREIGN KEY (instance_id) REFERENCES yt_aws_resources(instance_id) ON DELETE SET NULL
);

-- 6. Indexes for log correlation performance
CREATE INDEX IF NOT EXISTS idx_log_timestamp ON yt_log_entries(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_log_instance_id ON yt_log_entries(instance_id);
CREATE INDEX IF NOT EXISTS idx_log_source_ip ON yt_log_entries(source_ip);
CREATE INDEX IF NOT EXISTS idx_resource_private_ip ON yt_aws_resources(private_ip);
CREATE INDEX IF NOT EXISTS idx_resource_public_ip ON yt_aws_resources(public_ip);

-- 7. View for cost analysis with log correlation
CREATE OR REPLACE VIEW vw_resource_cost_analysis AS
SELECT 
    r.instance_id,
    r.instance_type,
    r.state,
    r.environment,
    
    -- Pricing
    p.hourly_cost,
    p.monthly_cost,
    
    -- Log activity (last 24 hours)
    COUNT(l.id) as log_entries_24h,
    COUNT(CASE WHEN l.level = 'ERROR' THEN 1 END) as error_count_24h,
    
    -- Resource utilization correlation
    CASE 
        WHEN COUNT(l.id) = 0 THEN 'No Activity'
        WHEN COUNT(l.id) < 10 THEN 'Low Activity'
        WHEN COUNT(l.id) < 100 THEN 'Medium Activity'
        ELSE 'High Activity'
    END as activity_level,
    
    -- Cost efficiency score
    CASE 
        WHEN COUNT(l.id) = 0 AND p.hourly_cost > 0.1 THEN 0.1 -- Idle expensive resource
        WHEN COUNT(l.id) > 0 THEN LEAST(1.0, COUNT(l.id)::DECIMAL / 100) -- Activity-based score
        ELSE 0.5
    END as efficiency_score

FROM vw_resource_pricing_lookup p
LEFT JOIN yt_log_entries l ON p.instance_id = l.instance_id 
    AND l.timestamp >= NOW() - INTERVAL '24 hours'
GROUP BY r.instance_id, r.instance_type, r.state, r.environment, p.hourly_cost, p.monthly_cost;

-- 8. Function to update resource identifiers from AWS metadata
CREATE OR REPLACE FUNCTION update_resource_identifiers() RETURNS INTEGER AS $$
DECLARE
    resource_count INTEGER := 0;
BEGIN
    -- Insert or update resource identifiers
    INSERT INTO yt_resource_identifiers (
        instance_id, private_ip, public_ip, hostname, 
        application_name, service_name, custom_identifier
    )
    SELECT 
        r.instance_id,
        r.private_ip,
        r.public_ip,
        r.private_dns,
        r.tags->>'Application' as application_name,
        r.tags->>'Service' as service_name,
        COALESCE(r.tags->>'Name', r.instance_id) as custom_identifier
    FROM yt_aws_resources r
    WHERE r.sync_status = 'active'
    ON CONFLICT (instance_id) DO UPDATE SET
        private_ip = EXCLUDED.private_ip,
        public_ip = EXCLUDED.public_ip,
        hostname = EXCLUDED.hostname,
        application_name = EXCLUDED.application_name,
        service_name = EXCLUDED.service_name,
        custom_identifier = EXCLUDED.custom_identifier,
        updated_at = NOW();
    
    GET DIAGNOSTICS resource_count = ROW_COUNT;
    RETURN resource_count;
END;
$$ LANGUAGE plpgsql;

SELECT 'Resource identification strategy created successfully' as status;