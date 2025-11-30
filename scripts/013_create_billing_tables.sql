-- Stripe billing integration tables
-- Migration: 013_create_billing_tables.sql

-- Billing customers (links tenants to Stripe customers)
CREATE TABLE IF NOT EXISTS yt_billing_customers (
    id SERIAL PRIMARY KEY,
    tenant_id INTEGER NOT NULL REFERENCES yt_tenants(id) ON DELETE CASCADE,
    stripe_customer_id VARCHAR(255) NOT NULL UNIQUE,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    tax_id VARCHAR(100), -- GSTIN, VAT, etc.
    billing_email TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(tenant_id)
);

CREATE INDEX idx_billing_customers_tenant ON yt_billing_customers(tenant_id);
CREATE INDEX idx_billing_customers_stripe ON yt_billing_customers(stripe_customer_id);

-- Stripe events (webhook event log)
CREATE TABLE IF NOT EXISTS yt_stripe_events (
    id SERIAL PRIMARY KEY,
    stripe_event_id VARCHAR(255) NOT NULL UNIQUE,
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    processed BOOLEAN NOT NULL DEFAULT false,
    processed_at TIMESTAMP WITH TIME ZONE,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_stripe_events_type ON yt_stripe_events(event_type);
CREATE INDEX idx_stripe_events_processed ON yt_stripe_events(processed);
CREATE INDEX idx_stripe_events_stripe_id ON yt_stripe_events(stripe_event_id);

-- Billing invoices
CREATE TABLE IF NOT EXISTS yt_billing_invoices (
    id SERIAL PRIMARY KEY,
    tenant_id INTEGER NOT NULL REFERENCES yt_tenants(id) ON DELETE CASCADE,
    stripe_invoice_id VARCHAR(255) NOT NULL UNIQUE,
    stripe_customer_id VARCHAR(255) NOT NULL,
    amount_cents INTEGER NOT NULL,
    tax_cents INTEGER NOT NULL DEFAULT 0,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    paid BOOLEAN NOT NULL DEFAULT false,
    paid_at TIMESTAMP WITH TIME ZONE,
    pdf_url TEXT,
    invoice_number VARCHAR(100),
    period_start TIMESTAMP WITH TIME ZONE,
    period_end TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_billing_invoices_tenant ON yt_billing_invoices(tenant_id);
CREATE INDEX idx_billing_invoices_stripe ON yt_billing_invoices(stripe_invoice_id);
CREATE INDEX idx_billing_invoices_paid ON yt_billing_invoices(paid);

-- Billing products and prices (Stripe price mapping)
CREATE TABLE IF NOT EXISTS yt_billing_products_prices (
    id SERIAL PRIMARY KEY,
    plan_slug VARCHAR(50) NOT NULL UNIQUE, -- free, professional, enterprise, financial
    stripe_price_id VARCHAR(255) NOT NULL UNIQUE,
    stripe_product_id VARCHAR(255),
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    amount_cents INTEGER NOT NULL,
    recurring_interval VARCHAR(20) NOT NULL, -- month, year
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_billing_products_plan ON yt_billing_products_prices(plan_slug);
CREATE INDEX idx_billing_products_stripe_price ON yt_billing_products_prices(stripe_price_id);

-- Subscriptions (active subscription tracking)
CREATE TABLE IF NOT EXISTS yt_billing_subscriptions (
    id SERIAL PRIMARY KEY,
    tenant_id INTEGER NOT NULL REFERENCES yt_tenants(id) ON DELETE CASCADE,
    stripe_subscription_id VARCHAR(255) NOT NULL UNIQUE,
    stripe_customer_id VARCHAR(255) NOT NULL,
    stripe_price_id VARCHAR(255) NOT NULL,
    plan_slug VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL, -- active, canceled, past_due, trialing, etc.
    current_period_start TIMESTAMP WITH TIME ZONE,
    current_period_end TIMESTAMP WITH TIME ZONE,
    cancel_at_period_end BOOLEAN NOT NULL DEFAULT false,
    canceled_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    UNIQUE(tenant_id)
);

CREATE INDEX idx_billing_subscriptions_tenant ON yt_billing_subscriptions(tenant_id);
CREATE INDEX idx_billing_subscriptions_stripe ON yt_billing_subscriptions(stripe_subscription_id);
CREATE INDEX idx_billing_subscriptions_status ON yt_billing_subscriptions(status);

-- Update trigger for updated_at
CREATE OR REPLACE FUNCTION update_billing_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_billing_customers_updated_at
    BEFORE UPDATE ON yt_billing_customers
    FOR EACH ROW
    EXECUTE FUNCTION update_billing_updated_at();

CREATE TRIGGER trigger_update_billing_products_updated_at
    BEFORE UPDATE ON yt_billing_products_prices
    FOR EACH ROW
    EXECUTE FUNCTION update_billing_updated_at();

CREATE TRIGGER trigger_update_billing_subscriptions_updated_at
    BEFORE UPDATE ON yt_billing_subscriptions
    FOR EACH ROW
    EXECUTE FUNCTION update_billing_updated_at();

-- Insert default price mappings (placeholders - update with actual Stripe price IDs)
INSERT INTO yt_billing_products_prices (plan_slug, stripe_price_id, stripe_product_id, currency, amount_cents, recurring_interval)
VALUES
    ('professional', 'price_placeholder_professional_monthly', 'prod_placeholder', 'USD', 9900, 'month'),
    ('enterprise', 'price_placeholder_enterprise_monthly', 'prod_placeholder', 'USD', 49900, 'month'),
    ('financial', 'price_placeholder_financial_monthly', 'prod_placeholder', 'USD', 199900, 'month')
ON CONFLICT (plan_slug) DO NOTHING;

SELECT 'Billing tables created successfully' as status;

