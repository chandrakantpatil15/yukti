# PLATFORM ARCHITECTURE - CRITICAL REFERENCE

## ⚠️ IMPORTANT: ALWAYS USE DOCKER

**The platform runs ENTIRELY in Docker containers, NOT locally!**

### Architecture Flow
```
Development → Docker Compose → EKS (Kubernetes) → Production
```

### Why Docker?
1. **Consistency**: Same environment in dev, staging, production
2. **EKS Deployment**: Docker images deploy directly to Kubernetes pods
3. **Isolation**: Each service runs in its own container
4. **Scalability**: Easy to scale with Kubernetes HPA

---

## Services (All in Docker)

| Service | Container | Port | Image |
|---------|-----------|------|-------|
| Backend | yukti-backend | 8081 | Go API |
| Frontend | yukti-frontend | 3000 | React |
| PostgreSQL | yukti-postgres | 5432 | postgres:15 |
| ML Service | yukti-ml | 8000 | Python FastAPI |
| Prometheus | yukti-prometheus | 9090 | Monitoring |
| Grafana | yukti-grafana | 3001 | Dashboards |

---

## Development Workflow

### 1. Make Code Changes
Edit files in:
- `internal/` - Backend Go code
- `frontend/src/` - Frontend React code
- `ml-service/` - ML Python code

### 2. Rebuild Docker Images
```bash
# Rebuild ALL services
docker-compose up -d --build

# Rebuild specific service
docker-compose up -d --build backend
docker-compose up -d --build frontend
```

### 3. Test Changes
```bash
# View logs
docker-compose logs -f backend
docker-compose logs -f frontend

# Check status
docker-compose ps

# Access services
# Frontend: http://localhost:3000
# Backend: http://localhost:8081
```

### 4. Deploy to EKS
```bash
# Build production images
docker build -t yukti-backend:latest .
docker build -t yukti-frontend:latest ./frontend

# Push to ECR
# Deploy to Kubernetes
kubectl apply -f k8s/
```

---

## Common Commands

### Start Platform
```bash
make start
# OR
docker-compose up -d
```

### Stop Platform
```bash
make stop
# OR
docker-compose down
```

### Rebuild After Changes
```bash
docker-compose up -d --build
```

### View Logs
```bash
docker-compose logs -f
```

### Restart Service
```bash
docker-compose restart backend
docker-compose restart frontend
```

---

## ❌ DO NOT RUN LOCALLY

**Never run these commands:**
- ❌ `./bin/api` (local backend)
- ❌ `npm start` in frontend/ (local React)
- ❌ Local PostgreSQL connection

**Always use Docker:**
- ✅ `docker-compose up -d --build`
- ✅ `make start`
- ✅ Docker containers for everything

---

## Database Access

**PostgreSQL runs in Docker container:**

```bash
# Connect to database
docker exec -it yukti-postgres psql -U yukti -d yukti_finops

# Run SQL file
docker exec -i yukti-postgres psql -U yukti -d yukti_finops < scripts/seed.sql
```

---

## Troubleshooting

### Changes not reflecting?
```bash
# Rebuild the service
docker-compose up -d --build backend
docker-compose up -d --build frontend
```

### Port conflicts?
```bash
# Stop all containers
docker-compose down

# Check ports
lsof -i :8081
lsof -i :3000

# Restart
docker-compose up -d
```

### Database issues?
```bash
# Check PostgreSQL logs
docker-compose logs postgres

# Restart database
docker-compose restart postgres
```

---

## Deployment Path

```
Local Changes
    ↓
Docker Build
    ↓
Docker Compose Test
    ↓
Push to ECR
    ↓
Deploy to EKS
    ↓
Production
```

---

**REMEMBER: Everything in Docker → Deploy to EKS → Production**

**Last Updated**: Session 13 - Platform architecture clarified
