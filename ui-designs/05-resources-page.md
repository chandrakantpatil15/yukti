# Feature: Resources Page

## Priority: MEDIUM (IMPLEMENTED ✅)

## What It Does
Displays all AWS resources (EC2, RDS, S3) discovered during scans with detailed metadata, tags, and cost information.

## Visual Reference
```
┌─────────────────────────────────────────────────────────────┐
│  YUKTI FINOPS - RESOURCES                    [Scan] [Profile]│
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  Filters:                                                     │
│  Type: [All ▼] [EC2] [RDS] [S3]                             │
│  Region: [All ▼] [us-east-1] [us-west-2] [eu-west-1]        │
│  Status: [All ▼] [Running] [Stopped] [Available]            │
│                                                               │
│  ┌─────────────────────────────────────────────────────┐    │
│  │  EC2 Instance                                       │    │
│  │  i-0a046ebb489ff3cd7                                │    │
│  │                                                      │    │
│  │  Region: us-east-1 | Type: t3.large                 │    │
│  │  Status: running | CPU: 23%                         │    │
│  │  Monthly Cost: $52.56                               │    │
│  │                                                      │    │
│  │  Tags:                                               │    │
│  │  • Name: web-server-1                               │    │
│  │  • Environment: production                          │    │
│  │  • Team: backend                                    │    │
│  │                                                      │    │
│  │  [View Details] [Stop] [Terminate]                  │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                               │
│  ┌─────────────────────────────────────────────────────┐    │
│  │  RDS Instance                                       │    │
│  │  database-1                                         │    │
│  │                                                      │    │
│  │  Region: us-west-2 | Engine: postgres               │    │
│  │  Status: available | Storage: 100 GB                │    │
│  │  Monthly Cost: $89.20                               │    │
│  │                                                      │    │
│  │  Tags:                                               │    │
│  │  • Name: production-db                              │    │
│  │  • Environment: production                          │    │
│  │                                                      │    │
│  │  [View Details] [Modify] [Create Snapshot]          │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                               │
│  ┌─────────────────────────────────────────────────────┐    │
│  │  S3 Bucket                                          │    │
│  │  my-app-assets-bucket                               │    │
│  │                                                      │    │
│  │  Region: us-east-1 | Size: 45.2 GB                  │    │
│  │  Versioning: Enabled | Encryption: AES256           │    │
│  │  Monthly Cost: $1.04                                │    │
│  │                                                      │    │
│  │  Tags:                                               │    │
│  │  • Name: assets-bucket                              │    │
│  │  • Project: web-app                                 │    │
│  │                                                      │    │
│  │  [View Details] [Configure Lifecycle]               │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                               │
│  Showing 847 resources | Total Cost: $12,450/month          │
│                                                               │
└─────────────────────────────────────────────────────────────┘
```

## User Flow
1. User clicks "Resources" in sidebar
2. Page loads all resources for tenant
3. User applies filters (type, region, status)
4. Resources list updates in real-time
5. User clicks "View Details" to see full metadata
6. Resource detail panel opens (right side)
7. Shows all metadata, tags, cost breakdown
8. User can perform actions (stop, terminate, modify)

## Data Requirements

### Input
- JWT token (from localStorage)
- Tenant ID (extracted from JWT)
- Filters: resource_type, region, status

### Output
```json
{
  "resources": [
    {
      "id": 29,
      "resource_id": "i-0a046ebb489ff3cd7",
      "resource_type": "ec2",
      "region": "us-east-1",
      "status": "running",
      "metadata": {
        "instance_type": "t3.large",
        "cpu_utilization": 23.5,
        "memory_utilization": 67.2,
        "network_in": 1024000,
        "network_out": 512000,
        "public_ip": "54.123.45.67",
        "private_ip": "10.0.1.25",
        "vpc_id": "vpc-12345678",
        "subnet_id": "subnet-87654321",
        "security_groups": ["sg-11111111", "sg-22222222"],
        "iam_instance_profile": "EC2-ReadOnly-Role",
        "monitoring": "enabled",
        "ebs_optimized": true,
        "root_device_type": "ebs",
        "virtualization_type": "hvm",
        "architecture": "x86_64",
        "hypervisor": "xen",
        "launch_time": "2025-01-15T08:30:00Z"
      },
      "tags": {
        "Name": "web-server-1",
        "Environment": "production",
        "Team": "backend",
        "CostCenter": "engineering"
      },
      "monthly_cost": 52.56,
      "created_at": "2025-01-31T10:00:00Z",
      "updated_at": "2025-01-31T12:00:00Z"
    }
  ],
  "total_count": 847,
  "total_cost": 12450.00
}
```

## API Endpoints

### GET /api/v1/resources
**Headers**:
```
Authorization: Bearer <JWT_TOKEN>
```

**Query Params**:
- `resource_type` (optional: ec2, rds, s3)
- `region` (optional: us-east-1, us-west-2, etc.)
- `status` (optional: running, stopped, available)
- `page` (default: 1)
- `limit` (default: 50)

**Response (200)**:
```json
{
  "resources": [
    {
      "id": 29,
      "resource_id": "i-0a046ebb489ff3cd7",
      "resource_type": "ec2",
      "region": "us-east-1",
      "status": "running",
      "metadata": {...},
      "tags": {...},
      "monthly_cost": 52.56
    }
  ],
  "total_count": 847,
  "page": 1,
  "limit": 50
}
```

### GET /api/v1/resources/:resource_id
**Headers**:
```
Authorization: Bearer <JWT_TOKEN>
```

**Response (200)**:
```json
{
  "id": 29,
  "resource_id": "i-0a046ebb489ff3cd7",
  "resource_type": "ec2",
  "region": "us-east-1",
  "status": "running",
  "metadata": {
    "instance_type": "t3.large",
    "cpu_utilization": 23.5,
    "memory_utilization": 67.2,
    "public_ip": "54.123.45.67",
    "private_ip": "10.0.1.25",
    "security_groups": ["sg-11111111"],
    "launch_time": "2025-01-15T08:30:00Z"
  },
  "tags": {
    "Name": "web-server-1",
    "Environment": "production"
  },
  "monthly_cost": 52.56,
  "cost_breakdown": {
    "compute": 45.00,
    "storage": 5.56,
    "data_transfer": 2.00
  }
}
```

### GET /api/v1/resources/stats
**Headers**:
```
Authorization: Bearer <JWT_TOKEN>
```

**Response (200)**:
```json
{
  "total_resources": 847,
  "by_type": {
    "ec2": 450,
    "rds": 120,
    "s3": 277
  },
  "by_region": {
    "us-east-1": 520,
    "us-west-2": 200,
    "eu-west-1": 127
  },
  "total_monthly_cost": 12450.00
}
```

## Database Tables

### yt_tenant_resources
- `id` (serial, primary key)
- `tenant_id` (integer)
- `resource_id` (varchar, AWS resource ID)
- `resource_type` (varchar: ec2, rds, s3)
- `region` (varchar)
- `status` (varchar: running, stopped, available)
- `metadata` (jsonb, all AWS metadata)
- `tags` (jsonb, AWS tags)
- `monthly_cost` (decimal)
- `created_at` (timestamp)
- `updated_at` (timestamp)

## UI Components

### Page
- **Path**: `/resources`
- **File**: `frontend/src/pages/Resources.tsx`

### Components Used
- FilterBar (type, region, status dropdowns)
- ResourceCard (individual resource card)
- ResourcePanel (right-side detail panel)
- ResourceInfoTab (metadata display)
- ResourceTagsTab (tags display)
- ResourceCostTab (cost breakdown)
- Pagination (page navigation)

## Business Rules
1. Only authenticated users can access
2. Data filtered by tenant_id from JWT
3. Resources sorted by monthly_cost (descending)
4. Pagination: 50 resources per page
5. Metadata stored as JSONB (flexible schema)
6. Tags stored as JSONB (key-value pairs)
7. Monthly cost estimated from instance type + usage
8. Resources updated on every scan

## Security Features
- ✅ JWT-based authentication
- ✅ Tenant isolation (no cross-tenant data)
- ✅ Read-only access (no modify/delete actions yet)
- ✅ Metadata sanitized (no sensitive data exposed)

## Implementation Status
- ✅ Frontend: `frontend/src/pages/Resources.tsx`
- ✅ Frontend: `frontend/src/components/ResourcePanel/ResourcePanel.tsx`
- ✅ Frontend: `frontend/src/components/ResourceDetails/ResourceInfoTab.tsx`
- ✅ Frontend: `frontend/src/components/ResourceDetails/ResourceTagsTab.tsx`
- ✅ Backend: `internal/api/handlers/resources.go`
- ✅ Backend: `internal/scanner/aws_scanner.go` (resource discovery)
- ✅ Database: `yt_tenant_resources` table
- ✅ Testing: Manual testing complete
- ✅ Deployment: Live in Docker container

## Known Issues
- Cost breakdown shows mock data (need real AWS Cost Explorer)
- Actions (stop, terminate) not implemented (read-only for now)

## Future Enhancements
- Add resource actions (stop, start, terminate, modify)
- Add cost history chart (last 30 days per resource)
- Add resource recommendations (right-sizing)
- Add bulk operations (stop all idle, delete all unattached)
- Add resource search (by name, tag, ID)
- Add export to CSV
