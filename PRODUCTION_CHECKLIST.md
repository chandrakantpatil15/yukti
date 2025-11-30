# Production Deployment Checklist

## 🔒 Security

- [ ] Change admin key from `yukti-admin-2024` to secure random key
- [ ] Implement proper JWT/OAuth authentication
- [ ] Enable HTTPS/TLS for all endpoints
- [ ] Configure CORS to allow only specific domains
- [ ] Set up API rate limiting per user (not just per IP)
- [ ] Enable SQL injection protection (parameterized queries ✅)
- [ ] Add XSS protection headers
- [ ] Implement CSRF tokens
- [ ] Set up WAF (Web Application Firewall)
- [ ] Enable database encryption at rest
- [ ] Rotate database credentials
- [ ] Set up secrets management (AWS Secrets Manager / HashiCorp Vault)

## 🗄️ Database

- [ ] Set up database backups (automated daily)
- [ ] Configure point-in-time recovery
- [ ] Set up read replicas for scaling
- [ ] Enable connection pooling
- [ ] Add database monitoring and alerts
- [ ] Optimize indexes for common queries
- [ ] Set up database migration strategy
- [ ] Configure retention policies for audit logs

## 🚀 Infrastructure

- [ ] Set up Kubernetes cluster (EKS/GKE/AKS)
- [ ] Configure auto-scaling (HPA)
- [ ] Set up load balancer
- [ ] Configure health checks
- [ ] Set up CDN for frontend assets
- [ ] Configure DNS and domain
- [ ] Set up SSL certificates (Let's Encrypt / ACM)
- [ ] Configure ingress controller

## 📊 Monitoring

- [ ] Set up Prometheus metrics collection
- [ ] Configure Grafana dashboards
- [ ] Set up alerting (PagerDuty / Opsgenie)
- [ ] Enable application logging (ELK / CloudWatch)
- [ ] Set up error tracking (Sentry / Rollbar)
- [ ] Configure uptime monitoring
- [ ] Set up performance monitoring (APM)
- [ ] Enable distributed tracing

## 🧪 Testing

- [ ] Run full integration tests
- [ ] Perform load testing
- [ ] Security penetration testing
- [ ] Test disaster recovery procedures
- [ ] Validate backup restoration
- [ ] Test auto-scaling behavior
- [ ] Verify multi-tenant isolation
- [ ] Test all API endpoints

## 🔄 CI/CD

- [ ] Set up GitHub Actions / GitLab CI
- [ ] Configure automated testing
- [ ] Set up staging environment
- [ ] Configure blue-green deployment
- [ ] Set up rollback procedures
- [ ] Enable automated security scanning
- [ ] Configure container scanning
- [ ] Set up dependency vulnerability scanning

## 📝 Documentation

- [ ] Update API documentation
- [ ] Create runbooks for common issues
- [ ] Document deployment procedures
- [ ] Create incident response plan
- [ ] Document backup/restore procedures
- [ ] Create user guides
- [ ] Document architecture decisions

## 💰 Cost Optimization

- [ ] Right-size compute resources
- [ ] Configure auto-scaling policies
- [ ] Set up cost alerts
- [ ] Enable Reserved Instances / Savings Plans
- [ ] Configure resource tagging
- [ ] Set up cost allocation reports

## 🎯 Performance

- [ ] Enable caching (Redis / Memcached)
- [ ] Optimize database queries
- [ ] Configure CDN caching
- [ ] Enable gzip compression
- [ ] Optimize frontend bundle size
- [ ] Implement lazy loading
- [ ] Set up database connection pooling

## 📧 Notifications

- [ ] Configure email service (SES / SendGrid)
- [ ] Set up Slack/Teams webhooks
- [ ] Configure SMS alerts (Twilio)
- [ ] Set up audit log notifications

## 🔐 Compliance

- [ ] GDPR compliance review
- [ ] SOC 2 compliance preparation
- [ ] Data retention policies
- [ ] Privacy policy updates
- [ ] Terms of service
- [ ] Cookie consent

## 🌍 Multi-Region (Optional)

- [ ] Set up multi-region deployment
- [ ] Configure global load balancer
- [ ] Set up database replication
- [ ] Configure CDN for global distribution
- [ ] Test failover procedures

## Environment Variables to Set

### Backend
```bash
DATABASE_URL=postgresql://user:pass@host:5432/yukti?sslmode=require
JWT_SECRET=<secure-random-key>
ADMIN_KEY=<secure-random-key>
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=<from-secrets-manager>
AWS_SECRET_ACCESS_KEY=<from-secrets-manager>
CORS_ALLOWED_ORIGINS=https://app.yukti.com
RATE_LIMIT_PER_MINUTE=100
LOG_LEVEL=info
SENTRY_DSN=<sentry-url>
```

### Frontend
```bash
REACT_APP_API_URL=https://api.yukti.com
REACT_APP_ENVIRONMENT=production
REACT_APP_SENTRY_DSN=<sentry-url>
```

### ML Service
```bash
MODEL_PATH=/app/models
LOG_LEVEL=info
```

## Pre-Launch Testing

### Load Testing
```bash
# Test with 1000 concurrent users
ab -n 10000 -c 1000 https://api.yukti.com/health

# Test admin endpoints
ab -n 1000 -c 100 -H "X-Admin-Key: <key>" https://api.yukti.com/api/admin/customers
```

### Security Testing
```bash
# SQL injection test
sqlmap -u "https://api.yukti.com/api/customers/dashboard?tenant_id=test"

# XSS test
# Test all input fields with XSS payloads

# CSRF test
# Verify CSRF tokens on all POST/PUT/DELETE requests
```

## Launch Day Checklist

- [ ] Verify all services are running
- [ ] Check all health endpoints
- [ ] Verify monitoring dashboards
- [ ] Test critical user flows
- [ ] Verify backup systems
- [ ] Check alert configurations
- [ ] Have rollback plan ready
- [ ] Team on standby for issues

## Post-Launch

- [ ] Monitor error rates
- [ ] Check performance metrics
- [ ] Review audit logs
- [ ] Verify backups running
- [ ] Check cost metrics
- [ ] Gather user feedback
- [ ] Plan next iteration
