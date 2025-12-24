# UI Designs - Index

This folder contains detailed specifications for all implemented features in Yukti FinOps platform.

## Purpose
- **Standardized Format**: All features documented in consistent template
- **Developer Reference**: Clear specs for implementation and maintenance
- **Onboarding**: New developers can understand features quickly
- **Future Development**: Template for new feature specifications

## Document Structure
Each feature document includes:
1. **Priority**: HIGH/MEDIUM/LOW + Implementation status
2. **What It Does**: 1-2 sentence description
3. **Visual Reference**: ASCII mockup of UI
4. **User Flow**: Step-by-step user journey
5. **Data Requirements**: Input/output formats
6. **API Endpoints**: Complete API specs with examples
7. **Database Tables**: Schema and relationships
8. **UI Components**: Page paths and component files
9. **Business Rules**: Validation, permissions, limits
10. **Security Features**: Authentication, authorization, audit
11. **Implementation Status**: Files, testing, deployment
12. **Future Enhancements**: Planned improvements

## Implemented Features

### Core User Features
1. **[Login Page](01-login-page.md)** - User authentication with JWT
2. **[Dashboard Page](02-dashboard-page.md)** - Main metrics and overview
3. **[Hidden Costs Page](03-hidden-costs-page.md)** - Cost optimization findings
4. **[Onboarding Page](04-onboarding-page.md)** - AWS account connection
5. **[Resources Page](05-resources-page.md)** - AWS resource inventory

### Admin Features
6. **[Admin Portal](06-admin-portal.md)** - Platform administration

## Planned Features (Not Yet Documented)
- Whitelists Page (approval workflow)
- IaC Generator Page (Terraform/CloudFormation)
- Budgets Page (budget tracking and alerts)
- Profile Page (user settings)
- Team Management Page (RBAC, invitations)
- Cost Analytics Page (ClickHouse-powered)
- Resource Utilization Page (right-sizing)

## How to Use This Folder

### For Developers
1. Read feature spec before implementing changes
2. Update spec when adding new functionality
3. Use template for new features

### For Product Managers
1. Review specs to understand current state
2. Add new feature specs using template
3. Update priorities and enhancements

### For QA/Testing
1. Use user flows for test scenarios
2. Verify API endpoints match specs
3. Check business rules are enforced

## Template for New Features
See any existing document for template structure. Key sections:
- Priority + Status
- Visual mockup
- User flow (numbered steps)
- API endpoints (request/response examples)
- Database schema
- Business rules
- Security features

## Maintenance
- Update specs when features change
- Mark deprecated features
- Add migration notes for breaking changes
- Keep API examples current

---

**Last Updated**: Session 25 - Phase 24 Complete (RBAC Implementation)
**Total Features Documented**: 6
**Total API Endpoints**: 50+
**Total Database Tables**: 15+
