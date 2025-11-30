# Resource Whitelisting - Technical Specification

## Overview
Allow customers to exclude business-critical resources from optimization recommendations with audit trails, approval workflows, and expiry management.

---

## Core Requirements

### Whitelisting Capabilities
1. **Whitelist by Resource**: Specific EC2 instance, RDS database, S3 bucket
2. **Whitelist by Tag**: All resources with `env:production` or `critical:true`
3. **Whitelist by Service**: All Lambda functions, all ECS tasks
4. **Whitelist by Recommendation Type**: Exclude "terminate" but allow "right-size"
5. **Temporary Whitelisting**: Auto-expire after 30/60/90 days
6. **Scheduled Whitelisting**: Active only during business hours (9am-5pm Mon-Fri)

### Business Rules
- **Reason Required**: Must provide business justification (min 20 characters)
- **Approval Workflow**: Whitelists >$1K/month impact require manager approval
- **Expiry Reminders**: Email 7 days before expiry with re-approval option
- **Audit Trail**: Track who whitelisted, when, why, and when removed
- **Impact Visibility**: Show monthly cost impact of whitelisted resources

---

## Database Schema

```sql
-- scripts/009_create_whitelisting_tables.sql

CREATE TABLE yt_whitelists (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES yt_tenants(id) ON DELETE CASCADE,
    
    -- What is whitelisted
    whitelist_type VARCHAR(20) NOT NULL, -- 'resource', 'tag', 'service', 'recommendation_type'
    resource_arn VARCHAR(500), -- For resource-level whitelisting
    tag_key VARCHAR(128), -- For tag-based whitelisting
    tag_value VARCHAR(256),
    service_name VARCHAR(50), -- For service-level whitelisting (ec2, rds, lambda)
    recommendation_type VARCHAR(50), -- For recommendation-type whitelisting (terminate, right_size)
    
    -- Why whitelisted
    reason TEXT NOT NULL,
    business_justification TEXT,
    cost_impact_monthly DECIMAL(12,2), -- Estimated monthly cost of NOT optimizing
    
    -- Who and when
    created_by VARCHAR(100) NOT NULL, -- User email
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    approved_by VARCHAR(100), -- Manager email (if approval required)
    approved_at TIMESTAMP,
    
    -- Expiry
    expires_at TIMESTAMP, -- NULL = permanent
    expiry_reminder_sent BOOLEAN DEFAULT FALSE,
    
    -- Schedule (for time-based whitelisting)
    schedule_enabled BOOLEAN DEFAULT FALSE,
    schedule_cron VARCHAR(100), -- Cron expression: "0 9-17 * * 1-5" (9am-5pm Mon-Fri)
    schedule_timezone VARCHAR(50) DEFAULT 'UTC',
    
    -- Status
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- active, expired, revoked, pending_approval
    revoked_by VARCHAR(100),
    revoked_at TIMESTAMP,
    revoked_reason TEXT,
    
    -- Metadata
    metadata JSONB, -- Additional context (e.g., ticket number, compliance requirement)
    
    CONSTRAINT valid_whitelist_type CHECK (whitelist_type IN ('resource', 'tag', 'service', 'recommendation_type')),
    CONSTRAINT valid_status CHECK (status IN ('active', 'expired', 'revoked', 'pending_approval'))
);

CREATE INDEX idx_whitelists_tenant ON yt_whitelists(tenant_id);
CREATE INDEX idx_whitelists_resource ON yt_whitelists(resource_arn);
CREATE INDEX idx_whitelists_tag ON yt_whitelists(tag_key, tag_value);
CREATE INDEX idx_whitelists_status ON yt_whitelists(status);
CREATE INDEX idx_whitelists_expires ON yt_whitelists(expires_at) WHERE expires_at IS NOT NULL;

-- Audit log for whitelist changes
CREATE TABLE yt_whitelist_audit (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    whitelist_id UUID NOT NULL REFERENCES yt_whitelists(id) ON DELETE CASCADE,
    action VARCHAR(20) NOT NULL, -- created, approved, expired, revoked, extended
    performed_by VARCHAR(100) NOT NULL,
    performed_at TIMESTAMP NOT NULL DEFAULT NOW(),
    details JSONB,
    
    CONSTRAINT valid_action CHECK (action IN ('created', 'approved', 'expired', 'revoked', 'extended'))
);

CREATE INDEX idx_whitelist_audit_whitelist ON yt_whitelist_audit(whitelist_id);
CREATE INDEX idx_whitelist_audit_performed_at ON yt_whitelist_audit(performed_at);

-- Approval requests
CREATE TABLE yt_whitelist_approvals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    whitelist_id UUID NOT NULL REFERENCES yt_whitelists(id) ON DELETE CASCADE,
    requested_by VARCHAR(100) NOT NULL,
    requested_at TIMESTAMP NOT NULL DEFAULT NOW(),
    approver_email VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, approved, rejected
    approved_at TIMESTAMP,
    rejection_reason TEXT,
    
    CONSTRAINT valid_approval_status CHECK (status IN ('pending', 'approved', 'rejected'))
);

CREATE INDEX idx_whitelist_approvals_whitelist ON yt_whitelist_approvals(whitelist_id);
CREATE INDEX idx_whitelist_approvals_status ON yt_whitelist_approvals(status);
```

---

## Go Implementation

### Models
```go
// internal/whitelist/models.go
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
    ID                     string          `json:"id"`
    TenantID               string          `json:"tenant_id"`
    WhitelistType          WhitelistType   `json:"whitelist_type"`
    ResourceARN            *string         `json:"resource_arn,omitempty"`
    TagKey                 *string         `json:"tag_key,omitempty"`
    TagValue               *string         `json:"tag_value,omitempty"`
    ServiceName            *string         `json:"service_name,omitempty"`
    RecommendationType     *string         `json:"recommendation_type,omitempty"`
    Reason                 string          `json:"reason"`
    BusinessJustification  string          `json:"business_justification"`
    CostImpactMonthly      float64         `json:"cost_impact_monthly"`
    CreatedBy              string          `json:"created_by"`
    CreatedAt              time.Time       `json:"created_at"`
    ApprovedBy             *string         `json:"approved_by,omitempty"`
    ApprovedAt             *time.Time      `json:"approved_at,omitempty"`
    ExpiresAt              *time.Time      `json:"expires_at,omitempty"`
    ExpiryReminderSent     bool            `json:"expiry_reminder_sent"`
    ScheduleEnabled        bool            `json:"schedule_enabled"`
    ScheduleCron           *string         `json:"schedule_cron,omitempty"`
    ScheduleTimezone       string          `json:"schedule_timezone"`
    Status                 WhitelistStatus `json:"status"`
    RevokedBy              *string         `json:"revoked_by,omitempty"`
    RevokedAt              *time.Time      `json:"revoked_at,omitempty"`
    RevokedReason          *string         `json:"revoked_reason,omitempty"`
    Metadata               map[string]any  `json:"metadata,omitempty"`
}

type CreateWhitelistRequest struct {
    WhitelistType          WhitelistType  `json:"whitelist_type" validate:"required"`
    ResourceARN            *string        `json:"resource_arn,omitempty"`
    TagKey                 *string        `json:"tag_key,omitempty"`
    TagValue               *string        `json:"tag_value,omitempty"`
    ServiceName            *string        `json:"service_name,omitempty"`
    RecommendationType     *string        `json:"recommendation_type,omitempty"`
    Reason                 string         `json:"reason" validate:"required,min=20"`
    BusinessJustification  string         `json:"business_justification"`
    ExpiresInDays          *int           `json:"expires_in_days,omitempty"` // 30, 60, 90, or null for permanent
    ScheduleEnabled        bool           `json:"schedule_enabled"`
    ScheduleCron           *string        `json:"schedule_cron,omitempty"`
    ScheduleTimezone       string         `json:"schedule_timezone"`
    Metadata               map[string]any `json:"metadata,omitempty"`
}
```

### Service
```go
// internal/whitelist/service.go
package whitelist

import (
    "context"
    "database/sql"
    "time"
)

type Service struct {
    db            *sql.DB
    costEstimator *CostEstimator
    notifier      *Notifier
}

func (s *Service) CreateWhitelist(ctx context.Context, tenantID, userEmail string, req CreateWhitelistRequest) (*Whitelist, error) {
    // 1. Validate request
    if err := s.validateRequest(req); err != nil {
        return nil, err
    }
    
    // 2. Estimate cost impact
    costImpact, err := s.costEstimator.EstimateCostImpact(ctx, tenantID, req)
    if err != nil {
        return nil, err
    }
    
    // 3. Determine if approval required
    requiresApproval := costImpact > 1000.00 // >$1K/month
    status := StatusActive
    if requiresApproval {
        status = StatusPendingApproval
    }
    
    // 4. Calculate expiry
    var expiresAt *time.Time
    if req.ExpiresInDays != nil {
        expiry := time.Now().AddDate(0, 0, *req.ExpiresInDays)
        expiresAt = &expiry
    }
    
    // 5. Insert whitelist
    whitelist := &Whitelist{
        TenantID:              tenantID,
        WhitelistType:         req.WhitelistType,
        ResourceARN:           req.ResourceARN,
        TagKey:                req.TagKey,
        TagValue:              req.TagValue,
        ServiceName:           req.ServiceName,
        RecommendationType:    req.RecommendationType,
        Reason:                req.Reason,
        BusinessJustification: req.BusinessJustification,
        CostImpactMonthly:     costImpact,
        CreatedBy:             userEmail,
        CreatedAt:             time.Now(),
        ExpiresAt:             expiresAt,
        ScheduleEnabled:       req.ScheduleEnabled,
        ScheduleCron:          req.ScheduleCron,
        ScheduleTimezone:      req.ScheduleTimezone,
        Status:                status,
        Metadata:              req.Metadata,
    }
    
    if err := s.insertWhitelist(ctx, whitelist); err != nil {
        return nil, err
    }
    
    // 6. Create approval request if needed
    if requiresApproval {
        if err := s.createApprovalRequest(ctx, whitelist.ID, userEmail); err != nil {
            return nil, err
        }
        s.notifier.SendApprovalRequest(whitelist)
    }
    
    // 7. Audit log
    s.auditLog(ctx, whitelist.ID, "created", userEmail, nil)
    
    return whitelist, nil
}

func (s *Service) IsResourceWhitelisted(ctx context.Context, tenantID, resourceARN string, recommendationType string) (bool, error) {
    // Check if resource is whitelisted by:
    // 1. Direct resource ARN match
    // 2. Tag-based whitelist (check resource tags)
    // 3. Service-level whitelist (extract service from ARN)
    // 4. Recommendation type whitelist
    // 5. Check if whitelist is active (not expired, not revoked)
    // 6. Check if within schedule (if schedule enabled)
    
    query := `
        SELECT COUNT(*) FROM yt_whitelists
        WHERE tenant_id = $1
        AND status = 'active'
        AND (expires_at IS NULL OR expires_at > NOW())
        AND (
            (whitelist_type = 'resource' AND resource_arn = $2)
            OR (whitelist_type = 'recommendation_type' AND recommendation_type = $3)
            OR (whitelist_type = 'service' AND service_name = $4)
        )
    `
    
    serviceName := extractServiceFromARN(resourceARN)
    var count int
    err := s.db.QueryRowContext(ctx, query, tenantID, resourceARN, recommendationType, serviceName).Scan(&count)
    return count > 0, err
}

func (s *Service) ProcessExpiringWhitelists(ctx context.Context) error {
    // Run daily cron job to:
    // 1. Find whitelists expiring in 7 days
    // 2. Send reminder emails
    // 3. Mark expired whitelists
    
    // Find expiring in 7 days
    expiringWhitelists, err := s.findExpiringWhitelists(ctx, 7)
    if err != nil {
        return err
    }
    
    for _, wl := range expiringWhitelists {
        if !wl.ExpiryReminderSent {
            s.notifier.SendExpiryReminder(wl)
            s.markReminderSent(ctx, wl.ID)
        }
    }
    
    // Mark expired whitelists
    expiredWhitelists, err := s.findExpiredWhitelists(ctx)
    if err != nil {
        return err
    }
    
    for _, wl := range expiredWhitelists {
        s.markExpired(ctx, wl.ID)
        s.auditLog(ctx, wl.ID, "expired", "system", nil)
    }
    
    return nil
}
```

### Cost Estimator
```go
// internal/whitelist/cost_estimator.go
package whitelist

import "context"

type CostEstimator struct {
    db *sql.DB
}

func (e *CostEstimator) EstimateCostImpact(ctx context.Context, tenantID string, req CreateWhitelistRequest) (float64, error) {
    switch req.WhitelistType {
    case WhitelistTypeResource:
        return e.estimateResourceCost(ctx, tenantID, *req.ResourceARN)
    case WhitelistTypeTag:
        return e.estimateTagCost(ctx, tenantID, *req.TagKey, *req.TagValue)
    case WhitelistTypeService:
        return e.estimateServiceCost(ctx, tenantID, *req.ServiceName)
    case WhitelistTypeRecommendationType:
        return e.estimateRecommendationTypeCost(ctx, tenantID, *req.RecommendationType)
    default:
        return 0, nil
    }
}

func (e *CostEstimator) estimateResourceCost(ctx context.Context, tenantID, resourceARN string) (float64, error) {
    // Query recommendations for this resource
    // Sum potential_savings_monthly
    query := `
        SELECT COALESCE(SUM(potential_savings_monthly), 0)
        FROM yt_tenant_recommendations
        WHERE tenant_id = $1 AND resource_arn = $2 AND status = 'open'
    `
    var totalSavings float64
    err := e.db.QueryRowContext(ctx, query, tenantID, resourceARN).Scan(&totalSavings)
    return totalSavings, err
}
```

---

## API Endpoints

### Create Whitelist
```http
POST /api/v1/whitelists
Authorization: Bearer <jwt_token>

Request:
{
  "whitelist_type": "resource",
  "resource_arn": "arn:aws:ec2:us-east-1:123456789012:instance/i-abc123",
  "reason": "Production database server, business-critical for customer transactions",
  "business_justification": "Downtime would cost $10K/hour in lost revenue",
  "expires_in_days": 90,
  "metadata": {
    "ticket": "JIRA-1234",
    "approved_by_cfo": true
  }
}

Response:
{
  "id": "wl_abc123",
  "status": "active",
  "cost_impact_monthly": 450.00,
  "message": "Resource whitelisted successfully. $450/month in recommendations will be suppressed."
}
```

### List Whitelists
```http
GET /api/v1/whitelists
Authorization: Bearer <jwt_token>

Query Parameters:
- status: active, expired, revoked, pending_approval
- whitelist_type: resource, tag, service, recommendation_type
- expiring_soon: true (expires in <30 days)

Response:
{
  "whitelists": [
    {
      "id": "wl_abc123",
      "whitelist_type": "resource",
      "resource_arn": "arn:aws:ec2:us-east-1:123456789012:instance/i-abc123",
      "reason": "Production database server",
      "cost_impact_monthly": 450.00,
      "created_by": "john@company.com",
      "created_at": "2025-01-01T10:00:00Z",
      "expires_at": "2025-04-01T10:00:00Z",
      "days_until_expiry": 75,
      "status": "active"
    }
  ],
  "total_cost_impact": 12450.00,
  "summary": {
    "active": 47,
    "expiring_soon": 8,
    "pending_approval": 3
  }
}
```

### Revoke Whitelist
```http
DELETE /api/v1/whitelists/{id}
Authorization: Bearer <jwt_token>

Request:
{
  "reason": "Resource decommissioned, no longer needed"
}

Response:
{
  "success": true,
  "message": "Whitelist revoked. Recommendations will resume for this resource."
}
```

### Extend Whitelist
```http
POST /api/v1/whitelists/{id}/extend
Authorization: Bearer <jwt_token>

Request:
{
  "extend_by_days": 90,
  "reason": "Project extended, still business-critical"
}

Response:
{
  "success": true,
  "new_expiry": "2025-07-01T10:00:00Z"
}
```

### Approve Whitelist
```http
POST /api/v1/whitelists/{id}/approve
Authorization: Bearer <jwt_token>

Response:
{
  "success": true,
  "message": "Whitelist approved and activated"
}
```

---

## UI Components

### Whitelisting Modal
```jsx
// frontend/src/components/WhitelistModal.js
function WhitelistModal({ resource, onClose }) {
  const [reason, setReason] = useState('');
  const [expiresInDays, setExpiresInDays] = useState(90);
  
  return (
    <Modal>
      <h2>Whitelist Resource</h2>
      <ResourceInfo resource={resource} />
      <CostImpactWarning impact={resource.potential_savings} />
      
      <TextArea
        label="Reason (required)"
        value={reason}
        onChange={setReason}
        minLength={20}
        placeholder="Explain why this resource should be excluded from recommendations"
      />
      
      <Select
        label="Expiry"
        value={expiresInDays}
        onChange={setExpiresInDays}
        options={[
          { value: 30, label: '30 days' },
          { value: 60, label: '60 days' },
          { value: 90, label: '90 days' },
          { value: null, label: 'Permanent (requires approval)' }
        ]}
      />
      
      <Button onClick={handleWhitelist}>Whitelist Resource</Button>
    </Modal>
  );
}
```

### Whitelists Dashboard
```jsx
// frontend/src/pages/Whitelists.js
function WhitelistsDashboard() {
  return (
    <div>
      <WhitelistsSummary total={47} expiringSoon={8} pendingApproval={3} />
      <CostImpactCard totalImpact={12450} />
      <WhitelistsTable whitelists={whitelists} onRevoke={handleRevoke} onExtend={handleExtend} />
      <ExpiringWhitelistsAlert whitelists={expiringWhitelists} />
    </div>
  );
}
```

---

## Notification Templates

### Approval Request Email
```
Subject: Whitelist Approval Required - $1,250/month impact

Hi [Manager],

[User] has requested to whitelist a resource that will suppress $1,250/month in cost optimization recommendations.

Resource: arn:aws:rds:us-east-1:123456789012:db:prod-db
Reason: Production database, business-critical
Expiry: 90 days

[Approve] [Reject] [View Details]
```

### Expiry Reminder Email
```
Subject: Whitelist Expiring in 7 Days

Hi [User],

Your whitelist for [Resource] will expire in 7 days on [Date].

After expiry, cost optimization recommendations will resume for this resource.

[Extend Whitelist] [Let it Expire] [View Details]
```

---

## Pricing Strategy

### Feature Availability
- **FREE**: 0 whitelists
- **PROFESSIONAL**: 10 whitelists
- **ENTERPRISE**: Unlimited whitelists + approval workflows
- **FINANCIAL**: Unlimited + scheduled whitelisting + custom policies

---

**Last Updated**: January 2025  
**Owner**: Product & Engineering Teams  
**Status**: In development, targeting Q1 2025 release
