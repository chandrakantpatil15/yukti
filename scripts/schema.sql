-- Yukti Database Schema
-- PostgreSQL schema for AWS cost optimization

-- AWS Pricing Table (Daily refresh)
CREATE TABLE IF NOT EXISTS aws_pricing (
    id SERIAL PRIMARY KEY,
    service VARCHAR(50) NOT NULL,
    region VARCHAR(20) NOT NULL,
    instance_type VARCHAR(50),
    pricing_unit VARCHAR(20) NOT NULL,
    price_per_unit DECIMAL(10,6) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    effective_date DATE NOT NULL DEFAULT CURRENT_DATE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(service, region, instance_type, pricing_unit, effective_date)
);

-- AWS Resources Inventory
CREATE TABLE IF NOT EXISTS aws_resources (
    id SERIAL PRIMARY KEY,
    resource_id VARCHAR(100) NOT NULL UNIQUE,
    resource_type VARCHAR(50) NOT NULL,
    service VARCHAR(50) NOT NULL,
    region VARCHAR(20) NOT NULL,
    instance_type VARCHAR(50),
    status VARCHAR(20) NOT NULL,
    launch_time TIMESTAMP,
    tags JSONB,
    monthly_cost DECIMAL(10,2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Resource Metrics (Aggregated from Prometheus)
CREATE TABLE IF NOT EXISTS resource_metrics (
    id SERIAL PRIMARY KEY,
    resource_id VARCHAR(100) NOT NULL,
    metric_name VARCHAR(50) NOT NULL,
    metric_value DECIMAL(10,4) NOT NULL,
    metric_timestamp TIMESTAMP NOT NULL,
    aggregation_period VARCHAR(10) DEFAULT '1h',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (resource_id) REFERENCES aws_resources(resource_id)
);

-- Cost Optimization Recommendations
CREATE TABLE IF NOT EXISTS optimization_recommendations (
    id SERIAL PRIMARY KEY,
    resource_id VARCHAR(100) NOT NULL,
    recommendation_type VARCHAR(50) NOT NULL,
    current_cost DECIMAL(10,2) NOT NULL,
    optimized_cost DECIMAL(10,2) NOT NULL,
    potential_savings DECIMAL(10,2) NOT NULL,
    confidence_score DECIMAL(3,2) NOT NULL,
    recommendation_details JSONB,
    status VARCHAR(20) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (resource_id) REFERENCES aws_resources(resource_id)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_pricing_lookup ON aws_pricing(service, region, instance_type);
CREATE INDEX IF NOT EXISTS idx_resources_type ON aws_resources(resource_type, status);
CREATE INDEX IF NOT EXISTS idx_metrics_resource_time ON resource_metrics(resource_id, metric_timestamp);
CREATE INDEX IF NOT EXISTS idx_recommendations_resource ON optimization_recommendations(resource_id, status);

-- Partitioning for metrics table (by month)
-- This will be implemented later for large datasets