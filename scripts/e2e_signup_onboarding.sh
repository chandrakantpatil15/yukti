#!/usr/bin/env bash
set -euo pipefail

# E2E test: Sign up -> Verify -> Configure AWS (skip verification) -> Check Dashboard
# Usage: ./scripts/e2e_signup_onboarding.sh [api_base]
# Default API base: http://localhost:8081

API_BASE=${1:-http://localhost:8081}
EMAIL="e2e+test$(date +%s)@example.com"
PASSWORD="TestPassw0rd!"
COMPANY_NAME="E2E Corp"

echo "Using API base: $API_BASE"
echo "Creating test user: $EMAIL"

# 1) Signup
signup_resp=$(curl -s -X POST "$API_BASE/api/v1/auth/signup" \
  -H "Content-Type: application/json" \
  -d "{\"email\": \"$EMAIL\", \"password\": \"$PASSWORD\", \"company_name\": \"$COMPANY_NAME\"}")

echo "Signup response: $signup_resp"

otp=$(echo "$signup_resp" | jq -r '.otp_code // empty')
tenant_id=$(echo "$signup_resp" | jq -r '.tenant_id // empty')
user_id=$(echo "$signup_resp" | jq -r '.user_id // empty')

if [ -z "$otp" ]; then
  echo "OTP not returned in signup response. Ensure server is running with JWT_SECRET unset for dev mode or check logs for OTP. Exiting."
  exit 1
fi

echo "Received OTP: $otp"
echo "Tenant ID: $tenant_id, User ID: $user_id"

# 2) Verify email
verify_resp=$(curl -s -X POST "$API_BASE/api/v1/auth/verify-email" \
  -H "Content-Type: application/json" \
  -d "{\"email\": \"$EMAIL\", \"code\": \"$otp\"}")

echo "Verify response: $verify_resp"

token=$(echo "$verify_resp" | jq -r '.token // empty')

if [ -z "$token" ]; then
  echo "Verification failed or token not returned. Exiting."
  exit 1
fi

echo "Received token: ${token:0:20}..."

# 3) Configure AWS onboarding (skip verification)
# Ensure SKIP_AWS_VERIFICATION=true is set in server env or set external_id/role to valid format.

ACCOUNT_ID="123456789012"
ROLE_ARN="arn:aws:iam::123456789012:role/YuktiDemoRole"

echo "Configuring AWS (tenant: $tenant_id)"
config_resp=$(curl -s -X POST "$API_BASE/api/onboarding/aws" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $token" \
  -d "{\"tenant_id\": \"$tenant_id\", \"account_id\": \"$ACCOUNT_ID\", \"role_arn\": \"$ROLE_ARN\", \"external_id\": \"\", \"regions\": [\"us-east-1\", \"us-west-2\"]}")

echo "Configure AWS response: $config_resp"

# 4) Check onboarding status
status_resp=$(curl -s -X GET "$API_BASE/api/onboarding/status?tenant_id=$tenant_id" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $token")

echo "Onboarding status: $status_resp"

# 5) Call dashboard (tenant scoped)
dashboard_resp=$(curl -s -X GET "$API_BASE/api/customers/dashboard" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $token")

echo "Dashboard response: $dashboard_resp"

# Summary
echo "\nE2E Summary"
echo "Email: $EMAIL"
echo "Tenant ID: $tenant_id"
echo "User ID: $user_id"
echo "Token (truncated): ${token:0:20}..."

echo "Done."
