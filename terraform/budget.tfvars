# Budget-friendly configuration for $100 AWS credit
aws_region = "us-east-1"
environment = "yukti-demo"

# Reduced instance counts for budget
instance_count = 2
rds_count = 1

# Cost-saving options
enable_multi_az = false
enable_detailed_monitoring = true  # Keep for testing findings

# Tagging for cost tracking
cost_center = "finops-demo"
owner_email = "demo@yukti.com"