# Comprehensive Logging Implementation

## Overview
Added comprehensive logging to 10 critical files for operations engineers to quickly identify and fix issues.

## Logging Levels Used

- **[INFO]** - Normal operations, successful actions
- **[WARN]** - Warning conditions, potential issues
- **[ERROR]** - Error conditions, failed operations
- **[DEBUG]** - Detailed debugging information
- **[FATAL]** - Critical errors that stop the application

## Files with Logging Added

### 1. cmd/main.go
**Purpose**: Application entry point
**Logs Added**:
- Application startup banner
- Configuration loading
- Database connection status
- Server initialization
- Port and endpoint information
- Fatal errors with context

**Example Logs**:
```
[INFO] ========================================
[INFO] Yukti FinOps Platform Starting...
[INFO] ========================================
[INFO] Loading configuration...
[INFO] Configuration loaded successfully
[INFO] Connecting to database...
[INFO] Database connection established
[INFO] Server starting on port 8080
[INFO] Health check: http://localhost:8080/health
```

### 2. internal/api/server.go
**Purpose**: API server initialization
**Logs Added**:
- Server initialization
- Route setup
- CORS configuration
- Server start/stop events
- Server errors

**Example Logs**:
```
[INFO] Initializing API server...
[INFO] Setting up routes...
[INFO] API server initialized successfully
[INFO] Configuring CORS...
[INFO] Starting HTTP server on :8080
[ERROR] Server failed to start: address already in use
```

### 3. internal/api/routes/routes.go
**Purpose**: Route registration
**Logs Added**:
- Route setup start
- Middleware initialization
- Handler initialization
- Public routes registration
- Admin routes registration
- Customer routes registration
- Protected routes registration
- Health check requests

**Example Logs**:
```
[INFO] Setting up API routes...
[DEBUG] Initializing middleware...
[DEBUG] Initializing handlers...
[DEBUG] Registering public routes...
[DEBUG] Registering admin routes...
[DEBUG] Health check from IP: 192.168.1.1
[INFO] All routes registered successfully
```

### 4. internal/database/database.go
**Purpose**: Database connection management
**Logs Added**:
- Connection attempts
- Connection success/failure
- GORM initialization
- Auto-migration start/completion
- Migration errors

**Example Logs**:
```
[INFO] Connecting to database...
[DEBUG] Opening GORM connection to PostgreSQL
[INFO] Running database auto-migration...
[INFO] Database migration completed successfully
[INFO] Database connection established successfully
[ERROR] Failed to connect to database: connection refused
```

### 5. internal/api/handlers/admin.go
**Purpose**: Admin API handlers
**Logs Added**:
- GetCustomers requests with IP
- Query execution
- Row scanning errors
- Customer count
- GetMetrics requests
- Metric calculation errors
- Metric values
- Impersonation requests
- Impersonation logging
- Audit log failures

**Example Logs**:
```
[INFO] Admin GetCustomers called from IP: 192.168.1.1
[INFO] Successfully fetched 3 customers
[INFO] Admin GetMetrics called from IP: 192.168.1.1
[INFO] Metrics: customers=3, trials=1, savings=4434.20, mrr=198.00
[WARN] Admin admin@yukti.com impersonating tenant tenant-001 from IP 192.168.1.1
[ERROR] Failed to query customers: connection lost
```

### 6. internal/api/handlers/audit.go
**Purpose**: Audit log viewing
**Logs Added**:
- Audit log requests with IP
- Query limit
- Query execution
- Row scanning errors
- Result count

**Example Logs**:
```
[INFO] GetAuditLogs called from IP: 192.168.1.1
[DEBUG] Fetching audit logs with limit: 100
[INFO] Successfully fetched 15 audit logs
[WARN] Failed to scan audit log row: invalid data
[ERROR] Failed to query audit logs: table not found
```

### 7. internal/api/handlers/customers.go
**Purpose**: Customer API handlers
**Logs Added**:
- GetDashboard requests with tenant and IP
- Missing tenant_id warnings
- Data fetching for findings, budget, RI
- Dashboard data summary
- GetFindings requests with filters
- Query execution
- Result count
- CreateCustomer requests
- Validation failures
- Customer creation success

**Example Logs**:
```
[INFO] GetDashboard called for tenant: tenant-001 from IP: 192.168.1.1
[DEBUG] Fetching findings for tenant: tenant-001
[DEBUG] Fetching budget for tenant: tenant-001
[INFO] Dashboard data for tenant tenant-001: savings=486.20, findings=3, budget=15000.00
[INFO] GetFindings called for tenant: tenant-001, category: Storage, severity: High from IP: 192.168.1.1
[INFO] Returning 3 findings for tenant tenant-001
[INFO] CreateCustomer called from IP: 192.168.1.1
[INFO] Creating new customer: Acme Corp (tenant: tenant-abc123, email: admin@acme.com)
[INFO] Successfully created customer: Acme Corp with tenant_id: tenant-abc123
[WARN] GetDashboard called without tenant_id
[ERROR] Failed to get findings for tenant tenant-001: connection timeout
```

### 8. internal/api/middleware/admin_auth.go
**Purpose**: Admin authentication
**Logs Added**:
- Auth check with path and IP
- Unauthorized attempts
- Successful authentication
- Audit logging
- Audit log failures

**Example Logs**:
```
[DEBUG] Admin auth check for /api/admin/customers from IP: 192.168.1.1
[INFO] Admin authenticated for /api/admin/customers
[DEBUG] Logging admin action: user=admin@yukti.com, action=admin_access, path=/api/admin/customers
[WARN] Unauthorized admin access attempt from IP: 192.168.1.100, path: /api/admin/customers
[ERROR] Failed to log admin action to audit table: connection lost
```

### 9. internal/api/middleware/tenant_isolation.go
**Purpose**: Tenant isolation validation
**Logs Added**:
- Tenant validation with tenant_id, path, IP
- Missing tenant_id warnings
- Invalid tenant_id warnings
- Validation success
- Database errors

**Example Logs**:
```
[DEBUG] Tenant validation for tenant_id: tenant-001, path: /api/customers/dashboard, IP: 192.168.1.1
[INFO] Tenant tenant-001 validated successfully
[WARN] Missing tenant_id in request from IP: 192.168.1.1, path: /api/customers/dashboard
[WARN] Invalid tenant_id invalid-tenant attempted from IP: 192.168.1.100, path: /api/customers/dashboard
[ERROR] Failed to validate tenant tenant-001: database connection lost
```

### 10. internal/api/middleware/ratelimit.go
**Purpose**: Rate limiting
**Logs Added**:
- Rate limiter initialization
- Rate limit exceeded warnings
- Cleanup operations

**Example Logs**:
```
[INFO] Rate limiter initialized: 100 requests per minute
[WARN] Rate limit exceeded for client: 192.168.1.1, path: /api/customers/findings
[DEBUG] Rate limiter cleanup: removed 25 expired entries
```

## Log Format

All logs follow this format:
```
[LEVEL] Message with context
```

Examples:
```
[INFO] Server starting on port 8080
[WARN] Rate limit exceeded for client: 192.168.1.1
[ERROR] Failed to connect to database: connection refused
[DEBUG] Fetching findings for tenant: tenant-001
[FATAL] Server failed to start: port already in use
```

## Benefits for Operations Engineers

### 1. Quick Issue Identification
- Every request logged with IP address
- Every error logged with context
- Every database operation logged

### 2. Security Monitoring
- All admin actions logged
- Unauthorized access attempts logged
- Tenant isolation violations logged
- Rate limit violations logged

### 3. Performance Debugging
- Query execution logged
- Response times can be calculated
- Resource usage patterns visible

### 4. Audit Trail
- Complete request/response flow
- Admin impersonation tracked
- Customer actions tracked

## How to Use Logs

### View Real-time Logs
```bash
# All logs
docker-compose logs -f backend

# Only errors
docker-compose logs -f backend | grep ERROR

# Only warnings
docker-compose logs -f backend | grep WARN

# Specific tenant
docker-compose logs -f backend | grep "tenant-001"

# Admin actions
docker-compose logs -f backend | grep "Admin"
```

### Search for Issues
```bash
# Find all errors in last hour
docker-compose logs --since 1h backend | grep ERROR

# Find rate limit violations
docker-compose logs backend | grep "Rate limit exceeded"

# Find unauthorized access
docker-compose logs backend | grep "Unauthorized"

# Find database errors
docker-compose logs backend | grep "database"
```

### Monitor Specific Operations
```bash
# Monitor customer creation
docker-compose logs -f backend | grep "CreateCustomer"

# Monitor impersonation
docker-compose logs -f backend | grep "impersonat"

# Monitor tenant validation
docker-compose logs -f backend | grep "Tenant validation"
```

## Log Levels Guide

### When to Check Each Level

**[INFO]** - Normal operations
- Check for: Application flow, successful operations
- Frequency: Always visible

**[WARN]** - Potential issues
- Check for: Rate limits, validation failures, missing data
- Frequency: Review daily

**[ERROR]** - Actual problems
- Check for: Failed operations, database errors, API failures
- Frequency: Review immediately

**[DEBUG]** - Detailed information
- Check for: Troubleshooting specific issues
- Frequency: Only when debugging

**[FATAL]** - Critical failures
- Check for: Application crashes, startup failures
- Frequency: Immediate alert

## Production Recommendations

1. **Log Aggregation**: Send logs to ELK/Splunk/CloudWatch
2. **Alerting**: Set up alerts for ERROR and FATAL logs
3. **Retention**: Keep logs for 30-90 days
4. **Monitoring**: Dashboard for log metrics
5. **Analysis**: Regular log analysis for patterns

## Next Steps

1. Add structured logging (JSON format)
2. Add request ID tracking
3. Add performance metrics
4. Add business metrics
5. Add custom log levels per environment
