-- Create basic schema for Yukti FinOps
CREATE TABLE IF NOT EXISTS cost_centers (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    budget_usd DECIMAL(15,2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS projects (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    cost_center_id INTEGER REFERENCES cost_centers(id),
    owner_email VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS resources (
    id SERIAL PRIMARY KEY,
    resource_id VARCHAR(255) UNIQUE NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    instance_type VARCHAR(50),
    region VARCHAR(50),
    status VARCHAR(50),
    project_id INTEGER REFERENCES projects(id),
    environment VARCHAR(50),
    launch_time TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS resource_costs (
    id SERIAL PRIMARY KEY,
    resource_id INTEGER REFERENCES resources(id),
    date DATE NOT NULL,
    cost_usd DECIMAL(10,4) NOT NULL,
    usage_hours INTEGER,
    data_source VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS optimization_recommendations (
    id SERIAL PRIMARY KEY,
    resource_id INTEGER REFERENCES resources(id),
    recommendation_type VARCHAR(100) NOT NULL,
    current_cost DECIMAL(10,4),
    optimized_cost DECIMAL(10,4),
    potential_savings DECIMAL(10,4),
    confidence DECIMAL(3,2),
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS aws_pricings (
    id SERIAL PRIMARY KEY,
    instance_type VARCHAR(50) NOT NULL,
    region VARCHAR(50) NOT NULL,
    os VARCHAR(50) NOT NULL,
    price_per_hour DECIMAL(10,4),
    ri_1yr_no_upfront DECIMAL(10,4),
    ri_1yr_partial_upfront DECIMAL(10,4),
    spot_price_avg DECIMAL(10,4),
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(instance_type, region, os)
);

CREATE TABLE IF NOT EXISTS resource_metrics (
    id SERIAL PRIMARY KEY,
    resource_id INTEGER REFERENCES resources(id),
    timestamp TIMESTAMP NOT NULL,
    cpu_utilization DECIMAL(5,2),
    memory_utilization DECIMAL(5,2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS resource_tags (
    id SERIAL PRIMARY KEY,
    resource_id INTEGER REFERENCES resources(id),
    key VARCHAR(255) NOT NULL,
    value VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert basic data
INSERT INTO cost_centers (code, name, budget_usd) VALUES 
('ENG', 'Engineering', 50000.00),
('MKT', 'Marketing', 25000.00),
('OPS', 'Operations', 30000.00)
ON CONFLICT (code) DO NOTHING;

INSERT INTO projects (code, name, cost_center_id, owner_email) VALUES 
('WEB', 'Web Application', 1, 'eng-lead@company.com'),
('API', 'API Services', 1, 'api-team@company.com'),
('DATA', 'Data Pipeline', 3, 'data-team@company.com')
ON CONFLICT (code) DO NOTHING;