package whitelist

import "time"

type WhitelistType string

const (
	WhitelistTypeResource          WhitelistType = "resource"
	WhitelistTypeTag               WhitelistType = "tag"
	WhitelistTypeService           WhitelistType = "service"
	WhitelistTypeRecommendationType WhitelistType = "recommendation_type"
)

type WhitelistStatus string

const (
	StatusActive          WhitelistStatus = "active"
	StatusExpired         WhitelistStatus = "expired"
	StatusRevoked         WhitelistStatus = "revoked"
	StatusPendingApproval WhitelistStatus = "pending_approval"
)

type Whitelist struct {
	ID                    string          `json:"id"`
	TenantID              string          `json:"tenant_id"`
	WhitelistType         WhitelistType   `json:"whitelist_type"`
	ResourceARN           *string         `json:"resource_arn,omitempty"`
	TagKey                *string         `json:"tag_key,omitempty"`
	TagValue              *string         `json:"tag_value,omitempty"`
	ServiceName           *string         `json:"service_name,omitempty"`
	RecommendationType    *string         `json:"recommendation_type,omitempty"`
	Reason                string          `json:"reason"`
	BusinessJustification string          `json:"business_justification"`
	CostImpactMonthly     float64         `json:"cost_impact_monthly"`
	CreatedBy             string          `json:"created_by"`
	CreatedAt             time.Time       `json:"created_at"`
	ApprovedBy            *string         `json:"approved_by,omitempty"`
	ApprovedAt            *time.Time      `json:"approved_at,omitempty"`
	ExpiresAt             *time.Time      `json:"expires_at,omitempty"`
	Status                WhitelistStatus `json:"status"`
}

type CreateWhitelistRequest struct {
	WhitelistType         WhitelistType  `json:"whitelist_type"`
	ResourceARN           *string        `json:"resource_arn,omitempty"`
	TagKey                *string        `json:"tag_key,omitempty"`
	TagValue              *string        `json:"tag_value,omitempty"`
	ServiceName           *string        `json:"service_name,omitempty"`
	RecommendationType    *string        `json:"recommendation_type,omitempty"`
	Reason                string         `json:"reason"`
	BusinessJustification string         `json:"business_justification"`
	ExpiresInDays         *int           `json:"expires_in_days,omitempty"`
}
