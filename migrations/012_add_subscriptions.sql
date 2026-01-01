-- Add subscription management table
CREATE TABLE IF NOT EXISTS yt_subscriptions (
    id SERIAL PRIMARY KEY,
    tenant_id INTEGER NOT NULL REFERENCES yt_tenants(id) ON DELETE CASCADE,
    plan VARCHAR(50) DEFAULT 'trial',
    trial_days INTEGER DEFAULT 30,  -- Configurable trial period
    current_period_end TIMESTAMP DEFAULT NOW() + INTERVAL '30 days',
    grace_until TIMESTAMP,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(tenant_id)
);

-- Create index for fast lookups
CREATE INDEX IF NOT EXISTS idx_subscriptions_tenant ON yt_subscriptions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_active ON yt_subscriptions(is_active, current_period_end);

-- Configuration table for global settings
CREATE TABLE IF NOT EXISTS yt_platform_config (
    id SERIAL PRIMARY KEY,
    config_key VARCHAR(100) UNIQUE NOT NULL,
    config_value TEXT NOT NULL,
    description TEXT,
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Insert default configuration
INSERT INTO yt_platform_config (config_key, config_value, description) VALUES
('trial_days', '30', 'Default trial period in days'),
('grace_days', '7', 'Grace period after subscription expires'),
('max_users_per_tenant', '10', 'Maximum users per tenant on free plan'),
('max_aws_accounts', '1', 'Maximum AWS accounts on trial')
ON CONFLICT (config_key) DO NOTHING;

-- Auto-create subscription when tenant is created
CREATE OR REPLACE FUNCTION create_default_subscription()
RETURNS TRIGGER AS $$
DECLARE
    trial_days_config INTEGER;
BEGIN
    -- Get trial days from config
    SELECT config_value::INTEGER INTO trial_days_config
    FROM yt_platform_config
    WHERE config_key = 'trial_days';
    
    -- Default to 30 if not found
    IF trial_days_config IS NULL THEN
        trial_days_config := 30;
    END IF;
    
    INSERT INTO yt_subscriptions (tenant_id, plan, trial_days, current_period_end)
    VALUES (NEW.id, 'trial', trial_days_config, NOW() + (trial_days_config || ' days')::INTERVAL)
    ON CONFLICT (tenant_id) DO NOTHING;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_create_subscription ON yt_tenants;
CREATE TRIGGER trigger_create_subscription
AFTER INSERT ON yt_tenants
FOR EACH ROW
EXECUTE FUNCTION create_default_subscription();

-- Create subscriptions for existing tenants
INSERT INTO yt_subscriptions (tenant_id, plan, trial_days, current_period_end)
SELECT id, 'trial', 30, NOW() + INTERVAL '30 days'
FROM yt_tenants
WHERE id NOT IN (SELECT tenant_id FROM yt_subscriptions)
ON CONFLICT (tenant_id) DO NOTHING;
