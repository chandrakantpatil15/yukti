#!/bin/bash

# Create Yukti Platform IAM User (simulates Yukti's AWS account)

set -e

USER_NAME="yukti-platform-user"
POLICY_NAME="YuktiAssumeRolePolicy"

echo "=========================================="
echo "Creating Yukti Platform IAM User"
echo "=========================================="

# Create IAM user
echo "Creating IAM user: $USER_NAME..."
aws iam create-user --user-name "$USER_NAME" 2>/dev/null || echo "⚠️  User already exists"

# Create policy to allow AssumeRole
cat > /tmp/yukti-assume-policy.json <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "sts:AssumeRole",
      "Resource": "arn:aws:iam::144403604430:role/YuktiTestReadOnlyRole"
    }
  ]
}
EOF

# Create policy
echo "Creating AssumeRole policy..."
POLICY_ARN=$(aws iam create-policy \
  --policy-name "$POLICY_NAME" \
  --policy-document file:///tmp/yukti-assume-policy.json \
  --query 'Policy.Arn' \
  --output text 2>/dev/null) || POLICY_ARN="arn:aws:iam::144403604430:policy/$POLICY_NAME"

echo "Policy ARN: $POLICY_ARN"

# Attach policy to user
echo "Attaching policy to user..."
aws iam attach-user-policy \
  --user-name "$USER_NAME" \
  --policy-arn "$POLICY_ARN" 2>/dev/null || echo "⚠️  Policy already attached"

# Create access key
echo "Creating access key..."
aws iam create-access-key --user-name "$USER_NAME" --query 'AccessKey.[AccessKeyId,SecretAccessKey]' --output text > /tmp/yukti-user-creds.txt 2>/dev/null || echo "⚠️  Access key may already exist"

if [ -f /tmp/yukti-user-creds.txt ]; then
  ACCESS_KEY=$(cat /tmp/yukti-user-creds.txt | awk '{print $1}')
  SECRET_KEY=$(cat /tmp/yukti-user-creds.txt | awk '{print $2}')
  
  echo ""
  echo "✅ Setup Complete!"
  echo ""
  echo "Update docker-compose.yml with these credentials:"
  echo "  AWS_ACCESS_KEY_ID: $ACCESS_KEY"
  echo "  AWS_SECRET_ACCESS_KEY: $SECRET_KEY"
  echo "  SKIP_AWS_VERIFICATION: false"
  echo ""
else
  echo ""
  echo "⚠️  Could not create new access key (may already exist)"
  echo "List existing keys: aws iam list-access-keys --user-name $USER_NAME"
fi
