# Port Configuration Flow - Yukti Platform

## Complete Flow Diagram

```
┌─────────────────────────────────────────────────────────────┐
│ 1. .env.ports (Source of Truth)                            │
│    BACKEND_PORT=8081                                        │
│    FRONTEND_PORT=3000                                       │
└─────────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────────┐
│ 2. docker-compose.yml (Reads .env.ports)                   │
│    backend:                                                 │
│      env_file:                                              │
│        - .env.ports  ← Loads all variables from file       │
│      environment:                                           │
│        PORT: ${BACKEND_PORT:-8081}  ← Uses loaded value    │
│      ports:                                                 │
│        - "${BACKEND_PORT:-8081}:${BACKEND_PORT:-8081}"     │
└─────────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────────┐
│ 3. Docker Container Starts                                  │
│    Environment variables set:                               │
│    PORT=8081                                                │
│    BACKEND_PORT=8081                                        │
│    FRONTEND_PORT=3000                                       │
└─────────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────────┐
│ 4. Application Code Reads Environment                       │
│    Go:     port := os.Getenv("PORT")                       │
│    React:  process.env.REACT_APP_API_URL                   │
└─────────────────────────────────────────────────────────────┘
```

---

## Detailed Breakdown

### Step 1: `.env.ports` File

```bash
# File: .env.ports
BACKEND_PORT=8081
FRONTEND_PORT=3000
POSTGRES_PORT=5432
ML_SERVICE_PORT=8000
PROMETHEUS_PORT=9090
GRAFANA_PORT=3001
```

**Purpose**: Single source of truth for all port numbers

---

### Step 2: `docker-compose.yml` Configuration

```yaml
backend:
  env_file:
    - .env.ports  # ← Docker Compose loads this file
  environment:
    PORT: ${BACKEND_PORT:-8081}  # ← Uses BACKEND_PORT from .env.ports
    CORS_ALLOWED_ORIGINS: http://localhost:${FRONTEND_PORT:-3000}
  ports:
    - "${BACKEND_PORT:-8081}:${BACKEND_PORT:-8081}"
```

**How it works:**
1. `env_file: - .env.ports` → Docker Compose reads the file
2. `${BACKEND_PORT:-8081}` → Uses value from .env.ports (8081)
3. `:-8081` → Fallback if .env.ports doesn't exist (safety net)

---

### Step 3: Container Environment Variables

When container starts, it has these environment variables:

```bash
# Inside yukti-backend container
$ env | grep PORT
PORT=8081
BACKEND_PORT=8081
FRONTEND_PORT=3000
POSTGRES_PORT=5432
```

**How to verify:**
```bash
docker exec yukti-backend env | grep PORT
```

---

### Step 4: Application Code

#### Backend (Go)

```go
// cmd/main.go
port := os.Getenv("PORT")  // Reads PORT=8081 from environment
if port == "" {
    log.Fatal("PORT not set")  // Fails if not found
}
server.Run(":" + port)  // Starts on :8081
```

#### Frontend (React)

```typescript
// frontend/src/services/api.ts
const API_BASE_URL = process.env.REACT_APP_API_URL!;
// Reads REACT_APP_API_URL=http://localhost:8081 from environment
```

---

## How Code Finds Ports

### Question: "How does code know where to find ports?"

**Answer**: Through environment variables set by Docker Compose!

```
.env.ports
    ↓ (loaded by docker-compose.yml)
Environment Variables in Container
    ↓ (read by application code)
os.Getenv("PORT") or process.env.REACT_APP_API_URL
```

---

## Example: Backend Port Lookup

### 1. You set in `.env.ports`:
```bash
BACKEND_PORT=8081
```

### 2. Docker Compose reads it:
```yaml
environment:
  PORT: ${BACKEND_PORT:-8081}  # Becomes PORT=8081
```

### 3. Container starts with:
```bash
PORT=8081
```

### 4. Go code reads it:
```go
port := os.Getenv("PORT")  // Gets "8081"
```

### 5. Server starts:
```
[INFO] Server starting on port 8081
```

---

## Verification Commands

### Check .env.ports
```bash
cat .env.ports
```

### Check docker-compose resolves variables
```bash
docker-compose config | grep PORT
```

### Check container environment
```bash
docker exec yukti-backend env | grep PORT
```

### Check application logs
```bash
docker-compose logs backend | grep "Server starting"
```

---

## What Happens If .env.ports Missing?

### Scenario 1: .env.ports exists
```bash
✅ BACKEND_PORT=8081 (from .env.ports)
✅ PORT=8081 (set in container)
✅ Server starts on 8081
```

### Scenario 2: .env.ports missing
```bash
⚠️  BACKEND_PORT=8081 (fallback from :-8081)
✅ PORT=8081 (set in container)
✅ Server starts on 8081 (using fallback)
```

### Scenario 3: .env.ports missing AND no fallback
```bash
❌ PORT= (empty)
❌ Server fails: "PORT not set"
```

---

## Current Setup (Best Practice)

We use **fallbacks** in docker-compose.yml for safety:

```yaml
PORT: ${BACKEND_PORT:-8081}
#                    ↑
#                    Fallback if .env.ports missing
```

**Why?**
- ✅ Works even if .env.ports accidentally deleted
- ✅ Clear default values visible in docker-compose.yml
- ✅ Easy to change: just edit .env.ports
- ✅ Fail-safe: won't break if file missing

---

## Summary

### How Code Finds Ports:

1. **`.env.ports`** - You define: `BACKEND_PORT=8081`
2. **`docker-compose.yml`** - Loads file: `env_file: - .env.ports`
3. **Docker sets env var** - Container gets: `PORT=8081`
4. **Code reads env var** - `os.Getenv("PORT")` returns `"8081"`
5. **Server starts** - Listens on port 8081

### To Change Ports:

1. Edit `.env.ports`
2. Run `docker-compose up -d --build`
3. Done! ✅

---

**Last Updated**: Session 13
**Maintainer**: DevOps Team
