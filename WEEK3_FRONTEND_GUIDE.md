# Week 3: Frontend Team Management - Implementation Guide

## Overview
Week 3 focuses on implementing the frontend UI for team management, user invitations, and tenant switching.

## Timeline: 5 Days

### Day 1-2: Team Management Page
### Day 3: Invite User Modal
### Day 4: Accept Invitation Page
### Day 5: Role Context & Guards

---

## Day 1-2: Team Management Page

### Create Team Page Component
**File**: `frontend/src/pages/Team.tsx`

**Features:**
- List active team members
- List pending invitations
- Show role badges (Owner, Admin, Editor, Viewer)
- Search/filter members
- Actions: Update role, Remove user, Resend/Revoke invitation

**API Calls:**
```typescript
// Get team members
const { data: members } = useQuery('team-members', () => 
  api.get('/api/v1/team/members')
);

// Get pending invitations
const { data: invitations } = useQuery('team-invitations', () => 
  api.get('/api/v1/team/invitations')
);
```

**UI Structure:**
```tsx
<div className="team-page">
  <header>
    <h1>Team Management</h1>
    <button onClick={openInviteModal}>Invite User</button>
  </header>
  
  <section className="active-members">
    <h2>Team Members ({members.length})</h2>
    {members.map(member => (
      <MemberCard 
        key={member.user_id}
        member={member}
        onUpdateRole={handleUpdateRole}
        onRemove={handleRemove}
      />
    ))}
  </section>
  
  <section className="pending-invitations">
    <h2>Pending Invitations ({invitations.length})</h2>
    {invitations.map(invite => (
      <InvitationCard
        key={invite.id}
        invitation={invite}
        onResend={handleResend}
        onRevoke={handleRevoke}
      />
    ))}
  </section>
</div>
```

### Create Member Card Component
**File**: `frontend/src/components/Team/MemberCard.tsx`

**Props:**
```typescript
interface MemberCardProps {
  member: {
    user_id: string;
    email: string;
    first_name?: string;
    last_name?: string;
    role: string;
    is_active: boolean;
    joined_at: string;
  };
  onUpdateRole: (userId: string, newRole: string) => void;
  onRemove: (userId: string) => void;
}
```

**UI:**
```tsx
<div className="member-card">
  <div className="member-info">
    <Avatar email={member.email} />
    <div>
      <h3>{member.first_name} {member.last_name}</h3>
      <p>{member.email}</p>
    </div>
  </div>
  
  <RoleBadge role={member.role} />
  
  <div className="member-actions">
    {canManageUser(currentRole, member.role) && (
      <>
        <button onClick={() => onUpdateRole(member.user_id)}>
          Change Role
        </button>
        {member.role !== 'owner' && (
          <button onClick={() => onRemove(member.user_id)}>
            Remove
          </button>
        )}
      </>
    )}
  </div>
</div>
```

### Create Invitation Card Component
**File**: `frontend/src/components/Team/InvitationCard.tsx`

**UI:**
```tsx
<div className="invitation-card">
  <div className="invite-info">
    <Mail className="icon" />
    <div>
      <h3>{invitation.email}</h3>
      <p>Invited as {invitation.role}</p>
      <p className="expires">Expires: {formatDate(invitation.expires_at)}</p>
    </div>
  </div>
  
  <div className="invite-actions">
    <button onClick={() => onResend(invitation.id)}>
      Resend
    </button>
    <button onClick={() => onRevoke(invitation.id)}>
      Revoke
    </button>
  </div>
</div>
```

---

## Day 3: Invite User Modal

### Create Invite Modal Component
**File**: `frontend/src/components/Team/InviteModal.tsx`

**Features:**
- Email input with validation
- Role selector dropdown
- Optional custom message
- Send invitation

**UI:**
```tsx
<Modal isOpen={isOpen} onClose={onClose}>
  <h2>Invite Team Member</h2>
  
  <form onSubmit={handleSubmit}>
    <div className="form-group">
      <label>Email Address</label>
      <input
        type="email"
        value={email}
        onChange={(e) => setEmail(e.target.value)}
        placeholder="user@example.com"
        required
      />
    </div>
    
    <div className="form-group">
      <label>Role</label>
      <select value={role} onChange={(e) => setRole(e.target.value)}>
        <option value="viewer">Viewer - Read-only access</option>
        <option value="editor">Editor - Can view and take actions</option>
        <option value="admin">Admin - Full access except billing</option>
      </select>
    </div>
    
    <div className="form-actions">
      <button type="button" onClick={onClose}>Cancel</button>
      <button type="submit" disabled={loading}>
        {loading ? 'Sending...' : 'Send Invitation'}
      </button>
    </div>
  </form>
</Modal>
```

**API Call:**
```typescript
const inviteUser = useMutation(
  (data: { email: string; role: string }) =>
    api.post('/api/v1/team/invite', data),
  {
    onSuccess: () => {
      queryClient.invalidateQueries('team-invitations');
      toast.success('Invitation sent!');
      onClose();
    },
    onError: (error) => {
      toast.error(error.message);
    }
  }
);
```

---

## Day 4: Accept Invitation Page

### Create Accept Invite Page
**File**: `frontend/src/pages/AcceptInvite.tsx`

**Features:**
- Parse token from URL query parameter
- Fetch invitation details (public endpoint)
- Show invitation info (tenant name, role)
- Accept button (requires login)
- Handle expired/invalid invitations

**Flow:**
```
1. User clicks email link → /accept-invite?token=abc123
2. Frontend fetches invitation details (public API)
3. If not logged in → Redirect to login with return URL
4. If logged in → Show accept button
5. User clicks accept → POST /api/v1/team/accept-invite
6. Success → Redirect to dashboard with new tenant
```

**UI:**
```tsx
<div className="accept-invite-page">
  {loading && <Spinner />}
  
  {error && (
    <div className="error-card">
      <AlertCircle />
      <h2>Invalid Invitation</h2>
      <p>{error}</p>
    </div>
  )}
  
  {invitation && (
    <div className="invitation-card">
      <CheckCircle className="success-icon" />
      <h1>You've been invited!</h1>
      <p>
        You've been invited to join <strong>{invitation.tenant_name}</strong>
        as a <strong>{invitation.role}</strong>.
      </p>
      
      {!isLoggedIn ? (
        <div>
          <p>Please log in to accept this invitation.</p>
          <button onClick={() => navigate('/login')}>
            Log In
          </button>
        </div>
      ) : (
        <button onClick={handleAccept} disabled={accepting}>
          {accepting ? 'Accepting...' : 'Accept Invitation'}
        </button>
      )}
    </div>
  )}
</div>
```

**API Calls:**
```typescript
// Get invitation details (public)
const { data: invitation } = useQuery(
  ['invitation', token],
  () => api.get(`/api/v1/team/invite-details?token=${token}`),
  { enabled: !!token }
);

// Accept invitation (requires auth)
const acceptInvite = useMutation(
  () => api.post('/api/v1/team/accept-invite', { token }),
  {
    onSuccess: (data) => {
      toast.success(`Welcome to ${data.tenant_name}!`);
      navigate('/dashboard');
    }
  }
);
```

---

## Day 5: Role Context & Guards

### Create Role Context
**File**: `frontend/src/contexts/RoleContext.tsx`

**Purpose:**
- Store current user's role
- Provide permission checking functions
- Update on tenant switch

**Implementation:**
```typescript
interface RoleContextType {
  role: string;
  hasPermission: (permission: string) => boolean;
  canManageUser: (targetRole: string) => boolean;
}

const RoleContext = createContext<RoleContextType | null>(null);

export const RoleProvider: React.FC = ({ children }) => {
  const [role, setRole] = useState<string>('viewer');
  
  // Get current user on mount
  useEffect(() => {
    api.get('/api/auth/current-user').then(res => {
      setRole(res.data.user.role);
    });
  }, []);
  
  const hasPermission = (permission: string) => {
    const permissions = ROLE_PERMISSIONS[role] || [];
    return permissions.includes(permission);
  };
  
  const canManageUser = (targetRole: string) => {
    if (role === 'owner' && targetRole !== 'owner') return true;
    if (role === 'admin' && ['editor', 'viewer'].includes(targetRole)) return true;
    return false;
  };
  
  return (
    <RoleContext.Provider value={{ role, hasPermission, canManageUser }}>
      {children}
    </RoleContext.Provider>
  );
};

export const useRole = () => {
  const context = useContext(RoleContext);
  if (!context) throw new Error('useRole must be used within RoleProvider');
  return context;
};
```

### Create Role Guard Component
**File**: `frontend/src/components/Auth/RoleGuard.tsx`

**Purpose:**
- Conditionally render based on role
- Hide UI elements user can't access

**Implementation:**
```typescript
interface RoleGuardProps {
  allowedRoles: string[];
  children: React.ReactNode;
  fallback?: React.ReactNode;
}

export const RoleGuard: React.FC<RoleGuardProps> = ({
  allowedRoles,
  children,
  fallback = null
}) => {
  const { role } = useRole();
  
  if (allowedRoles.includes(role)) {
    return <>{children}</>;
  }
  
  return <>{fallback}</>;
};
```

**Usage:**
```tsx
<RoleGuard allowedRoles={['owner', 'admin']}>
  <button onClick={handleInviteUser}>Invite User</button>
</RoleGuard>

<RoleGuard allowedRoles={['owner', 'admin', 'editor']}>
  <button onClick={handleUpdateFinding}>Update Finding</button>
</RoleGuard>
```

### Create Tenant Selector Component
**File**: `frontend/src/components/TenantSelector.tsx`

**Features:**
- Show current tenant
- Dropdown with all user's tenants
- Switch tenant on selection

**UI:**
```tsx
<div className="tenant-selector">
  <button onClick={() => setIsOpen(!isOpen)}>
    <Building className="icon" />
    <span>{currentTenant?.tenant_name}</span>
    <ChevronDown />
  </button>
  
  {isOpen && (
    <div className="tenant-dropdown">
      {tenants.map(tenant => (
        <div
          key={tenant.tenant_id}
          className={`tenant-option ${tenant.tenant_id === currentTenant?.tenant_id ? 'active' : ''}`}
          onClick={() => handleSwitch(tenant.tenant_id)}
        >
          <div>
            <strong>{tenant.tenant_name}</strong>
            <span className="role-badge">{tenant.role}</span>
          </div>
          {tenant.tenant_id === currentTenant?.tenant_id && (
            <Check className="check-icon" />
          )}
        </div>
      ))}
    </div>
  )}
</div>
```

**API Call:**
```typescript
const switchTenant = useMutation(
  (tenantId: string) =>
    api.post('/api/auth/switch-tenant', { tenant_id: tenantId }),
  {
    onSuccess: (data) => {
      // Update localStorage
      localStorage.setItem('token', data.token);
      localStorage.setItem('refresh_token', data.refresh_token);
      
      // Reload page to refresh all data
      window.location.reload();
    }
  }
);
```

---

## Permission Matrix (Frontend)

```typescript
const ROLE_PERMISSIONS = {
  owner: [
    'view_aws', 'manage_aws', 'scan_resources',
    'view_findings', 'manage_findings',
    'view_whitelists', 'manage_whitelists',
    'view_budgets', 'manage_budgets',
    'generate_iac',
    'view_team', 'manage_team',
    'view_billing', 'manage_billing'
  ],
  admin: [
    'view_aws', 'manage_aws', 'scan_resources',
    'view_findings', 'manage_findings',
    'view_whitelists', 'manage_whitelists',
    'view_budgets', 'manage_budgets',
    'generate_iac',
    'view_team', 'manage_team',
    'view_billing' // Can view but not manage
  ],
  editor: [
    'view_aws', 'scan_resources',
    'view_findings', 'manage_findings',
    'view_whitelists', 'manage_whitelists',
    'view_budgets',
    'generate_iac',
    'view_team'
  ],
  viewer: [
    'view_aws',
    'view_findings',
    'view_whitelists',
    'view_budgets',
    'view_team'
  ]
};
```

---

## Integration with Existing Pages

### Update App.tsx
```tsx
<RoleProvider>
  <Router>
    <Routes>
      <Route path="/team" element={
        <RoleGuard allowedRoles={['owner', 'admin']}>
          <Team />
        </RoleGuard>
      } />
      <Route path="/accept-invite" element={<AcceptInvite />} />
      {/* ... other routes */}
    </Routes>
  </Router>
</RoleProvider>
```

### Update Sidebar
```tsx
<nav>
  <TenantSelector />
  
  <NavLink to="/dashboard">Dashboard</NavLink>
  <NavLink to="/resources">Resources</NavLink>
  
  <RoleGuard allowedRoles={['owner', 'admin']}>
    <NavLink to="/team">Team</NavLink>
  </RoleGuard>
  
  <RoleGuard allowedRoles={['owner', 'admin']}>
    <NavLink to="/billing">Billing</NavLink>
  </RoleGuard>
</nav>
```

### Update Login.tsx
```tsx
// After successful login
const handleLogin = async (credentials) => {
  const response = await api.post('/api/auth/login', credentials);
  
  // Store tokens
  localStorage.setItem('token', response.data.token);
  localStorage.setItem('refresh_token', response.data.refresh_token);
  
  // Store tenants for tenant selector
  localStorage.setItem('tenants', JSON.stringify(response.data.tenants));
  
  // Redirect to dashboard
  navigate('/dashboard');
};
```

---

## Testing Checklist

### Team Page
- [ ] List shows all team members
- [ ] Role badges display correctly
- [ ] Invite button visible for owner/admin only
- [ ] Update role works
- [ ] Remove user works (except owner)
- [ ] Pending invitations list shows
- [ ] Resend invitation works
- [ ] Revoke invitation works

### Invite Modal
- [ ] Email validation works
- [ ] Role selector shows correct options
- [ ] Submit sends invitation
- [ ] Success message displays
- [ ] Modal closes after success
- [ ] Error handling works

### Accept Invitation
- [ ] Token parsed from URL
- [ ] Invitation details display
- [ ] Login redirect works if not authenticated
- [ ] Accept button works
- [ ] Success redirects to dashboard
- [ ] Expired invitation shows error
- [ ] Invalid token shows error

### Role Guards
- [ ] UI elements hidden based on role
- [ ] Viewer can't see admin features
- [ ] Editor can't manage team
- [ ] Admin can't manage billing
- [ ] Owner sees all features

### Tenant Selector
- [ ] Shows current tenant
- [ ] Lists all user's tenants
- [ ] Switch tenant works
- [ ] Page reloads after switch
- [ ] New tenant data loads

---

## API Integration Summary

### Endpoints Used
```typescript
// Team Management
GET    /api/v1/team/members
GET    /api/v1/team/invitations
POST   /api/v1/team/invite
POST   /api/v1/team/accept-invite
GET    /api/v1/team/invite-details?token=xxx
PUT    /api/v1/team/members/:id/role
DELETE /api/v1/team/members/:id
POST   /api/v1/team/invitations/:id/resend
DELETE /api/v1/team/invitations/:id

// Auth
POST   /api/auth/switch-tenant
GET    /api/auth/current-user
```

---

## Styling Guide

### Color Scheme
```css
:root {
  --role-owner: #8b5cf6;    /* Purple */
  --role-admin: #3b82f6;    /* Blue */
  --role-editor: #10b981;   /* Green */
  --role-viewer: #6b7280;   /* Gray */
}
```

### Role Badges
```css
.role-badge {
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 600;
  text-transform: uppercase;
}

.role-badge.owner { background: var(--role-owner); color: white; }
.role-badge.admin { background: var(--role-admin); color: white; }
.role-badge.editor { background: var(--role-editor); color: white; }
.role-badge.viewer { background: var(--role-viewer); color: white; }
```

---

## Success Criteria

✅ **Week 3 Complete When:**
- Team page displays all members and invitations
- Users can invite new members
- Invitations can be accepted via email link
- Role-based UI rendering works
- Tenant selector allows switching
- All permission checks enforced
- No console errors
- Responsive design works

---

**Estimated Effort**: 5 days (1 frontend developer)  
**Dependencies**: Week 1 & 2 backend complete ✅  
**Status**: Ready to implement

