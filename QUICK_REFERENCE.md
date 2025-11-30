# Yukti - Quick Reference Guide

## 🚀 Quick Start Commands

```bash
# Start everything
make start

# Load test data
make seed

# Open admin dashboard
make admin

# View logs
make logs

# Stop everything
make stop

# Rebuild after code changes
docker-compose down && docker-compose up -d --build
```

## 🔑 Admin Access

**Admin Key:** `yukti-admin-2024`
**Admin User:** `admin@yukti.com`

Use these in API calls:
```bash
curl -H "X-Admin-Key: yukti-admin-2024" -H "X-Admin-User: admin@yukti.com" http://localhost:8080/api/admin/customers
```

## 🌐 URLs

- **Frontend:** http://localhost:3000
- **Backend API:** http://localhost:8080
- **ML Service:** http://localhost:8000
- **Prometheus:** http://localhost:9090
- **Grafana:** http://localhost:3001 (admin/admin)

## 👥 Test Customers

| Company | Tenant ID | Email | Status |
|---------|-----------|-------|--------|
| Acme Corp | tenant-001 | admin@acme.com | completed |
| TechStart Inc | tenant-002 | cto@techstart.io | in_progress |
| CloudScale LLC | tenant-003 | ops@cloudscale.com | completed |

## 📊 Key Features

### Admin Dashboard
- View all customers
- See platform metrics (MRR, savings, trials)
- Impersonate customers (logged in audit)
- Search customers

### Customer Dashboard
- Total savings
- Findings count
- Budget usage
- RI/SP savings potential

### Hidden Costs
- 77 detectors across 10 categories
- Filter by category and severity
- Click for detailed remediation
- Generate IaC for fixes

### Audit Logs (Security Team)
- All admin actions tracked
- Impersonation monitoring
- IP address logging
- Real-time activity feed

## 🔒 Security Features

1. **Admin Authentication** - All admin endpoints require X-Admin-Key
2. **Tenant Isolation** - Validates tenant_id exists before data access
3. **Audit Logging** - Every admin action logged with timestamp, IP, user
4. **Input Validation** - All inputs validated before processing
5. **Error Boundaries** - Frontend crashes handled gracefully

## 🐛 Troubleshooting

### Containers won't start
```bash
# Check Docker is running
docker ps

# Check logs
docker-compose logs backend
docker-compose logs frontend

# Restart everything
docker-compose down
docker-compose up -d --build
```

### Database issues
```bash
# Connect to database
docker exec -it yukti-postgres psql -U yukti -d yukti

# Check tables
\dt

# Check customers
SELECT * FROM yt_customers;

# Check audit logs
SELECT * FROM yt_audit_logs ORDER BY created_at DESC LIMIT 10;
```

### Frontend not loading
```bash
# Check frontend logs
docker-compose logs frontend

# Rebuild frontend
docker-compose up -d --build frontend

# Check browser console for errors
```

### API returning errors
```bash
# Test health endpoint
curl http://localhost:8080/health

# Test with admin key
curl -H "X-Admin-Key: yukti-admin-2024" http://localhost:8080/api/admin/customers

# Check backend logs
docker-compose logs backend | tail -50
```

## 📝 Common Tasks

### Add a new customer
```bash
curl -X POST http://localhost:8080/api/customers \
  -H "Content-Type: application/json" \
  -d '{"company_name": "New Co", "email": "admin@newco.com"}'
```

### Impersonate a customer
1. Go to http://localhost:3000/admin
2. Click "View" on any customer
3. You'll be redirected to their dashboard
4. Action is logged in audit logs

### View audit logs
1. Go to http://localhost:3000/audit-logs
2. See all admin actions
3. Filter by date, user, or action type

### Check savings for a tenant
```bash
curl "http://localhost:8080/api/customers/dashboard?tenant_id=tenant-001"
```

## 🎯 Testing Scenarios

### Test Multi-Tenant Isolation
```bash
# Valid tenant (should work)
curl "http://localhost:8080/api/customers/dashboard?tenant_id=tenant-001"

# Invalid tenant (should return 403)
curl "http://localhost:8080/api/customers/dashboard?tenant_id=invalid"
```

### Test Admin Authentication
```bash
# Without key (should return 401)
curl http://localhost:8080/api/admin/customers

# With key (should work)
curl -H "X-Admin-Key: yukti-admin-2024" http://localhost:8080/api/admin/customers
```

### Test Audit Logging
```bash
# Impersonate a customer
curl -X POST http://localhost:8080/api/admin/impersonate \
  -H "X-Admin-Key: yukti-admin-2024" \
  -H "X-Admin-User: admin@yukti.com" \
  -H "Content-Type: application/json" \
  -d '{"tenant_id": "tenant-001"}'

# Check it was logged
curl -H "X-Admin-Key: yukti-admin-2024" "http://localhost:8080/api/admin/audit-logs?limit=5"
```

## 📚 Documentation

- **API Examples:** See `API_EXAMPLES.md`
- **Deployment Guide:** See `DEPLOYMENT_GUIDE.md`
- **Testing Guide:** See `TESTING_GUIDE.md`
- **Production Checklist:** See `PRODUCTION_CHECKLIST.md`
- **Fixes Applied:** See `AUTONOMOUS_FIXES_APPLIED.md`

## 🆘 Getting Help

1. Check logs: `docker-compose logs [service]`
2. Check database: `docker exec -it yukti-postgres psql -U yukti -d yukti`
3. Run test script: `./test_everything.sh`
4. Check browser console for frontend errors
5. Verify all containers running: `docker-compose ps`
