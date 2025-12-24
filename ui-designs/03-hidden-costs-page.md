# Feature: Hidden Costs Page

## Priority: HIGH (IMPLEMENTED ✅)

## What It Does
Displays all cost optimization findings with filters (severity, category, status) and detailed recommendations.

## Visual Reference
```
┌─────────────────────────────────────────────────────────────┐
│  YUKTI FINOPS - HIDDEN COSTS                 [Scan] [Profile]│
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  Filters:                                                     │
│  Severity: [All ▼] [Critical] [High] [Medium] [Low]         │
│  Category: [All ▼] [Compute] [Storage] [Network] [Database] │
│  Status:   [All ▼] [Open] [In Progress] [Resolved]          │
│                                                               │
│  ┌─────────────────────────────────────────────────────┐    │
│  │  🔴 CRITICAL - Unencrypted RDS instances            │    │
│  │  Category: Database | Resources: 2                  │    │
│  │  Potential Savings: $0/month (Security Risk)        │    │
│  │                                                      │    │
│  │  Description:                                        │    │
│  │  2 RDS instances are not encrypted at rest. This    │    │
│  │  violates compliance requirements (PCI-DSS, HIPAA). │    │
│  │                                                      │    │
│  │  Recommendation:                                     │    │
│  │  Enable encryption for RDS instances. Create        │    │
│  │  encrypted snapshot, restore to new instance.       │    │
│  │                                                      │    │
│  │  [View Resources] [Generate IaC] [Whitelist]        │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                               │
│  ┌─────────────────────────────────────────────────────┐    │
│  │  🔴 HIGH - Idle EC2 instances                       │    │
│  │  Category: Compute | Resources: 3                   │    │
│  │  Potential Savings: $120/month                      │    │
│  │                                                      │    │
│  │  Description:                                        │    │
│  │  3 EC2 instances with CPU utilization < 5% for      │    │
│  │  last 7 days. Likely unused or over-provisioned.    │    │
│  │                                                      │    │
│  │  Recommendation:                                     │    │
│  │  Stop or terminate idle instances. Consider         │    │
│  │  downsizing to smaller instance types.              │    │
│  │                                                      │    │
│  │  [View Resources] [Generate IaC] [Whitelist]        │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                               │
│  ┌─────────────────────────────────────────────────────┐    │
│  │  🟡 MEDIUM - Unattached EBS volumes                 │    │
│  │  Category: Storage | Resources: 2                   │    │
│  │  Potential Savings: $80/month                       │    │
│  │                                                      │    │
│  │  Description:                                        │    │
│  │  2 EBS volumes are not attached to any EC2          │    │
│  │  instance. You're paying for unused storage.        │    │
│  │                                                      │    │
│  │  Recommendation:                                     │    │
│  │  Delete unattached volumes or create snapshots      │    │
│  │  for backup before deletion.                        │    │
│  │                                                      │    │
│  │  [View Resources] [Generate IaC] [Whitelist]        │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                               │
│  Showing 7 findings | Total Savings: $425.60/month          │
│                                                               │
└─────────────────────────────────────────────────────────────┘
```

## User Flow
1. User clicks "Hidden Costs" in sidebar
2. Page loads all findings for tenant
3. User applies filters (severity, category, status)
4. Findings list updates in real-time
5. User clicks "View Resources" to see affected resources
6. User clicks "Generate IaC" to create Terraform fix
7. User clicks "Whitelist" to mark as accepted risk
8. Findings update status (open → in progress → resolved)

## Data Requirements

### Input
- JWT token (from localStorage)
- Tenant ID (extracted from JWT)
- Filters: severity, category, status

### Output
```json
{
  "findings": [
    {
      "id": 1,
      "title": "Idle EC2 instances",
      "description": "3 EC2 instances with CPU < 5%",
      "severity": "high",
      "category": "compute",
      "status": "open",
      "monthly_savings": 120.00,
      "resource_count": 3,
      "resources": [
        {
          "resource_id": "i-0a046ebb489ff3cd7",
          "resource_type": "ec2",
          "region": "us-east-1",
          "metadata": {
            "instance_type": "t3.large",
            "cpu_utilization": 2.3
          }
        }
      ],
      "recommendation": "Stop or terminate idle instances",
      "created_at": "2025-01-30T08:00:00Z"
    }
  ],
  "total_savings": 425.60,
  "total_count": 7
}
```

## API Endpoints

### GET /api/v1/customers/findings
**Headers**:
```
Authorization: Bearer <JWT_TOKEN>
```

**Query Params**:
- `severity` (optional: critical, high, medium, low)
- `category` (optional: compute, storage, network, database)
- `status` (optional: open, in_progress, resolved)

**Response (200)**:
```json
{
  "findings": [
    {
      "id": 1,
      "title": "Idle EC2 instances",
      "description": "3 EC2 instances with CPU < 5%",
      "severity": "high",
      "category": "compute",
      "status": "open",
      "monthly_savings": 120.00,
      "resource_count": 3,
      "created_at": "2025-01-30T08:00:00Z"
    }
  ]
}
```

### GET /api/v1/customers/findings/:id/resources
**Headers**:
```
Authorization: Bearer <JWT_TOKEN>
```

**Response (200)**:
```json
{
  "resources": [
    {
      "resource_id": "i-0a046ebb489ff3cd7",
      "resource_type": "ec2",
      "region": "us-east-1",
      "instance_type": "t3.large",
      "state": "running",
      "cpu_utilization": 2.3,
      "tags": {
        "Name": "web-server-1",
        "Environment": "production"
      }
    }
  ]
}
```

### POST /api/v1/whitelist/create
**Headers**:
```
Authorization: Bearer <JWT_TOKEN>
```

**Request**:
```json
{
  "finding_id": 1,
  "reason": "This instance is used for batch processing",
  "expires_at": "2025-12-31T23:59:59Z"
}
```

**Response (200)**:
```json
{
  "message": "Finding whitelisted successfully",
  "whitelist_id": 123
}
```

## Database Tables

### yt_hidden_cost_findings
- `id` (serial, primary key)
- `tenant_id` (integer)
- `title` (varchar)
- `description` (text)
- `severity` (varchar: critical, high, medium, low)
- `category` (varchar: compute, storage, network, database, etc.)
- `status` (varchar: open, in_progress, resolved)
- `monthly_savings` (decimal)
- `resource_count` (integer)
- `recommendation` (text)
- `created_at` (timestamp)
- `updated_at` (timestamp)

### yt_finding_resources
- `id` (serial, primary key)
- `finding_id` (integer, foreign key)
- `resource_id` (varchar)
- `resource_type` (varchar)
- `region` (varchar)
- `metadata` (jsonb)

### yt_whitelisted_findings
- `id` (serial, primary key)
- `tenant_id` (integer)
- `finding_id` (integer, foreign key)
- `reason` (text)
- `whitelisted_by` (integer, foreign key to yt_users)
- `expires_at` (timestamp)
- `created_at` (timestamp)

## UI Components

### Page
- **Path**: `/hidden-costs`
- **File**: `frontend/src/pages/HiddenCosts.tsx`

### Components Used
- FilterBar (severity, category, status dropdowns)
- FindingCard (individual finding with details)
- SeverityBadge (color-coded severity indicator)
- ResourceModal (shows affected resources)
- WhitelistModal (whitelist form)

## Business Rules
1. Only authenticated users can access
2. Data filtered by tenant_id from JWT
3. Findings sorted by severity (critical → high → medium → low)
4. Whitelisted findings hidden by default (toggle to show)
5. Resolved findings archived after 30 days
6. Only admins/owners can whitelist findings
7. Whitelist requires reason (min 10 chars)
8. Whitelist can have expiration date

## Security Features
- ✅ JWT-based authentication
- ✅ Tenant isolation (no cross-tenant data)
- ✅ Role-based access (whitelist requires admin/owner)
- ✅ Audit logging (whitelist actions tracked)

## Implementation Status
- ✅ Frontend: `frontend/src/pages/HiddenCosts.tsx`
- ✅ Backend: `internal/api/handlers/customers.go` (GetFindings)
- ✅ Backend: `internal/api/handlers/whitelist.go` (CreateWhitelist)
- ✅ Database: All tables created and seeded
- ✅ Testing: Manual testing complete
- ✅ Deployment: Live in Docker container

## Known Issues
- Resource modal shows mock data (need real resource details API)
- IaC generation not fully implemented (placeholder button)

## Future Enhancements
- Add bulk whitelist (select multiple findings)
- Add finding history (track status changes)
- Add email notifications (new findings, status changes)
- Add export to CSV/PDF
- Add finding comments/notes
