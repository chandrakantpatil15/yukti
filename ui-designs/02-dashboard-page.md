# Feature: Dashboard Page

## Priority: HIGH (IMPLEMENTED ✅)

## What It Does
Main landing page showing tenant-specific metrics: total cost, potential savings, resource count, cost trends, and top findings.

## Visual Reference
```
┌─────────────────────────────────────────────────────────────┐
│  YUKTI FINOPS - DASHBOARD                    [Scan] [Profile]│
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │ Total Cost  │  │   Savings   │  │  Resources  │         │
│  │  $12,450    │  │   $425.60   │  │     847     │         │
│  │  This Month │  │  Potential  │  │   Scanned   │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
│                                                               │
│  AWS Connection Status                                       │
│  ┌─────────────────────────────────────────────────────┐    │
│  │  ✅ Connected to AWS Account: 424851482219         │    │
│  │  Last Scan: 2 hours ago                             │    │
│  │  [Scan Now] [Sync]                                  │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                               │
│  Cost Trend (Last 30 Days)                                   │
│  ┌─────────────────────────────────────────────────────┐    │
│  │  $500 ┤                                    ╱╲       │    │
│  │       ┤                          ╱╲      ╱  ╲      │    │
│  │  $400 ┤                    ╱╲  ╱  ╲    ╱    ╲     │    │
│  │       ┤              ╱╲  ╱  ╲╱    ╲  ╱      ╲    │    │
│  │  $300 ┤        ╱╲  ╱  ╲╱            ╲╱        ╲   │    │
│  │       └────────────────────────────────────────────│    │
│  │         Jan 1        Jan 15        Jan 30          │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                               │
│  Top Findings (7)                                            │
│  ┌─────────────────────────────────────────────────────┐    │
│  │  🔴 HIGH   Idle EC2 instances (3)                   │    │
│  │            Save $120/month                          │    │
│  │            [View Details] [Fix Now]                 │    │
│  │                                                      │    │
│  │  🟡 MEDIUM Unattached EBS volumes (2)               │    │
│  │            Save $80/month                           │    │
│  │            [View Details] [Fix Now]                 │    │
│  │                                                      │    │
│  │  🟢 LOW    Old snapshots (5)                        │    │
│  │            Save $50/month                           │    │
│  │            [View Details] [Fix Now]                 │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                               │
└─────────────────────────────────────────────────────────────┘
```

## User Flow
1. User logs in successfully
2. System redirects to /dashboard
3. Dashboard loads tenant-specific data from JWT token
4. Shows 3 stat cards (cost, savings, resources)
5. Shows AWS connection status
6. Shows cost trend chart (last 30 days)
7. Shows top 7 findings with severity badges
8. User can click "Scan Now" to trigger AWS scan
9. User can click "View Details" to see finding details
10. Auto-refreshes every 60 seconds

## Data Requirements

### Input
- JWT token (from localStorage)
- Tenant ID (extracted from JWT)

### Output
```json
{
  "total_cost": 12450.00,
  "potential_savings": 425.60,
  "resource_count": 847,
  "aws_connection": {
    "connected": true,
    "account_id": "424851482219",
    "last_scan": "2025-01-31T10:30:00Z"
  },
  "cost_trend": [
    {"date": "2025-01-01", "cost": 380.50},
    {"date": "2025-01-02", "cost": 420.30},
    ...
  ],
  "top_findings": [
    {
      "id": 1,
      "title": "Idle EC2 instances",
      "severity": "high",
      "count": 3,
      "monthly_savings": 120.00,
      "category": "compute"
    },
    ...
  ]
}
```

## API Endpoints

### GET /api/v1/customers/dashboard
**Headers**:
```
Authorization: Bearer <JWT_TOKEN>
```

**Response (200)**:
```json
{
  "total_cost": 12450.00,
  "potential_savings": 425.60,
  "resource_count": 847,
  "findings_count": 7
}
```

### GET /api/v1/onboarding/aws-connection
**Headers**:
```
Authorization: Bearer <JWT_TOKEN>
```

**Response (200)**:
```json
{
  "connected": true,
  "account_id": "424851482219",
  "role_arn": "arn:aws:iam::424851482219:role/YuktiFinOpsRole",
  "verified": true,
  "last_verified_at": "2025-01-31T10:30:00Z"
}
```

### GET /api/v1/customers/findings
**Headers**:
```
Authorization: Bearer <JWT_TOKEN>
```

**Query Params**:
- `limit=7` (top 7 findings)

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
      "monthly_savings": 120.00,
      "resource_count": 3,
      "created_at": "2025-01-30T08:00:00Z"
    }
  ]
}
```

### POST /api/v1/scan/trigger
**Headers**:
```
Authorization: Bearer <JWT_TOKEN>
```

**Response (200)**:
```json
{
  "message": "Scan started successfully",
  "scan_id": "scan_123456"
}
```

## Database Tables

### yt_customers
- `id` (serial, primary key)
- `tenant_id` (integer, unique)
- `company_name` (varchar)
- `total_cost` (decimal)
- `potential_savings` (decimal)

### yt_hidden_cost_findings
- `id` (serial, primary key)
- `tenant_id` (integer)
- `title` (varchar)
- `description` (text)
- `severity` (varchar: critical, high, medium, low)
- `category` (varchar)
- `monthly_savings` (decimal)
- `created_at` (timestamp)

### yt_tenant_resources
- `id` (serial, primary key)
- `tenant_id` (integer)
- `resource_id` (varchar)
- `resource_type` (varchar: ec2, rds, s3)
- `region` (varchar)
- `metadata` (jsonb)

### yt_aws_connections
- `id` (serial, primary key)
- `tenant_id` (integer, unique)
- `account_id` (varchar)
- `role_arn` (varchar)
- `verified` (boolean)
- `last_verified_at` (timestamp)

## UI Components

### Page
- **Path**: `/dashboard`
- **File**: `frontend/src/pages/Dashboard.tsx`

### Components Used
- StatCard (total cost, savings, resources)
- AWSConnectionStatus (connection indicator)
- CostTrendChart (line chart with recharts)
- FindingsList (top findings with severity badges)
- Sidebar (navigation)

## Business Rules
1. Only authenticated users can access
2. Data filtered by tenant_id from JWT
3. Auto-refresh every 60 seconds
4. Scan throttled to 1 per 5 minutes
5. Findings sorted by severity (critical → high → medium → low)
6. Cost trend shows last 30 days only
7. Top 7 findings displayed (rest on Hidden Costs page)

## Security Features
- ✅ JWT-based authentication
- ✅ Tenant isolation (no cross-tenant data)
- ✅ No tenant_id in query params (JWT only)
- ✅ Auto-logout on token expiration
- ✅ 401 handler redirects to login

## Implementation Status
- ✅ Frontend: `frontend/src/pages/Dashboard.tsx`
- ✅ Backend: `internal/api/handlers/customers.go` (GetDashboard)
- ✅ Backend: `internal/api/handlers/onboarding.go` (GetAWSConnection)
- ✅ Backend: `internal/api/handlers/scan.go` (TriggerScan)
- ✅ Database: All tables created and seeded
- ✅ Testing: Manual testing complete
- ✅ Deployment: Live in Docker container

## Known Issues
- Cost trend chart shows mock data (need real AWS Cost Explorer integration)
- Auto-refresh can be optimized (use WebSocket instead of polling)

## Future Enhancements
- Add date range selector (30D, 90D, 1Y)
- Add cost breakdown by service
- Add resource utilization metrics
- Add anomaly detection alerts
