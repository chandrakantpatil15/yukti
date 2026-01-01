# Yukti FinOps - Product Offerings

## Two Product Models

### Product 1: Self-Hosted (Customer Infrastructure)
**Target**: Enterprise customers with strict data sovereignty requirements

**Deployment**:
- Customer deploys Yukti on their own AWS/cloud infrastructure
- All data stays within customer's VPC
- Customer manages infrastructure, scaling, backups
- Yukti provides Docker images + Kubernetes manifests

**Architecture**:
```
Customer AWS Account
├── VPC (Customer-owned)
│   ├── Private Subnet
│   │   ├── Yukti Backend (EKS/ECS)
│   │   ├── PostgreSQL (RDS)
│   │   └── Redis (ElastiCache)
│   └── Public Subnet
│       ├── ALB (Load Balancer)
│       └── Yukti Frontend (S3 + CloudFront)
└── IAM Role (ReadOnlyAccess to scan resources)
```

**Pricing**:
- One-time license fee: $10,000 - $50,000
- Annual support: 20% of license fee
- No per-user fees
- Customer pays for their own infrastructure costs

**Pros**:
- ✅ Complete data control
- ✅ No data leaves customer environment
- ✅ Compliance-friendly (HIPAA, SOC2, PCI-DSS)
- ✅ Customizable deployment

**Cons**:
- ❌ Customer manages infrastructure
- ❌ Higher upfront cost
- ❌ Slower updates/features

---

### Product 2: SaaS Multi-Tenant (Yukti-Hosted)
**Target**: SMBs, startups, mid-market companies

**Deployment**:
- Yukti hosts platform on Yukti's AWS infrastructure
- Multi-tenant architecture (shared infrastructure, isolated data)
- Yukti manages all infrastructure, scaling, backups, updates
- Customers access via web browser

**Architecture**:
```
Yukti AWS Account (144403604430)
├── Production VPC
│   ├── Private Subnet
│   │   ├── Yukti Backend (EKS cluster)
│   │   ├── PostgreSQL (RDS Multi-AZ)
│   │   ├── ClickHouse (Analytics)
│   │   └── Redis (ElastiCache cluster)
│   └── Public Subnet
│       ├── ALB (Load Balancer)
│       ├── CloudFront (CDN)
│       └── S3 (Static assets)
├── IAM User: yukti-platform-user
│   └── AssumeRole permissions to customer accounts
└── Cross-Account Access
    ├── Customer Account 1 (IAM Role: YuktiReadOnlyRole)
    ├── Customer Account 2 (IAM Role: YuktiReadOnlyRole)
    └── Customer Account N (IAM Role: YuktiReadOnlyRole)
```

**Data Isolation**:
- **PostgreSQL**: Tenant ID in every table (row-level isolation)
- **ClickHouse**: Separate databases per tenant (tenant_1, tenant_2, etc.)
- **Redis**: Tenant-prefixed keys (tenant:18:session:xyz)
- **S3**: Tenant-specific folders (s3://yukti-data/tenant-18/)

**Pricing**:
- FREE: $0/month (1 AWS account, 30-day data retention)
- PROFESSIONAL: $99/month (3 AWS accounts, 90-day retention)
- ENTERPRISE: $499/month (10 AWS accounts, 2-year retention)
- FINANCIAL: $1,999/month (Unlimited accounts, custom retention)

**Pros**:
- ✅ Zero infrastructure management
- ✅ Instant setup (5 minutes)
- ✅ Automatic updates
- ✅ Lower cost for small teams
- ✅ Pay-as-you-grow

**Cons**:
- ❌ Data stored in Yukti's infrastructure
- ❌ Shared infrastructure (noisy neighbor risk)
- ❌ Less customization

---

## Feature Comparison

| Feature | Self-Hosted | SaaS Multi-Tenant |
|---------|-------------|-------------------|
| **Deployment** | Customer infrastructure | Yukti infrastructure |
| **Data Location** | Customer VPC | Yukti VPC |
| **Setup Time** | 2-4 hours | 5 minutes |
| **Infrastructure Cost** | Customer pays | Included in subscription |
| **Updates** | Manual (quarterly) | Automatic (weekly) |
| **Customization** | High | Limited |
| **Support** | Email + Slack | Email + Chat + Phone |
| **SLA** | Customer-managed | 99.9% uptime |
| **Compliance** | Customer-controlled | SOC2, ISO 27001 |
| **Scaling** | Customer-managed | Auto-scaling |
| **Backup** | Customer-managed | Automated daily |
| **Multi-Account** | Unlimited | Plan-based |
| **Data Retention** | Unlimited | Plan-based |
| **Cost** | $10K-$50K + 20% annual | $0-$1,999/month |

---

## Target Customers

### Self-Hosted Best For:
- **Financial Services**: Banks, insurance, investment firms
- **Healthcare**: HIPAA-compliant organizations
- **Government**: Public sector, defense contractors
- **Large Enterprises**: Fortune 500 with strict data policies
- **Regulated Industries**: PCI-DSS, SOX compliance required

### SaaS Multi-Tenant Best For:
- **Startups**: Fast-growing tech companies
- **SMBs**: Small to medium businesses
- **Mid-Market**: Companies with $100K-$1M AWS spend
- **SaaS Companies**: Cloud-native businesses
- **E-commerce**: Online retailers

---

## Migration Path

**Self-Hosted → SaaS**:
- Export data from self-hosted PostgreSQL
- Import to Yukti SaaS tenant
- Reconfigure IAM roles to trust Yukti account
- Decommission self-hosted infrastructure

**SaaS → Self-Hosted**:
- Provide Docker images + Kubernetes manifests
- Customer deploys to their infrastructure
- Export data from Yukti SaaS
- Import to customer's PostgreSQL
- Reconfigure IAM roles to trust customer account

---

## Current Status

**Product 1 (Self-Hosted)**: ✅ Ready for deployment
- Docker images available
- Kubernetes manifests ready
- Documentation complete
- Tested on AWS EKS

**Product 2 (SaaS Multi-Tenant)**: ✅ Live in production
- Platform running on AWS (144403604430)
- Multi-tenant architecture implemented
- RBAC with 4 roles (owner, admin, editor, viewer)
- Cross-account IAM integration working
- 16-region AWS scanning operational
- 77 cost detectors active

---

## Recommended Go-To-Market Strategy

**Phase 1 (Months 1-6)**: Focus on SaaS Multi-Tenant
- Lower barrier to entry
- Faster customer acquisition
- Prove product-market fit
- Build case studies

**Phase 2 (Months 7-12)**: Introduce Self-Hosted
- Target enterprise customers from SaaS waitlist
- Leverage SaaS success stories
- Higher deal sizes
- Longer sales cycles

**Phase 3 (Year 2+)**: Hybrid Model
- Offer both products
- Upsell SaaS customers to self-hosted
- Cross-sell self-hosted customers to SaaS for dev/test environments
