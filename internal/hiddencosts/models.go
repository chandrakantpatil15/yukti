package hiddencosts

import "time"

type Category string

const (
	CategoryDataTransfer    Category = "Data Transfer Costs"
	CategoryStorage         Category = "Storage Lifecycle Waste"
	CategoryCompute         Category = "Compute Waste"
	CategoryDatabase        Category = "Database Inefficiencies"
	CategoryNetworking      Category = "Networking Costs"
	CategoryManagedServices Category = "Managed Service Premiums"
	CategoryServerless      Category = "Serverless Inefficiencies"
	CategoryContainer       Category = "Container Waste"
	CategoryKubernetes      Category = "Kubernetes Optimization"
	CategoryEOL             Category = "End-of-Life Software"
)

type Severity string

const (
	SeverityCritical Severity = "Critical"
	SeverityHigh     Severity = "High"
	SeverityMedium   Severity = "Medium"
	SeverityLow      Severity = "Low"
	SeverityInfo     Severity = "Info"
)

type Finding struct {
	ID               string     `json:"id"`
	TenantID         string     `json:"tenant_id"`
	DetectorName     string     `json:"detector"`
	Category         Category   `json:"category"`
	Severity         Severity   `json:"severity"`
	Title            string     `json:"title"`
	Description      string     `json:"description"`
	ResourceARN      string     `json:"resource_arn"`
	EstimatedCost    float64    `json:"estimated_cost"`
	EstimatedSavings float64    `json:"estimated_savings"`
	Confidence       float64    `json:"confidence"`
	Recommendation   string     `json:"recommendation"`
	RemediationSteps []string   `json:"remediation_steps"`
	IaCCode          string     `json:"iac_code"`
	DetectedAt       time.Time  `json:"detected_at"`
	SuppressedUntil  *time.Time `json:"suppressed_until,omitempty"`
}

type Resource struct {
	ARN      string
	Type     string
	Region   string
	Tags     map[string]string
	Metadata map[string]interface{}
}
