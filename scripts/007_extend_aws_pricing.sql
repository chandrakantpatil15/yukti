-- Alter yt_aws_pricing table to support multiple AWS services
ALTER TABLE yt_aws_pricing ADD COLUMN IF NOT EXISTS service_code VARCHAR(50) NOT NULL DEFAULT 'AmazonEC2';
ALTER TABLE yt_aws_pricing ADD COLUMN IF NOT EXISTS resource_type VARCHAR(100);
ALTER TABLE yt_aws_pricing ADD COLUMN IF NOT EXISTS attributes JSONB DEFAULT '{}';
ALTER TABLE yt_aws_pricing ADD COLUMN IF NOT EXISTS pricing_unit VARCHAR(20) DEFAULT 'Hrs';
ALTER TABLE yt_aws_pricing ADD COLUMN IF NOT EXISTS pricing_currency VARCHAR(3) DEFAULT 'USD';

-- Drop existing unique constraint
ALTER TABLE yt_aws_pricing DROP CONSTRAINT IF EXISTS yt_aws_pricing_instance_type_region_os_key;

-- Add new unique constraint
ALTER TABLE yt_aws_pricing ADD CONSTRAINT yt_aws_pricing_unique_key 
    UNIQUE (service_code, resource_type, region, os);

-- Add indexes for performance
CREATE INDEX IF NOT EXISTS idx_yt_aws_pricing_service ON yt_aws_pricing(service_code);
CREATE INDEX IF NOT EXISTS idx_yt_aws_pricing_resource ON yt_aws_pricing(resource_type);
CREATE INDEX IF NOT EXISTS idx_yt_aws_pricing_attributes ON yt_aws_pricing USING gin (attributes);

-- Update existing EC2 data
UPDATE yt_aws_pricing 
SET resource_type = instance_type,
    attributes = jsonb_build_object(
        'instanceType', instance_type,
        'vcpu', vcpu,
        'memoryGb', memory_gb,
        'storage', storage,
        'networkPerformance', network_performance
    )
WHERE service_code = 'AmazonEC2';

-- Comments for documentation
COMMENT ON COLUMN yt_aws_pricing.service_code IS 'AWS service identifier (e.g., AmazonEC2, AmazonRDS)';
COMMENT ON COLUMN yt_aws_pricing.resource_type IS 'Type of resource within the service';
COMMENT ON COLUMN yt_aws_pricing.attributes IS 'Service-specific attributes in JSONB format';
COMMENT ON COLUMN yt_aws_pricing.pricing_unit IS 'Unit of measurement for pricing (e.g., Hrs, GB, Requests)';
COMMENT ON COLUMN yt_aws_pricing.pricing_currency IS 'Currency code for pricing (e.g., USD)';

SELECT 'yt_aws_pricing table updated successfully' as status;