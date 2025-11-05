-- Create yt_aws_pricing table for AWS Pricing API data with 24-hour caching
DROP TABLE IF EXISTS yt_aws_pricing CASCADE;

CREATE TABLE yt_aws_pricing (
    id SERIAL PRIMARY KEY,
    instance_type VARCHAR(50) NOT NULL,
    region VARCHAR(50) NOT NULL,
    os VARCHAR(20) NOT NULL DEFAULT 'Linux',
    
    -- Instance specifications
    vcpu INTEGER,
    memory_gb DECIMAL(8,2),
    storage VARCHAR(100),
    network_performance VARCHAR(50),
    
    -- Pricing data from AWS Pricing API
    on_demand_price_usd DECIMAL(10,6) NOT NULL,
    reserved_1yr_no_upfront DECIMAL(10,6),
    reserved_1yr_partial_upfront DECIMAL(10,6),
    reserved_3yr_no_upfront DECIMAL(10,6),
    spot_price_avg DECIMAL(10,6),
    
    -- 24-hour cache management
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    is_active BOOLEAN DEFAULT true,
    
    -- Unique constraint
    UNIQUE(instance_type, region, os)
);

-- Indexes for performance
CREATE INDEX idx_yt_aws_pricing_instance_type ON yt_aws_pricing(instance_type);
CREATE INDEX idx_yt_aws_pricing_region ON yt_aws_pricing(region);
CREATE INDEX idx_yt_aws_pricing_cache ON yt_aws_pricing(last_updated) WHERE is_active = true;

-- Cache management function
CREATE OR REPLACE FUNCTION is_pricing_cache_expired() RETURNS BOOLEAN AS $$
BEGIN
    RETURN NOT EXISTS (
        SELECT 1 FROM yt_aws_pricing 
        WHERE last_updated > NOW() - INTERVAL '24 hours'
        AND is_active = true
        LIMIT 1
    );
END;
$$ LANGUAGE plpgsql;

SELECT 'yt_aws_pricing table created successfully' as status;