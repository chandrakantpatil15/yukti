-- Add resources table to existing database
CREATE TABLE IF NOT EXISTS resources (
    id SERIAL PRIMARY KEY,
    resource_id VARCHAR(255) UNIQUE NOT NULL,
    instance_type VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL,
    region VARCHAR(50) NOT NULL,
    monthly_cost DECIMAL(10,2) NOT NULL,
    cpu_utilization INTEGER NOT NULL,
    environment VARCHAR(50) NOT NULL,
    project VARCHAR(100),
    owner VARCHAR(100),
    cost_center VARCHAR(50),
    backup_schedule VARCHAR(50),
    application VARCHAR(100),
    tags TEXT NOT NULL,
    associated_resources TEXT NOT NULL,
    security_compliance TEXT NOT NULL,
    billing_breakdown TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_resources_environment ON resources(environment);
CREATE INDEX IF NOT EXISTS idx_resources_status ON resources(status);
CREATE INDEX IF NOT EXISTS idx_resources_instance_type ON resources(instance_type);
CREATE INDEX IF NOT EXISTS idx_resources_project ON resources(project);
CREATE INDEX IF NOT EXISTS idx_resources_owner ON resources(owner);
CREATE INDEX IF NOT EXISTS idx_resources_cost_center ON resources(cost_center);