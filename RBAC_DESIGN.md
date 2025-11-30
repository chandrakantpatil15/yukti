# 🔐 Role-Based Access Control (RBAC) Design

## Overview
Multi-user tenant system where tenant owners can invite users with different permission levels.

---

## 👥 User Roles

### 1. **Tenant Owner** (Creator)
- First user who signs up
- Full control over tenant
- Cannot be deleted
- Can manage all users

**Permissions**:
- ✅ Invite/remove users
- ✅ Manage AWS connections
- ✅ View/edit all resources
- ✅ Approve/reject findings
- ✅ Generate IaC
- ✅ Manage budgets
- ✅ View audit logs
- ✅ Billing & subscription

### 2. **Admin**
- Full access except billing
- Can manage other users (except owner)

**Permissions**:
- ✅ Invite/remove users (except owner)
- ✅ Manage AWS connections
- ✅ View/edit all resources
- ✅ Approve/reject findings
- ✅ Generate IaC
- ✅ Manage budgets
- ✅ View audit logs
- ❌ Billing & subscription

### 3. **Editor**
- Can view and take actions
- Cannot manage users or settings

**Permissions**:
- ❌ Invite/remove users
- ❌ Manage AWS connections
- ✅ View all resources
- ✅ Approve/reject findings
- ✅ Generate IaC
- ✅ Whitelist resources
- ❌ Manage budgets
- ❌ View audit logs

### 4. **Viewer** (Read-Only)
- View-only access
- Cannot make changes

**Permissions**:
- ❌ Invite/remove users
- ❌ Manage AWS connections
- ✅ View all resources
- ✅ View findings
- ❌ Approve/reject findings
- ❌ Generate IaC
- ❌ Whitelist resources
- ❌ Manage budgets
- ❌ View audit logs

---

## 📊 Permission Matrix

| Feature | Owner | Admin | Editor | Viewer |
|---------|-------|-------|--------|--------|
| **User Management** |
| Invite users | ✅ | ✅ | ❌ | ❌ |
| Remove users | ✅ | ✅* | ❌ | ❌ |
| Change user roles | ✅ | ✅* | ❌ | ❌ |
| **AWS Integration** |
| Add AWS account | ✅ | ✅ | ❌ | ❌ |
| Remove AWS account | ✅ | ✅ | ❌ | ❌ |
| Trigger scan | ✅ | ✅ | ✅ | ❌ |
| **Resources** |
| View resources | ✅ | ✅ | ✅ | ✅ |
| View metrics | ✅ | ✅ | ✅ | ✅ |
| **Cost Optimization** |
| View findings | ✅ | ✅ | ✅ | ✅ |
| Approve findings | ✅ | ✅ | ✅ | ❌ |
| Reject findings | ✅ | ✅ | ✅ | ❌ |
| Generate IaC | ✅ | ✅ | ✅ | ❌ |
| **Whitelisting** |
| Create whitelist | ✅ | ✅ | ✅ | ❌ |
| Remove whitelist | ✅ | ✅ | ✅ | ❌ |
| **Budgets** |
| View budgets | ✅ | ✅ | ✅ | ✅ |
| Create/edit budgets | ✅ | ✅ | ❌ | ❌ |
| **Audit & Security** |
| View audit logs | ✅ | ✅ | ❌ | ❌ |
| **Billing** |
| View billing | ✅ | ❌ | ❌ | ❌ |
| Manage subscription | ✅ | ❌ | ❌ | ❌ |

*Admin cannot remove/modify Owner

---

## 🔄 User Invitation Flow

### Step 1: Owner/Admin Invites User
```
1. Navigate to Team Settings
2. Click "Invite User"
3. Enter:
   - Email address
   - Role (Admin/Editor/Viewer)
   - Optional: Custom message
4. Click "Send Invitation"
```

**Backend**:
```sql
INSERT INTO yt_user_invitations (
  tenant_id,
  email,
  role,
  invited_by,
  invitation_token,
  expires_at,
  status
) VALUES (
  $tenant_id,
  $email,
  $role,
  $inviter_user_id,
  $random_token,
  NOW() + INTERVAL '7 days',
  'pending'
);
```

**Email Sent**:
```
Subject: You've been invited to join [Company Name] on Yukti

Hi,

[Inviter Name] has invited you to join their team on Yukti 
as a [Role].

Click here to accept: https://yukti.com/accept-invite?token=xxx

This invitation expires in 7 days.
```

### Step 2: User Accepts Invitation
```
1. Click invitation link
2. If existing user:
   - Login → Join tenant
3. If new user:
   - Signup → Verify email → Join tenant
4. Redirect to dashboard
```

**Backend**:
```sql
-- Verify token
SELECT * FROM yt_user_invitations 
WHERE invitation_token = $token 
AND status = 'pending' 
AND expires_at > NOW();

-- Create user-tenant relationship
INSERT INTO yt_tenant_users (
  tenant_id,
  user_id,
  role,
  joined_at
) VALUES ($tenant_id, $user_id, $role, NOW());

-- Update invitation status
UPDATE yt_user_invitations 
SET status = 'accepted', accepted_at = NOW()
WHERE id = $invitation_id;
```

### Step 3: User Accesses Tenant
```
1. Login
2. If user belongs to multiple tenants:
   - Show tenant selector
3. Select tenant
4. Dashboard loads with tenant context
```

---

## 🗄️ Database Schema

### New Tables

#### 1. `yt_tenant_users` (User-Tenant Relationship)
```sql
CREATE TABLE yt_tenant_users (
  id SERIAL PRIMARY KEY,
  tenant_id INTEGER NOT NULL REFERENCES yt_customers(id),
  user_id UUID NOT NULL REFERENCES yt_users(id),
  role VARCHAR(20) NOT NULL CHECK (role IN ('owner', 'admin', 'editor', 'viewer')),
  joined_at TIMESTAMP DEFAULT NOW(),
  created_at TIMESTAMP DEFAULT NOW(),
  UNIQUE(tenant_id, user_id)
);

CREATE INDEX idx_tenant_users_tenant ON yt_tenant_users(tenant_id);
CREATE INDEX idx_tenant_users_user ON yt_tenant_users(user_id);
```

#### 2. `yt_user_invitations` (Pending Invitations)
```sql
CREATE TABLE yt_user_invitations (
  id SERIAL PRIMARY KEY,
  tenant_id INTEGER NOT NULL REFERENCES yt_customers(id),
  email VARCHAR(255) NOT NULL,
  role VARCHAR(20) NOT NULL CHECK (role IN ('admin', 'editor', 'viewer')),
  invited_by UUID NOT NULL REFERENCES yt_users(id),
  invitation_token VARCHAR(255) UNIQUE NOT NULL,
  custom_message TEXT,
  status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'expired', 'revoked')),
  expires_at TIMESTAMP NOT NULL,
  accepted_at TIMESTAMP,
  created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_invitations_token ON yt_user_invitations(invitation_token);
CREATE INDEX idx_invitations_email ON yt_user_invitations(email);
```

#### 3. Update `yt_users` Table
```sql
-- Add fields for multi-tenant support
ALTER TABLE yt_users ADD COLUMN default_tenant_id INTEGER REFERENCES yt_customers(id);
ALTER TABLE yt_users ADD COLUMN last_active_tenant_id INTEGER REFERENCES yt_customers(id);
```

---

## 🎨 UI Components

### 1. Team Management Page (`/team`)

**Layout**:
```
┌─────────────────────────────────────────┐
│ Team Members                    [Invite]│
├─────────────────────────────────────────┤
│ Active Members (5)                      │
│                                         │
│ ┌─────────────────────────────────────┐│
│ │ 👤 John Doe (You)          [Owner] ││
│ │    john@company.com                 ││
│ │    Joined: Jan 1, 2024              ││
│ └─────────────────────────────────────┘│
│                                         │
│ ┌─────────────────────────────────────┐│
│ │ 👤 Jane Smith            [Admin] ▼ ││
│ │    jane@company.com                 ││
│ │    Joined: Jan 5, 2024    [Remove] ││
│ └─────────────────────────────────────┘│
│                                         │
│ Pending Invitations (2)                 │
│                                         │
│ ┌─────────────────────────────────────┐│
│ │ 📧 bob@company.com       [Editor]   ││
│ │    Invited: Jan 10, 2024            ││
│ │    Expires: Jan 17, 2024  [Resend] ││
│ │                          [Revoke]  ││
│ └─────────────────────────────────────┘│
└─────────────────────────────────────────┘
```

### 2. Invite User Modal

```
┌─────────────────────────────────────┐
│ Invite Team Member            [×]   │
├─────────────────────────────────────┤
│                                     │
│ Email Address *                     │
│ ┌─────────────────────────────────┐ │
│ │ user@company.com                │ │
│ └─────────────────────────────────┘ │
│                                     │
│ Role *                              │
│ ┌─────────────────────────────────┐ │
│ │ Editor                      ▼   │ │
│ └─────────────────────────────────┘ │
│                                     │
│ ○ Admin - Full access except billing│
│ ● Editor - Can view and take actions│
│ ○ Viewer - Read-only access         │
│                                     │
│ Custom Message (Optional)           │
│ ┌─────────────────────────────────┐ │
│ │ Welcome to the team!            │ │
│ │                                 │ │
│ └─────────────────────────────────┘ │
│                                     │
│        [Cancel]  [Send Invitation]  │
└─────────────────────────────────────┘
```

### 3. Tenant Selector (Multi-Tenant Users)

```
┌─────────────────────────────────────┐
│ Select Workspace                    │
├─────────────────────────────────────┤
│                                     │
│ ┌─────────────────────────────────┐ │
│ │ 🏢 Acme Corp            [Owner] │ │
│ │    5 members • 12 resources     │ │
│ └─────────────────────────────────┘ │
│                                     │
│ ┌─────────────────────────────────┐ │
│ │ 🏢 TechStart Inc       [Editor] │ │
│ │    3 members • 8 resources      │ │
│ └─────────────────────────────────┘ │
│                                     │
│ ┌─────────────────────────────────┐ │
│ │ 🏢 CloudScale LLC     [Viewer]  │ │
│ │    10 members • 45 resources    │ │
│ └─────────────────────────────────┘ │
│                                     │
└─────────────────────────────────────┘
```

### 4. Role Badge Component

```tsx
// Owner badge
<span className="px-2 py-1 bg-purple-100 text-purple-800 rounded-full text-xs font-semibold">
  👑 Owner
</span>

// Admin badge
<span className="px-2 py-1 bg-blue-100 text-blue-800 rounded-full text-xs font-semibold">
  🔧 Admin
</span>

// Editor badge
<span className="px-2 py-1 bg-green-100 text-green-800 rounded-full text-xs font-semibold">
  ✏️ Editor
</span>

// Viewer badge
<span className="px-2 py-1 bg-gray-100 text-gray-800 rounded-full text-xs font-semibold">
  👁️ Viewer
</span>
```

---

## 🔌 API Endpoints

### User Management

#### 1. List Team Members
```
GET /api/v1/team/members
Authorization: Bearer {jwt_token}

Response:
{
  "success": true,
  "data": {
    "members": [
      {
        "user_id": "uuid",
        "email": "john@company.com",
        "name": "John Doe",
        "role": "owner",
        "joined_at": "2024-01-01T00:00:00Z",
        "last_active": "2024-01-15T10:30:00Z"
      }
    ],
    "pending_invitations": [
      {
        "id": 123,
        "email": "bob@company.com",
        "role": "editor",
        "invited_by": "John Doe",
        "invited_at": "2024-01-10T00:00:00Z",
        "expires_at": "2024-01-17T00:00:00Z"
      }
    ]
  }
}
```

#### 2. Invite User
```
POST /api/v1/team/invite
Authorization: Bearer {jwt_token}
Content-Type: application/json

Request:
{
  "email": "user@company.com",
  "role": "editor",
  "custom_message": "Welcome to the team!"
}

Response:
{
  "success": true,
  "message": "Invitation sent successfully",
  "data": {
    "invitation_id": 123,
    "expires_at": "2024-01-17T00:00:00Z"
  }
}
```

#### 3. Accept Invitation
```
POST /api/v1/team/accept-invite
Content-Type: application/json

Request:
{
  "token": "invitation_token_here"
}

Response:
{
  "success": true,
  "message": "Successfully joined team",
  "data": {
    "tenant_id": 1,
    "tenant_name": "Acme Corp",
    "role": "editor"
  }
}
```

#### 4. Remove User
```
DELETE /api/v1/team/members/{user_id}
Authorization: Bearer {jwt_token}

Response:
{
  "success": true,
  "message": "User removed successfully"
}
```

#### 5. Update User Role
```
PATCH /api/v1/team/members/{user_id}/role
Authorization: Bearer {jwt_token}
Content-Type: application/json

Request:
{
  "role": "admin"
}

Response:
{
  "success": true,
  "message": "User role updated successfully"
}
```

#### 6. Revoke Invitation
```
DELETE /api/v1/team/invitations/{invitation_id}
Authorization: Bearer {jwt_token}

Response:
{
  "success": true,
  "message": "Invitation revoked successfully"
}
```

#### 7. Resend Invitation
```
POST /api/v1/team/invitations/{invitation_id}/resend
Authorization: Bearer {jwt_token}

Response:
{
  "success": true,
  "message": "Invitation resent successfully"
}
```

---

## 🔐 Middleware & Authorization

### Role Check Middleware
```go
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userRole := c.GetString("user_role")
        
        allowed := false
        for _, role := range allowedRoles {
            if userRole == role {
                allowed = true
                break
            }
        }
        
        if !allowed {
            c.JSON(403, gin.H{
                "success": false,
                "error": "Insufficient permissions"
            })
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

### Usage in Routes
```go
// Only owner and admin can invite users
router.POST("/team/invite", 
    jwtAuth.RequireAuth, 
    RequireRole("owner", "admin"), 
    teamHandler.InviteUser)

// Only owner and admin can manage budgets
router.POST("/budgets", 
    jwtAuth.RequireAuth, 
    RequireRole("owner", "admin"), 
    budgetHandler.CreateBudget)

// Editor and above can approve findings
router.POST("/findings/:id/approve", 
    jwtAuth.RequireAuth, 
    RequireRole("owner", "admin", "editor"), 
    findingsHandler.ApproveFinding)

// All authenticated users can view resources
router.GET("/resources", 
    jwtAuth.RequireAuth, 
    resourceHandler.ListResources)
```

---

## 📱 Frontend Implementation

### 1. Role Context
```tsx
// contexts/RoleContext.tsx
interface RoleContextType {
  role: 'owner' | 'admin' | 'editor' | 'viewer';
  canInviteUsers: boolean;
  canManageAWS: boolean;
  canApproveFindings: boolean;
  canManageBudgets: boolean;
  canViewAuditLogs: boolean;
}

export const useRole = () => {
  const context = useContext(RoleContext);
  return context;
};
```

### 2. Permission-Based UI
```tsx
// Example: Conditional rendering based on role
const Dashboard = () => {
  const { canManageAWS, canInviteUsers } = useRole();
  
  return (
    <div>
      {canManageAWS && (
        <button onClick={triggerScan}>Scan Resources</button>
      )}
      
      {canInviteUsers && (
        <button onClick={openInviteModal}>Invite User</button>
      )}
    </div>
  );
};
```

### 3. Route Protection
```tsx
// components/Auth/RoleGuard.tsx
interface RoleGuardProps {
  allowedRoles: string[];
  children: React.ReactNode;
  fallback?: React.ReactNode;
}

export const RoleGuard: React.FC<RoleGuardProps> = ({ 
  allowedRoles, 
  children, 
  fallback 
}) => {
  const { role } = useRole();
  
  if (!allowedRoles.includes(role)) {
    return fallback || <Navigate to="/403" />;
  }
  
  return <>{children}</>;
};
```

---

## 🧪 Testing Scenarios

### Scenario 1: Owner Invites Admin
1. Login as owner
2. Navigate to Team page
3. Click "Invite User"
4. Enter email + select "Admin" role
5. Send invitation
6. Check email for invitation link
7. Click link (as new user)
8. Signup + verify email
9. Join tenant as Admin
10. Verify admin permissions

### Scenario 2: Editor Tries to Invite User
1. Login as editor
2. Navigate to Team page
3. "Invite User" button should be hidden/disabled
4. Try direct API call
5. Should receive 403 Forbidden

### Scenario 3: Multi-Tenant User
1. User belongs to 3 tenants
2. Login
3. See tenant selector
4. Select Tenant A (Owner role)
5. Full access to all features
6. Switch to Tenant B (Viewer role)
7. Read-only access only

---

## 📋 Implementation Checklist

### Phase 1: Database & Backend
- [ ] Create `yt_tenant_users` table
- [ ] Create `yt_user_invitations` table
- [ ] Update `yt_users` table
- [ ] Create team management handlers
- [ ] Implement role-based middleware
- [ ] Create invitation email templates
- [ ] Add audit logging for user actions

### Phase 2: API Endpoints
- [ ] POST /api/v1/team/invite
- [ ] GET /api/v1/team/members
- [ ] DELETE /api/v1/team/members/:id
- [ ] PATCH /api/v1/team/members/:id/role
- [ ] POST /api/v1/team/accept-invite
- [ ] DELETE /api/v1/team/invitations/:id
- [ ] POST /api/v1/team/invitations/:id/resend

### Phase 3: Frontend Components
- [ ] Team Management page
- [ ] Invite User modal
- [ ] Tenant Selector component
- [ ] Role Badge component
- [ ] RoleContext provider
- [ ] RoleGuard component
- [ ] Permission-based UI rendering

### Phase 4: Testing
- [ ] Unit tests for role middleware
- [ ] Integration tests for team APIs
- [ ] E2E tests for invitation flow
- [ ] E2E tests for multi-tenant switching
- [ ] Permission boundary tests

---

**Ready for implementation!** 🚀
