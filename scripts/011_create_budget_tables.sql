-- Budget tracking tables

CREATE TABLE IF NOT EXISTS yt_budgets (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    name VARCHAR(255) NOT NULL,
    amount DECIMAL(15,2) NOT NULL,
    period VARCHAR(20) NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE,
    alert_threshold DECIMAL(5,2) NOT NULL DEFAULT 80.00,
    current_spend DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_tenant_id (tenant_id),
    INDEX idx_status (status),
    FOREIGN KEY (tenant_id) REFERENCES yt_customers(tenant_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS yt_budget_alerts (
    id VARCHAR(36) PRIMARY KEY,
    budget_id VARCHAR(36) NOT NULL,
    tenant_id VARCHAR(36) NOT NULL,
    alert_type VARCHAR(50) NOT NULL,
    threshold DECIMAL(5,2) NOT NULL,
    current_spend DECIMAL(15,2) NOT NULL,
    message TEXT,
    triggered BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_budget_id (budget_id),
    INDEX idx_tenant_id (tenant_id),
    INDEX idx_triggered (triggered),
    FOREIGN KEY (budget_id) REFERENCES yt_budgets(id) ON DELETE CASCADE,
    FOREIGN KEY (tenant_id) REFERENCES yt_customers(tenant_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS yt_cost_data (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    date DATE NOT NULL,
    service VARCHAR(100) NOT NULL,
    cost DECIMAL(15,2) NOT NULL,
    usage_type VARCHAR(100),
    region VARCHAR(50),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_tenant_date (tenant_id, date),
    INDEX idx_service (service),
    FOREIGN KEY (tenant_id) REFERENCES yt_customers(tenant_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS yt_ri_recommendations (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    service VARCHAR(50) NOT NULL,
    instance_type VARCHAR(50) NOT NULL,
    region VARCHAR(50) NOT NULL,
    term VARCHAR(20) NOT NULL,
    payment_option VARCHAR(50) NOT NULL,
    estimated_monthly_cost DECIMAL(15,2) NOT NULL,
    estimated_savings DECIMAL(15,2) NOT NULL,
    recommended_quantity INT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_tenant_id (tenant_id),
    FOREIGN KEY (tenant_id) REFERENCES yt_customers(tenant_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS yt_sp_recommendations (
    id VARCHAR(36) PRIMARY KEY,
    tenant_id VARCHAR(36) NOT NULL,
    service VARCHAR(50) NOT NULL,
    term VARCHAR(20) NOT NULL,
    payment_option VARCHAR(50) NOT NULL,
    hourly_commitment DECIMAL(10,4) NOT NULL,
    estimated_monthly_cost DECIMAL(15,2) NOT NULL,
    estimated_savings DECIMAL(15,2) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_tenant_id (tenant_id),
    FOREIGN KEY (tenant_id) REFERENCES yt_customers(tenant_id) ON DELETE CASCADE
);
