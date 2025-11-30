# Docker Quick Reference - Yukti Platform

## ⚠️ CRITICAL: Always Use Docker

**Never run services locally!** All development happens in Docker containers.

---

## Quick Commands

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

### Rebuild After Code Changes
```bash
# Rebuild specific service
docker-compose up -d --build backend
docker-compose up -d --build frontend

# Rebuild all
docker-compose up -d --build
```

### View Logs
```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f backend
docker-compose logs -f frontend
```

### Check Status
```bash
docker-compose ps
```

### Restart Service
```bash
docker-compose restart backend
docker-compose restart frontend
```

---

## Service URLs

| Service | URL | Container |
|---------|-----|-----------|
| Frontend | http://localhost:3000 | yukti-frontend |
| Backend API | http://localhost:8081 | yukti-backend |
| ML Service | http://localhost:8000 | yukti-ml |
| Prometheus | http://localhost:9090 | yukti-prometheus |
| Grafana | http://localhost:3001 | yukti-grafana |
| PostgreSQL | localhost:5432 | yukti-postgres |

---

## Database Access

```bash
# Connect to PostgreSQL
docker exec -it yukti-postgres psql -U yukti -d yukti_finops

# Run SQL file
docker exec -i yukti-postgres psql -U yukti -d yukti_finops < scripts/seed.sql

# Backup database
docker exec yukti-postgres pg_dump -U yukti yukti_finops > backup.sql
```

---

## Development Workflow

### 1. Make Code Changes
Edit files in:
- `internal/` - Backend Go code
- `frontend/src/` - Frontend React code
- `ml-service/` - ML Python code

### 2. Rebuild Container
```bash
docker-compose up -d --build backend
```

### 3. View Logs
```bash
docker-compose logs -f backend
```

### 4. Test Changes
Open http://localhost:3000

---

## Troubleshooting

### Container won't start
```bash
# Check logs
docker-compose logs backend

# Check status
docker-compose ps

# Restart
docker-compose restart backend
```

### Port already in use
```bash
# Stop all containers
docker-compose down

# Check what's using port
lsof -i :8081

# Kill process
kill -9 <PID>

# Restart
docker-compose up -d
```

### Changes not reflecting
```bash
# Force rebuild
docker-compose up -d --build --force-recreate backend

# Clear cache
docker-compose build --no-cache backend
```

### Database issues
```bash
# Check PostgreSQL logs
docker-compose logs postgres

# Restart database
docker-compose restart postgres

# Reset database (WARNING: deletes data)
docker-compose down -v
docker-compose up -d
```

---

## Clean Up

### Remove containers
```bash
docker-compose down
```

### Remove containers + volumes (deletes data)
```bash
docker-compose down -v
```

### Remove images
```bash
docker-compose down --rmi all
```

### Full cleanup
```bash
docker-compose down -v --rmi all
docker system prune -a
```

---

## Deployment Path

```
Local Changes
    ↓
Docker Compose (Test)
    ↓
Build Production Images
    ↓
Push to ECR
    ↓
Deploy to EKS
    ↓
Production
```

---

## ❌ DO NOT DO

- ❌ Run `./bin/api` (local backend)
- ❌ Run `npm start` in frontend/ (local React)
- ❌ Use local PostgreSQL
- ❌ Mix Docker and local services

## ✅ ALWAYS DO

- ✅ Use `docker-compose up -d --build`
- ✅ Use `make start` / `make stop`
- ✅ Test in Docker containers
- ✅ View logs with `docker-compose logs -f`

---

**Last Updated**: Session 13
**Reference**: See PLATFORM_ARCHITECTURE.md for details
