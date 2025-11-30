# 👨‍💼 Admin Flow Design (Platform Admin)

## Overview
Platform-level admin interface for Yukti team to manage all tenants, monitor platform health, and provide support.

---

## 🎯 Admin vs Tenant Owner

| Feature | Platform Admin | Tenant Owner |
|---------|---------------|--------------|
| Scope | All tenants | Single tenant |
| Access | Platform-wide | Tenant-specific |
| User Management | All users | Tenant users only |
| Billing | All subscriptions | Own subscription |
| Support | Can impersonate | Cannot impersonate |
| Analytics | Platform metrics | Tenant metrics |

---

## 🔐 Admin Authentication

### Admin User Creation
```sql
-- Create admin user
INSERT INTO yt_users (
  id,
  email,
  password_hash,
  role,
  is_platform_admin,
  created_at
) VALUES (
  gen_random_uuid(),
  'admin@yukti.com',
  $hashed_password,
  'platform_admin',
  true,
  NOW()
);
```

### Admin Login Flow
```
1. Navigate to /admin/login (separate from user login)
2. Enter admin credentials
3. 2FA verification (required for admins)
4. Admin dashboard access
```

---

## 📊 Admin Dashboard

### Layout
```
┌─────────────────────────────────────────────────────────┐
│ Yukti Admin Portal                    [Admin] [Logout]  │
├─────────────────────────────────────────────────────────┤
│                                                          │
│ Platform Overview                                        │
│                                                          │
│ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐   │
│ │ Tenants  │ │  Users   │ │ Revenue  │ │ Findings │   │
│ │   245    │ │  1,234   │ │ $45.2K   │ │  12,456  │   │
│ │  +12%    │ │  +8%     │ │  +15%    │ │  +23%    │   │
│ └──────────┘ └──────────┘ └──────────┘ └──────────┘   │
│                                                          │
│ Recent Activity                                          │
│ ┌────────────────────────────────────────────────────┐  │
│ │ 🆕 New tenant: Acme Corp (5 min ago)              │  │
│ │ 💰 Subscription upgraded: TechStart ($99→$499)    │  │
│ │ ⚠️  Support ticket: CloudScale (#1234)            │  │
│ │ 🔍 Scan completed: DataCorp (1,234 resources)     │  │
│ └────────────────────────────────────────────────────┘  │
│                                                          │
│ Platform Health                                          │
│ ┌────────────────────────────────────────────────────┐  │
│ │ API Response Time: 145ms ✅                        │  │
│ │ Database Connections: 45/100 ✅                    │  │
│ │ Scanner Queue: 3 pending ✅                        │  │
│ │ Error Rate: 0.02% ✅                               │  │
│ └────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

---

## 🏢 Tenant Management

### Tenant List View
```
┌─────────────────────────────────────────────────────────┐
│ All Tenants                    [Search] [Filter] [Export]│
├─────────────────────────────────────────────────────────┤
│                                                          │
│ ┌────────────────────────────────────────────────────┐  │
│ │ 🏢 Acme Corp                          [Impersonate]│  │
│ │    ID: 1 • Owner: john@acme.com                    │  │
│ │    Plan: Professional ($99/mo)                     │  │
│ │    Users: 5 • Resources: 234 • Findings: 45       │  │
│ │    Created: Jan 1, 2024 • Last Active: 2h ago     │  │
│ │    Status: ✅ Active • Billing: ✅ Current         │  │
│ │    [View Details] [Suspend] [Delete]               │  │
│ └────────────────────────────────────────────────────┘  │
│                                                          │
│ ┌────────────────────────────────────────────────────┐  │
│ │ 🏢 TechStart Inc                      [Impersonate]│  │
│ │    ID: 2 • Owner: jane@techstart.com               │  │
│ │    Plan: Enterprise ($499/mo)                      │  │
│ │    Users: 12 • Resources: 1,234 • Findings: 234   │  │
│ │    Created: Jan 5, 2024 • Last Active: 10m ago    │  │
│ │    Status: ✅ Active • Billing: ⚠️  Overdue        │  │
│ │    [View Details] [Suspend] [Delete]               │  │
│ └────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

### Tenant Detail View
```
┌─────────────────────────────────────────────────────────┐
│ ← Back to Tenants                                        │
├─────────────────────────────────────────────────────────┤
│                                                          │
│ 🏢 Acme Corp (ID: 1)                    [Impersonate]   │
│                                                          │
│ ┌─────────────────────────────────────────────────────┐ │
│ │ Overview                                            │ │
│ │                                                     │ │
│ │ Owner: john@acme.com                                │ │
│ │ Company: Acme Corporation                           │ │
│ │ Created: January 1, 2024                            │ │
│ │ Last Active: 2 hours ago                            │ │
│ │ Status: ✅ Active                                    │ │
│ └─────────────────────────────────────────────────────┘ │
│                                                          │
│ ┌─────────────────────────────────────────────────────┐ │
│ │ Subscription                                        │ │
│ │                                                     │ │
│ │ Plan: Professional ($99/month)                      │ │
│ │ Billing Cycle: Monthly                              │ │
│ │ Next Billing: February 1, 2024                      │ │
│ │ Payment Method: •••• 4242                           │ │
│ │ Status: ✅ Current                                   │ │
│ │                                                     │ │
│ │ [Upgrade Plan] [Change Billing] [Cancel]            │ │
│ └─────────────────────────────────────────────────────┘ │
│                                                          │
│ ┌─────────────────────────────────────────────────────┐ │
│ │ Usage Statistics                                    │ │
│ │                                                     │ │
│ │ Users: 5/10                                         │ │
│ │ AWS Accounts: 2/5                                   │ │
│ │ Resources Scanned: 234                              │ │
│ │ Findings Generated: 45                              │ │
│ │ Savings Identified: $12,450/month                   │ │
│ │ API Calls (30d): 12,456                             │ │
│ └─────────────────────────────────────────────────────┘ │
│                                                          │
│ ┌─────────────────────────────────────────────────────┐ │
│ │ Actions                                             │ │
│ │                                                     │ │
│ │ [Impersonate User] [Reset Password]                 │ │
│ │ [Suspend Account] [Delete Account]                  │ │
│ │ [View Audit Logs] [Export Data]                     │ │
│ └─────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────┘
```

---

## 👥 User Management (All Users)

### All Users View
```
┌─────────────────────────────────────────────────────────┐
│ All Users                      [Search] [Filter] [Export]│
├─────────────────────────────────────────────────────────┤
│                                                          │
│ Filters: [All Tenants ▼] [All Roles ▼] [Active ▼]      │
│                                                          │
│ ┌────────────────────────────────────────────────────┐  │
│ │ 👤 John Doe                                        │  │
│ │    john@acme.com                                   │  │
│ │    Tenant: Acme Corp (Owner)                       │  │
│ │    Created: Jan 1, 2024 • Last Login: 2h ago      │  │
│ │    Status: ✅ Active • Email: ✅ Verified          │  │
│ │    [View Details] [Reset Password] [Suspend]       │  │
│ └────────────────────────────────────────────────────┘  │
│                                                          │
│ ┌────────────────────────────────────────────────────┐  │
│ │ 👤 Jane Smith                                      │  │
│ │    jane@techstart.com                              │  │
│ │    Tenant: TechStart Inc (Owner)                   │  │
│ │    Created: Jan 5, 2024 • Last Login: 10m ago     │  │
│ │    Status: ✅ Active • Email: ✅ Verified          │  │
│ │    [View Details] [Reset Password] [Suspend]       │  │
│ └────────────────────────────────────────────────────┘  │
│                                                          │
│ ┌────────────────────────────────────────────────────┐  │
│ │ 👤 Bob Wilson                                      │  │
│ │    bob@acme.com                                    │  │
│ │    Tenant: Acme Corp (Editor)                      │  │
│ │    Created: Jan 10, 2024 • Last Login: 1d ago     │  │
│ │    Status: ⚠️  Suspended • Email: ✅ Verified      │  │
│ │    [View Details] [Reset Password] [Activate]      │  │
│ └────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

---

## 🔍 Impersonation Feature

### Impersonate Flow
```
1. Admin clicks "Impersonate" on tenant
2. Confirmation modal:
   ┌─────────────────────────────────────┐
   │ Impersonate User              [×]   │
   ├─────────────────────────────────────┤
   │                                     │
   │ ⚠️  You are about to impersonate:   │
   │                                     │
   │ User: john@acme.com                 │
   │ Tenant: Acme Corp                   │
   │ Role: Owner                         │
   │                                     │
   │ This action will be logged in       │
   │ audit logs for security purposes.   │
   │                                     │
   │ Reason for impersonation:           │
   │ ┌─────────────────────────────────┐ │
   │ │ Customer support request #1234  │ │
   │ └─────────────────────────────────┘ │
   │                                     │
   │      [Cancel]  [Start Impersonation]│
   └─────────────────────────────────────┘

3. Admin sees tenant's view with banner:
   ┌─────────────────────────────────────────────────────┐
   │ ⚠️  IMPERSONATION MODE: Viewing as john@acme.com    │
   │ Reason: Customer support #1234    [End Impersonation]│
   └─────────────────────────────────────────────────────┘

4. All actions logged to audit trail
5. Admin clicks "End Impersonation"
6. Returns to admin dashboard
```

### Impersonation Audit Log
```sql
INSERT INTO yt_admin_audit_logs (
  admin_user_id,
  action,
  target_user_id,
  target_tenant_id,
  reason,
  ip_address,
  user_agent,
  created_at
) VALUES (
  $admin_id,
  'impersonate_start',
  $target_user_id,
  $target_tenant_id,
  $reason,
  $ip,
  $user_agent,
  NOW()
);
```

---

## 💰 Billing & Revenue

### Revenue Dashboard
```
┌─────────────────────────────────────────────────────────┐
│ Revenue Overview                                         │
├─────────────────────────────────────────────────────────┤
│                                                          │
│ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐   │
│ │   MRR    │ │   ARR    │ │  Churn   │ │   LTV    │   │
│ │ $45,234  │ │ $542,808 │ │   2.3%   │ │ $12,450  │   │
│ │  +15%    │ │  +18%    │ │  -0.5%   │ │  +12%    │   │
│ └──────────┘ └──────────┘ └──────────┘ └──────────┘   │
│                                                          │
│ Revenue by Plan                                          │
│ ┌────────────────────────────────────────────────────┐  │
│ │ Free:         45 tenants  •  $0/mo                 │  │
│ │ Professional: 120 tenants •  $11,880/mo            │  │
│ │ Enterprise:   80 tenants  •  $39,920/mo            │  │
│ └────────────────────────────────────────────────────┘  │
│                                                          │
│ Overdue Payments                                         │
│ ┌────────────────────────────────────────────────────┐  │
│ │ 🏢 TechStart Inc  •  $499  •  15 days overdue      │  │
│ │ 🏢 CloudScale LLC •  $99   •  5 days overdue       │  │
│ └────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

---

## 📈 Platform Analytics

### Analytics Dashboard
```
┌─────────────────────────────────────────────────────────┐
│ Platform Analytics                    [Last 30 Days ▼]  │
├─────────────────────────────────────────────────────────┤
│                                                          │
│ User Growth                                              │
│ ┌────────────────────────────────────────────────────┐  │
│ │     📈 +234 new users                              │  │
│ │                                                    │  │
│ │  1500│                                    ●        │  │
│ │  1000│                          ●    ●             │  │
│ │   500│              ●      ●                       │  │
│ │     0└──────────────────────────────────────────  │  │
│ │       Jan 1    Jan 10   Jan 20   Jan 30           │  │
│ └────────────────────────────────────────────────────┘  │
│                                                          │
│ Resource Scans                                           │
│ ┌────────────────────────────────────────────────────┐  │
│ │ Total Scans: 1,234                                 │  │
│ │ Resources Discovered: 45,678                       │  │
│ │ Findings Generated: 12,456                         │  │
│ │ Total Savings Identified: $2.4M/month              │  │
│ └────────────────────────────────────────────────────┘  │
│                                                          │
│ Top Tenants by Usage                                     │
│ ┌────────────────────────────────────────────────────┐  │
│ │ 1. Acme Corp       • 1,234 resources • 234 scans  │  │
│ │ 2. TechStart Inc   • 987 resources  • 189 scans   │  │
│ │ 3. CloudScale LLC  • 756 resources  • 145 scans   │  │
│ └────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

---

## 🎫 Support Tickets

### Support Dashboard
```
┌─────────────────────────────────────────────────────────┐
│ Support Tickets                [New] [Filter] [Export]   │
├─────────────────────────────────────────────────────────┤
│                                                          │
│ Open Tickets (12)                                        │
│                                                          │
│ ┌────────────────────────────────────────────────────┐  │
│ │ 🔴 #1234 - Cannot connect AWS account              │  │
│ │    Acme Corp • john@acme.com                       │  │
│ │    Priority: High • Created: 2h ago                │  │
│ │    [View] [Impersonate] [Assign to Me]             │  │
│ └────────────────────────────────────────────────────┘  │
│                                                          │
│ ┌────────────────────────────────────────────────────┐  │
│ │ 🟡 #1235 - Billing question                        │  │
│ │    TechStart Inc • jane@techstart.com              │  │
│ │    Priority: Medium • Created: 5h ago              │  │
│ │    [View] [Impersonate] [Assign to Me]             │  │
│ └────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

---

## 🔧 System Configuration

### Platform Settings
```
┌─────────────────────────────────────────────────────────┐
│ Platform Configuration                                   │
├─────────────────────────────────────────────────────────┤
│                                                          │
│ ┌─────────────────────────────────────────────────────┐ │
│ │ Feature Flags                                       │ │
│ │                                                     │ │
│ │ ☑ Multi-tenant support                              │ │
│ │ ☑ AWS integration                                   │ │
│ │ ☑ CloudWatch metrics                                │ │
│ │ ☐ Azure integration (coming soon)                   │ │
│ │ ☐ GCP integration (coming soon)                     │ │
│ └─────────────────────────────────────────────────────┘ │
│                                                          │
│ ┌─────────────────────────────────────────────────────┐ │
│ │ Rate Limits                                         │ │
│ │                                                     │ │
│ │ API calls per minute: 100                           │ │
│ │ Scans per day: 10                                   │ │
│ │ Max resources per tenant: 10,000                    │ │
│ └─────────────────────────────────────────────────────┘ │
│                                                          │
│ ┌─────────────────────────────────────────────────────┐ │
│ │ Email Configuration                                 │ │
│ │                                                     │ │
│ │ Provider: AWS SES                                   │ │
│ │ From: noreply@yukti.com                             │ │
│ │ Daily limit: 50,000                                 │ │
│ │ Status: ✅ Active                                    │ │
│ └─────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────┘
```

---

## 🔐 API Endpoints (Admin)

### Admin Authentication
```
POST /api/admin/login
POST /api/admin/2fa/verify
POST /api/admin/logout
```

### Tenant Management
```
GET    /api/admin/tenants
GET    /api/admin/tenants/:id
POST   /api/admin/tenants/:id/suspend
POST   /api/admin/tenants/:id/activate
DELETE /api/admin/tenants/:id
POST   /api/admin/tenants/:id/impersonate
POST   /api/admin/tenants/:id/end-impersonation
```

### User Management
```
GET    /api/admin/users
GET    /api/admin/users/:id
POST   /api/admin/users/:id/reset-password
POST   /api/admin/users/:id/suspend
POST   /api/admin/users/:id/activate
DELETE /api/admin/users/:id
```

### Analytics
```
GET /api/admin/analytics/overview
GET /api/admin/analytics/revenue
GET /api/admin/analytics/usage
GET /api/admin/analytics/growth
```

### Audit Logs
```
GET /api/admin/audit-logs
GET /api/admin/audit-logs/impersonations
GET /api/admin/audit-logs/tenant/:id
GET /api/admin/audit-logs/user/:id
```

---

## 📋 Implementation Priority

### Phase 1: Core Admin Features (Week 1-2)
- [ ] Admin authentication with 2FA
- [ ] Tenant list and detail views
- [ ] User management (list, suspend, activate)
- [ ] Basic impersonation
- [ ] Audit logging

### Phase 2: Analytics & Monitoring (Week 3)
- [ ] Platform analytics dashboard
- [ ] Revenue tracking
- [ ] Usage statistics
- [ ] System health monitoring

### Phase 3: Support Features (Week 4)
- [ ] Support ticket system
- [ ] Advanced impersonation with reason tracking
- [ ] Tenant data export
- [ ] Bulk operations

### Phase 4: Advanced Features (Week 5+)
- [ ] Feature flags management
- [ ] Rate limit configuration
- [ ] Email template editor
- [ ] Automated alerts
- [ ] Platform API documentation

---

**Admin portal design complete!** 👨‍💼✅
