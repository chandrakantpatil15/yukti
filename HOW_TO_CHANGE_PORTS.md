# How to Change Ports - Yukti Platform

## Quick Guide

**Change ports in ONE place, apply everywhere!**

### Step 1: Edit `.env.ports`

```bash
# Open the file
nano .env.ports

# Change the ports you need
BACKEND_PORT=8082      # Change from 8081 to 8082
FRONTEND_PORT=3001     # Change from 3000 to 3001
POSTGRES_PORT=5433     # Change from 5432 to 5433
ML_SERVICE_PORT=8001   # Change from 8000 to 8001
PROMETHEUS_PORT=9091   # Change from 9090 to 9091
GRAFANA_PORT=3002      # Change from 3001 to 3002
```

### Step 2: Run Update Script

```bash
./scripts/update-ports.sh
```

This will automatically update:
- ✅ `docker-compose.yml` - All service ports
- ✅ `frontend/src/services/api.ts` - API URL
- ✅ `internal/config/config.go` - Backend default port
- ✅ `cmd/main.go` - Backend default port
- ✅ `README.md` - Documentation

### Step 3: Rebuild Containers

```bash
docker-compose down
docker-compose up -d --build
```

### Step 4: Verify

```bash
# Check services are running on new ports
docker-compose ps

# Test backend
curl http://localhost:8082/health

# Test frontend
open http://localhost:3001
```

---

## Manual Method (If Script Fails)

### 1. Update `.env.ports`
Edit the file with your new ports.

### 2. Update `docker-compose.yml`
Ports are automatically read from `.env.ports` using `${BACKEND_PORT:-8081}` syntax.

### 3. Update Frontend API URL
```bash
# frontend/src/services/api.ts
const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8082';
```

### 4. Update Backend Config
```bash
# internal/config/config.go
Port: getEnv("PORT", "8082"),

# cmd/main.go
if port == "" {
    port = "8082"
}
```

### 5. Rebuild
```bash
docker-compose up -d --build
```

---

## Port Conflict Resolution

### Check What's Using a Port

```bash
# Check port 8081
lsof -i :8081

# Check port 3000
lsof -i :3000
```

### Kill Process Using Port

```bash
# Find process ID
lsof -i :8081

# Kill it
kill -9 <PID>
```

### Use Different Ports

If ports 8081/3000 are taken, use alternatives:
- Backend: 8082, 8083, 8084
- Frontend: 3001, 3002, 3003

---

## Common Port Configurations

### Default (Current)
```
Backend:    8081
Frontend:   3000
PostgreSQL: 5432
ML Service: 8000
Prometheus: 9090
Grafana:    3001
```

### Alternative 1 (All +1)
```
Backend:    8082
Frontend:   3001
PostgreSQL: 5433
ML Service: 8001
Prometheus: 9091
Grafana:    3002
```

### Alternative 2 (High Ports)
```
Backend:    9081
Frontend:   9000
PostgreSQL: 9432
ML Service: 9800
Prometheus: 9090
Grafana:    9001
```

---

## Troubleshooting

### Ports not updating?

```bash
# 1. Stop all containers
docker-compose down

# 2. Check .env.ports is correct
cat .env.ports

# 3. Rebuild with no cache
docker-compose build --no-cache

# 4. Start fresh
docker-compose up -d
```

### Frontend can't connect to backend?

```bash
# Check REACT_APP_API_URL in docker-compose.yml
docker-compose config | grep REACT_APP_API_URL

# Should show: http://localhost:<BACKEND_PORT>
```

### Backend not listening on correct port?

```bash
# Check backend logs
docker-compose logs backend | grep "Server starting"

# Should show: Server starting on port <BACKEND_PORT>
```

---

## Files That Use Ports

### Automatically Updated by Script
- ✅ `docker-compose.yml`
- ✅ `frontend/src/services/api.ts`
- ✅ `internal/config/config.go`
- ✅ `cmd/main.go`
- ✅ `README.md`

### Automatically Read from .env.ports
- ✅ All Docker containers
- ✅ Environment variables

### May Need Manual Update
- ⚠️ `PORT_CONFIGURATION.md` - Documentation only
- ⚠️ `DOCKER_QUICK_REFERENCE.md` - Documentation only
- ⚠️ `.amazonq/rules/session-progress.md` - Documentation only

---

## Best Practices

1. **Always use `.env.ports`** - Single source of truth
2. **Run update script** - Ensures consistency
3. **Rebuild containers** - Apply changes
4. **Test all services** - Verify everything works
5. **Update documentation** - Keep docs in sync

---

## Example: Changing Backend Port

```bash
# 1. Edit .env.ports
echo "BACKEND_PORT=8082" > .env.ports

# 2. Run update script
./scripts/update-ports.sh

# 3. Rebuild
docker-compose down
docker-compose up -d --build

# 4. Test
curl http://localhost:8082/health
```

---

**Last Updated**: Session 13
**Maintainer**: DevOps Team
