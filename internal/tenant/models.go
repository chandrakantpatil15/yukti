package tenant

import "time"

type SubscriptionTier string

const (
	TierFree         SubscriptionTier = "FREE"
	TierProfessional SubscriptionTier = "PROFESSIONAL"
	TierEnterprise   SubscriptionTier = "ENTERPRISE"
)

type Tenant struct {
	ID               int              `json:"id"`
	TenantCode       string           `json:"tenant_code"`
	CompanyName      string           `json:"company_name"`
	SubscriptionTier SubscriptionTier `json:"subscription_tier"`
	Status           string           `json:"status"`
	CreatedAt        time.Time        `json:"created_at"`
	TrialEndsAt      *time.Time       `json:"trial_ends_at,omitempty"`
}

type AWSAccount struct {
	ID           int       `json:"id"`
	TenantID     int       `json:"tenant_id"`
	AccountID    string    `json:"account_id"`
	AccountName  string    `json:"account_name"`
	RoleARN      string    `json:"role_arn"`
	ExternalID   string    `json:"external_id"`
	Status       string    `json:"status"`
	LastSync     time.Time `json:"last_sync"`
	ErrorMessage string    `json:"error_message,omitempty"`
}

type OnboardingRequest struct {
	CompanyName string `json:"company_name" validate:"required"`
	Email       string `json:"email" validate:"required,email"`
	AWSAccounts []struct {
		AccountID   string `json:"account_id" validate:"required,len=12"`
		AccountName string `json:"account_name"`
	} `json:"aws_accounts" validate:"required,min=1"`
}

type OnboardingResponse struct {
	TenantCode       string           `json:"tenant_code"`
	RoleARN          string           `json:"role_arn"`
	ExternalID       string           `json:"external_id"`
	SetupScript      string           `json:"setup_script"`
	Accounts         []AWSAccount     `json:"accounts"`
	SubscriptionTier SubscriptionTier `json:"subscription_tier"`
}
