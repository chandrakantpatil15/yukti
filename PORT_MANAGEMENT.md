# Port Management - Yukti Platform

## ✅ Single Source of Truth: `.env.ports`

**All ports are defined in ONE place only!**

```
.env.ports (ONLY place with hardcoded ports)
    ↓
docker-compose.yml (reads from .env.ports)
    ↓
All services (read from environment variables)
```

---

## Port Configuration

### `.env.ports` - The ONLY file with hardcoded ports

```bash
# Yukti Platform - Centralized Port Configuration
# Change ports here and they will apply everywhere

# Backend API
BACKEND_PORT=8081

# Frontend
FRONTEND_PORT=3000

# Database
POSTGRES_PORT=5432

# ML Service
ML_SERVICE_PORT=8000

# Monitoring
PROMETHEUS_PORT=9090
GRAFANA_PORT=3001
```

---

## How It Works

### 1. `.env.ports` defines all ports (hardcoded)
```bash
BACKEND_PORT=8081
FRONTEND_PORT=3000
```

### 2. `docker-compose.yml` reads from `.env.ports`
```yaml
backend:
  env_file:
    - .env.ports
  environment:
    PORT: ${BACKEND_PORT:-8081}  # Reads from .env.ports
  ports:
    - "${BACKEND_PORT:-8081}:${BACKEND_PORT:-8081}"
```

### 3. Application code reads from environment
```go
// cmd/main.go - NO hardcoded fallback!
port := os.Getenv("PORT")
if port == "" {
    log.Fatal("PORT not set. Check .env.ports")
}
```

```typescript
// frontend/src/services/api.ts - NO hardcoded fallback!
const API_BASE_URL = process.env.REACT_APP_API_URL!;
```

---

## Files and Their Role

| File | Role | Hardcoded Ports? |
|------|------|------------------|
| `.env.ports` | **Source of truth** | ✅ YES (only here) |
| `docker-compose.yml` | Reads from .env.ports | ❌ NO |
| `cmd/main.go` | Reads from env var | ❌ NO |
| `internal/config/config.go` | Reads from env var | ❌ NO |
| `frontend/src/services/api.ts` | Reads from env var | ❌ NO |

---

## Changing Ports

### Method 1: Edit `.env.ports` (Recommended)

```bash
# 1. Edit the file
nano .env.ports

# Change:
BACKEND_PORT=8082
FRONTEND_PORT=3001

# 2. Rebuild
docker-compose down
docker-compose up -d --build
```

### Method 2: Override with Environment Variables

```bash
# Temporary override (doesn't change .env.ports)
BACKEND_PORT=8082 FRONTEND_PORT=3001 docker-compose up -d
```

---

## Validation

### Services will FAIL if ports not set

**Backend:**
```
[FATAL] PORT environment variable not set. Check .env.ports file.
```

**Frontend:**
```
TypeError: Cannot read property 'REACT_APP_API_URL' of undefined
```

This is **intentional** - forces you to use `.env.ports`!

---

## Benefits

### ✅ Advantages
1. **Single source of truth** - Change once, apply everywhere
2. **No hidden defaults** - All ports explicit in .env.ports
3. **Fail fast** - Services won't start with wrong config
4. **Easy to change** - Edit one file, rebuild
5. **Version controlled** - .env.ports in git

### ❌ No More
- ❌ Hardcoded ports scattered across files
- ❌ Forgotten fallback values
- ❌ Port conflicts from defaults
- ❌ Manual updates in multiple files

---

## Troubleshooting

### Service won't start?

```bash
# Check .env.ports exists
cat .env.ports

# Check docker-compose reads it
docker-compose config | grep PORT

# Check environment in container
docker exec yukti-backend env | grep PORT
```

### Port conflict?

```bash
# Check what's using port
lsof -i :8081

# Change port in .env.ports
nano .env.ports

# Rebuild
docker-compose down
docker-compose up -d --build
```

---

## Example: Complete Port Change

```bash
# 1. Edit .env.ports
cat > .env.ports << EOF
BACKEND_PORT=9081
FRONTEND_PORT=9000
POSTGRES_PORT=9432
ML_SERVICE_PORT=9800
PROMETHEUS_PORT=9090
GRAFANA_PORT=9001
EOF

# 2. Rebuild everything
docker-compose down
docker-compose up -d --build

# 3. Verify
curl http://localhost:9081/health
open http://localhost:9000
```

---

## Production Deployment

### Kubernetes ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: yukti-ports
data:
  BACKEND_PORT: "8081"
  FRONTEND_PORT: "3000"
  POSTGRES_PORT: "5432"
```

### EKS Deployment

```yaml
containers:
- name: backend
  env:
  - name: PORT
    valueFrom:
      configMapKeyRef:
        name: yukti-ports
        key: BACKEND_PORT
```

---

## Summary

**Rule: Only `.env.ports` has hardcoded port numbers!**

- ✅ `.env.ports` - Hardcoded (source of truth)
- ❌ All other files - Read from environment
- ❌ No fallback defaults - Fail if not set
- ✅ Change once - Apply everywhere

---

**Last Updated**: Session 13
**Maintainer**: DevOps Team
