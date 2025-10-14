-- Sample data for Yukti cost optimization testing

-- Insert EC2 pricing data for US-East-1
INSERT INTO aws_pricing (service, region, instance_type, pricing_unit, price_per_unit) VALUES
('EC2', 'us-east-1', 't3.nano', 'hour', 0.0052),
('EC2', 'us-east-1', 't3.micro', 'hour', 0.0104),
('EC2', 'us-east-1', 't3.small', 'hour', 0.0208),
('EC2', 'us-east-1', 't3.medium', 'hour', 0.0416),
('EC2', 'us-east-1', 't3.large', 'hour', 0.0832),
('EC2', 'us-east-1', 't3.xlarge', 'hour', 0.1664),
('EC2', 'us-east-1', 'm5.large', 'hour', 0.096),
('EC2', 'us-east-1', 'm5.xlarge', 'hour', 0.192),
('EC2', 'us-east-1', 'c5.large', 'hour', 0.085),
('EC2', 'us-east-1', 'c5.xlarge', 'hour', 0.17)
ON CONFLICT (service, region, instance_type, pricing_unit, effective_date) DO NOTHING;

-- Insert sample AWS resources
INSERT INTO aws_resources (resource_id, resource_type, service, region, instance_type, status, launch_time, tags, monthly_cost) VALUES
('i-1234567890abcdef0', 'instance', 'EC2', 'us-east-1', 't3.medium', 'running', '2024-01-15 10:30:00', '{"Environment": "production", "Team": "backend"}', 30.37),
('i-0987654321fedcba0', 'instance', 'EC2', 'us-east-1', 't3.large', 'running', '2024-01-10 14:20:00', '{"Environment": "staging", "Team": "frontend"}', 60.74),
('i-1111222233334444', 'instance', 'EC2', 'us-east-1', 'm5.xlarge', 'running', '2024-01-05 09:15:00', '{"Environment": "production", "Team": "data"}', 140.16),
('i-5555666677778888', 'instance', 'EC2', 'us-east-1', 't3.small', 'stopped', '2024-01-20 16:45:00', '{"Environment": "development", "Team": "qa"}', 0.00),
('i-9999aaaabbbbcccc', 'instance', 'EC2', 'us-east-1', 'c5.large', 'running', '2024-01-12 11:30:00', '{"Environment": "production", "Team": "api"}', 62.05),
('i-ddddeeeeffffgggg', 'instance', 'EC2', 'us-east-1', 't3.micro', 'running', '2024-01-25 08:00:00', '{"Environment": "development", "Team": "mobile"}', 7.59),
('i-hhhhjjjjkkkkllll', 'instance', 'EC2', 'us-east-1', 't3.xlarge', 'running', '2024-01-08 13:45:00', '{"Environment": "staging", "Team": "ml"}', 121.49),
('i-mmmmnnnnoooopp', 'instance', 'EC2', 'us-east-1', 't3.nano', 'stopped', '2024-01-30 07:20:00', '{"Environment": "testing", "Team": "devops"}', 0.00)
ON CONFLICT (resource_id) DO NOTHING;

-- Insert sample resource metrics (CPU utilization)
INSERT INTO resource_metrics (resource_id, metric_name, metric_value, metric_timestamp) VALUES
-- High utilization instances (good candidates)
('i-1234567890abcdef0', 'cpu_utilization', 75.5, NOW() - INTERVAL '1 hour'),
('i-1234567890abcdef0', 'cpu_utilization', 78.2, NOW() - INTERVAL '2 hours'),
('i-1234567890abcdef0', 'cpu_utilization', 72.8, NOW() - INTERVAL '3 hours'),

-- Low utilization instances (optimization candidates)
('i-0987654321fedcba0', 'cpu_utilization', 15.2, NOW() - INTERVAL '1 hour'),
('i-0987654321fedcba0', 'cpu_utilization', 12.8, NOW() - INTERVAL '2 hours'),
('i-0987654321fedcba0', 'cpu_utilization', 18.5, NOW() - INTERVAL '3 hours'),

-- Very low utilization (downsize candidates)
('i-1111222233334444', 'cpu_utilization', 8.5, NOW() - INTERVAL '1 hour'),
('i-1111222233334444', 'cpu_utilization', 6.2, NOW() - INTERVAL '2 hours'),
('i-1111222233334444', 'cpu_utilization', 9.8, NOW() - INTERVAL '3 hours'),

-- Good utilization
('i-9999aaaabbbbcccc', 'cpu_utilization', 65.4, NOW() - INTERVAL '1 hour'),
('i-9999aaaabbbbcccc', 'cpu_utilization', 68.9, NOW() - INTERVAL '2 hours'),
('i-9999aaaabbbbcccc', 'cpu_utilization', 62.1, NOW() - INTERVAL '3 hours'),

-- Development instance (low usage expected)
('i-ddddeeeeffffgggg', 'cpu_utilization', 25.3, NOW() - INTERVAL '1 hour'),
('i-ddddeeeeffffgggg', 'cpu_utilization', 22.7, NOW() - INTERVAL '2 hours'),
('i-ddddeeeeffffgggg', 'cpu_utilization', 28.1, NOW() - INTERVAL '3 hours')
ON CONFLICT DO NOTHING;

-- Insert sample optimization recommendations
INSERT INTO optimization_recommendations (resource_id, recommendation_type, current_cost, optimized_cost, potential_savings, confidence_score, recommendation_details) VALUES
('i-0987654321fedcba0', 'downsize', 60.74, 30.37, 30.37, 0.85, '{"from": "t3.large", "to": "t3.medium", "reason": "Low CPU utilization (avg 15.5%)", "risk": "low"}'),
('i-1111222233334444', 'downsize', 140.16, 60.74, 79.42, 0.92, '{"from": "m5.xlarge", "to": "t3.large", "reason": "Very low CPU utilization (avg 8.2%)", "risk": "low"}'),
('i-5555666677778888', 'terminate', 15.18, 0.00, 15.18, 0.95, '{"reason": "Instance stopped for 10+ days", "action": "terminate_unused", "risk": "medium"}'),
('i-hhhhjjjjkkkkllll', 'schedule', 121.49, 60.75, 60.74, 0.78, '{"reason": "Staging environment", "schedule": "9AM-6PM weekdays", "savings": "50%", "risk": "low"}')
ON CONFLICT DO NOTHING;

-- Insert memory utilization metrics
INSERT INTO resource_metrics (resource_id, metric_name, metric_value, metric_timestamp) VALUES
('i-1234567890abcdef0', 'memory_utilization', 68.5, NOW() - INTERVAL '1 hour'),
('i-0987654321fedcba0', 'memory_utilization', 25.2, NOW() - INTERVAL '1 hour'),
('i-1111222233334444', 'memory_utilization', 15.8, NOW() - INTERVAL '1 hour'),
('i-9999aaaabbbbcccc', 'memory_utilization', 72.4, NOW() - INTERVAL '1 hour'),
('i-ddddeeeeffffgggg', 'memory_utilization', 45.3, NOW() - INTERVAL '1 hour')
ON CONFLICT DO NOTHING;