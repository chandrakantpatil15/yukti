#!/bin/bash

# Setup Test IAM Role for Yukti Platform Testing
# This simulates a customer giving Yukti read-only access to their AWS account

set -e

ACCOUNT_ID="144403604430"
ROLE_NAME="YuktiTestReadOnlyRole"
EXTERNAL_ID="yukti-test-12345"

echo "=========================================="
echo "Setting up Yukti Test Role"
echo "=========================================="
echo "Account ID: $ACCOUNT_ID"
echo "Role Name: $ROLE_NAME"
echo "External ID: $EXTERNAL_ID"
echo ""

# Create trust policy
cat > /tmp/yukti-trust-policy.json <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:aws:iam::${ACCOUNT_ID}:root"
      },
      "Action": "sts:AssumeRole",
      "Condition": {
        "StringEquals": {
          "sts:ExternalId": "${EXTERNAL_ID}"
        }
      }
    }
  ]
}
EOF

echo "✅ Trust policy created"
echo ""

# Create IAM role
echo "Creating IAM role: $ROLE_NAME..."
aws iam create-role \
  --role-name "$ROLE_NAME" \
  --assume-role-policy-document file:///tmp/yukti-trust-policy.json \
  --description "Test role for Yukti FinOps platform - Read-only access" \
  2>/dev/null || echo "⚠️  Role already exists, continuing..."

echo "✅ IAM role created/verified"
echo ""

# Attach ReadOnlyAccess policy
echo "Attaching ReadOnlyAccess policy..."
aws iam attach-role-policy \
  --role-name "$ROLE_NAME" \
  --policy-arn "arn:aws:iam::aws:policy/ReadOnlyAccess" \
  2>/dev/null || echo "⚠️  Policy already attached"

echo "✅ ReadOnlyAccess policy attached"
echo ""

# Wait for role to propagate
echo "Waiting 10 seconds for IAM role to propagate..."
sleep 10

# Test AssumeRole
echo "Testing AssumeRole..."
ROLE_ARN="arn:aws:iam::${ACCOUNT_ID}:role/${ROLE_NAME}"

aws sts assume-role \
  --role-arn "$ROLE_ARN" \
  --role-session-name "yukti-test-session" \
  --external-id "$EXTERNAL_ID" \
  --duration-seconds 900 \
  --query 'Credentials.[AccessKeyId,SecretAccessKey,SessionToken]' \
  --output text > /tmp/yukti-test-creds.txt

if [ $? -eq 0 ]; then
  echo "✅ AssumeRole test SUCCESSFUL!"
else
  echo "❌ AssumeRole test FAILED!"
  exit 1
fi

echo ""
echo "=========================================="
echo "✅ Setup Complete!"
echo "=========================================="
echo ""
echo "Role ARN: $ROLE_ARN"
echo "External ID: $EXTERNAL_ID"
echo ""
echo "Next steps:"
echo "1. Update docker-compose.yml: SKIP_AWS_VERIFICATION=false"
echo "2. Rebuild backend: docker-compose up -d --build backend"
echo "3. Test onboarding with:"
echo "   - Account ID: $ACCOUNT_ID"
echo "   - Role Name: $ROLE_NAME"
echo ""
echo "To delete this test role later:"
echo "  aws iam detach-role-policy --role-name $ROLE_NAME --policy-arn arn:aws:iam::aws:policy/ReadOnlyAccess"
echo "  aws iam delete-role --role-name $ROLE_NAME"
echo ""
