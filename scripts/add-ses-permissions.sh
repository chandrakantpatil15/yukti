#!/bin/bash
# Add SES permissions to yukti-platform-user

echo "Adding SES permissions to yukti-platform-user..."

aws iam put-user-policy \
  --user-name yukti-platform-user \
  --policy-name YuktiSESPolicy \
  --policy-document '{
    "Version": "2012-10-17",
    "Statement": [
      {
        "Effect": "Allow",
        "Action": [
          "ses:SendEmail",
          "ses:SendRawEmail"
        ],
        "Resource": "*"
      }
    ]
  }'

echo "✅ SES permissions added successfully!"
echo ""
echo "Verify with:"
echo "aws iam list-user-policies --user-name yukti-platform-user"
