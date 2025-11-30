package aws

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type RoleVerifier struct {
	cfg aws.Config
}

type VerificationResult struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	ErrorCode    string `json:"error_code,omitempty"`
	ErrorDetails string `json:"error_details,omitempty"`
}

func NewRoleVerifier(ctx context.Context) (*RoleVerifier, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &RoleVerifier{cfg: cfg}, nil
}

// VerifyRoleAccess attempts to assume the role and validates connectivity
func (v *RoleVerifier) VerifyRoleAccess(ctx context.Context, roleARN, externalID string) *VerificationResult {
	// Create STS client
	stsClient := sts.NewFromConfig(v.cfg)

	// Attempt to assume role
	input := &sts.AssumeRoleInput{
		RoleArn:         aws.String(roleARN),
		RoleSessionName: aws.String("yukti-verification"),
		ExternalId:      aws.String(externalID),
		DurationSeconds: aws.Int32(900), // 15 minutes minimum
	}

	output, err := stsClient.AssumeRole(ctx, input)
	if err != nil {
		return v.parseAssumeRoleError(err)
	}

	// Verify we got valid credentials
	if output.Credentials == nil {
		return &VerificationResult{
			Success:   false,
			Message:   "Failed to obtain credentials from AssumeRole",
			ErrorCode: "NO_CREDENTIALS",
		}
	}

	// Test the assumed role credentials by calling GetCallerIdentity
	assumedCfg := v.cfg.Copy()
	assumedCfg.Credentials = aws.NewCredentialsCache(
		stscreds.NewAssumeRoleProvider(stsClient, roleARN, func(o *stscreds.AssumeRoleOptions) {
			o.ExternalID = aws.String(externalID)
		}),
	)

	testSTS := sts.NewFromConfig(assumedCfg)
	identity, err := testSTS.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return &VerificationResult{
			Success:      false,
			Message:      "Role assumed but failed to verify identity",
			ErrorCode:    "IDENTITY_VERIFICATION_FAILED",
			ErrorDetails: err.Error(),
		}
	}

	return &VerificationResult{
		Success: true,
		Message: fmt.Sprintf("Successfully verified role access. Identity: %s", *identity.Arn),
	}
}

// parseAssumeRoleError provides user-friendly error messages
func (v *RoleVerifier) parseAssumeRoleError(err error) *VerificationResult {
	errStr := err.Error()

	// Access Denied - most common error
	if strings.Contains(errStr, "AccessDenied") || strings.Contains(errStr, "is not authorized") {
		return &VerificationResult{
			Success:      false,
			Message:      "Access denied. Please check the trust policy on your IAM role.",
			ErrorCode:    "ACCESS_DENIED",
			ErrorDetails: "The IAM role trust policy must allow Yukti's AWS account to assume the role. Ensure the trust policy includes the correct AWS account ID and external ID.",
		}
	}

	// Invalid External ID
	if strings.Contains(errStr, "ExternalId") || strings.Contains(errStr, "external id") {
		return &VerificationResult{
			Success:      false,
			Message:      "External ID mismatch. Please use the exact external ID provided by Yukti.",
			ErrorCode:    "INVALID_EXTERNAL_ID",
			ErrorDetails: "The external ID in your IAM role trust policy does not match the one provided. Copy the external ID exactly as shown.",
		}
	}

	// Role not found
	if strings.Contains(errStr, "NoSuchEntity") || strings.Contains(errStr, "does not exist") {
		return &VerificationResult{
			Success:      false,
			Message:      "IAM role not found. Please verify the role ARN is correct.",
			ErrorCode:    "ROLE_NOT_FOUND",
			ErrorDetails: "The role ARN format should be: arn:aws:iam::ACCOUNT_ID:role/ROLE_NAME",
		}
	}

	// Malformed ARN
	if strings.Contains(errStr, "InvalidParameterValue") || strings.Contains(errStr, "malformed") {
		return &VerificationResult{
			Success:      false,
			Message:      "Invalid role ARN format.",
			ErrorCode:    "INVALID_ARN",
			ErrorDetails: "The role ARN must be in the format: arn:aws:iam::123456789012:role/YourRoleName",
		}
	}

	// Network/timeout errors
	if strings.Contains(errStr, "timeout") || strings.Contains(errStr, "connection") {
		return &VerificationResult{
			Success:      false,
			Message:      "Network error. Please check your internet connection and try again.",
			ErrorCode:    "NETWORK_ERROR",
			ErrorDetails: errStr,
		}
	}

	// Generic error
	return &VerificationResult{
		Success:      false,
		Message:      "Failed to verify role access. Please check your AWS configuration.",
		ErrorCode:    "VERIFICATION_FAILED",
		ErrorDetails: errStr,
	}
}

// ValidateRoleARN performs basic validation on the role ARN format
func ValidateRoleARN(roleARN string) error {
	if roleARN == "" {
		return fmt.Errorf("role ARN cannot be empty")
	}

	if !strings.HasPrefix(roleARN, "arn:aws:iam::") {
		return fmt.Errorf("role ARN must start with 'arn:aws:iam::'")
	}

	if !strings.Contains(roleARN, ":role/") {
		return fmt.Errorf("role ARN must contain ':role/'")
	}

	parts := strings.Split(roleARN, ":")
	if len(parts) < 6 {
		return fmt.Errorf("invalid role ARN format")
	}

	return nil
}

// ValidateAccountID checks if the account ID is valid
func ValidateAccountID(accountID string) error {
	if accountID == "" {
		return fmt.Errorf("account ID cannot be empty")
	}

	if len(accountID) != 12 {
		return fmt.Errorf("account ID must be exactly 12 digits")
	}

	for _, c := range accountID {
		if c < '0' || c > '9' {
			return fmt.Errorf("account ID must contain only digits")
		}
	}

	return nil
}
