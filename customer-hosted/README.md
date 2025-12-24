# Yukti Customer-Hosted Architecture

**Status**: 🚧 Work in Progress  
**Target**: Self-hosted deployment model  
**Timeline**: To be implemented after SaaS MVP is complete

---

## 🎯 Overview

This folder contains the **customer-hosted version** of Yukti FinOps platform.

**Key Differences from SaaS**:
- Runs entirely in customer's environment (Kubernetes/Docker)
- No multi-tenancy (single tenant per deployment)
- License validation via external API
- Data never leaves customer environment
- Local IAM role (no cross-account)

---

## 📁 Structure

```
customer-hosted/
├── license-api/          # License validation service (runs in YOUR cloud)
│   ├── cmd/
│   ├── internal/
│   ├── migrations/
│   ├── Dockerfile
│   └── README.md
│
├── platform/             # Main Yukti platform (runs in CUSTOMER cloud)
│   ├── backend/          # Go API (modified for single-tenant)
│   ├── frontend/         # React UI (same as SaaS)
│   ├── helm/             # Kubernetes Helm chart
│   ├── docker-compose.yml
│   └── README.md
│
└── docs/                 # Customer-hosted documentation
    ├── ARCHITECTURE.md
    ├── DEPLOYMENT.md
    ├── LICENSE_MODEL.md
    └── TRUST_BY_DESIGN.md
```

---

## 🔄 Architecture

### **SaaS Model** (Current - in root folder)
```
┌─────────────────────────────────────────┐
│         YUKTI CLOUD (Your AWS)          │
├─────────────────────────────────────────┤
│  Multi-tenant Platform                  │
│  Cross-account IAM roles                │
│  Centralized database                   │
└─────────────────────────────────────────┘
```

### **Customer-Hosted Model** (This folder)
```
┌─────────────────────────────────────────┐
│      CUSTOMER KUBERNETES CLUSTER        │
├─────────────────────────────────────────┤
│  Yukti Platform (Single-tenant)         │
│  Local IAM role                         │
│  Local database                         │
│  ↓ (License check only)                │
└─────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────┐
│      YOUR CLOUD (license.yukti.io)      │
├─────────────────────────────────────────┤
│  License API                            │
│  - Validate license key                 │
│  - Issue JWT (24h)                      │
│  - No customer data                     │
└─────────────────────────────────────────┘
```

---

## 🚀 Components

### 1. License API (Your Cloud)
**Purpose**: Validate customer licenses  
**Location**: `license-api/`  
**Tech**: Go + PostgreSQL  
**Endpoints**:
- `POST /validate` - Validate license key
- `POST /activate` - Activate on customer instance
- `GET /status` - Check license status
- `POST /refresh` - Refresh JWT

### 2. Platform (Customer Cloud)
**Purpose**: Main FinOps platform  
**Location**: `platform/`  
**Tech**: Go + React + PostgreSQL  
**Changes from SaaS**:
- Remove multi-tenancy
- Remove cross-account IAM
- Add license validation
- Use local IAM role

---

## 📋 Implementation Status

### License API
- [ ] Database schema
- [ ] License validation logic
- [ ] JWT issuance
- [ ] API endpoints
- [ ] Docker deployment
- [ ] Documentation

### Platform
- [ ] Remove multi-tenancy
- [ ] Remove cross-account IAM
- [ ] Add license check on startup
- [ ] Helm chart
- [ ] Docker Compose
- [ ] Documentation

### Documentation
- [ ] Architecture guide
- [ ] Deployment guide
- [ ] License model explanation
- [ ] Trust by design document

---

## 🎯 Next Steps

1. **Complete SaaS MVP first** (current root folder)
2. **Build License API** (this folder)
3. **Adapt platform for single-tenant** (this folder)
4. **Create Helm chart** (this folder)
5. **Write documentation** (this folder)

---

## 💡 Usage

### For Development
```bash
# Work on SaaS version (current)
cd /Users/chandrakantpatil/workspace/yukti
docker-compose up -d

# Work on customer-hosted version (future)
cd /Users/chandrakantpatil/workspace/yukti/customer-hosted
# (to be implemented)
```

### For Deployment
```bash
# SaaS deployment (current)
# Deploy to your AWS account

# Customer-hosted deployment (future)
# Customer deploys to their Kubernetes cluster
helm install yukti ./customer-hosted/platform/helm
```

---

## 📝 Notes

- This folder is **future work** - not blocking current SaaS development
- Code will be adapted from root folder when ready
- License API is the first component to build
- Platform adaptation comes after License API is ready

---

**Last Updated**: January 31, 2025  
**Status**: Planning phase - implementation pending
