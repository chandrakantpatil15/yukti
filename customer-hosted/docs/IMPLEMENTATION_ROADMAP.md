# Customer-Hosted Implementation Roadmap

**Goal**: Transform Yukti from SaaS to customer-hosted model  
**Timeline**: 8 weeks (after SaaS MVP complete)  
**Status**: Planning phase

---

## 🎯 Overview

This roadmap outlines the steps to build a **customer-hosted version** of Yukti FinOps platform.

**Key Principle**: "Yukti runs in your environment, keeps your data with you, and asks for nothing except a license check."

---

## 📅 8-Week Implementation Plan

### **Week 1-2: License API** (Foundation)

#### Week 1: Database & Core Logic
**Goal**: Build license validation service

**Tasks**:
- [ ] Create database schema (3 tables)
  - `license_keys` - Customer licenses
  - `activations` - Active instances
  - `feature_flags` - Plan features
- [ ] Implement models (`internal/models/license.go`)
- [ ] Implement database layer (`internal/database/db.go`)
- [ ] Implement JWT generation (`internal/auth/jwt.go`)
- [ ] Write unit tests

**Deliverable**: Working license validation logic

#### Week 2: API & Deployment
**Goal**: Deploy License API to production

**Tasks**:
- [ ] Implement API handlers (`internal/api/handlers.go`)
- [ ] Add rate limiting (10/min per key)
- [ ] Add monitoring (Prometheus metrics)
- [ ] Create Dockerfile
- [ ] Deploy to AWS (ECS or Lambda)
- [ ] Set up domain (license.yukti.io)
- [ ] Write API documentation

**Deliverable**: Live License API at https://license.yukti.io

---

### **Week 3-5: Platform Adaptation** (Core Changes)

#### Week 3: Remove Multi-Tenancy
**Goal**: Convert to single-tenant architecture

**Tasks**:
- [ ] Remove `tenant_id` from all database tables
- [ ] Remove tenant isolation middleware
- [ ] Remove cross-account IAM role assumption
- [ ] Update all queries (remove tenant filtering)
- [ ] Update all API handlers
- [ ] Test with single customer data

**Deliverable**: Single-tenant platform working locally

#### Week 4: Add License Validation
**Goal**: Integrate License API

**Tasks**:
- [ ] Add license validation on startup
- [ ] Cache JWT for 24 hours
- [ ] Implement JWT refresh logic
- [ ] Add offline grace period (7 days)
- [ ] Lock features on expiry (read-only mode)
- [ ] Add "License Expired" UI banner
- [ ] Test license expiry flow

**Deliverable**: Platform validates license and locks features

#### Week 5: Local IAM Role
**Goal**: Use local IAM role instead of cross-account

**Tasks**:
- [ ] Remove cross-account IAM code
- [ ] Use EC2 instance profile (AWS)
- [ ] Use Workload Identity (GCP)
- [ ] Use Managed Identity (Azure)
- [ ] Update AWS SDK initialization
- [ ] Test with local IAM role
- [ ] Update documentation

**Deliverable**: Platform uses local IAM role

---

### **Week 6-7: Helm Chart** (Deployment)

#### Week 6: Helm Chart Structure
**Goal**: Create Kubernetes deployment

**Tasks**:
- [ ] Create Helm chart structure
  - `Chart.yaml`
  - `values.yaml`
  - `templates/deployment.yaml`
  - `templates/service.yaml`
  - `templates/ingress.yaml`
  - `templates/configmap.yaml`
  - `templates/secret.yaml`
- [ ] Add PostgreSQL StatefulSet
- [ ] Add backend Deployment
- [ ] Add frontend Deployment
- [ ] Configure resource limits
- [ ] Add health checks

**Deliverable**: Working Helm chart

#### Week 7: Helm Chart Testing
**Goal**: Test on real Kubernetes cluster

**Tasks**:
- [ ] Test on local Kubernetes (minikube)
- [ ] Test on AWS EKS
- [ ] Test on GCP GKE
- [ ] Test on Azure AKS
- [ ] Test upgrades (helm upgrade)
- [ ] Test rollbacks (helm rollback)
- [ ] Publish to Helm registry

**Deliverable**: Production-ready Helm chart

---

### **Week 8: Documentation** (Polish)

#### Week 8: Complete Documentation
**Goal**: Write all customer-facing docs

**Tasks**:
- [ ] Deployment guide (Kubernetes)
- [ ] Deployment guide (Docker Compose)
- [ ] Architecture guide (trust by design)
- [ ] Troubleshooting guide
- [ ] FAQ
- [ ] Video tutorial (5 minutes)
- [ ] Migration guide (SaaS → Self-hosted)

**Deliverable**: Complete documentation

---

## 📊 Progress Tracking

### **Milestones**

| Milestone | Week | Status | Deliverable |
|-----------|------|--------|-------------|
| License API Live | 2 | ⏳ Pending | https://license.yukti.io |
| Single-Tenant Platform | 3 | ⏳ Pending | Working locally |
| License Integration | 4 | ⏳ Pending | Features locked on expiry |
| Local IAM Role | 5 | ⏳ Pending | No cross-account |
| Helm Chart | 6 | ⏳ Pending | Deployable to K8s |
| Helm Chart Tested | 7 | ⏳ Pending | Production-ready |
| Documentation | 8 | ⏳ Pending | Complete guides |

### **Success Criteria**

**License API**:
- [ ] Validates license keys
- [ ] Issues JWTs (24h expiry)
- [ ] Tracks activations
- [ ] Handles 1000 req/min
- [ ] 99.9% uptime

**Platform**:
- [ ] Deploys to Kubernetes in <10 minutes
- [ ] Works with local IAM role
- [ ] Validates license on startup
- [ ] Locks features on expiry
- [ ] No customer data sent to Yukti cloud

**Documentation**:
- [ ] Deployment guide (step-by-step)
- [ ] Architecture guide (trust by design)
- [ ] Troubleshooting guide (common issues)
- [ ] Video tutorial (5 minutes)

---

## 🚧 Blockers & Risks

### **Potential Blockers**

1. **License API Downtime**
   - **Risk**: Customer instances can't validate
   - **Mitigation**: 7-day offline grace period

2. **Kubernetes Complexity**
   - **Risk**: Customers struggle with Helm
   - **Mitigation**: Docker Compose alternative

3. **IAM Role Permissions**
   - **Risk**: Customers give wrong permissions
   - **Mitigation**: Clear IAM policy in docs

4. **License Piracy**
   - **Risk**: Customers share license keys
   - **Mitigation**: Track activations, limit instances

### **Risk Mitigation**

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| License API down | High | Low | Offline grace period |
| Helm complexity | Medium | Medium | Docker Compose option |
| Wrong IAM permissions | Medium | High | Clear documentation |
| License piracy | Low | Medium | Activation tracking |

---

## 💰 Cost Analysis

### **SaaS Model** (Current)
- **Your Cost**: $500-$2000/month (infra)
- **Customer Cost**: $99-$1999/month (subscription)
- **Your Margin**: 50-80%

### **Customer-Hosted Model** (Future)
- **Your Cost**: $50-$100/month (License API only)
- **Customer Cost**: $299-$999/month (license) + their infra
- **Your Margin**: 90-95%

**Key Insight**: Customer-hosted is MORE profitable (lower infra cost).

---

## 🎯 Go/No-Go Decision Points

### **After Week 2** (License API)
**Question**: Is License API working and deployed?
- ✅ Go: Continue to platform adaptation
- ❌ No-Go: Fix License API first

### **After Week 5** (Platform Adaptation)
**Question**: Does platform work with license validation?
- ✅ Go: Continue to Helm chart
- ❌ No-Go: Fix platform integration

### **After Week 7** (Helm Chart)
**Question**: Can customers deploy to Kubernetes?
- ✅ Go: Launch customer-hosted offering
- ❌ No-Go: Simplify deployment

---

## 📝 Next Steps

### **Immediate** (This Week)
1. ✅ Create `customer-hosted/` folder structure
2. ✅ Write architecture documentation
3. ✅ Write implementation roadmap
4. ⏳ Get feedback on approach

### **Short-Term** (Next Month)
1. ⏳ Finish SaaS MVP (current root folder)
2. ⏳ Build License API (Week 1-2)
3. ⏳ Test with pilot customer

### **Long-Term** (Next Quarter)
1. ⏳ Complete platform adaptation (Week 3-5)
2. ⏳ Create Helm chart (Week 6-7)
3. ⏳ Launch customer-hosted offering

---

## 🤝 Team Requirements

### **Skills Needed**
- Go development (License API, platform changes)
- Kubernetes/Helm (deployment)
- AWS/GCP/Azure (IAM roles, cloud APIs)
- PostgreSQL (database changes)
- Technical writing (documentation)

### **Time Commitment**
- **Full-time**: 1 developer for 8 weeks
- **Part-time**: 2 developers for 4 weeks each

---

**Last Updated**: January 31, 2025  
**Status**: Planning complete - ready to start Week 1
