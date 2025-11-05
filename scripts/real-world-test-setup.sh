#!/bin/bash

# Real-World Testing: 20 Spot Instances with Stress Testing
# This creates realistic workload patterns for Yukti FinOps testing

echo "🚀 YUKTI FINOPS - REAL WORLD TEST SETUP"
echo "======================================"

# Configuration
INSTANCE_COUNT=20
INSTANCE_TYPE="t3.medium"  # Cost-effective for testing
REGION="us-east-1"
KEY_NAME="yukti-test-key"
SECURITY_GROUP="yukti-test-sg"

# Create security group if it doesn't exist
echo "🔒 Setting up security group..."
aws ec2 create-security-group \
    --group-name $SECURITY_GROUP \
    --description "Yukti FinOps Testing Security Group" \
    --region $REGION 2>/dev/null || echo "Security group already exists"

# Allow SSH access
aws ec2 authorize-security-group-ingress \
    --group-name $SECURITY_GROUP \
    --protocol tcp \
    --port 22 \
    --cidr 0.0.0.0/0 \
    --region $REGION 2>/dev/null || echo "SSH rule already exists"

# Create user data script for stress testing
cat > /tmp/stress-test-userdata.sh << 'EOF'
#!/bin/bash
yum update -y
yum install -y stress htop

# Install CloudWatch agent for detailed metrics
wget https://s3.amazonaws.com/amazoncloudwatch-agent/amazon_linux/amd64/latest/amazon-cloudwatch-agent.rpm
rpm -U ./amazon-cloudwatch-agent.rpm

# CloudWatch agent configuration
cat > /opt/aws/amazon-cloudwatch-agent/etc/amazon-cloudwatch-agent.json << 'CWCONFIG'
{
    "metrics": {
        "namespace": "YuktiFinOps/Testing",
        "metrics_collected": {
            "cpu": {
                "measurement": ["cpu_usage_idle", "cpu_usage_iowait", "cpu_usage_user", "cpu_usage_system"],
                "metrics_collection_interval": 60
            },
            "disk": {
                "measurement": ["used_percent"],
                "metrics_collection_interval": 60,
                "resources": ["*"]
            },
            "diskio": {
                "measurement": ["io_time"],
                "metrics_collection_interval": 60,
                "resources": ["*"]
            },
            "mem": {
                "measurement": ["mem_used_percent"],
                "metrics_collection_interval": 60
            },
            "netstat": {
                "measurement": ["tcp_established", "tcp_time_wait"],
                "metrics_collection_interval": 60
            },
            "swap": {
                "measurement": ["swap_used_percent"],
                "metrics_collection_interval": 60
            }
        }
    }
}
CWCONFIG

# Start CloudWatch agent
/opt/aws/amazon-cloudwatch-agent/bin/amazon-cloudwatch-agent-ctl \
    -a fetch-config -m ec2 -c file:/opt/aws/amazon-cloudwatch-agent/etc/amazon-cloudwatch-agent.json -s

# Create stress testing script
cat > /home/ec2-user/stress-pattern.sh << 'STRESS'
#!/bin/bash

# Yukti FinOps Stress Testing Pattern
# 10 min stress -> 5 min idle -> repeat

echo "🔥 Starting Yukti FinOps stress testing pattern..."

while true; do
    echo "$(date): Starting 10-minute stress test..."
    
    # High CPU stress (80-90% utilization)
    stress --cpu 2 --timeout 600s &
    
    # Memory stress (60-70% utilization)  
    stress --vm 1 --vm-bytes 512M --timeout 600s &
    
    # Light I/O stress
    stress --io 1 --timeout 600s &
    
    # Wait for stress tests to complete
    wait
    
    echo "$(date): Stress test completed. Starting 5-minute idle period..."
    sleep 300  # 5 minutes idle
    
    echo "$(date): Idle period completed. Restarting cycle..."
done
STRESS

chmod +x /home/ec2-user/stress-pattern.sh

# Start stress testing in background
nohup /home/ec2-user/stress-pattern.sh > /home/ec2-user/stress.log 2>&1 &

# Add tags for identification
INSTANCE_ID=$(curl -s http://169.254.169.254/latest/meta-data/instance-id)
aws ec2 create-tags --resources $INSTANCE_ID --tags \
    Key=Name,Value=YuktiFinOps-Test-$(hostname) \
    Key=Environment,Value=testing \
    Key=Project,Value=yukti-finops \
    Key=TestPattern,Value=stress-cycle \
    Key=Owner,Value=yukti-team

echo "✅ Stress testing setup completed on $(hostname)"
EOF

# Launch spot instances
echo "🚀 Launching $INSTANCE_COUNT spot instances..."

# Get latest Amazon Linux 2 AMI
AMI_ID=$(aws ec2 describe-images \
    --owners amazon \
    --filters "Name=name,Values=amzn2-ami-hvm-*-x86_64-gp2" \
    --query 'Images | sort_by(@, &CreationDate) | [-1].ImageId' \
    --output text \
    --region $REGION)

echo "📦 Using AMI: $AMI_ID"

# Create spot instance request
aws ec2 request-spot-instances \
    --spot-price "0.05" \
    --instance-count $INSTANCE_COUNT \
    --type "one-time" \
    --launch-specification "{
        \"ImageId\": \"$AMI_ID\",
        \"InstanceType\": \"$INSTANCE_TYPE\",
        \"KeyName\": \"$KEY_NAME\",
        \"SecurityGroups\": [\"$SECURITY_GROUP\"],
        \"UserData\": \"$(base64 -w 0 /tmp/stress-test-userdata.sh)\",
        \"IamInstanceProfile\": {
            \"Name\": \"CloudWatchAgentServerRole\"
        }
    }" \
    --region $REGION

echo "⏳ Waiting for instances to launch..."
sleep 60

# Show launched instances
echo "📊 Launched instances:"
aws ec2 describe-instances \
    --filters "Name=tag:Project,Values=yukti-finops" "Name=instance-state-name,Values=running" \
    --query 'Reservations[].Instances[].{InstanceId:InstanceId,InstanceType:InstanceType,State:State.Name,PublicIP:PublicIpAddress}' \
    --output table \
    --region $REGION

echo ""
echo "🎉 REAL-WORLD TEST SETUP COMPLETE!"
echo "================================="
echo "✅ $INSTANCE_COUNT spot instances launched"
echo "✅ Stress testing pattern: 10min stress → 5min idle → repeat"
echo "✅ CloudWatch detailed monitoring enabled"
echo "✅ Custom metrics namespace: YuktiFinOps/Testing"
echo ""
echo "📊 Expected workload patterns:"
echo "   • Batch/Intermittent: High CPU spikes then idle"
echo "   • Variable utilization: 80-90% → 0-5% cycles"
echo "   • Memory patterns: 60-70% → low usage"
echo "   • I/O patterns: Periodic disk activity"
echo ""
echo "⏰ Wait 30-60 minutes for sufficient CloudWatch data"
echo "🔍 Then run: make sync-all-data && make assess-daily"

# Cleanup script
cat > /tmp/cleanup-test-instances.sh << 'CLEANUP'
#!/bin/bash
echo "🧹 Cleaning up Yukti FinOps test instances..."
aws ec2 describe-instances \
    --filters "Name=tag:Project,Values=yukti-finops" "Name=instance-state-name,Values=running" \
    --query 'Reservations[].Instances[].InstanceId' \
    --output text | xargs -n1 aws ec2 terminate-instances --instance-ids
echo "✅ Cleanup initiated"
CLEANUP

chmod +x /tmp/cleanup-test-instances.sh
echo "🧹 Cleanup script created: /tmp/cleanup-test-instances.sh"