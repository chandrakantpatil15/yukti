# Yukti FinOps - Production Deployment Guide

## Prerequisites

- Docker 20.10+
- Kubernetes 1.24+
- PostgreSQL 14+
- Redis 7+
- Python 3.11+
- Go 1.21+
- Node.js 18+

## Architecture Overview

```
┌─────────────────────────────────────────────────────────┐
│                    Load Balancer (ALB)                  │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│                    Frontend (React)                     │
│                  CloudFront + S3                        │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│                  API Gateway (Go)                       │
│              3 replicas, auto-scaling                   │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│                ML Service (Python)                      │
│              2 replicas, auto-scaling                   │
└─────────────────────────────────────────────────────────┘
                          ↓
┌──────────────────┐  ┌──────────────────┐
│   PostgreSQL     │  │      Redis       │
│   RDS Multi-AZ   │  │   ElastiCache    │
└──────────────────┘  └──────────────────┘
```

## Environment Setup

### 1. Database Setup (PostgreSQL)

```bash
# Create RDS PostgreSQL instance
aws rds create-db-instance \
  --db-instance-identifier yukti-prod \
  --db-instance-class db.r5.large \
  --engine postgres \
  --engine-version 14.9 \
  --master-username yukti_admin \
  --master-user-password <SECURE_PASSWORD> \
  --allocated-storage 100 \
  --storage-type gp3 \
  --multi-az \
  --backup-retention-period 30 \
  --preferred-backup-window "03:00-04:00" \
  --preferred-maintenance-window "sun:04:00-sun:05:00"

# Run migrations
export DATABASE_URL="postgres://yukti_admin:<PASSWORD>@yukti-prod.xxx.rds.amazonaws.com:5432/yukti"
psql $DATABASE_URL -f scripts/001_create_yt_aws_pricing.sql
psql $DATABASE_URL -f scripts/002_create_yt_aws_resources.sql
# ... run all migration scripts
```

### 2. Redis Setup (ElastiCache)

```bash
# Create ElastiCache Redis cluster
aws elasticache create-replication-group \
  --replication-group-id yukti-redis \
  --replication-group-description "Yukti ML cache" \
  --engine redis \
  --cache-node-type cache.r5.large \
  --num-cache-clusters 2 \
  --automatic-failover-enabled \
  --multi-az-enabled
```

### 3. Secrets Management

```bash
# Store secrets in AWS Secrets Manager
aws secretsmanager create-secret \
  --name yukti/prod/database \
  --secret-string '{"username":"yukti_admin","password":"<PASSWORD>"}'

aws secretsmanager create-secret \
  --name yukti/prod/jwt-secret \
  --secret-string '{"secret":"<RANDOM_256_BIT_KEY>"}'

aws secretsmanager create-secret \
  --name yukti/prod/encryption-key \
  --secret-string '{"key":"<RANDOM_256_BIT_KEY>"}'
```

## Docker Build

### Backend (Go API)

```dockerfile
# Dockerfile.api
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o yukti-api cmd/api-server.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/yukti-api .
EXPOSE 8080
CMD ["./yukti-api"]
```

```bash
# Build and push
docker build -f Dockerfile.api -t yukti/api:1.0.0 .
docker tag yukti/api:1.0.0 <ECR_REPO>/yukti/api:1.0.0
docker push <ECR_REPO>/yukti/api:1.0.0
```

### ML Service (Python)

```dockerfile
# Dockerfile.ml
FROM python:3.11-slim
WORKDIR /app
COPY ml-service/requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
COPY ml-service/ .
EXPOSE 8000
CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8000"]
```

```bash
# Build and push
docker build -f Dockerfile.ml -t yukti/ml-service:1.0.0 .
docker tag yukti/ml-service:1.0.0 <ECR_REPO>/yukti/ml-service:1.0.0
docker push <ECR_REPO>/yukti/ml-service:1.0.0
```

### Frontend (React)

```bash
# Build production bundle
cd frontend
npm install
npm run build

# Upload to S3
aws s3 sync build/ s3://yukti-frontend-prod/ --delete

# Invalidate CloudFront cache
aws cloudfront create-invalidation \
  --distribution-id <DISTRIBUTION_ID> \
  --paths "/*"
```

## Kubernetes Deployment

### 1. Create Namespace

```yaml
# k8s/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: yukti-prod
```

### 2. ConfigMap

```yaml
# k8s/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: yukti-config
  namespace: yukti-prod
data:
  DATABASE_HOST: "yukti-prod.xxx.rds.amazonaws.com"
  DATABASE_PORT: "5432"
  DATABASE_NAME: "yukti"
  REDIS_HOST: "yukti-redis.xxx.cache.amazonaws.com"
  REDIS_PORT: "6379"
  ML_SERVICE_URL: "http://ml-service:8000"
  API_PORT: "8080"
```

### 3. Secrets

```yaml
# k8s/secrets.yaml
apiVersion: v1
kind: Secret
metadata:
  name: yukti-secrets
  namespace: yukti-prod
type: Opaque
data:
  database-password: <BASE64_ENCODED>
  jwt-secret: <BASE64_ENCODED>
  encryption-key: <BASE64_ENCODED>
```

### 4. API Deployment

```yaml
# k8s/api-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: yukti-api
  namespace: yukti-prod
spec:
  replicas: 3
  selector:
    matchLabels:
      app: yukti-api
  template:
    metadata:
      labels:
        app: yukti-api
    spec:
      containers:
      - name: api
        image: <ECR_REPO>/yukti/api:1.0.0
        ports:
  - containerPort: 8080
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: yukti-secrets
              key: database-url
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: yukti-secrets
              key: jwt-secret
        resources:
          requests:
            memory: "512Mi"
            cpu: "500m"
          limits:
            memory: "1Gi"
            cpu: "1000m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: yukti-api
  namespace: yukti-prod
spec:
  selector:
    app: yukti-api
  ports:
  - port: 8080
    targetPort: 8080
  type: ClusterIP
---
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: yukti-api-hpa
  namespace: yukti-prod
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: yukti-api
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

### 5. ML Service Deployment

```yaml
# k8s/ml-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ml-service
  namespace: yukti-prod
spec:
  replicas: 2
  selector:
    matchLabels:
      app: ml-service
  template:
    metadata:
      labels:
        app: ml-service
    spec:
      containers:
      - name: ml
        image: <ECR_REPO>/yukti/ml-service:1.0.0
        ports:
        - containerPort: 8000
        envFrom:
        - configMapRef:
            name: yukti-config
        resources:
          requests:
            memory: "1Gi"
            cpu: "1000m"
          limits:
            memory: "2Gi"
            cpu: "2000m"
---
apiVersion: v1
kind: Service
metadata:
  name: ml-service
  namespace: yukti-prod
spec:
  selector:
    app: ml-service
  ports:
  - port: 8000
    targetPort: 8000
  type: ClusterIP
```

### 6. Ingress (ALB)

```yaml
# k8s/ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: yukti-ingress
  namespace: yukti-prod
  annotations:
    kubernetes.io/ingress.class: alb
    alb.ingress.kubernetes.io/scheme: internet-facing
    alb.ingress.kubernetes.io/target-type: ip
    alb.ingress.kubernetes.io/certificate-arn: <ACM_CERTIFICATE_ARN>
    alb.ingress.kubernetes.io/listen-ports: '[{"HTTP": 80}, {"HTTPS": 443}]'
    alb.ingress.kubernetes.io/ssl-redirect: '443'
spec:
  rules:
  - host: api.yukti.io
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: yukti-api
            port:
              number: 8080
```

## Deploy to Kubernetes

```bash
# Apply all configurations
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/secrets.yaml
kubectl apply -f k8s/api-deployment.yaml
kubectl apply -f k8s/ml-deployment.yaml
kubectl apply -f k8s/ingress.yaml

# Verify deployment
kubectl get pods -n yukti-prod
kubectl get svc -n yukti-prod
kubectl get ingress -n yukti-prod

# Check logs
kubectl logs -f deployment/yukti-api -n yukti-prod
kubectl logs -f deployment/ml-service -n yukti-prod
```

## Monitoring & Logging

### Prometheus

```yaml
# k8s/prometheus-servicemonitor.yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: yukti-api
  namespace: yukti-prod
spec:
  selector:
    matchLabels:
      app: yukti-api
  endpoints:
  - port: metrics
    interval: 30s
```

### CloudWatch Logs

```bash
# Install Fluent Bit
kubectl apply -f https://raw.githubusercontent.com/aws/amazon-cloudwatch-container-insights/latest/k8s-deployment-manifest-templates/deployment-mode/daemonset/container-insights-monitoring/fluent-bit/fluent-bit.yaml
```

## Backup & Disaster Recovery

### Database Backups

```bash
# Automated RDS snapshots (already configured)
# Manual snapshot
aws rds create-db-snapshot \
  --db-instance-identifier yukti-prod \
  --db-snapshot-identifier yukti-prod-manual-$(date +%Y%m%d)

# Point-in-time recovery enabled (30-day retention)
```

### Application Backups

```bash
# Backup Kubernetes configs
kubectl get all -n yukti-prod -o yaml > backup-$(date +%Y%m%d).yaml

# Backup to S3
aws s3 cp backup-$(date +%Y%m%d).yaml s3://yukti-backups/k8s/
```

## Rollback Procedure

```bash
# Rollback API deployment
kubectl rollout undo deployment/yukti-api -n yukti-prod

# Rollback to specific revision
kubectl rollout history deployment/yukti-api -n yukti-prod
kubectl rollout undo deployment/yukti-api --to-revision=2 -n yukti-prod

# Rollback ML service
kubectl rollout undo deployment/ml-service -n yukti-prod
```

## Health Checks

```bash
# API health
curl https://api.yukti.io/health

# ML service health (internal)
kubectl exec -it deployment/yukti-api -n yukti-prod -- curl http://ml-service:8000/health

# Database connectivity
kubectl exec -it deployment/yukti-api -n yukti-prod -- psql $DATABASE_URL -c "SELECT 1"
```

## Scaling

### Manual Scaling

```bash
# Scale API
kubectl scale deployment/yukti-api --replicas=5 -n yukti-prod

# Scale ML service
kubectl scale deployment/ml-service --replicas=3 -n yukti-prod
```

### Auto-scaling (Already configured via HPA)

```bash
# Check HPA status
kubectl get hpa -n yukti-prod

# Describe HPA
kubectl describe hpa yukti-api-hpa -n yukti-prod
```

## Security Hardening

### Network Policies

```yaml
# k8s/network-policy.yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: yukti-api-policy
  namespace: yukti-prod
spec:
  podSelector:
    matchLabels:
      app: yukti-api
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: ingress-controller
    ports:
    - protocol: TCP
      port: 8090
  egress:
  - to:
    - podSelector:
        matchLabels:
          app: ml-service
    ports:
    - protocol: TCP
      port: 8091
```

### Pod Security Policy

```yaml
# k8s/pod-security-policy.yaml
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: yukti-restricted
spec:
  privileged: false
  allowPrivilegeEscalation: false
  requiredDropCapabilities:
  - ALL
  runAsUser:
    rule: MustRunAsNonRoot
  seLinux:
    rule: RunAsAny
  fsGroup:
    rule: RunAsAny
```

## Troubleshooting

### Common Issues

**Pods not starting:**
```bash
kubectl describe pod <POD_NAME> -n yukti-prod
kubectl logs <POD_NAME> -n yukti-prod
```

**Database connection issues:**
```bash
# Check security groups
# Verify RDS endpoint
# Test connectivity from pod
kubectl exec -it <POD_NAME> -n yukti-prod -- nc -zv <RDS_ENDPOINT> 5432
```

**High memory usage:**
```bash
kubectl top pods -n yukti-prod
kubectl top nodes
```

## Production Checklist

- [ ] Database backups enabled (30-day retention)
- [ ] Multi-AZ deployment for RDS
- [ ] Redis cluster with failover
- [ ] SSL/TLS certificates configured
- [ ] Secrets stored in AWS Secrets Manager
- [ ] CloudWatch logging enabled
- [ ] Prometheus monitoring configured
- [ ] Auto-scaling enabled (HPA)
- [ ] Network policies applied
- [ ] Pod security policies enforced
- [ ] Health checks configured
- [ ] Rollback procedure tested
- [ ] Disaster recovery plan documented
- [ ] Load testing completed
- [ ] Security audit passed

## Support

- Runbook: https://docs.yukti.io/runbook
- On-call: +1-800-YUKTI-911
- Slack: #yukti-ops
