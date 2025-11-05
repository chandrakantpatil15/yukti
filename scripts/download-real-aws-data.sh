#!/bin/bash
set -e

echo "📥 Downloading real AWS cost data samples..."

# Create data directory
mkdir -p data/samples

# Download AWS Sample Cost & Usage Report
echo "Downloading AWS CUR sample data..."
curl -L "https://s3.amazonaws.com/cur-report-samples/sample-cur-report.csv" \
  -o data/samples/aws-cur-sample.csv

# Download Netflix cost optimization data (if available)
echo "Downloading Netflix cost patterns..."
curl -L "https://raw.githubusercontent.com/Netflix/security_monkey/develop/security_monkey/tests/test_data/aws_cost_sample.json" \
  -o data/samples/netflix-cost-sample.json || echo "Netflix data not available"

# Download Cloud Custodian resource samples
echo "Downloading Cloud Custodian resource data..."
curl -L "https://raw.githubusercontent.com/cloud-custodian/cloud-custodian/master/tests/data/ec2-instances.json" \
  -o data/samples/ec2-instances-sample.json || echo "Cloud Custodian data not available"

# Download AWS pricing data
echo "Downloading AWS pricing data..."
curl -L "https://pricing.us-east-1.amazonaws.com/offers/v1.0/aws/AmazonEC2/current/index.json" \
  -o data/samples/aws-ec2-pricing.json

echo "✅ Real AWS data downloaded to data/samples/"
echo "📊 Files available:"
ls -la data/samples/