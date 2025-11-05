#!/bin/bash

# Real-World Testing: 20 Spot Instances with Stress Testing (Fixed)
echo "🚀 YUKTI FINOPS - REAL WORLD TEST SETUP (FIXED)"
echo "=============================================="

# Configuration
INSTANCE_COUNT=5  # Start with 5 for testing
INSTANCE_TYPE="t3.micro"  # Free tier eligible
REGION="us-east-1"
SECURITY_GROUP="yukti-test-sg"

# Create security group if it doesn't exist
echo "🔒 Setting up security group..."
aws ec2 create-security-group \
    --group-name $SECURITY_GROUP \
    --description "Yukti FinOps Testing Security Group" \
    --region $REGION 2>/dev/null || echo "Security group already exists"

# Create user data script (base64 encoded inline)
USER_DATA=$(cat << 'EOF' | base64
#!/bin/bash
yum update -y
yum install -y stress htop

# Simple stress testing script
cat > /home/ec2-user/stress-pattern.sh << 'STRESS'
#!/bin/bash
echo "🔥 Starting Yukti FinOps stress testing..."
while true; do
    echo "$(date): Starting 10-minute stress test..."
    stress --cpu 1 --timeout 600s &
    wait
    echo "$(date): Starting 5-minute idle period..."
    sleep 300
done
STRESS

chmod +x /home/ec2-user/stress-pattern.sh
nohup /home/ec2-user/stress-pattern.sh > /home/ec2-user/stress.log 2>&1 &

# Tag instance
INSTANCE_ID=$(curl -s http://169.254.169.254/latest/meta-data/instance-id)
aws ec2 create-tags --resources $INSTANCE_ID --tags \
    Key=Name,Value=YuktiFinOps-Test \
    Key=Project,Value=yukti-finops \
    Key=Environment,Value=testing

echo "✅ Setup completed"
EOF
)

# Get latest Amazon Linux 2 AMI
AMI_ID=$(aws ec2 describe-images \
    --owners amazon \
    --filters "Name=name,Values=amzn2-ami-hvm-*-x86_64-gp2" \
    --query 'Images | sort_by(@, &CreationDate) | [-1].ImageId' \
    --output text \
    --region $REGION)

echo "📦 Using AMI: $AMI_ID"

# Launch regular instances (not spot for simplicity)
echo "🚀 Launching $INSTANCE_COUNT instances..."

aws ec2 run-instances \
    --image-id $AMI_ID \
    --count $INSTANCE_COUNT \
    --instance-type $INSTANCE_TYPE \
    --security-groups $SECURITY_GROUP \
    --user-data "$USER_DATA" \
    --tag-specifications "ResourceType=instance,Tags=[{Key=Name,Value=YuktiFinOps-Test},{Key=Project,Value=yukti-finops},{Key=Environment,Value=testing}]" \
    --region $REGION

echo "⏳ Waiting for instances to launch..."
sleep 30

# Show launched instances
echo "📊 Launched instances:"
aws ec2 describe-instances \
    --filters "Name=tag:Project,Values=yukti-finops" "Name=instance-state-name,Values=running,pending" \
    --query 'Reservations[].Instances[].{InstanceId:InstanceId,InstanceType:InstanceType,State:State.Name,PublicIP:PublicIpAddress}' \
    --output table \
    --region $REGION

echo ""
echo "🎉 REAL-WORLD TEST SETUP COMPLETE!"
echo "================================="
echo "✅ $INSTANCE_COUNT instances launched"
echo "✅ Stress testing pattern: 10min stress → 5min idle → repeat"
echo ""
echo "⏰ Wait 15-30 minutes for CloudWatch data"
echo "🔍 Then run: make sync-all-data && make assess-daily"

# Create cleanup script
cat > /tmp/cleanup-test-instances.sh << 'CLEANUP'
#!/bin/bash
echo "🧹 Cleaning up Yukti FinOps test instances..."
INSTANCE_IDS=$(aws ec2 describe-instances \
    --filters "Name=tag:Project,Values=yukti-finops" "Name=instance-state-name,Values=running" \
    --query 'Reservations[].Instances[].InstanceId' \
    --output text)

if [ -n "$INSTANCE_IDS" ]; then
    aws ec2 terminate-instances --instance-ids $INSTANCE_IDS
    echo "✅ Cleanup initiated for instances: $INSTANCE_IDS"
else
    echo "ℹ️  No running test instances found"
fi
CLEANUP

chmod +x /tmp/cleanup-test-instances.sh
echo "🧹 Cleanup script created: /tmp/cleanup-test-instances.sh"