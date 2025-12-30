-- Migration: Billing System
-- Description: Add billing management tables for customer subscriptions
-- Version: 011
-- Date: 2025-01-31

-- Billing records table
CREATE TABLE IF NOT EXISTS yt_billing (
    id SERIAL PRIMARY KEY,
    tenant_id INT NOT NULL REFERENCES yt_customers(tenant_id) ON DELETE CASCADE,
    plan VARCHAR(50) NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    status VARCHAR(20) DEFAULT 'pending',
    due_date DATE NOT NULL,
    paid_date DATE,
    invoice_url TEXT,
    notes TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_billing_tenant ON yt_billing(tenant_id);
CREATE INDEX idx_billing_status ON yt_billing(status);
CREATE INDEX idx_billing_due_date ON yt_billing(due_date);
CREATE INDEX idx_billing_created_at ON yt_billing(created_at);

-- Trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_billing_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER billing_updated_at_trigger
    BEFORE UPDATE ON yt_billing
    FOR EACH ROW
    EXECUTE FUNCTION update_billing_updated_at();

-- Insert sample billing data for testing
INSERT INTO yt_billing (tenant_id, plan, amount, status, due_date) VALUES
(18, 'pro', 299.00, 'paid', '2025-01-15'),
(18, 'pro', 299.00, 'pending', '2025-02-15'),
(18, 'pro', 299.00, 'pending', '2025-03-15');

-- Comments
COMMENT ON TABLE yt_billing IS 'Customer billing records and invoices';
COMMENT ON COLUMN yt_billing.tenant_id IS 'Reference to customer tenant';
COMMENT ON COLUMN yt_billing.plan IS 'Subscription plan (free, pro, enterprise)';
COMMENT ON COLUMN yt_billing.amount IS 'Billing amount in specified currency';
COMMENT ON COLUMN yt_billing.status IS 'Payment status (pending, paid, overdue, cancelled)';
COMMENT ON COLUMN yt_billing.due_date IS 'Payment due date';
COMMENT ON COLUMN yt_billing.paid_date IS 'Actual payment date';
COMMENT ON COLUMN yt_billing.invoice_url IS 'URL to generated invoice PDF';
