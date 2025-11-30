package onboarding

import "time"

type OnboardingStatus string

const (
	StatusPending    OnboardingStatus = "pending"
	StatusInProgress OnboardingStatus = "in_progress"
	StatusCompleted  OnboardingStatus = "completed"
	StatusFailed     OnboardingStatus = "failed"
)

type OnboardingStep string

const (
	StepAWSConnection     OnboardingStep = "aws_connection"
	StepMetricsIntegration OnboardingStep = "metrics_integration"
	StepInitialScan       OnboardingStep = "initial_scan"
	StepReviewFindings    OnboardingStep = "review_findings"
)

type Customer struct {
	ID              string           `json:"id"`
	TenantID        string           `json:"tenant_id"`
	CompanyName     string           `json:"company_name"`
	Email           string           `json:"email"`
	OnboardingStatus OnboardingStatus `json:"onboarding_status"`
	CurrentStep     OnboardingStep   `json:"current_step"`
	CreatedAt       time.Time        `json:"created_at"`
	CompletedAt     *time.Time       `json:"completed_at,omitempty"`
}

type AWSConnection struct {
	TenantID        string    `json:"tenant_id"`
	AccountID       string    `json:"account_id"`
	RoleARN         string    `json:"role_arn"`
	ExternalID      string    `json:"external_id"`
	Regions         []string  `json:"regions"`
	Verified        bool      `json:"verified"`
	LastVerifiedAt  time.Time `json:"last_verified_at"`
}

type MetricsIntegration struct {
	TenantID   string                 `json:"tenant_id"`
	Source     string                 `json:"source"`
	Endpoint   string                 `json:"endpoint"`
	Credentials map[string]string     `json:"credentials"`
	Config     map[string]interface{} `json:"config"`
	Verified   bool                   `json:"verified"`
	CreatedAt  time.Time              `json:"created_at"`
}
