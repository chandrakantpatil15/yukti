# Yukti Database Setup

## Quick Start

```bash
# Run the setup script
./scripts/setup-local-db.sh
```

## What the script does:

1. **Installs PostgreSQL + Redis** (if not already installed)
2. **Creates database and user**
3. **Sets up schema** with 4 main tables
4. **Inserts sample data** for testing

## Database Schema

### Tables Created:
- `aws_pricing` - Daily AWS pricing data (EC2, RDS, S3)
- `aws_resources` - Your AWS resource inventory 
- `resource_metrics` - Aggregated metrics from Prometheus
- `optimization_recommendations` - Generated cost savings suggestions

### Sample Data Included:
- **10 EC2 instance types** with pricing for us-east-1
- **8 sample EC2 instances** with different utilization patterns
- **CPU/Memory metrics** showing optimization opportunities
- **4 cost optimization recommendations** with potential savings

## Connection Details

After setup:
- **PostgreSQL**: `jdbc:postgresql://localhost:5432/yukti`
- **Username**: `yukti_user`
- **Password**: `yukti_pass`
- **Redis**: `localhost:6379`

## Testing the Setup

```bash
# Test PostgreSQL connection
psql -h localhost -U yukti_user -d yukti -c "SELECT COUNT(*) FROM aws_resources;"

# Test Redis connection  
redis-cli ping

# View sample recommendations
psql -h localhost -U yukti_user -d yukti -c "SELECT * FROM optimization_recommendations;"
```

## Sample Data Overview

The script creates realistic test data:
- **Over-provisioned instances** (low CPU usage) → Downsize recommendations
- **Stopped instances** → Termination recommendations  
- **Staging environments** → Scheduling recommendations
- **Production instances** → Properly sized (no recommendations)

This gives you immediate data to test cost optimization algorithms!