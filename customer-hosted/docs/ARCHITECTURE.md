# Customer-Hosted Architecture

**Document**: Complete architecture for self-hosted Yukti FinOps  
**Audience**: Technical decision makers, platform engineers  
**Status**: Design phase - implementation pending

---

## 🎯 Core Principle

> "Yukti runs in your environment, keeps your data with you, and asks for nothing except a license check."

---

## 🏗️ System Architecture

### **High-Level Overview**

```
┌─────────────────────────────────────────────────────────────┐
│           CUSTOMER KUBERNETES CLUSTER                       │
│           (Your AWS / GCP / Azure / On-Prem)               │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  YUKTI PLATFORM (Single-Tenant)                      │  │
│  ├──────────────────────────────────────────────────────┤  │
│  │                                                       │  │
│  │  Frontend (React)                                     │  │
│  │  ↓                                                    │  │
│  │  Backend (Go)                                         │  │
│  │  ↓                                                    │  │
│  │  PostgreSQL (Local)                                   │  │
│  │                                                       │  │
│  │  ↓ (Local IAM Role / Service Account)               │  │
│  │  AWS / GCP / Azure APIs                              │  │
│  │  - Read-only access                                   │  │
│  │  - Cost data                                          │  │
│  │  - Resource metadata                                  │  │
│  │  - CloudWatch metrics                                 │  │
│  │                                                       │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                             │
│  ↓ (License Check - HTTPS)                                 │
│  Only sends: license_key, instance_id, version             │
│                                                             │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│           YUKTI CLOUD (license.yukti.io)                    │
│           (Your AWS Account)                                │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  LICENSE API (Minimal Service)                       │  │
│  ├──────────────────────────────────────────────────────┤  │
│  │                                                       │  │
│  │  Endpoints:                                           │  │
│  │  - POST /validate (validate license key)             │  │
│  │  - GET /status (check license status)                │  │
│  │  - POST /refresh (refresh JWT)                       │  │
│  │                                                       │  │
│  │  Database:                                            │  │
│  │  - license_keys (customer licenses)                  │  │
│  │  - activations (active instances)                    │  │
│  │  - feature_flags (plan features)                     │  │
│  │                                                       │  │
│  │  ⚠️ NO CUSTOMER DATA STORED                          │  │
│  │  ⚠️ NO AWS COST DATA                                 │  │
│  │  ⚠️ NO RESOURCE METADATA                             │  │
│  │                                                       │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## 🔄 Data Flow

### **1. Initial Deployment**

```
Customer Admin
    ↓
    1. Purchases license from yukti.io
    ↓
    2. Receives license key via email
       YUKTI-A1B2-C3D4-E5F6
    ↓
    3. Deploys Helm chart to Kubernetes
       helm install yukti ./helm \
         --set license.key=YUKTI-A1B2-C3D4-E5F6
    ↓
    4. Yukti platform starts
    ↓
    5. Calls License API to validate
       POST /api/v1/license/validate
       {
         "license_key": "YUKTI-A1B2-C3D4-E5F6",
         "instance_id": "k8s-cluster-hash",
         "version": "1.0.0"
       }
    ↓
    6. License API responds with JWT
       {
         "valid": true,
         "jwt": "eyJhbGc...",
         "expires_at": "2025-02-01T10:00:00Z",
         "features": {...}
       }
    ↓
    7. Yukti platform unlocks features
    ↓
    8. Customer accesses dashboard
```

### **2. Daily Operations**

```
Yukti Platform (Customer Environment)
    ↓
    Every 24 hours:
    - Refresh JWT from License API
    - Continue operations
    ↓
    Every hour:
    - Scan AWS resources (local IAM role)
    - Collect CloudWatch metrics
    - Run 77 cost detectors
    - Store findings in local PostgreSQL
    ↓
    User accesses dashboard:
    - All data served from local database
    - No external API calls
    - Fast response times
```

### **3. License Expiry**

```
License expires
    ↓
    Yukti platform attempts JWT refresh
    ↓
    License API returns 401 Unauthorized
    ↓
    Yukti platform enters read-only mode:
    - Dashboard still accessible
    - Historical data visible
    - No new scans
    - No IaC generation
    - Banner: "License expired. Renew at yukti.io"
    ↓
    Customer renews license
    ↓
    Yukti platform validates new key
    ↓
    Full features restored
```

---

## 🔐 Security Model

### **Trust by Design**

#### **What Customer Trusts**
1. ✅ Data stays in their environment
2. ✅ Read-only AWS access
3. ✅ Open-source code (can audit)
4. ✅ Minimal external communication

#### **What Customer Doesn't Need to Trust**
1. ❌ Yukti won't steal data (data never leaves)
2. ❌ Yukti won't modify resources (read-only)
3. ❌ Yukti won't share data (no data to share)

### **IAM Permissions (AWS Example)**

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ec2:Describe*",
        "rds:Describe*",
        "s3:ListBucket",
        "s3:GetBucketLocation",
        "cloudwatch:GetMetricStatistics",
        "ce:GetCostAndUsage"
      ],
      "Resource": "*"
    }
  ]
}
```

**Key Points**:
- Read-only (no `Create`, `Delete`, `Modify`)
- No secrets access
- No data plane access (only metadata)

### **Network Security**

```
Yukti Platform
    ↓ (Outbound HTTPS only)
    license.yukti.io:443
    ↓ (Payload)
    {
      "license_key": "YUKTI-...",
      "instance_id": "k8s-...",
      "version": "1.0.0"
    }
```

**What's NOT sent**:
- ❌ AWS cost data
- ❌ Resource metadata
- ❌ CloudWatch metrics
- ❌ Customer names
- ❌ Account IDs
- ❌ Any PII

---

## 📦 Deployment Models

### **1. Kubernetes (Recommended)**

```bash
# Add Helm repo
helm repo add yukti https://charts.yukti.io

# Install
helm install yukti yukti/yukti \
  --set license.key=YUKTI-XXXX-XXXX-XXXX \
  --set aws.region=us-east-1 \
  --set ingress.enabled=true \
  --set ingress.host=yukti.company.com
```

**Resources Created**:
- Deployment (backend, frontend)
- StatefulSet (PostgreSQL)
- Service (ClusterIP, LoadBalancer)
- Ingress (HTTPS)
- ConfigMap (configuration)
- Secret (license key, JWT secret)

### **2. Docker Compose (Development)**

```bash
# Clone repo
git clone https://github.com/yukti/customer-hosted

# Configure
cp .env.example .env
# Edit .env with license key

# Start
docker-compose up -d

# Access
open http://localhost:3000
```

### **3. AWS ECS (Alternative)**

```bash
# Deploy via Terraform
cd terraform/aws-ecs
terraform init
terraform apply \
  -var="license_key=YUKTI-XXXX-XXXX-XXXX" \
  -var="region=us-east-1"
```

---

## 🎛️ Configuration

### **Environment Variables**

```bash
# License
LICENSE_KEY=YUKTI-XXXX-XXXX-XXXX
LICENSE_API_URL=https://license.yukti.io

# AWS (if using AWS)
AWS_REGION=us-east-1
AWS_ROLE_ARN=arn:aws:iam::123456789012:role/YuktiReadOnly

# Database
DATABASE_URL=postgres://yukti:password@postgres:5432/yukti

# Features
ENABLE_CUSTOM_DETECTORS=true
MAX_DETECTORS=77
SCAN_INTERVAL=1h

# Offline Mode
OFFLINE_GRACE_PERIOD=168h  # 7 days
```

### **Helm Values**

```yaml
# values.yaml
license:
  key: "YUKTI-XXXX-XXXX-XXXX"
  apiUrl: "https://license.yukti.io"

aws:
  region: "us-east-1"
  roleArn: "arn:aws:iam::123456789012:role/YuktiReadOnly"

database:
  persistence:
    enabled: true
    size: 50Gi
    storageClass: gp3

ingress:
  enabled: true
  host: "yukti.company.com"
  tls:
    enabled: true
    secretName: "yukti-tls"

resources:
  backend:
    requests:
      cpu: 500m
      memory: 1Gi
    limits:
      cpu: 2000m
      memory: 4Gi
```

---

## 📊 Comparison: SaaS vs Customer-Hosted

| Aspect | SaaS (Current) | Customer-Hosted (This Folder) |
|--------|----------------|-------------------------------|
| **Data Location** | Your cloud | Customer cloud |
| **Multi-Tenancy** | Yes (shared DB) | No (single tenant) |
| **IAM Access** | Cross-account | Local role |
| **Compliance** | Your responsibility | Customer responsibility |
| **Deployment** | You deploy | Customer deploys |
| **Updates** | Automatic | Manual (Helm upgrade) |
| **Cost** | Your infra cost | Customer infra cost |
| **Trust** | Customer trusts you | Customer trusts architecture |
| **Pricing** | $99-$1999/mo | $299-$999/mo (license) |

---

## 🚀 Migration Path (SaaS → Customer-Hosted)

### **Phase 1: Build License API** (2 weeks)
- [ ] Database schema
- [ ] API endpoints
- [ ] JWT generation
- [ ] Deploy to production

### **Phase 2: Adapt Platform** (3 weeks)
- [ ] Remove multi-tenancy
- [ ] Remove cross-account IAM
- [ ] Add license validation
- [ ] Use local IAM role
- [ ] Test end-to-end

### **Phase 3: Create Helm Chart** (2 weeks)
- [ ] Chart structure
- [ ] Templates (deployment, service, ingress)
- [ ] Values configuration
- [ ] Test on Kubernetes

### **Phase 4: Documentation** (1 week)
- [ ] Deployment guide
- [ ] Architecture guide
- [ ] Trust by design doc
- [ ] Troubleshooting guide

**Total**: 8 weeks

---

## 💡 Key Decisions

### **Why Separate License API?**
- Customer doesn't run it (runs in your cloud)
- Minimal attack surface
- Easy to secure
- Can't be bypassed easily

### **Why Local IAM Role?**
- No cross-account complexity
- Customer controls permissions
- Easier to audit
- More secure

### **Why Single-Tenant?**
- Simpler architecture
- No tenant isolation bugs
- Better performance
- Customer preference

### **Why Helm Chart?**
- Industry standard
- Easy to customize
- Version control
- Rollback support

---

## 📝 Next Steps

1. **Finish SaaS MVP** (current root folder)
2. **Build License API** (this folder)
3. **Test with pilot customer**
4. **Iterate based on feedback**
5. **Launch customer-hosted offering**

---

**Last Updated**: January 31, 2025  
**Status**: Design complete - ready for implementation
