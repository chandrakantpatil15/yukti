# 🗺️ Implementation Roadmap - Multi-User & Admin Features

## 📊 Overview

**Goal**: Implement complete RBAC system with multi-user tenants and platform admin portal

**Timeline**: 5-6 weeks  
**Complexity**: High  
**Priority**: High (Required for production)

---

## 🎯 Phase 1: Database & Backend Foundation (Week 1)

### Day 1-2: Database Schema
**Tasks**:
- [ ] Create `yt_tenant_users` table
- [ ] Create `yt_user_invitations` table  
- [ ] Create `yt_admin_audit_logs` table
- [ ] Update `yt_users` table (add multi-tenant fields)
- [ ] Create migration scripts
- [ ] Add indexes for performance

**Files to Create**:
- `migrations/008_multi_user_rbac.sql`
- `migrations/009_admin_audit_logs.sql`

**Testing**:
```bash
# Run migrations
psql -U yukti -d yukti_finops -f migrations/008_multi_user_rbac.sql

# Verify tables
psql -U yukti -d yukti_finops -c "\dt yt_*"
```

### Day 3-4: Role-Based Middleware
**Tasks**:
- [ ] Create role check middleware
- [ ] Update JWT to include role
- [ ] Create permission helper functions
- [ ] Add role validation

**Files to Create**:
- `internal/api/middleware/role_auth.go`
- `internal/models/permissions.go`

**Code Example**:
```go
// internal/api/middleware/role_auth.go
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userRole := c.GetString("user_role")
        // Check if user has required role
    }
}
```

### Day 5: Team Management Handlers
**Tasks**:
- [ ] Create team handler
- [ ] Implement invite user endpoint
- [ ] Implement list members endpoint
- [ ] Implement remove user endpoint
- [ ] Implement update role endpoint

**Files to Create**:
- `internal/api/handlers/team.go`
- `internal/services/invitation.go`

---

## 🎯 Phase 2: Invitation System (Week 2)

### Day 1-2: Invitation Backend
**Tasks**:
- [ ] Create invitation service
- [ ] Generate invitation tokens
- [ ] Send invitation emails
- [ ] Handle invitation acceptance
- [ ] Handle invitation expiration

**Files to Create**:
- `internal/services/invitation_service.go`
- `internal/email/templates/invitation.html`

### Day 3-4: Invitation API Endpoints
**Tasks**:
- [ ] POST /api/v1/team/invite
- [ ] POST /api/v1/team/accept-invite
- [ ] DELETE /api/v1/team/invitations/:id
- [ ] POST /api/v1/team/invitations/:id/resend
- [ ] GET /api/v1/team/invitations

**Testing**:
```bash
# Test invitation flow
curl -X POST http://localhost:8081/api/v1/team/invite \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"email":"user@test.com","role":"editor"}'
```

### Day 5: Multi-Tenant User Support
**Tasks**:
- [ ] Update login to support multiple tenants
- [ ] Create tenant selector logic
- [ ] Update JWT to include active tenant
- [ ] Handle tenant switching

---

## 🎯 Phase 3: Frontend - Team Management (Week 3)

### Day 1-2: Team Management Page
**Tasks**:
- [ ] Create `/team` page
- [ ] List active members
- [ ] List pending invitations
- [ ] Show role badges
- [ ] Add search/filter

**Files to Create**:
- `frontend/src/pages/Team.tsx`
- `frontend/src/components/Team/MemberCard.tsx`
- `frontend/src/components/Team/InvitationCard.tsx`

### Day 3: Invite User Modal
**Tasks**:
- [ ] Create invite modal component
- [ ] Email input with validation
- [ ] Role selector dropdown
- [ ] Custom message textarea
- [ ] Send invitation API call

**Files to Create**:
- `frontend/src/components/Team/InviteModal.tsx`

### Day 4: Accept Invitation Flow
**Tasks**:
- [ ] Create `/accept-invite` page
- [ ] Parse invitation token from URL
- [ ] Show invitation details
- [ ] Handle acceptance (new/existing user)
- [ ] Redirect to dashboard

**Files to Create**:
- `frontend/src/pages/AcceptInvite.tsx`

### Day 5: Role Context & Guards
**Tasks**:
- [ ] Create RoleContext provider
- [ ] Create RoleGuard component
- [ ] Update all pages with role checks
- [ ] Hide/show UI based on permissions

**Files to Create**:
- `frontend/src/contexts/RoleContext.tsx`
- `frontend/src/components/Auth/RoleGuard.tsx`

---

## 🎯 Phase 4: Admin Portal Backend (Week 4)

### Day 1-2: Admin Authentication
**Tasks**:
- [ ] Create admin login endpoint
- [ ] Implement 2FA for admins
- [ ] Create admin JWT tokens
- [ ] Admin session management

**Files to Create**:
- `internal/api/handlers/admin_auth.go`
- `internal/services/admin_2fa.go`

### Day 3-4: Admin Tenant Management
**Tasks**:
- [ ] GET /api/admin/tenants (list all)
- [ ] GET /api/admin/tenants/:id (details)
- [ ] POST /api/admin/tenants/:id/suspend
- [ ] POST /api/admin/tenants/:id/activate
- [ ] DELETE /api/admin/tenants/:id

**Files to Create**:
- `internal/api/handlers/admin_tenants.go`

### Day 5: Impersonation Feature
**Tasks**:
- [ ] POST /api/admin/tenants/:id/impersonate
- [ ] Create impersonation JWT
- [ ] Log impersonation to audit trail
- [ ] POST /api/admin/end-impersonation

**Files to Create**:
- `internal/services/impersonation.go`
- `internal/audit/impersonation_logger.go`

---

## 🎯 Phase 5: Admin Portal Frontend (Week 5)

### Day 1-2: Admin Dashboard
**Tasks**:
- [ ] Create `/admin` route
- [ ] Platform overview cards
- [ ] Recent activity feed
- [ ] Platform health indicators
- [ ] Quick actions

**Files to Create**:
- `frontend/src/pages/Admin/Dashboard.tsx`
- `frontend/src/components/Admin/PlatformStats.tsx`

### Day 3: Tenant Management UI
**Tasks**:
- [ ] Tenant list view
- [ ] Tenant detail view
- [ ] Suspend/activate actions
- [ ] Impersonate button
- [ ] Search and filters

**Files to Create**:
- `frontend/src/pages/Admin/Tenants.tsx`
- `frontend/src/pages/Admin/TenantDetail.tsx`
- `frontend/src/components/Admin/TenantCard.tsx`

### Day 4: User Management UI
**Tasks**:
- [ ] All users list
- [ ] User detail view
- [ ] Suspend/activate users
- [ ] Reset password
- [ ] Filters by tenant/role

**Files to Create**:
- `frontend/src/pages/Admin/Users.tsx`
- `frontend/src/components/Admin/UserCard.tsx`

### Day 5: Impersonation UI
**Tasks**:
- [ ] Impersonation confirmation modal
- [ ] Impersonation banner
- [ ] End impersonation button
- [ ] Audit log display

**Files to Create**:
- `frontend/src/components/Admin/ImpersonationBanner.tsx`
- `frontend/src/components/Admin/ImpersonationModal.tsx`

---

## 🎯 Phase 6: Testing & Polish (Week 6)

### Day 1-2: Backend Testing
**Tasks**:
- [ ] Unit tests for role middleware
- [ ] Integration tests for team APIs
- [ ] Test invitation flow
- [ ] Test impersonation
- [ ] Test permission boundaries

**Files to Create**:
- `internal/api/handlers/team_test.go`
- `internal/services/invitation_test.go`

### Day 3: Frontend Testing
**Tasks**:
- [ ] E2E test: Invite user flow
- [ ] E2E test: Accept invitation
- [ ] E2E test: Role permissions
- [ ] E2E test: Tenant switching
- [ ] E2E test: Admin impersonation

**Files to Create**:
- `frontend/src/__tests__/team.test.tsx`
- `frontend/src/__tests__/admin.test.tsx`

### Day 4: Documentation
**Tasks**:
- [ ] API documentation for team endpoints
- [ ] API documentation for admin endpoints
- [ ] User guide for team management
- [ ] Admin guide for platform management

**Files to Create**:
- `docs/API_TEAM_MANAGEMENT.md`
- `docs/API_ADMIN_PORTAL.md`
- `docs/USER_GUIDE_TEAMS.md`
- `docs/ADMIN_GUIDE.md`

### Day 5: Final Polish
**Tasks**:
- [ ] UI/UX improvements
- [ ] Performance optimization
- [ ] Security audit
- [ ] Bug fixes
- [ ] Deployment preparation

---

## 📋 Complete Checklist

### Database (10 tasks)
- [ ] Create yt_tenant_users table
- [ ] Create yt_user_invitations table
- [ ] Create yt_admin_audit_logs table
- [ ] Update yt_users table
- [ ] Add indexes
- [ ] Create migration scripts
- [ ] Test migrations
- [ ] Seed test data
- [ ] Backup strategy
- [ ] Performance tuning

### Backend (25 tasks)
- [ ] Role-based middleware
- [ ] Permission helper functions
- [ ] Team handler (invite/remove/update)
- [ ] Invitation service
- [ ] Invitation email templates
- [ ] Accept invitation endpoint
- [ ] Resend invitation endpoint
- [ ] Revoke invitation endpoint
- [ ] List team members endpoint
- [ ] Multi-tenant login support
- [ ] Tenant switching logic
- [ ] Admin authentication
- [ ] Admin 2FA
- [ ] Admin tenant management
- [ ] Admin user management
- [ ] Impersonation service
- [ ] Impersonation audit logging
- [ ] Admin analytics endpoints
- [ ] Admin audit log endpoints
- [ ] Rate limiting
- [ ] Input validation
- [ ] Error handling
- [ ] Logging
- [ ] Unit tests
- [ ] Integration tests

### Frontend (20 tasks)
- [ ] Team management page
- [ ] Member card component
- [ ] Invitation card component
- [ ] Invite modal component
- [ ] Accept invitation page
- [ ] Role context provider
- [ ] Role guard component
- [ ] Permission-based UI rendering
- [ ] Tenant selector component
- [ ] Admin dashboard
- [ ] Admin tenant list
- [ ] Admin tenant detail
- [ ] Admin user list
- [ ] Impersonation modal
- [ ] Impersonation banner
- [ ] Admin analytics page
- [ ] Admin audit logs page
- [ ] Role badge component
- [ ] E2E tests
- [ ] UI polish

### Documentation (8 tasks)
- [ ] RBAC design document ✅
- [ ] Admin flow design ✅
- [ ] API documentation (team)
- [ ] API documentation (admin)
- [ ] User guide (teams)
- [ ] Admin guide
- [ ] Migration guide
- [ ] Security best practices

---

## 🚀 Quick Start Guide

### For Developers

**Week 1 - Start Here**:
```bash
# 1. Create database tables
psql -U yukti -d yukti_finops -f migrations/008_multi_user_rbac.sql

# 2. Create role middleware
touch internal/api/middleware/role_auth.go

# 3. Create team handler
touch internal/api/handlers/team.go

# 4. Run tests
go test ./internal/api/handlers/...
```

**Week 2 - Invitation System**:
```bash
# 1. Create invitation service
touch internal/services/invitation_service.go

# 2. Create email template
touch internal/email/templates/invitation.html

# 3. Test invitation flow
curl -X POST http://localhost:8081/api/v1/team/invite \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"email":"test@test.com","role":"editor"}'
```

**Week 3 - Frontend**:
```bash
# 1. Create team page
touch frontend/src/pages/Team.tsx

# 2. Create invite modal
touch frontend/src/components/Team/InviteModal.tsx

# 3. Test in browser
npm run dev
# Navigate to http://localhost:3000/team
```

---

## 📊 Success Metrics

### Technical Metrics
- [ ] All API endpoints return < 200ms
- [ ] Database queries optimized (< 50ms)
- [ ] Zero security vulnerabilities
- [ ] 90%+ test coverage
- [ ] Zero breaking changes

### User Metrics
- [ ] Users can invite team members
- [ ] Invitations delivered within 1 minute
- [ ] Role permissions enforced correctly
- [ ] Admin can impersonate users
- [ ] All actions logged to audit trail

---

## 🎉 Completion Criteria

**Phase 1-2 Complete When**:
- ✅ Database schema created
- ✅ Role middleware working
- ✅ Invitation system functional
- ✅ All backend tests passing

**Phase 3 Complete When**:
- ✅ Team management page working
- ✅ Users can invite/remove members
- ✅ Role-based UI rendering works
- ✅ All frontend tests passing

**Phase 4-5 Complete When**:
- ✅ Admin portal accessible
- ✅ Admins can manage all tenants
- ✅ Impersonation working
- ✅ Audit logs complete

**Phase 6 Complete When**:
- ✅ All tests passing (90%+ coverage)
- ✅ Documentation complete
- ✅ Security audit passed
- ✅ Ready for production deployment

---

**Total Estimated Effort**: 5-6 weeks (1 developer full-time)  
**Priority**: High  
**Risk**: Medium (complex feature, requires careful testing)  
**Dependencies**: Current platform must be stable

**Ready to start implementation!** 🚀
