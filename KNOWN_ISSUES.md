# Known Issues & Development Gaps

**Last Updated**: 2025-11-12  
**Status**: Development Phase

---

## 🔴 Critical Issues (Blocking Production)

### 1. Missing AWS Account Configuration
**Issue**: `REACT_APP_YUKTI_AWS_ACCOUNT` is set to placeholder `123456789012`  
**Impact**: CloudFormation templates reference non-existent AWS account  
**Fix Required**: 
- Create Yukti platform AWS account
- Update `.env.development` and `.env.production` with real account ID
- Test cross-account IAM role assumption

**Files**:
- `frontend/.env.development`
- `frontend/.env.production`
- `frontend/src/components/Onboarding/OnboardingAws.tsx`

---

### 2. Missing Stripe Configuration
**Issue**: `STRIPE_SECRET_KEY` environment variable not set  
**Impact**: Billing functionality disabled  
**Fix Required**:
- Create Stripe account
- Add `STRIPE_SECRET_KEY` to backend environment
- Configure webhook endpoint
- Test payment flow

**Files**:
- `docker-compose.yml` (backend environment)
- `internal/api/handlers/billing.go`

---

### 3. Onboarding Flow Incomplete
**Issue**: Frontend calls `/api/onboarding/aws-connection` but backend expects different payload  
**Impact**: AWS connection setup fails  
**Status**: ✅ **FIXED** - Added route alias and external ID generation  
**Remaining**:
- Frontend needs to call `/api/onboarding/external-id` to get backend-generated external ID
- Update `OnboardingAws.tsx` to fetch external ID on mount

**Files**:
- `frontend/src/components/Onboarding/OnboardingAws.tsx`
- `internal/api/handlers/onboarding.go`

---

## 🟡 Medium Priority Issues

### 4. Missing Database Table: `yt_metrics_integrations`
**Issue**: Onboarding service references table that doesn't exist  
**Impact**: Metrics integration step will fail  
**Fix Required**:
```sql
CREATE TABLE yt_metrics_integrations (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    source VARCHAR(50) NOT NULL,
    endpoint TEXT,
    verified BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(tenant_id)
);
```

**Files**:
- `internal/onboarding/service.go` (line 59)
- Need new migration: `scripts/018_metrics_integrations.sql`

---

### 5. CORS Configuration Warning
**Issue**: `CORS_ALLOWED_ORIGINS` not set, using default  
**Impact**: May block requests from production frontend domain  
**Fix Required**:
- Add `CORS_ALLOWED_ORIGINS` to docker-compose.yml
- Set to production domain when deploying

**Files**:
- `docker-compose.yml`
- `cmd/main.go`

---

### 6. JWT Secret Hardcoded
**Issue**: JWT secret is `yukti-secret-key-change-in-production`  
**Impact**: Security vulnerability if deployed as-is  
**Fix Required**:
- Generate secure random secret (32+ characters)
- Store in environment variable
- Rotate periodically

**Files**:
- `docker-compose.yml`
- `.env.example`

---

### 7. External ID Generation on Frontend
**Issue**: External ID generated client-side with `Math.random()`  
**Impact**: Not cryptographically secure, predictable  
**Status**: ✅ **FIXED** - Backend now generates secure external ID  
**Remaining**: Update frontend to fetch from backend

**Files**:
- `frontend/src/components/Onboarding/OnboardingAws.tsx` (line 25)
- `internal/onboarding/service.go` (line 100)

---

## 🟢 Low Priority Issues

### 8. Missing Environment Variables Documentation
**Issue**: No centralized list of required environment variables  
**Fix Required**: Create `.env.example` with all variables

**Required Variables**:
```bash
# Backend
DATABASE_URL=postgresql://user:pass@host:5432/db
PORT=8081
JWT_SECRET=<secure-random-string>
AWS_REGION=us-east-1
ENVIRONMENT=development
STRIPE_SECRET_KEY=sk_test_xxx
CORS_ALLOWED_ORIGINS=http://localhost:3000,https://app.yukti.io

# Frontend
REACT_APP_API_URL=http://localhost:8081
REACT_APP_YUKTI_AWS_ACCOUNT=<your-aws-account-id>
NODE_ENV=development
```

---

### 9. No Health Check for ML Service
**Issue**: ML service has no health endpoint  
**Impact**: Can't verify service is running  
**Fix Required**: Add `/health` endpoint to `ml-service/api_ml.py`

---

### 10. Missing API Endpoints

#### a. Resource Details Endpoint
**Frontend calls**: `/api/v1/resources/{id}`  
**Backend has**: `/api/v1/resources/details?resource_id={id}`  
**Fix**: Add route alias or update frontend

#### b. IaC Generation Endpoint
**Frontend calls**: `/api/v1/iac/generate`  
**Backend has**: No route registered  
**Fix**: Add IaC handler and route

---

### 11. Mock Data in Development
**Issue**: Onboarding validation always succeeds (2-second delay)  
**Impact**: Can't test real AWS connection failures  
**Status**: Intentional for development  
**Action**: Remove mock before production

**Files**:
- `frontend/src/components/Onboarding/OnboardingAws.tsx` (line 177)

---

### 12. No Volume Mount for Frontend Hot Reload
**Issue**: Frontend requires full rebuild for code changes  
**Impact**: Slow development iteration  
**Fix**: Add volume mount in docker-compose.yml
```yaml
frontend:
  volumes:
    - ./frontend/src:/app/src
    - ./frontend/public:/app/public
```

---

## 📋 Missing Features (Not Bugs)

### 13. Email Verification Not Implemented
**Issue**: OTP codes generated but no email sent  
**Impact**: Users can't verify email  
**Required**:
- Configure SMTP (SendGrid, AWS SES, etc.)
- Implement email sending in `internal/api/handlers/auth.go`

---

### 14. AWS STS AssumeRole Not Implemented
**Issue**: Backend doesn't actually assume customer IAM roles  
**Impact**: Can't fetch real AWS data  
**Required**:
- Implement STS client in `internal/aws/`
- Add role assumption logic
- Cache temporary credentials

---

### 15. Hidden Cost Detectors Not Connected
**Issue**: 77 detectors exist but not triggered by onboarding  
**Impact**: No findings generated after onboarding  
**Required**:
- Trigger scan after AWS connection
- Schedule periodic scans
- Store findings in `yt_hidden_cost_findings`

---

## 🔧 Technical Debt

### 16. No Error Boundaries in Frontend
**Status**: ✅ **FIXED** - ErrorBoundary component exists  
**Remaining**: Add to more components

---

### 17. No Request Validation
**Issue**: Backend handlers don't validate request payloads  
**Impact**: Invalid data can cause panics  
**Fix**: Add validation middleware or use struct tags

---

### 18. No Rate Limiting on Auth Endpoints
**Issue**: Login/signup endpoints not rate limited  
**Impact**: Vulnerable to brute force attacks  
**Fix**: Apply rate limiter middleware to auth routes

---

### 19. SQL Injection Risk in Dynamic Queries
**Issue**: Some handlers build SQL strings dynamically  
**Status**: Most use parameterized queries ✅  
**Action**: Audit all SQL queries

---

### 20. No Logging Strategy
**Issue**: Inconsistent logging (some use log.Printf, some don't log)  
**Fix**: Implement structured logging (zerolog, zap)

---

## 🧪 Testing Gaps

### 21. No Unit Tests
**Coverage**: 0%  
**Required**: Add tests for critical paths (auth, billing, detectors)

---

### 22. No Integration Tests
**Required**: Test full onboarding flow, API endpoints

---

### 23. No E2E Tests
**Required**: Test frontend → backend → database flow

---

## 📊 Monitoring Gaps

### 24. Prometheus Metrics Not Exposed
**Issue**: Backend doesn't expose `/metrics` endpoint  
**Fix**: Add prometheus middleware

---

### 25. No Alerting Rules
**Required**: Configure Grafana alerts for errors, high latency

---

## 🔐 Security Issues

### 26. Passwords Stored as Plain Text in Logs
**Issue**: Login requests logged with passwords  
**Fix**: Sanitize logs, never log sensitive data

---

### 27. No HTTPS in Development
**Issue**: All traffic over HTTP  
**Fix**: Add TLS certificates for local development

---

### 28. No Input Sanitization
**Issue**: User inputs not sanitized  
**Impact**: XSS vulnerability  
**Fix**: Sanitize all user inputs on backend

---

## 🚀 Deployment Blockers

### 29. No CI/CD Pipeline
**Required**: GitHub Actions or similar for automated testing/deployment

---

### 30. No Kubernetes Manifests
**Issue**: `make k8s-deploy` references non-existent files  
**Required**: Create k8s/ directory with deployments, services, ingress

---

### 31. No Database Migration Strategy
**Issue**: Manual SQL scripts, no version tracking  
**Fix**: Use golang-migrate or similar tool

---

### 32. No Secrets Management
**Issue**: Secrets in docker-compose.yml  
**Fix**: Use AWS Secrets Manager, HashiCorp Vault, or Kubernetes secrets

---

## ✅ Recently Fixed Issues

1. ✅ Login navigation not working (localStorage key mismatch)
2. ✅ Onboarding 404 error (missing route)
3. ✅ External ID generation (moved to backend)
4. ✅ Navigation showing on onboarding page (added hideNavigation prop)
5. ✅ NODE_ENV not set (added to docker-compose)
6. ✅ Token storage inconsistency (centralized with setToken)

---

## 📝 Action Items Priority

### Immediate (Before Demo)
1. Fix onboarding external ID fetch from backend
2. Create `yt_metrics_integrations` table
3. Add mock AWS account ID to .env
4. Test full onboarding flow end-to-end

### Short Term (Before Beta)
1. Implement real AWS STS AssumeRole
2. Connect hidden cost detectors to onboarding
3. Add Stripe configuration
4. Implement email verification

### Medium Term (Before Production)
1. Add comprehensive error handling
2. Implement rate limiting on all endpoints
3. Add unit and integration tests
4. Set up CI/CD pipeline
5. Create Kubernetes manifests
6. Implement proper secrets management

### Long Term (Post-Launch)
1. Add monitoring and alerting
2. Implement database migration tool
3. Add E2E tests
4. Performance optimization
5. Security audit

---

**Note**: This document should be updated as issues are discovered and resolved.
