# Yukti License API

**Purpose**: Validate customer licenses for self-hosted Yukti deployments  
**Deployment**: Runs in YOUR cloud (license.yukti.io)  
**Tech Stack**: Go 1.23 + PostgreSQL 15 + JWT

---

## 🎯 What It Does

This is a **minimal microservice** that:
1. Validates license keys
2. Issues short-lived JWTs (24h expiry)
3. Tracks activations (which customer instances are active)
4. Enforces feature flags
5. **Does NOT store customer data** (only license metadata)

---

## 🏗️ Architecture

```
Customer Instance (Self-Hosted Yukti)
    ↓ (on startup + every 24h)
    POST /api/v1/license/validate
    {
      "license_key": "YUKTI-XXXX-XXXX-XXXX",
      "instance_id": "k8s-cluster-hash",
      "version": "1.0.0"
    }
    ↓
License API (Your Cloud)
    ↓ (validates key)
    ↓ (checks expiry, limits)
    ↓ (issues JWT)
    ↓
Response
    {
      "valid": true,
      "jwt": "eyJhbGc...",
      "expires_at": "2025-02-01T10:00:00Z",
      "features": {
        "max_detectors": 77,
        "support_level": "pro",
        "custom_detectors": true
      }
    }
```

---

## 📊 Database Schema

### `license_keys`
```sql
CREATE TABLE license_keys (
    id SERIAL PRIMARY KEY,
    license_key VARCHAR(255) UNIQUE NOT NULL,
    customer_email VARCHAR(255) NOT NULL,
    plan VARCHAR(50) NOT NULL,  -- community, pro, enterprise
    max_instances INT DEFAULT 1,
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

### `activations`
```sql
CREATE TABLE activations (
    id SERIAL PRIMARY KEY,
    license_key_id INT REFERENCES license_keys(id),
    instance_id VARCHAR(255) NOT NULL,
    instance_version VARCHAR(50),
    last_check_in TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(license_key_id, instance_id)
);
```

### `feature_flags`
```sql
CREATE TABLE feature_flags (
    id SERIAL PRIMARY KEY,
    plan VARCHAR(50) NOT NULL,
    feature_name VARCHAR(100) NOT NULL,
    enabled BOOLEAN DEFAULT true,
    metadata JSONB,
    UNIQUE(plan, feature_name)
);
```

---

## 🔌 API Endpoints

### 1. Validate License
```http
POST /api/v1/license/validate
Content-Type: application/json

{
  "license_key": "YUKTI-XXXX-XXXX-XXXX",
  "instance_id": "k8s-cluster-abc123",
  "version": "1.0.0"
}
```

**Response (200 OK)**:
```json
{
  "valid": true,
  "jwt": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2025-02-01T10:00:00Z",
  "features": {
    "max_detectors": 77,
    "support_level": "pro",
    "custom_detectors": true,
    "max_resources": -1
  }
}
```

**Response (401 Unauthorized)**:
```json
{
  "valid": false,
  "error": "LICENSE_EXPIRED",
  "message": "License expired on 2025-01-15. Please renew.",
  "renewal_url": "https://yukti.io/renew"
}
```

### 2. Check Status
```http
GET /api/v1/license/status
Authorization: Bearer <JWT>
```

**Response (200 OK)**:
```json
{
  "license_key": "YUKTI-****-****-XXXX",
  "plan": "pro",
  "expires_at": "2025-12-31T23:59:59Z",
  "days_remaining": 334,
  "active_instances": 1,
  "max_instances": 3
}
```

### 3. Refresh JWT
```http
POST /api/v1/license/refresh
Authorization: Bearer <OLD_JWT>
```

**Response (200 OK)**:
```json
{
  "jwt": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2025-02-02T10:00:00Z"
}
```

### 4. Deactivate Instance
```http
DELETE /api/v1/license/deactivate
Authorization: Bearer <JWT>

{
  "instance_id": "k8s-cluster-abc123"
}
```

**Response (200 OK)**:
```json
{
  "message": "Instance deactivated successfully"
}
```

---

## 🔐 Security

### JWT Claims
```json
{
  "iss": "license.yukti.io",
  "sub": "YUKTI-XXXX-XXXX-XXXX",
  "exp": 1738406400,
  "iat": 1738320000,
  "features": {
    "max_detectors": 77,
    "support_level": "pro",
    "custom_detectors": true
  }
}
```

### Rate Limiting
- 10 requests per minute per license key
- 100 requests per hour per IP

### Offline Grace Period
- If License API is unreachable, customer instance continues working for **7 days**
- After 7 days, features are locked (read-only mode)

---

## 🚀 Deployment

### Local Development
```bash
cd customer-hosted/license-api

# Start PostgreSQL
docker-compose up -d postgres

# Run migrations
psql -U yukti -d yukti_license < migrations/001_init.sql

# Start API
go run cmd/main.go
```

### Production (Docker)
```bash
# Build image
docker build -t yukti-license-api:latest .

# Run container
docker run -d \
  -p 8080:8080 \
  -e DATABASE_URL=postgres://... \
  -e JWT_SECRET=... \
  yukti-license-api:latest
```

### Production (Kubernetes)
```bash
# Deploy to your cluster
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
kubectl apply -f k8s/ingress.yaml
```

---

## 📋 License Plans

### Community (Free)
- 5 detectors
- 1 instance
- Community support (forum)
- No custom detectors

### Pro ($299/month)
- All 77 detectors
- 3 instances
- Email support
- Custom detectors

### Enterprise ($999/month)
- All 77 detectors
- Unlimited instances
- Slack support + SLA
- Custom detectors
- Priority features

---

## 🧪 Testing

### Generate Test License
```bash
go run scripts/generate_license.go \
  --email test@example.com \
  --plan pro \
  --expires 2025-12-31
```

Output:
```
License Key: YUKTI-A1B2-C3D4-E5F6
Plan: pro
Expires: 2025-12-31
```

### Validate License
```bash
curl -X POST http://localhost:8080/api/v1/license/validate \
  -H "Content-Type: application/json" \
  -d '{
    "license_key": "YUKTI-A1B2-C3D4-E5F6",
    "instance_id": "test-instance",
    "version": "1.0.0"
  }'
```

---

## 📊 Monitoring

### Metrics (Prometheus)
- `license_validations_total` - Total validation requests
- `license_validations_failed` - Failed validations
- `active_instances` - Currently active instances
- `license_expirations_7d` - Licenses expiring in 7 days

### Alerts
- License API down
- High validation failure rate (>10%)
- Database connection issues
- JWT secret rotation needed

---

## 🔧 Configuration

### Environment Variables
```bash
# Database
DATABASE_URL=postgres://user:pass@localhost:5432/yukti_license

# JWT
JWT_SECRET=your-secret-key-here
JWT_EXPIRY=24h

# Server
PORT=8080
ENV=production

# Rate Limiting
RATE_LIMIT_PER_MINUTE=10
RATE_LIMIT_PER_HOUR=100

# Offline Grace Period
OFFLINE_GRACE_PERIOD=168h  # 7 days
```

---

## 📝 Implementation Status

- [ ] Database schema (migrations/001_init.sql)
- [ ] Models (internal/models/license.go)
- [ ] Database layer (internal/database/db.go)
- [ ] JWT generation (internal/auth/jwt.go)
- [ ] API handlers (internal/api/handlers.go)
- [ ] Main entry point (cmd/main.go)
- [ ] Docker setup (Dockerfile, docker-compose.yml)
- [ ] License generator script (scripts/generate_license.go)
- [ ] Tests (internal/api/handlers_test.go)
- [ ] Documentation (this file)

---

## 🎯 Next Steps

1. Implement database schema
2. Build API handlers
3. Add JWT generation
4. Create license generator script
5. Write tests
6. Deploy to production

---

**Last Updated**: January 31, 2025  
**Status**: Planning phase - ready for implementation
