-- Customer onboarding tables

CREATE TABLE IF NOT EXISTS yt_customers (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) UNIQUE NOT NULL,
    company_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    onboarding_status VARCHAR(50) NOT NULL DEFAULT 'pending',
    current_step VARCHAR(50) NOT NULL DEFAULT 'aws_connection',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,
    INDEX idx_tenant_id (tenant_id),
    INDEX idx_onboarding_status (onboarding_status)
);

CREATE TABLE IF NOT EXISTS yt_aws_connections (
    tenant_id VARCHAR(36) PRIMARY KEY,
    account_id VARCHAR(20) NOT NULL,
    role_arn VARCHAR(255) NOT NULL,
    external_id VARCHAR(255) NOT NULL,
    regions JSON NOT NULL,
    verified BOOLEAN NOT NULL DEFAULT FALSE,
    last_verified_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (tenant_id) REFERENCES yt_customers(tenant_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS yt_metrics_integrations (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    source VARCHAR(50) NOT NULL,
    endpoint VARCHAR(500) NOT NULL,
    verified BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_tenant_id (tenant_id),
    FOREIGN KEY (tenant_id) REFERENCES yt_customers(tenant_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS yt_onboarding_events (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    event_data JSON,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_tenant_id (tenant_id),
    INDEX idx_event_type (event_type),
    FOREIGN KEY (tenant_id) REFERENCES yt_customers(tenant_id) ON DELETE CASCADE
);
