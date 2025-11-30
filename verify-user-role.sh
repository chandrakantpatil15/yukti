#!/bin/bash

echo "=========================================="
echo "Verifying IAM Role in User Account"
echo "=========================================="
echo ""
echo "User Account: 424851482219"
echo "Role Name: YuktiReadOnlyRole"
echo ""

# You need to run this with credentials for account 424851482219
# Option 1: Configure a profile
#   aws configure --profile user-account
#   Then run: aws iam get-role --role-name YuktiReadOnlyRole --profile user-account

# Option 2: Set environment variables
#   export AWS_ACCESS_KEY_ID=<user-account-key>
#   export AWS_SECRET_ACCESS_KEY=<user-account-secret>
#   Then run: aws iam get-role --role-name YuktiReadOnlyRole

echo "To verify the role exists, run ONE of these:"
echo ""
echo "Option 1 - Using AWS Console:"
echo "  1. Login to AWS Console for account 424851482219"
echo "  2. Go to IAM → Roles"
echo "  3. Search for 'YuktiReadOnlyRole'"
echo "  4. Click on it and check the Trust Policy tab"
echo ""
echo "Option 2 - Using AWS CLI:"
echo "  aws iam get-role --role-name YuktiReadOnlyRole --profile user-account"
echo ""
echo "The trust policy MUST be:"
cat <<'EOF'
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {
      "AWS": "arn:aws:iam::144403604430:user/yukti-platform-user"
    },
    "Action": "sts:AssumeRole",
    "Condition": {
      "StringLike": {
        "sts:ExternalId": "yukti-*"
      }
    }
  }]
}
EOF

echo ""
echo "=========================================="
echo "Common Issues:"
echo "=========================================="
echo "1. Role doesn't exist → Create it in AWS Console"
echo "2. Wrong trust policy → Update trust policy"
echo "3. Wrong account → Make sure you're in account 424851482219"
echo "4. Missing ReadOnlyAccess policy → Attach it to the role"
echo ""
