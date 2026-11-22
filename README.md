# Yukti: Open-Source AWS FinOps Platform

**Find hidden AWS costs automatically, and get the Terraform to fix them.**

Yukti scans your AWS accounts with 77 cost detectors, forecasts spend with ML, and generates Terraform / CloudFormation to remediate waste, all through a read-only cross-account role.

![Yukti dashboard](docs/images/dashboard.png) <!-- TODO: add a real screenshot or demo GIF -->

> **Status: Beta.** Core scanning, dashboard, RBAC and IaC generation work. CI/CD, performance testing and Kubernetes deployment are still in progress. See the [Roadmap](#roadmap). Feedback and contributions are very welcome.

---

## Why Yukti?

Most cost tools show you *what you spent*. Yukti tells you *what you're wasting* and *how to stop it*.

- **77 detectors** (65 AWS + 12 Kubernetes): idle EC2, unattached EBS volumes, old snapshots, unencrypted RDS, missing S3 lifecycle policies, data-transfer waste and more
- **Fix, not just report**: generates Terraform and CloudFormation for each finding
- **ML-powered**: data-transfer forecasting, anomaly detection, savings-confidence scoring
- **Safe by design**: read-only IAM role, auto-generated External ID (prevents confused-deputy attacks), tenant isolation
- **Multi-region**: scans 16 AWS regions
- **Self-hostable**: runs with a single `docker-compose` command

| | Yukti | OpenCost | Infracost | Cloud Custodian |
|---|---|---|---|---|
| Waste detection on live AWS | ✅ | ❌ | ❌ | ✅ |
| Auto-generated IaC fixes | ✅ | ❌ | ❌ | ❌ |
| ML forecasting / anomaly detection | ✅ | ❌ | ❌ | ❌ |
| Web dashboard + RBAC | ✅ | Partial | Cloud only | ❌ |

<!-- TODO: double-check this table is accurate before publishing. Wrong claims about other tools will hurt trust. -->

## Quick Start (2 minutes)

Requires Docker and Docker Compose.

```bash
git clone https://github.com/chandrakantpatil15/yukti.git
cd yukti
cp .env.example .env
make start
```

Then open:

- Dashboard: http://localhost:3000
- API: http://localhost:8081
- Grafana: http://localhost:3001

The demo ships with seed data (3 sample tenants, sample findings), so you can explore the dashboard before connecting a real AWS account.

## Connect Your AWS Account

1. Create an IAM role with `ReadOnlyAccess` that trusts your Yukti account
2. Paste the Account ID and Role ARN into onboarding (an External ID is generated for you)
3. Yukti verifies with STS `AssumeRole`, then starts scanning

Details: [AWS_ROLE_TESTING.md](AWS_ROLE_TESTING.md)

## How It Works

```
React + TypeScript  -->  Go API (Gin)  -->  PostgreSQL
                              |
                              +--> AWS Scanner (16 regions, STS AssumeRole)
                              +--> 77 Detectors --> Findings
                              +--> ML Service (FastAPI + scikit-learn)
                              +--> IaC Generator (Terraform / CloudFormation)
```

**Tech stack:** Go 1.23, React 18 + TypeScript + Tailwind, PostgreSQL 15, Python FastAPI, Prometheus + Grafana, Docker.

Full architecture: [PLATFORM_ARCHITECTURE.md](PLATFORM_ARCHITECTURE.md)

## Roadmap

- [x] AWS multi-region scanning (EC2, RDS, S3, CloudWatch)
- [x] 77 detectors, ML models, IaC generation
- [x] RBAC (owner / admin / editor / viewer) and admin portal
- [ ] CI/CD pipeline and performance testing
- [ ] Kubernetes deployment (Helm chart)
- [ ] Multi-account AWS support
- [ ] Azure support
- [ ] GCP support

## Contributing

Contributions are welcome, including small ones.

- Look for issues labeled [`good first issue`](https://github.com/chandrakantpatil15/yukti/labels/good%20first%20issue) and [`help wanted`](https://github.com/chandrakantpatil15/yukti/labels/help%20wanted)
- Read [CONTRIBUTING.md](CONTRIBUTING.md)
- Ideas, questions, and feedback: open a [Discussion](https://github.com/chandrakantpatil15/yukti/discussions)

**Easy ways to help:** add a new detector, improve docs, test against your own AWS account and report what you find.

## Documentation

- [API Documentation](API_DOCUMENTATION.md): 50+ endpoints
- [Deployment Guide](DEPLOYMENT_GUIDE.md): Docker and Kubernetes
- [Hidden Cost Detection Spec](HIDDEN_COST_DETECTION_SPEC.md): all 77 detectors
- [CHANGELOG](CHANGELOG.md)

## License

See [LICENSE](LICENSE).

---

If Yukti saves you money, please ⭐ the repo. It helps other people find it.
