#!/bin/bash

EMAIL="chandrakantpatil1594@gmail.com"
REGION="us-east-1"

echo "🚀 Setting up AWS SES with Gmail..."

# Verify email
aws ses verify-email-identity --email-address $EMAIL --region $REGION

echo "✅ Verification email sent to $EMAIL"
echo "📧 Check your Gmail inbox and click the verification link!"
echo ""
read -p "Press Enter after clicking verification link..."

# Check status
STATUS=$(aws ses get-identity-verification-attributes \
  --identities $EMAIL \
  --region $REGION \
  --query "VerificationAttributes.\"$EMAIL\".VerificationStatus" \
  --output text)

if [ "$STATUS" = "Success" ]; then
  echo "✅ Email verified successfully!"
else
  echo "❌ Email not verified yet. Status: $STATUS"
  exit 1
fi

# Send test email
echo "📤 Sending test email..."
aws ses send-email \
  --from $EMAIL \
  --destination ToAddresses=$EMAIL \
  --message Subject={Data="Yukti Test",Charset=utf-8},Body={Text={Data="Test successful!",Charset=utf-8}} \
  --region $REGION

echo "✅ Test email sent! Check your Gmail."
echo "🔄 Restart backend: docker-compose restart backend"
