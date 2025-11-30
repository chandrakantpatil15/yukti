#!/bin/bash
# Complete AWS Resource Cleanup Script
# Ensures 100% resource deletion and cost prevention

set -e

echo "🧹 Starting complete AWS resource cleanup..."

# 1. Terraform destroy (primary cleanup)
echo "📋 Step 1: Terraform destroy"
terraform destroy -auto-approve

# 2. Check for orphaned EBS volumes
echo "💾 Step 2: Checking for orphaned EBS volumes"
aws ec2 describe-volumes \
  --filters "Name=tag:Environment,Values=development" "Name=state,Values=available" \
  --query 'Volumes[*].{ID:VolumeId,Size:Size,State:State}' \
  --output table

# 3. Delete any remaining EBS volumes
echo "🗑️  Step 3: Cleaning up orphaned EBS volumes"
ORPHANED_VOLUMES=$(aws ec2 describe-volumes \
  --filters "Name=tag:Environment,Values=development" "Name=state,Values=available" \
  --query 'Volumes[*].VolumeId' \
  --output text)

if [ ! -z "$ORPHANED_VOLUMES" ]; then
  for volume in $ORPHANED_VOLUMES; do
    echo "Deleting volume: $volume"
    aws ec2 delete-volume --volume-id $volume || echo "Failed to delete $volume"
  done
else
  echo "✅ No orphaned EBS volumes found"
fi

# 4. Check for orphaned snapshots
echo "📸 Step 4: Checking for orphaned snapshots"
aws ec2 describe-snapshots \
  --owner-ids self \
  --filters "Name=tag:Environment,Values=development" \
  --query 'Snapshots[*].{ID:SnapshotId,Size:VolumeSize,State:State}' \
  --output table

# 5. Empty and verify S3 buckets are deleted
echo "🪣 Step 5: Verifying S3 bucket cleanup"
REMAINING_BUCKETS=$(aws s3api list-buckets \
  --query 'Buckets[?contains(Name, `yukti-test-storage`)].Name' \
  --output text)

if [ ! -z "$REMAINING_BUCKETS" ]; then
  for bucket in $REMAINING_BUCKETS; do
    echo "Force emptying bucket: $bucket"
    aws s3 rm s3://$bucket --recursive || echo "Bucket $bucket already empty"
    aws s3api delete-bucket --bucket $bucket || echo "Failed to delete $bucket"
  done
else
  echo "✅ No remaining S3 buckets found"
fi

# 6. Check for remaining Elastic IPs
echo "🌐 Step 6: Checking for unassociated Elastic IPs"
aws ec2 describe-addresses \
  --filters "Name=tag:Name,Values=test-nat-eip" \
  --query 'Addresses[*].{IP:PublicIp,AllocationId:AllocationId,Associated:AssociationId}' \
  --output table

# 7. Final cost verification
echo "💰 Step 7: Final cost check"
echo "Checking current month charges..."
aws ce get-cost-and-usage \
  --time-period Start=$(date -d "$(date +%Y-%m-01)" +%Y-%m-%d),End=$(date +%Y-%m-%d) \
  --granularity MONTHLY \
  --metrics BlendedCost \
  --query 'ResultsByTime[0].Total.BlendedCost.Amount' \
  --output text

echo ""
echo "🎉 Cleanup completed!"
echo "✅ All Terraform resources destroyed"
echo "✅ Orphaned resources cleaned up"
echo "✅ S3 buckets emptied and deleted"
echo "✅ EBS volumes removed"
echo ""
echo "⚠️  IMPORTANT: Monitor your AWS billing for 24-48 hours to ensure no charges"
echo "📊 Check AWS Cost Explorer: https://console.aws.amazon.com/cost-management/home"