# 🚀 Yukti Platform - Port Allocation Reference

## Service Port Configuration

| Service | Internal Port | External Port | URL | Status |
|---------|---------------|---------------|-----|--------|
| **Frontend** | 3000 | 3000 | http://localhost:3000 | ✅ Active |
| **Backend API** | 8080 | **8081** | http://localhost:8081 | 🔧 Fixing |
| **ML Service** | 8000 | 8000 | http://localhost:8000 | ✅ Active |
| **PostgreSQL** | 5432 | 5432 | localhost:5432 | ✅ Active |
| **Prometheus** | 9090 | 9090 | http://localhost:9090 | ✅ Active |
| **Grafana** | 3000 | 3001 | http://localhost:3001 | ✅ Active |

## Key API Endpoints

### Backend API (Port 8081)
- **Health**: `GET http://localhost:8081/health`
- **Auth Signup**: `POST http://localhost:8081/api/auth/signup`
- **Auth Login**: `POST http://localhost:8081/api/auth/login`
- **Auth Verify**: `POST http://localhost:8081/api/auth/verify`
- **Admin API**: `http://localhost:8081/api/admin/*`
- **Customer API**: `http://localhost:8081/api/customers/*`

### ML Service (Port 8000)
- **Health**: `GET http://localhost:8000/health`
- **Anomaly Detection**: `POST http://localhost:8000/api/ml/anomaly-detect`
- **Forecasting**: `POST http://localhost:8000/api/ml/forecast`

## Port Conflicts Avoided
- **Port 8080**: Reserved for Jenkins (system conflict)
- **Port 3001**: Used for Grafana (not 3000 to avoid frontend conflict)

## Docker Compose Mapping
```yaml
backend:
  ports:
    - "8081:8080"  # External:Internal

ml-service:
  ports:
    - "8000:8000"  # External:Internal

frontend:
  ports:
    - "3000:3000"  # External:Internal
```

## Environment Variables
- **Backend**: `PORT=8080` (internal), `REACT_APP_API_URL=http://localhost:8081`
- **ML Service**: `PORT=8000`
- **Database**: `DATABASE_URL=postgresql://yukti@host.docker.internal:5432/yukti_finops`

---
**Last Updated**: November 2024  
**Note**: Backend runs internally on 8080 but is accessible externally on 8081