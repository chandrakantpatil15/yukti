# Yukti FinOps - API Documentation

## Base URL
```
Production: https://api.yukti.io
Staging: https://staging-api.yukti.io
Local: http://localhost:8080
```

## Authentication

### API Key Authentication
```http
GET /api/v1/resources
Headers:
  X-API-Key: {tenant-code}_{api-key}
```

### JWT Authentication
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@company.com",
  "password": "secure_password"
}

Response:
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2024-11-07T10:00:00Z"
}
```

## Endpoints

### Health Check
```http
GET /health

Response: 200 OK
{
  "status": "healthy",
  "version": "1.0.0",
  "timestamp": "2024-11-06T10:00:00Z"
}
```

### Resources

#### List Resources
```http
GET /api/v1/resources?page=1&per_page=50&type=ec2&region=us-east-1

Response: 200 OK
{
  "success": true,
  "data": [
    {
      "id": 1,
      "resource_id": "i-1234567890",
      "resource_type": "ec2",
      "instance_type": "t3.medium",
      "region": "us-east-1",
      "state": "running",
      "tags": {
        "Environment": "production",
        "Project": "web-app"
      },
      "monthly_cost": 520.00,
      "utilization": {
        "cpu": 15.5,
        "memory": 42.3
      }
    }
  ],
  "meta": {
    "page": 1,
    "per_page": 50,
    "total": 127,
    "total_pages": 3
  }
}
```

#### Get Resource Details
```http
GET /api/v1/resources/{resource_id}

Response: 200 OK
{
  "success": true,
  "data": {
    "id": 1,
    "resource_id": "i-1234567890",
    "details": {
      "launch_time": "2024-01-15T10:00:00Z",
      "vpc_id": "vpc-12345",
      "subnet_id": "subnet-67890",
      "security_groups": ["sg-11111"],
      "iam_role": "ec2-web-role"
    },
    "utilization_history": {
      "7_day": {"cpu_avg": 15.5, "memory_avg": 42.3},
      "30_day": {"cpu_avg": 18.2, "memory_avg": 45.1}
    },
    "recommendations": [
      {
        "type": "downsize",
        "from": "t3.medium",
        "to": "t3.small",
        "savings": 260.00,
        "confidence": 0.92
      }
    ]
  }
}
```

#### Get Resource Statistics
```http
GET /api/v1/resources/stats

Response: 200 OK
{
  "success": true,
  "data": {
    "total_resources": 127,
    "total_cost": 45230.00,
    "breakdown": [
      {
        "type": "ec2",
        "count": 87,
        "cost": 25000.00
      },
      {
        "type": "rds",
        "count": 23,
        "cost": 12000.00
      }
    ]
  }
}
```

### Recommendations

#### List Recommendations
```http
GET /api/v1/recommendations?status=pending&severity=high

Response: 200 OK
{
  "success": true,
  "data": {
    "recommendations": [
      {
        "id": 1,
        "type": "downsize",
        "severity": "high",
        "resource_id": "i-1234567890",
        "current_cost": 520.00,
        "optimized_cost": 260.00,
        "monthly_savings": 260.00,
        "confidence_score": 0.92,
        "reason": "CPU utilization below 20% for 30 days",
        "status": "pending",
        "created_at": "2024-11-01T10:00:00Z"
      }
    ],
    "total_savings": 12496.00,
    "count": 23
  }
}
```

#### Accept Recommendation
```http
POST /api/v1/recommendations/{id}/accept

Response: 200 OK
{
  "success": true,
  "data": {
    "recommendation_id": 1,
    "status": "accepted",
    "iac_script": "terraform { ... }",
    "deployment_instructions": "Run: terraform apply"
  }
}
```

#### Reject Recommendation
```http
POST /api/v1/recommendations/{id}/reject
Content-Type: application/json

{
  "reason": "Business critical resource"
}

Response: 200 OK
{
  "success": true,
  "message": "Recommendation rejected"
}
```

### Whitelisting

#### Whitelist Resource
```http
POST /api/v1/resources/{resource_id}/whitelist
Content-Type: application/json

{
  "reason": "business_critical",
  "custom_reason": "Production database - needs headroom for Black Friday",
  "scope": "all",
  "duration": "90_days",
  "approved_by": "manager@company.com"
}

Response: 201 Created
{
  "success": true,
  "data": {
    "whitelist_id": 1,
    "resource_id": "db-prod-master",
    "reason": "business_critical",
    "expires_at": "2025-02-06T10:00:00Z",
    "whitelisted_by": "john.doe@company.com",
    "whitelisted_at": "2024-11-06T10:00:00Z"
  }
}
```

#### List Whitelisted Resources
```http
GET /api/v1/whitelists?expiring_soon=true

Response: 200 OK
{
  "success": true,
  "data": [
    {
      "id": 1,
      "resource_id": "db-prod-master",
      "reason": "business_critical",
      "custom_reason": "Production database",
      "expires_at": "2025-02-06T10:00:00Z",
      "whitelisted_by": "john.doe@company.com",
      "cost_impact": 1800.00
    }
  ],
  "total_whitelisted": 12,
  "total_cost_impact": 8450.00
}
```

#### Remove Whitelist
```http
DELETE /api/v1/whitelists/{id}

Response: 200 OK
{
  "success": true,
  "message": "Whitelist removed"
}
```

### ML Forecasting

#### Cost Forecast
```http
POST /api/v1/ml/forecast
Content-Type: application/json

{
  "tenant_id": 1,
  "forecast_days": 30,
  "historical_data": [
    {"date": "2024-10-01", "cost": 42000},
    {"date": "2024-10-02", "cost": 42500}
  ]
}

Response: 200 OK
{
  "tenant_id": 1,
  "forecast": [
    {
      "date": "2024-11-01",
      "predicted_cost": 47800.00,
      "confidence": 0.85
    }
  ],
  "total_predicted_cost": 47800.00,
  "trend": "increasing",
  "model": "linear_regression"
}
```

#### Anomaly Detection
```http
POST /api/v1/ml/anomaly-detect
Content-Type: application/json

{
  "tenant_id": 1,
  "threshold": 2.0,
  "historical_data": [...]
}

Response: 200 OK
{
  "tenant_id": 1,
  "anomalies": [
    {
      "date": "2024-10-15",
      "cost": 52000.00,
      "expected_cost": 42000.00,
      "deviation": 10000.00,
      "severity": "high",
      "z_score": 3.2
    }
  ],
  "anomaly_count": 2,
  "baseline_cost": 42000.00
}
```

### Reports

#### Generate Executive Report
```http
POST /api/v1/reports/executive
Content-Type: application/json

{
  "period": "monthly",
  "format": "pdf",
  "include_charts": true,
  "email_to": ["cfo@company.com"]
}

Response: 202 Accepted
{
  "success": true,
  "data": {
    "report_id": "rep_abc123",
    "status": "generating",
    "estimated_time": "30 seconds",
    "download_url": "https://api.yukti.io/reports/rep_abc123/download"
  }
}
```

#### Download Report
```http
GET /api/v1/reports/{report_id}/download

Response: 200 OK
Content-Type: application/pdf
Content-Disposition: attachment; filename="yukti-report-2024-11.pdf"

[PDF Binary Data]
```

### Hidden Cost Detection (Premium)

#### Scan for Hidden Costs
```http
POST /api/v1/hidden-costs/scan
Content-Type: application/json

{
  "categories": ["data_transfer", "storage", "networking", "licensing"]
}

Response: 200 OK
{
  "success": true,
  "data": {
    "total_hidden_costs": 18500.00,
    "findings": [
      {
        "category": "data_transfer",
        "type": "nat_gateway_fees",
        "current_cost": 3200.00,
        "optimized_cost": 0.00,
        "savings": 3200.00,
        "severity": "high",
        "description": "NAT Gateway data processing fees",
        "solution": "Replace with VPC Endpoints",
        "iac_script_available": true
      }
    ]
  }
}
```

## Rate Limiting

```
Free Tier: 100 requests/minute
Professional: 500 requests/minute
Enterprise: 2000 requests/minute
Financial: Unlimited
```

Rate limit headers:
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1699272000
```

## Error Responses

### 400 Bad Request
```json
{
  "success": false,
  "error": "Invalid request parameters",
  "details": {
    "field": "forecast_days",
    "message": "Must be between 1 and 90"
  }
}
```

### 401 Unauthorized
```json
{
  "success": false,
  "error": "Invalid or missing API key"
}
```

### 403 Forbidden
```json
{
  "success": false,
  "error": "Insufficient permissions",
  "required_tier": "ENTERPRISE"
}
```

### 429 Too Many Requests
```json
{
  "success": false,
  "error": "Rate limit exceeded",
  "retry_after": 60
}
```

### 500 Internal Server Error
```json
{
  "success": false,
  "error": "Internal server error",
  "request_id": "req_abc123"
}
```

## Webhooks

### Configure Webhook
```http
POST /api/v1/webhooks
Content-Type: application/json

{
  "url": "https://your-app.com/webhooks/yukti",
  "events": ["recommendation.created", "cost.anomaly", "whitelist.expiring"],
  "secret": "your_webhook_secret"
}
```

### Webhook Events

#### recommendation.created
```json
{
  "event": "recommendation.created",
  "timestamp": "2024-11-06T10:00:00Z",
  "data": {
    "recommendation_id": 1,
    "type": "downsize",
    "resource_id": "i-1234567890",
    "savings": 260.00
  }
}
```

#### cost.anomaly
```json
{
  "event": "cost.anomaly",
  "timestamp": "2024-11-06T10:00:00Z",
  "data": {
    "date": "2024-11-06",
    "cost": 52000.00,
    "expected_cost": 42000.00,
    "severity": "high"
  }
}
```

## SDKs

### JavaScript/TypeScript
```bash
npm install @yukti/sdk
```

```javascript
import { YuktiClient } from '@yukti/sdk';

const client = new YuktiClient({
  apiKey: 'your-api-key',
  baseUrl: 'https://api.yukti.io'
});

const resources = await client.resources.list({ page: 1 });
const forecast = await client.ml.forecast({ days: 30 });
```

### Python
```bash
pip install yukti-sdk
```

```python
from yukti import YuktiClient

client = YuktiClient(api_key='your-api-key')

resources = client.resources.list(page=1)
forecast = client.ml.forecast(days=30)
```

### Go
```bash
go get github.com/yukti-finops/sdk-go
```

```go
import "github.com/yukti-finops/sdk-go"

client := yukti.NewClient("your-api-key")
resources, _ := client.Resources.List(1, 50)
forecast, _ := client.ML.Forecast(30)
```

## Support

- Documentation: https://docs.yukti.io
- API Status: https://status.yukti.io
- Support: support@yukti.io
- Emergency: +1-800-YUKTI-911
