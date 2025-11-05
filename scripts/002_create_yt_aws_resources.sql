-- Create yt_aws_resources table for real AWS EC2 inventory
DROP TABLE IF EXISTS yt_aws_resources CASCADE;

CREATE TABLE yt_aws_resources (
    id SERIAL PRIMARY KEY,
    instance_id VARCHAR(50) NOT NULL UNIQUE,
    instance_type VARCHAR(50) NOT NULL,
    region VARCHAR(50) NOT NULL,
    availability_zone VARCHAR(50),
    
    -- Instance state and details
    state VARCHAR(20) NOT NULL, -- running, stopped, terminated, pending
    platform VARCHAR(20) DEFAULT 'linux',
    architecture VARCHAR(20) DEFAULT 'x86_64',
    
    -- Lifecycle information
    launch_time TIMESTAMP WITH TIME ZONE,
    uptime_hours INTEGER GENERATED ALWAYS AS (
        CASE 
            WHEN state = 'running' AND launch_time IS NOT NULL 
            THEN EXTRACT(EPOCH FROM (NOW() - launch_time))/3600 
            ELSE 0 
        END
    ) STORED,
    
    -- Organization and tagging
    environment VARCHAR(20), -- prod, staging, dev, test
    project_name VARCHAR(100),
    cost_center VARCHAR(50),
    owner VARCHAR(100),
    
    -- AWS tags as JSON for flexibility
    tags JSONB DEFAULT '{}',
    
    -- Sync metadata
    last_synced TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    sync_status VARCHAR(20) DEFAULT 'active', -- active, deleted, error
    
    -- Indexes for performance
    INDEX idx_instance_state (state),
    INDEX idx_instance_type (instance_type),
    INDEX idx_environment (environment),
    INDEX idx_last_synced (last_synced),
    INDEX idx_tags_gin (tags) USING GIN
);

-- Function to check if resource sync is needed (every hour)
CREATE OR REPLACE FUNCTION is_resource_sync_needed() RETURNS BOOLEAN AS $$
BEGIN
    RETURN NOT EXISTS (
        SELECT 1 FROM yt_aws_resources 
        WHERE last_synced > NOW() - INTERVAL '1 hour'
        AND sync_status = 'active'
        LIMIT 1
    );
END;
$$ LANGUAGE plpgsql;

-- View for active resources with pricing
CREATE OR REPLACE VIEW vw_active_resources_with_pricing AS
SELECT 
    r.*,
    p.on_demand_price_usd,
    p.spot_price_avg,
    p.reserved_1yr_no_upfront,
    p.vcpu,
    p.memory_gb,
    (p.on_demand_price_usd * 24 * 30) as estimated_monthly_cost
FROM yt_aws_resources r
LEFT JOIN yt_aws_pricing p ON r.instance_type = p.instance_type AND r.region = p.region
WHERE r.sync_status = 'active' AND r.state IN ('running', 'stopped');

SELECT 'yt_aws_resources table created successfully' as status;