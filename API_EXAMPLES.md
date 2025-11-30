# API Examples - Yukti FinOps Platform

## Authentication

All admin endpoints require the `X-Admin-Key` header:
```bash
X-Admin-Key: yukti-admin-2024
X-Admin-User: admin@yukti.com
```

## Admin Endpoints

### 1. Get All Customers
```bash
curl -X GET http://localhost:8080/api/admin/customers \
  -H "X-Admin-Key: yukti-admin-2024" \
  -H "X-Admin-User: admin@yukti.com"
```

**Response:**
```json
{
  "success": true,
  "customers": [
    {
      "id": "uuid",
      "tenant_id": "tenant-001",
      "company_name": "Acme Corp",
      "email": "admin@acme.com",
      "onboarding_status": "completed",
      "created_at": "2024-01-15",
      "total_savings": 15000,
      "findings_count": 12
    }
  ],
  "count": 3
}
```

### 2. Get Platform Metrics
```bash
curl -X GET http://localhost:8080/api/admin/metrics \
  -H "X-Admin-Key: yukti-admin-2024"
```

**Response:**
```json
{
  "success": true,
  "metrics": {
    "total_customers": 3,
    "total_savings": 45000,
    "active_trials": 1,
    "mrr": 198
  }
}
```

### 3. Impersonate Tenant (with audit logging)
```bash
curl -X POST http://localhost:8080/api/admin/impersonate \
  -H "X-Admin-Key: yukti-admin-2024" \
  -H "X-Admin-User: admin@yukti.com" \
  -H "Content-Type: application/json" \
  -d '{"tenant_id": "tenant-001"}'
```

**Response:**
```json
{
  "success": true,
  "tenant_id": "tenant-001"
}
```

### 4. Get Audit Logs (Security Team)
```bash
curl -X GET "http://localhost:8080/api/admin/audit-logs?limit=50" \
  -H "X-Admin-Key: yukti-admin-2024"
```

**Response:**
```json
{
  "success": true,
  "logs": [
    {
      "id": "uuid",
      "admin_user": "admin@yukti.com",
      "action": "impersonate_tenant",
      "resource_type": "customer",
      "tenant_id": "tenant-001",
      "ip_address": "192.168.1.1",
      "created_at": "2024-01-15 10:30:00"
    }
  ]
}
```

## Customer Endpoints

### 5. Get Customer Dashboard
```bash
curl -X GET "http://localhost:8080/api/customers/dashboard?tenant_id=tenant-001"
```

**Response:**
```json
{
  "success": true,
  "data": {
    "total_savings": 15000,
    "findings_count": 12,
    "budget_amount": 50000,
    "current_spend": 42000,
    "ri_savings": 8000
  }
}
```

### 6. Get Hidden Cost Findings
```bash
# All findings
curl -X GET "http://localhost:8080/api/customers/findings?tenant_id=tenant-001"

# Filter by category
curl -X GET "http://localhost:8080/api/customers/findings?tenant_id=tenant-001&category=Data%20Transfer"

# Filter by severity
curl -X GET "http://localhost:8080/api/customers/findings?tenant_id=tenant-001&severity=Critical"

# Combined filters
curl -X GET "http://localhost:8080/api/customers/findings?tenant_id=tenant-001&category=Compute&severity=High"
```

**Response:**
```json
{
  "success": true,
  "findings": [
    {
      "id": "uuid",
      "detector_name": "unattached-ebs-volumes",
      "category": "Storage",
      "severity": "High",
      "title": "Unattached EBS Volume Detected",
      "description": "EBS volume vol-123 is not attached to any instance",
      "resource_arn": "arn:aws:ec2:us-east-1:123:volume/vol-123",
      "estimated_savings": 1200,
      "confidence": 0.95,
      "created_at": "2024-01-15 10:00:00"
    }
  ]
}
```

### 7. Create New Customer (Onboarding)
```bash
curl -X POST http://localhost:8080/api/customers \
  -H "Content-Type: application/json" \
  -d '{
    "company_name": "New Startup Inc",
    "email": "admin@newstartup.com"
  }'
```

**Response:**
```json
{
  "success": true,
  "tenant_id": "tenant-abc123",
  "id": "uuid"
}
```

## Error Responses

### 401 Unauthorized (Missing Admin Key)
```json
{
  "error": "Unauthorized - Admin access required"
}
```

### 403 Forbidden (Invalid Tenant)
```json
{
  "error": "Invalid tenant_id"
}
```

### 400 Bad Request (Missing Parameters)
```json
{
  "success": false,
  "error": "tenant_id required"
}
```

### 500 Internal Server Error
```json
{
  "success": false,
  "error": "Failed to fetch customers"
}
```

## Testing Multi-Tenant Isolation

### Valid Tenant (Should Work)
```bash
curl -X GET "http://localhost:8080/api/customers/dashboard?tenant_id=tenant-001"
# Returns 200 OK with data
```

### Invalid Tenant (Should Fail)
```bash
curl -X GET "http://localhost:8080/api/customers/dashboard?tenant_id=invalid-tenant"
# Returns 403 Forbidden
```

### Missing Tenant (Should Fail)
```bash
curl -X GET "http://localhost:8080/api/customers/dashboard"
# Returns 400 Bad Request
```

## Health Check

```bash
curl -X GET http://localhost:8080/health
```

**Response:**
```json
{
  "status": "healthy"
}
```

## Rate Limiting

Protected endpoints have rate limiting:
- 100 requests per minute per IP
- Returns 429 Too Many Requests if exceeded

## CORS

All endpoints support CORS with:
- Allowed Origins: * (configure for production)
- Allowed Methods: GET, POST, PUT, DELETE, OPTIONS
- Allowed Headers: Content-Type, X-Admin-Key, X-Admin-User
