# 🎨 User Flows - Visual Diagrams

## 1. 🆕 New User Onboarding Flow

```
┌─────────────────────────────────────────────────────────────┐
│                    NEW USER JOURNEY                          │
└─────────────────────────────────────────────────────────────┘

Step 1: Signup
┌──────────────┐
│   /signup    │
│              │
│ Enter:       │
│ • Email      │
│ • Password   │
│ • Company    │
└──────┬───────┘
       │
       ▼
Step 2: Email Verification
┌──────────────┐
│   OTP Code   │
│              │
│ • Check email│
│ • Enter OTP  │
│ • Verify     │
└──────┬───────┘
       │
       ▼
Step 3: Onboarding (MANDATORY)
┌──────────────┐
│ /onboarding  │
│              │
│ Configure:   │
│ • AWS Acct   │
│ • IAM Role   │
│ • Verify     │
└──────┬───────┘
       │
       ▼
Step 4: Dashboard Access ✅
┌──────────────┐
│  /dashboard  │
│              │
│ • View data  │
│ • Scan AWS   │
│ • Findings   │
└──────────────┘
```

---

## 2. 🔄 Returning User Flow

```
┌─────────────────────────────────────────────────────────────┐
│                 RETURNING USER JOURNEY                       │
└─────────────────────────────────────────────────────────────┘

Login
┌──────────────┐
│   /login     │
│              │
│ Enter:       │
│ • Email      │
│ • Password   │
└──────┬───────┘
       │
       ▼
Check Onboarding Status
┌──────────────────────────────────────┐
│ Has AWS Connection?                  │
└──────┬───────────────────────┬───────┘
       │ NO                    │ YES
       ▼                       ▼
┌──────────────┐      ┌──────────────┐
│ /onboarding  │      │  /dashboard  │
│              │      │              │
│ Must complete│      │ Direct access│
│ AWS setup    │      │ ✅           │
└──────────────┘      └──────────────┘
```

---

## 3. 👥 Multi-User Tenant Flow (Future)

```
┌─────────────────────────────────────────────────────────────┐
│              TEAM COLLABORATION FLOW                         │
└─────────────────────────────────────────────────────────────┘

Owner Invites User
┌──────────────┐
│    /team     │
│              │
│ Click:       │
│ "Invite User"│
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ Invite Modal │
│              │
│ • Email      │
│ • Role       │
│ • Send       │
└──────┬───────┘
       │
       ▼
Email Sent
┌──────────────┐
│ 📧 Invitation│
│              │
│ Click link   │
│ to accept    │
└──────┬───────┘
       │
       ▼
New User?
┌──────────────────────────────────────┐
│ Has Account?                         │
└──────┬───────────────────────┬───────┘
       │ NO                    │ YES
       ▼                       ▼
┌──────────────┐      ┌──────────────┐
│   Signup     │      │    Login     │
│ + Verify     │      │              │
└──────┬───────┘      └──────┬───────┘
       │                     │
       └──────────┬──────────┘
                  ▼
         ┌──────────────┐
         │ Join Tenant  │
         │              │
         │ Role: Editor │
         └──────┬───────┘
                │
                ▼
         ┌──────────────┐
         │  /dashboard  │
         │              │
         │ Access based │
         │ on role      │
         └──────────────┘
```

---

## 4. 🔐 Role-Based Access Flow

```
┌─────────────────────────────────────────────────────────────┐
│                PERMISSION CHECK FLOW                         │
└─────────────────────────────────────────────────────────────┘

User Action
┌──────────────┐
│ User clicks  │
│ "Invite User"│
└──────┬───────┘
       │
       ▼
Check Role
┌──────────────────────────────────────┐
│ Role = Owner or Admin?               │
└──────┬───────────────────────┬───────┘
       │ YES                   │ NO
       ▼                       ▼
┌──────────────┐      ┌──────────────┐
│ Allow Action │      │ Show Error   │
│              │      │              │
│ Open modal   │      │ "Insufficient│
│ ✅           │      │ permissions" │
└──────────────┘      │ ❌           │
                      └──────────────┘

Permission Matrix:
┌─────────────────────────────────────────────┐
│ Feature        │ Owner │ Admin │ Editor │ V │
├────────────────┼───────┼───────┼────────┼───┤
│ Invite users   │  ✅   │  ✅   │   ❌   │ ❌│
│ Scan AWS       │  ✅   │  ✅   │   ✅   │ ❌│
│ View resources │  ✅   │  ✅   │   ✅   │ ✅│
│ Approve finds  │  ✅   │  ✅   │   ✅   │ ❌│
│ Manage billing │  ✅   │  ❌   │   ❌   │ ❌│
└─────────────────────────────────────────────┘
```

---

## 5. 👨💼 Admin Impersonation Flow

```
┌─────────────────────────────────────────────────────────────┐
│              ADMIN IMPERSONATION FLOW                        │
└─────────────────────────────────────────────────────────────┘

Admin Dashboard
┌──────────────┐
│ /admin       │
│              │
│ View all     │
│ tenants      │
└──────┬───────┘
       │
       ▼
Select Tenant
┌──────────────┐
│ Acme Corp    │
│              │
│ Click:       │
│ "Impersonate"│
└──────┬───────┘
       │
       ▼
Confirmation
┌──────────────┐
│ Confirm      │
│              │
│ • User       │
│ • Reason     │
│ • Confirm    │
└──────┬───────┘
       │
       ▼
Impersonation Active
┌─────────────────────────────────────┐
│ ⚠️  IMPERSONATION MODE              │
│ Viewing as: john@acme.com           │
│ [End Impersonation]                 │
└──────┬──────────────────────────────┘
       │
       ▼
┌──────────────┐
│ Tenant View  │
│              │
│ • Dashboard  │
│ • Resources  │
│ • Settings   │
└──────┬───────┘
       │
       ▼
End Impersonation
┌──────────────┐
│ Click "End"  │
│              │
│ Return to    │
│ admin portal │
└──────┬───────┘
       │
       ▼
Audit Log
┌──────────────┐
│ Log Entry:   │
│              │
│ • Admin ID   │
│ • Target user│
│ • Reason     │
│ • Duration   │
│ • Actions    │
└──────────────┘
```

---

## 6. 🔄 Tenant Switching Flow (Multi-Tenant User)

```
┌─────────────────────────────────────────────────────────────┐
│              TENANT SWITCHING FLOW                           │
└─────────────────────────────────────────────────────────────┘

User Login
┌──────────────┐
│   /login     │
│              │
│ Credentials  │
└──────┬───────┘
       │
       ▼
Multiple Tenants?
┌──────────────────────────────────────┐
│ User belongs to > 1 tenant?          │
└──────┬───────────────────────┬───────┘
       │ YES                   │ NO
       ▼                       ▼
┌──────────────┐      ┌──────────────┐
│ Show Selector│      │  /dashboard  │
│              │      │              │
│ ┌──────────┐│      │ Direct access│
│ │Acme Corp ││      │ ✅           │
│ │(Owner)   ││      └──────────────┘
│ └──────────┘│
│ ┌──────────┐│
│ │TechStart ││
│ │(Editor)  ││
│ └──────────┘│
│ ┌──────────┐│
│ │CloudScale││
│ │(Viewer)  ││
│ └──────────┘│
└──────┬───────┘
       │
       ▼
Select Tenant
┌──────────────┐
│ User selects │
│ "Acme Corp"  │
└──────┬───────┘
       │
       ▼
Load Tenant Context
┌──────────────┐
│ JWT updated  │
│              │
│ • Tenant ID  │
│ • Role       │
│ • Perms      │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│  /dashboard  │
│              │
│ Acme Corp    │
│ context      │
└──────────────┘

Switch Tenant
┌──────────────┐
│ Click tenant │
│ switcher     │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ Select new   │
│ tenant       │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ Reload with  │
│ new context  │
└──────────────┘
```

---

## 7. 📊 Complete User Journey Map

```
┌─────────────────────────────────────────────────────────────┐
│                 COMPLETE USER JOURNEY                        │
└─────────────────────────────────────────────────────────────┘

Day 1: Discovery & Signup
┌──────────┐   ┌──────────┐   ┌──────────┐   ┌──────────┐
│ Landing  │ → │  Signup  │ → │  Verify  │ → │Onboarding│
│   Page   │   │          │   │   Email  │   │          │
└──────────┘   └──────────┘   └──────────┘   └──────────┘

Day 1: First Scan
┌──────────┐   ┌──────────┐   ┌──────────┐   ┌──────────┐
│Dashboard │ → │   Scan   │ → │Resources │ → │ Findings │
│          │   │   AWS    │   │Discovered│   │Generated │
└──────────┘   └──────────┘   └──────────┘   └──────────┘

Week 1: Optimization
┌──────────┐   ┌──────────┐   ┌──────────┐   ┌──────────┐
│  Review  │ → │ Generate │ → │  Apply   │ → │  Track   │
│ Findings │   │   IaC    │   │  Changes │   │ Savings  │
└──────────┘   └──────────┘   └──────────┘   └──────────┘

Month 1: Team Growth
┌──────────┐   ┌──────────┐   ┌──────────┐   ┌──────────┐
│  Invite  │ → │  Users   │ → │  Assign  │ → │Collaborate│
│   Team   │   │   Join   │   │  Roles   │   │          │
└──────────┘   └──────────┘   └──────────┘   └──────────┘

Ongoing: Continuous Optimization
┌──────────┐   ┌──────────┐   ┌──────────┐   ┌──────────┐
│  Weekly  │ → │  Review  │ → │Implement │ → │  Report  │
│  Scans   │   │   New    │   │  Changes │   │ Savings  │
│          │   │ Findings │   │          │   │          │
└──────────┘   └──────────┘   └──────────┘   └──────────┘
```

---

## 8. 🎯 Decision Tree: User Actions

```
User Logs In
     │
     ├─ Has AWS Connection?
     │  ├─ NO → Redirect to /onboarding (MANDATORY)
     │  └─ YES → Continue
     │
     ├─ Multiple Tenants?
     │  ├─ YES → Show tenant selector
     │  └─ NO → Load default tenant
     │
     ├─ Load Dashboard
     │
     └─ User Action
        │
        ├─ Invite User?
        │  ├─ Role = Owner/Admin? → Allow ✅
        │  └─ Role = Editor/Viewer? → Deny ❌
        │
        ├─ Scan AWS?
        │  ├─ Role = Owner/Admin/Editor? → Allow ✅
        │  └─ Role = Viewer? → Deny ❌
        │
        ├─ Approve Finding?
        │  ├─ Role = Owner/Admin/Editor? → Allow ✅
        │  └─ Role = Viewer? → Deny ❌
        │
        └─ View Resources?
           └─ All Roles → Allow ✅
```

---

## 📋 Flow Summary

| Flow | Status | Complexity | Priority |
|------|--------|------------|----------|
| New User Onboarding | ✅ Implemented | Low | Critical |
| Returning User Login | ✅ Implemented | Low | Critical |
| Multi-User Tenant | 📋 Designed | High | High |
| Role-Based Access | 📋 Designed | High | High |
| Admin Impersonation | 📋 Designed | Medium | Medium |
| Tenant Switching | 📋 Designed | Medium | Medium |

---

**All user flows documented with visual diagrams!** 🎨✅
