-- Create yt_aws_resources table for real AWS EC2 inventory
CREATE TABLE yt_aws_resources (
    id SERIAL PRIMARY KEY,
    instance_id VARCHAR(50) NOT NULL UNIQUE,
    instance_type VARCHAR(50) NOT NULL,
    region VARCHAR(50) NOT NULL,
    availability_zone VARCHAR(50),
    
    -- Instance state and details
    state VARCHAR(20) NOT NULL,
    platform VARCHAR(20) DEFAULT 'linux',
    architecture VARCHAR(20) DEFAULT 'x86_64',
    
    -- Lifecycle information
    launch_time TIMESTAMP WITH TIME ZONE,
    
    -- Organization and tagging
    environment VARCHAR(20),
    project_name VARCHAR(100),
    cost_center VARCHAR(50),
    owner VARCHAR(100),
    
    -- AWS tags as JSON
    tags JSONB DEFAULT '{}',
    
    -- Sync metadata
    last_synced TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    sync_status VARCHAR(20) DEFAULT 'active'
);

-- Create indexes
CREATE INDEX idx_instance_state ON yt_aws_resources(state);
CREATE INDEX idx_instance_type ON yt_aws_resources(instance_type);
CREATE INDEX idx_environment ON yt_aws_resources(environment);
CREATE INDEX idx_last_synced ON yt_aws_resources(last_synced);
CREATE INDEX idx_tags_gin ON yt_aws_resources USING GIN(tags);

SELECT 'yt_aws_resources table created successfully' as status;