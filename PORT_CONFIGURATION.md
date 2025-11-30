# Port Configuration

## Standard Ports (All in Docker)

| Service | Port | URL | Container |
|---------|------|-----|----------|
| **Backend API** | 8081 | http://localhost:8081 | yukti-backend |
| **Frontend** | 3000 | http://localhost:3000 | yukti-frontend |
| **PostgreSQL** | 5432 | localhost:5432 | yukti-postgres |
| **ML Service** | 8000 | http://localhost:8000 | yukti-ml |
| **Prometheus** | 9090 | http://localhost:9090 | yukti-prometheus |
| **Grafana** | 3001 | http://localhost:3001 | yukti-grafana |

## Important Notes

1. **All services run in Docker containers**
2. **Backend runs on port 8081** (NOT 8080)
3. **Frontend runs on port 3000**
4. **PostgreSQL runs in Docker** (NOT local)

## Configuration Files

### Docker Compose
- `docker-compose.yml` - All service configurations
- Backend port: `8081:8081`
- Frontend port: `3000:3000`
- PostgreSQL port: `5432:5432`

### Backend Port
- `cmd/main.go` - Default port: `8081`
- `internal/config/config.go` - Default port: `8081`
- Environment variable: `PORT=8081`

### Frontend API URL
- `frontend/src/services/api.ts` - Default: `http://localhost:8081`
- Environment variable: `REACT_APP_API_URL=http://localhost:8081`

### Database (Docker Container)
- Connection: `postgres://yukti:yukti123@postgres:5432/yukti_finops`
- Database name: `yukti_finops`
- User: `yukti`
- Container: `yukti-postgres`

## How to Start Services (Docker)

### 1. Start All Services
```bash
# Using Makefile
make start

# OR using docker-compose directly
docker-compose up -d
```

Expected output:
```
Creating yukti-postgres ... done
Creating yukti-backend  ... done
Creating yukti-frontend ... done
Creating yukti-ml       ... done
Creating yukti-prometheus ... done
Creating yukti-grafana  ... done
```

### 2. Verify Services
```bash
docker-compose ps
```

Expected output:
```
NAME              STATUS    PORTS
yukti-backend     Up        0.0.0.0:8081->8081/tcp
yukti-frontend    Up        0.0.0.0:3000->3000/tcp
yukti-postgres    Up        0.0.0.0:5432->5432/tcp
```

### 3. View Logs
```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f backend
docker-compose logs -f frontend
```

## Troubleshooting

### Backend not responding
```bash
# Check container status
docker-compose ps backend

# View logs
docker-compose logs backend

# Restart container
docker-compose restart backend
```

### Frontend can't connect to backend
```bash
# Check if backend container is running
docker-compose ps

# Check backend logs for errors
docker-compose logs backend

# Verify network connectivity
docker exec yukti-frontend curl http://backend:8081/health
```

### Port already in use
```bash
# Stop all containers
docker-compose down

# Check what's using the port
lsof -i :8081
lsof -i :3000

# Kill process if needed
kill -9 <PID>

# Restart containers
docker-compose up -d
```

### Rebuild After Code Changes
```bash
# Rebuild specific service
docker-compose up -d --build backend
docker-compose up -d --build frontend

# Rebuild all services
docker-compose up -d --build
```

## Environment Variables

Environment variables are configured in `docker-compose.yml`:

```yaml
# Backend environment
backend:
  environment:
    - PORT=8081
    - DATABASE_URL=postgres://yukti:yukti123@postgres:5432/yukti_finops
    - JWT_SECRET=your-secret-key-here
    - CORS_ALLOWED_ORIGINS=http://localhost:3000
    - ENVIRONMENT=development

# Frontend environment
frontend:
  environment:
    - REACT_APP_API_URL=http://localhost:8081
```

## Deployment Path

```
Local Development (Docker)
    ↓
Docker Compose Testing
    ↓
Build Production Images
    ↓
Push to ECR
    ↓
Deploy to EKS (Kubernetes)
    ↓
Production
```

---

**Last Updated**: Session 13 - Docker-first development workflow
