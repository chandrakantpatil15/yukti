# License Activation Model - Customer-Hosted Product

## Overview
Customer-hosted Yukti installations require continuous license validation via JWT tokens from the SaaS platform. If token refresh fails for 15 minutes, the UI becomes non-responsive.

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│              YUKTI SAAS PLATFORM (License Server)           │
│              AWS Account: 144403604430                      │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  License Management Service                                 │
│  ├── Generate JWT tokens (15-min expiry)                    │
│  ├── Validate customer subscription                         │
│  ├── Track active installations                             │
│  └── Revoke licenses (payment failure, abuse)               │
│                                                             │
│  API Endpoint: https://license.yukti.io/api/v1/validate    │
│                                                             │
└─────────────────────────────────────────────────────────────┘
                            ↓ JWT Token (every 15 min)
┌─────────────────────────────────────────────────────────────┐
│         CUSTOMER-HOSTED INSTALLATION                        │
│         Customer AWS Account: 424851482219                  │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  License Validation Client                                  │
│  ├── Request token every 15 minutes                         │
│  ├── Store token in Redis (TTL: 15 min)                     │
│  ├── Validate token on every API request                    │
│  └── Disable UI if token expired                            │
│                                                             │
│  Customer Data (stays in customer VPC)                      │
│  ├── PostgreSQL (cost data, resources)                      │
│  ├── Redis (cache + license token)                          │
│  └── S3 (reports, exports)                                  │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## JWT Token Structure

### License Token (from SaaS Platform)
```json
{
  "iss": "yukti-license-server",
  "sub": "customer-acme-corp",
  "customer_id": "cust_abc123",
  "installation_id": "inst_xyz789",
  "license_type": "enterprise",
  "features": ["aws_scanning", "ml_predictions", "iac_generation"],
  "max_users": 50,
  "max_aws_accounts": 10,
  "iat": 1738368000,
  "exp": 1738368900,
  "nbf": 1738368000
}
```

**Token Expiry**: 15 minutes (900 seconds)  
**Refresh Interval**: Every 12 minutes (720 seconds) - 3-minute buffer  
**Grace Period**: 3 minutes after expiry before UI lockout

---

## License Validation Flow

### Step 1: Initial Activation (First-Time Setup)

```
Customer Admin
    ↓
1. Purchase license from Yukti SaaS
    ↓
2. Receive activation key: YUKTI-ENTERPRISE-ABC123-XYZ789
    ↓
3. Deploy customer-hosted installation
    ↓
4. Enter activation key in UI
    ↓
5. Backend calls SaaS API: POST /api/v1/license/activate
    ↓
6. SaaS validates key + creates installation record
    ↓
7. Returns first JWT token (15-min expiry)
    ↓
8. Customer installation stores token in Redis
    ↓
9. UI becomes active
```

**Activation API Request**:
```bash
POST https://license.yukti.io/api/v1/license/activate
Content-Type: application/json

{
  "activation_key": "YUKTI-ENTERPRISE-ABC123-XYZ789",
  "installation_id": "inst_xyz789",
  "customer_email": "admin@acmecorp.com",
  "deployment_region": "us-east-1"
}
```

**Activation API Response**:
```json
{
  "success": true,
  "license_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2025-02-01T10:15:00Z",
  "refresh_interval": 720,
  "customer_id": "cust_abc123",
  "license_type": "enterprise",
  "features": ["aws_scanning", "ml_predictions", "iac_generation"]
}
```

---

### Step 2: Continuous Token Refresh (Every 12 Minutes)

```
Customer Installation (Background Service)
    ↓
Every 12 minutes:
    ↓
1. Check Redis for current token
    ↓
2. If token expires in < 5 minutes:
    ↓
3. Call SaaS API: POST /api/v1/license/refresh
    ↓
4. SaaS validates:
   - Customer subscription active?
   - Payment up to date?
   - Installation not revoked?
    ↓
5. Generate new JWT token (15-min expiry)
    ↓
6. Return token to customer installation
    ↓
7. Store new token in Redis (TTL: 15 min)
    ↓
8. Continue normal operation
```

**Refresh API Request**:
```bash
POST https://license.yukti.io/api/v1/license/refresh
Authorization: Bearer <current_license_token>
Content-Type: application/json

{
  "installation_id": "inst_xyz789",
  "customer_id": "cust_abc123"
}
```

**Refresh API Response (Success)**:
```json
{
  "success": true,
  "license_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2025-02-01T10:30:00Z",
  "refresh_interval": 720
}
```

**Refresh API Response (Failure - Payment Issue)**:
```json
{
  "success": false,
  "error": "SUBSCRIPTION_EXPIRED",
  "message": "Your subscription expired on 2025-01-15. Please renew.",
  "action_url": "https://yukti.io/billing/renew",
  "grace_period_ends": "2025-02-05T00:00:00Z"
}
```

---

### Step 3: Token Validation (Every API Request)

```
User Action (Dashboard, Resources, etc.)
    ↓
Frontend → Backend API Request
    ↓
Backend Middleware: ValidateLicenseToken()
    ↓
1. Extract token from Redis
    ↓
2. Check token expiry
    ↓
3. Verify JWT signature (RSA public key from SaaS)
    ↓
4. Validate claims (customer_id, installation_id)
    ↓
5. Check feature access (e.g., ml_predictions)
    ↓
If valid:
    → Process request normally
    ↓
If expired (< 15 min):
    → Return 402 Payment Required
    → UI shows "License expired" modal
    ↓
If expired (> 15 min):
    → Return 403 Forbidden
    → UI becomes non-responsive
    → Show "Contact support" message
```

---

## UI States

### State 1: Active License (Token Valid)
```
┌─────────────────────────────────────────────────────────────┐
│  YUKTI FINOPS - DASHBOARD        [Scan] [Profile] [Logout]  │
├─────────────────────────────────────────────────────────────┤
│  🟢 License Active • Enterprise Plan • Expires: Jan 31, 2026│
│                                                             │
│  [Normal dashboard content...]                              │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

### State 2: Token Refresh Failed (Grace Period)
```
┌─────────────────────────────────────────────────────────────┐
│  ⚠️ LICENSE WARNING                                          │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Unable to validate license with Yukti servers.             │
│                                                             │
│  Possible causes:                                           │
│  • Network connectivity issue                               │
│  • Yukti license server maintenance                         │
│  • Subscription payment pending                             │
│                                                             │
│  Grace period: 3 minutes remaining                          │
│                                                             │
│  [Retry Now] [Contact Support] [View Details]               │
│                                                             │
│  Dashboard will become read-only if not resolved.           │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

### State 3: License Expired (UI Locked)
```
┌─────────────────────────────────────────────────────────────┐
│  🔴 LICENSE EXPIRED                                          │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Your Yukti license has expired.                            │
│                                                             │
│  Reason: Subscription payment failed                        │
│  Expired on: February 1, 2025 at 10:15 AM                   │
│                                                             │
│  To restore access:                                         │
│  1. Update payment method at yukti.io/billing               │
│  2. Contact support: support@yukti.io                       │
│  3. Call: +1 (555) 123-4567                                 │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │         Renew Subscription                          │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  Your data is safe and will be restored after renewal.      │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## Backend Implementation

### License Validation Middleware (Go)

```go
// internal/api/middleware/license_validator.go
package middleware

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
    
    "github.com/golang-jwt/jwt/v5"
    "github.com/redis/go-redis/v9"
)

type LicenseValidator struct {
    redisClient *redis.Client
    publicKey   []byte // RSA public key from SaaS platform
    saasURL     string
}

func (lv *LicenseValidator) ValidateLicense(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Get token from Redis
        token, err := lv.redisClient.Get(r.Context(), "license:token").Result()
        if err != nil {
            lv.handleExpiredLicense(w, "NO_TOKEN")
            return
        }
        
        // Parse JWT
        claims, err := lv.parseToken(token)
        if err != nil {
            lv.handleExpiredLicense(w, "INVALID_TOKEN")
            return
        }
        
        // Check expiry
        if time.Now().Unix() > claims.ExpiresAt.Unix() {
            lv.handleExpiredLicense(w, "TOKEN_EXPIRED")
            return
        }
        
        // Add license info to context
        ctx := context.WithValue(r.Context(), "license", claims)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func (lv *LicenseValidator) handleExpiredLicense(w http.ResponseWriter, reason string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusPaymentRequired)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": false,
        "error":   "LICENSE_EXPIRED",
        "reason":  reason,
        "message": "License validation failed. Please contact support.",
    })
}
```

---

### Token Refresh Service (Go)

```go
// internal/services/license_refresh.go
package services

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
    
    "github.com/redis/go-redis/v9"
)

type LicenseRefreshService struct {
    redisClient    *redis.Client
    saasURL        string
    installationID string
    customerID     string
}

func (lrs *LicenseRefreshService) StartRefreshLoop() {
    ticker := time.NewTicker(12 * time.Minute) // Refresh every 12 minutes
    
    for range ticker.C {
        if err := lrs.refreshToken(); err != nil {
            log.Printf("[ERROR] License refresh failed: %v", err)
        }
    }
}

func (lrs *LicenseRefreshService) refreshToken() error {
    // Get current token
    currentToken, _ := lrs.redisClient.Get(ctx, "license:token").Result()
    
    // Call SaaS API
    req := map[string]string{
        "installation_id": lrs.installationID,
        "customer_id":     lrs.customerID,
    }
    
    body, _ := json.Marshal(req)
    httpReq, _ := http.NewRequest("POST", lrs.saasURL+"/api/v1/license/refresh", bytes.NewBuffer(body))
    httpReq.Header.Set("Authorization", "Bearer "+currentToken)
    httpReq.Header.Set("Content-Type", "application/json")
    
    resp, err := http.DefaultClient.Do(httpReq)
    if err != nil {
        return fmt.Errorf("network error: %w", err)
    }
    defer resp.Body.Close()
    
    var result struct {
        Success      bool   `json:"success"`
        LicenseToken string `json:"license_token"`
        ExpiresAt    string `json:"expires_at"`
    }
    
    json.NewDecoder(resp.Body).Decode(&result)
    
    if !result.Success {
        return fmt.Errorf("refresh failed")
    }
    
    // Store new token in Redis
    lrs.redisClient.Set(ctx, "license:token", result.LicenseToken, 15*time.Minute)
    
    log.Printf("[INFO] License token refreshed successfully")
    return nil
}
```

---

### Frontend Token Check (React)

```typescript
// frontend/src/services/licenseValidator.ts
import { useEffect, useState } from 'react';
import api from './api';

export const useLicenseValidator = () => {
  const [licenseStatus, setLicenseStatus] = useState<'active' | 'warning' | 'expired'>('active');
  
  useEffect(() => {
    const checkLicense = async () => {
      try {
        const response = await api.get('/api/v1/license/status');
        
        if (response.success) {
          setLicenseStatus('active');
        } else if (response.error === 'LICENSE_EXPIRED') {
          setLicenseStatus('expired');
        }
      } catch (error: any) {
        if (error.response?.status === 402) {
          setLicenseStatus('warning');
        } else if (error.response?.status === 403) {
          setLicenseStatus('expired');
        }
      }
    };
    
    // Check every 30 seconds
    const interval = setInterval(checkLicense, 30000);
    checkLicense(); // Initial check
    
    return () => clearInterval(interval);
  }, []);
  
  return licenseStatus;
};
```

---

## SaaS Platform License Management

### Database Schema

```sql
-- License management tables
CREATE TABLE yt_license_customers (
  id SERIAL PRIMARY KEY,
  customer_id VARCHAR(50) UNIQUE NOT NULL,
  company_name VARCHAR(255) NOT NULL,
  email VARCHAR(255) NOT NULL,
  license_type VARCHAR(50) NOT NULL, -- free, professional, enterprise
  subscription_status VARCHAR(50) NOT NULL, -- active, expired, suspended
  subscription_start DATE NOT NULL,
  subscription_end DATE NOT NULL,
  max_installations INT DEFAULT 1,
  max_users INT DEFAULT 10,
  max_aws_accounts INT DEFAULT 3,
  created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE yt_license_installations (
  id SERIAL PRIMARY KEY,
  installation_id VARCHAR(50) UNIQUE NOT NULL,
  customer_id VARCHAR(50) REFERENCES yt_license_customers(customer_id),
  activation_key VARCHAR(100) UNIQUE NOT NULL,
  activated_at TIMESTAMP,
  last_token_refresh TIMESTAMP,
  deployment_region VARCHAR(50),
  status VARCHAR(50) DEFAULT 'active', -- active, revoked, suspended
  created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE yt_license_token_logs (
  id SERIAL PRIMARY KEY,
  installation_id VARCHAR(50) REFERENCES yt_license_installations(installation_id),
  token_issued_at TIMESTAMP DEFAULT NOW(),
  token_expires_at TIMESTAMP NOT NULL,
  refresh_success BOOLEAN DEFAULT true,
  failure_reason VARCHAR(255)
);
```

---

## Security Considerations

### Token Signing
- **Algorithm**: RSA-256 (asymmetric encryption)
- **Private Key**: Stored in SaaS platform (never shared)
- **Public Key**: Embedded in customer installation (for verification)

### Network Security
- **HTTPS Only**: All license API calls over TLS 1.3
- **Rate Limiting**: Max 10 refresh requests per minute per installation
- **IP Whitelisting**: Optional customer IP restriction

### Abuse Prevention
- **Installation Fingerprinting**: Track deployment region, instance ID
- **Concurrent Installation Limit**: Max 1 active installation per license key
- **Revocation**: Instant token invalidation for abuse/non-payment

---

## Failure Scenarios

### Scenario 1: Network Outage (Customer Side)
- **Grace Period**: 3 minutes after token expiry
- **UI Behavior**: Show warning banner, allow read-only access
- **Resolution**: Auto-recover when network restored

### Scenario 2: SaaS Platform Downtime
- **Fallback**: Use cached token for up to 1 hour (emergency mode)
- **UI Behavior**: Show "Offline mode" banner
- **Resolution**: Resume normal validation when SaaS recovers

### Scenario 3: Payment Failure
- **Grace Period**: 7 days after subscription end date
- **UI Behavior**: Show payment reminder banner
- **Resolution**: Restore access immediately after payment

### Scenario 4: License Revoked (Abuse)
- **Immediate**: Token refresh returns 403 Forbidden
- **UI Behavior**: Complete lockout, show contact support message
- **Resolution**: Manual review by Yukti support team

---

## Pricing Impact

**Self-Hosted License Fees**:
- **Startup**: $5,000/year (1 installation, 10 users, 3 AWS accounts)
- **Professional**: $15,000/year (2 installations, 50 users, 10 AWS accounts)
- **Enterprise**: $50,000/year (5 installations, unlimited users, unlimited AWS accounts)

**License includes**:
- ✅ Continuous token validation
- ✅ Software updates
- ✅ Email support
- ✅ 99.9% SaaS uptime SLA

---

## Customer Benefits

1. **No Piracy**: License validation prevents unauthorized usage
2. **Automatic Updates**: SaaS can push feature flags via token
3. **Usage Tracking**: SaaS knows active installations, user count
4. **Flexible Licensing**: Easy to upgrade/downgrade plans
5. **Data Sovereignty**: Customer data stays in their VPC

---

## Implementation Timeline

**Week 1**: SaaS license server (API endpoints, database)
**Week 2**: Customer installation client (token refresh service)
**Week 3**: Frontend integration (license status UI)
**Week 4**: Testing (network failures, payment scenarios)
**Week 5**: Documentation + customer onboarding guide
