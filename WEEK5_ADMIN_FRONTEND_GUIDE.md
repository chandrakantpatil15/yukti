# Week 5: Admin Portal Frontend - Implementation Guide

## Overview
Week 5 implements the admin portal frontend for platform administrators to manage tenants, users, and perform impersonation.

## Timeline: 5 Days

### Day 1: Admin Login & Dashboard
### Day 2: Tenant Management UI
### Day 3: User Management UI
### Day 4: Impersonation UI
### Day 5: Analytics & Polish

---

## Day 1: Admin Login & Dashboard

### 1. Admin Login Page
**File**: `frontend/src/pages/Admin/AdminLogin.tsx`

```typescript
import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import api from '../../services/api';

export default function AdminLogin() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const navigate = useNavigate();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError('');

    try {
      const response = await api.post('/api/admin/login', { email, password });
      localStorage.setItem('admin_token', response.data.token);
      localStorage.setItem('admin_user', JSON.stringify(response.data.admin));
      navigate('/admin/dashboard');
    } catch (err: any) {
      setError(err.response?.data?.error || 'Login failed');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <div className="max-w-md w-full bg-white p-8 rounded-lg shadow">
        <h2 className="text-2xl font-bold mb-6">Admin Portal</h2>
        
        {error && (
          <div className="bg-red-50 text-red-600 p-3 rounded mb-4">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit}>
          <div className="mb-4">
            <label className="block text-sm font-medium mb-2">Email</label>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="w-full px-3 py-2 border rounded"
              required
            />
          </div>

          <div className="mb-6">
            <label className="block text-sm font-medium mb-2">Password</label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="w-full px-3 py-2 border rounded"
              required
            />
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full bg-blue-600 text-white py-2 rounded hover:bg-blue-700 disabled:opacity-50"
          >
            {loading ? 'Logging in...' : 'Login'}
          </button>
        </form>
      </div>
    </div>
  );
}
```

### 2. Admin Dashboard
**File**: `frontend/src/pages/Admin/AdminDashboard.tsx`

```typescript
import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import api from '../../services/api';

interface PlatformStats {
  total_tenants: number;
  active_tenants: number;
  total_users: number;
  total_resources: number;
  total_findings: number;
  total_savings: number;
}

export default function AdminDashboard() {
  const [stats, setStats] = useState<PlatformStats | null>(null);
  const [loading, setLoading] = useState(true);
  const navigate = useNavigate();

  useEffect(() => {
    fetchStats();
  }, []);

  const fetchStats = async () => {
    try {
      const token = localStorage.getItem('admin_token');
      const response = await api.get('/api/admin/stats', {
        headers: { Authorization: `Bearer ${token}` }
      });
      setStats(response.data);
    } catch (err) {
      console.error('Failed to fetch stats:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleLogout = () => {
    localStorage.removeItem('admin_token');
    localStorage.removeItem('admin_user');
    navigate('/admin/login');
  };

  if (loading) return <div>Loading...</div>;

  return (
    <div className="min-h-screen bg-gray-50">
      <nav className="bg-white shadow px-6 py-4 flex justify-between items-center">
        <h1 className="text-xl font-bold">Admin Portal</h1>
        <button
          onClick={handleLogout}
          className="px-4 py-2 text-sm bg-gray-200 rounded hover:bg-gray-300"
        >
          Logout
        </button>
      </nav>

      <div className="p-6">
        <h2 className="text-2xl font-bold mb-6">Platform Overview</h2>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
          <StatCard
            title="Total Tenants"
            value={stats?.total_tenants || 0}
            subtitle={`${stats?.active_tenants || 0} active`}
          />
          <StatCard
            title="Total Users"
            value={stats?.total_users || 0}
          />
          <StatCard
            title="Total Resources"
            value={stats?.total_resources || 0}
          />
          <StatCard
            title="Total Findings"
            value={stats?.total_findings || 0}
          />
          <StatCard
            title="Total Savings"
            value={`$${(stats?.total_savings || 0).toFixed(2)}`}
          />
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <QuickAction
            title="Manage Tenants"
            description="View and manage all customer tenants"
            onClick={() => navigate('/admin/tenants')}
          />
          <QuickAction
            title="Manage Users"
            description="View and manage all platform users"
            onClick={() => navigate('/admin/users')}
          />
        </div>
      </div>
    </div>
  );
}

function StatCard({ title, value, subtitle }: any) {
  return (
    <div className="bg-white p-6 rounded-lg shadow">
      <h3 className="text-sm text-gray-600 mb-2">{title}</h3>
      <p className="text-3xl font-bold">{value}</p>
      {subtitle && <p className="text-sm text-gray-500 mt-1">{subtitle}</p>}
    </div>
  );
}

function QuickAction({ title, description, onClick }: any) {
  return (
    <div
      onClick={onClick}
      className="bg-white p-6 rounded-lg shadow cursor-pointer hover:shadow-lg transition"
    >
      <h3 className="text-lg font-bold mb-2">{title}</h3>
      <p className="text-gray-600">{description}</p>
    </div>
  );
}
```

### 3. Admin API Service
**File**: `frontend/src/services/adminApi.ts`

```typescript
import axios from 'axios';

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8081';

const adminApi = axios.create({
  baseURL: API_URL,
});

// Add auth token to all requests
adminApi.interceptors.request.use((config) => {
  const token = localStorage.getItem('admin_token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Handle 401 errors
adminApi.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('admin_token');
      localStorage.removeItem('admin_user');
      window.location.href = '/admin/login';
    }
    return Promise.reject(error);
  }
);

export default adminApi;
```

### 4. Admin Routes
**File**: Update `frontend/src/App.tsx`

```typescript
// Add admin routes
import AdminLogin from './pages/Admin/AdminLogin';
import AdminDashboard from './pages/Admin/AdminDashboard';
import AdminTenants from './pages/Admin/AdminTenants';
import AdminUsers from './pages/Admin/AdminUsers';

// In routes section:
<Route path="/admin/login" element={<AdminLogin />} />
<Route path="/admin/dashboard" element={<AdminDashboard />} />
<Route path="/admin/tenants" element={<AdminTenants />} />
<Route path="/admin/users" element={<AdminUsers />} />
```

---

## Day 2: Tenant Management UI

### 1. Tenant List Page
**File**: `frontend/src/pages/Admin/AdminTenants.tsx`

```typescript
import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import adminApi from '../../services/adminApi';

interface Tenant {
  id: string;
  name: string;
  status: string;
  user_count: number;
  resource_count: number;
  findings_count: number;
  monthly_savings: number;
  created_at: string;
}

export default function AdminTenants() {
  const [tenants, setTenants] = useState<Tenant[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState('');
  const navigate = useNavigate();

  useEffect(() => {
    fetchTenants();
  }, []);

  const fetchTenants = async () => {
    try {
      const response = await adminApi.get('/api/admin/tenants');
      setTenants(response.data.tenants || []);
    } catch (err) {
      console.error('Failed to fetch tenants:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleSuspend = async (tenantId: string) => {
    if (!confirm('Suspend this tenant?')) return;
    
    try {
      await adminApi.post(`/api/admin/tenants/${tenantId}/suspend`);
      fetchTenants();
    } catch (err) {
      alert('Failed to suspend tenant');
    }
  };

  const handleActivate = async (tenantId: string) => {
    try {
      await adminApi.post(`/api/admin/tenants/${tenantId}/activate`);
      fetchTenants();
    } catch (err) {
      alert('Failed to activate tenant');
    }
  };

  const handleImpersonate = (tenantId: string) => {
    navigate(`/admin/impersonate?tenant=${tenantId}`);
  };

  const filteredTenants = tenants.filter(t =>
    t.name.toLowerCase().includes(search.toLowerCase())
  );

  if (loading) return <div>Loading...</div>;

  return (
    <div className="min-h-screen bg-gray-50">
      <nav className="bg-white shadow px-6 py-4">
        <button
          onClick={() => navigate('/admin/dashboard')}
          className="text-blue-600 hover:underline"
        >
          ← Back to Dashboard
        </button>
      </nav>

      <div className="p-6">
        <h2 className="text-2xl font-bold mb-6">Tenant Management</h2>

        <input
          type="text"
          placeholder="Search tenants..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="w-full max-w-md px-4 py-2 border rounded mb-6"
        />

        <div className="bg-white rounded-lg shadow overflow-hidden">
          <table className="w-full">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                  Tenant
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                  Status
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                  Users
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                  Resources
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                  Savings
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {filteredTenants.map((tenant) => (
                <tr key={tenant.id}>
                  <td className="px-6 py-4">
                    <div className="font-medium">{tenant.name}</div>
                    <div className="text-sm text-gray-500">{tenant.id}</div>
                  </td>
                  <td className="px-6 py-4">
                    <span
                      className={`px-2 py-1 text-xs rounded ${
                        tenant.status === 'active'
                          ? 'bg-green-100 text-green-800'
                          : 'bg-red-100 text-red-800'
                      }`}
                    >
                      {tenant.status}
                    </span>
                  </td>
                  <td className="px-6 py-4">{tenant.user_count}</td>
                  <td className="px-6 py-4">{tenant.resource_count}</td>
                  <td className="px-6 py-4">
                    ${tenant.monthly_savings.toFixed(2)}
                  </td>
                  <td className="px-6 py-4">
                    <div className="flex gap-2">
                      {tenant.status === 'active' ? (
                        <button
                          onClick={() => handleSuspend(tenant.id)}
                          className="text-sm text-red-600 hover:underline"
                        >
                          Suspend
                        </button>
                      ) : (
                        <button
                          onClick={() => handleActivate(tenant.id)}
                          className="text-sm text-green-600 hover:underline"
                        >
                          Activate
                        </button>
                      )}
                      <button
                        onClick={() => handleImpersonate(tenant.id)}
                        className="text-sm text-blue-600 hover:underline"
                      >
                        Impersonate
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
```

---

## Day 3: User Management UI

### 1. User List Page
**File**: `frontend/src/pages/Admin/AdminUsers.tsx`

```typescript
import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import adminApi from '../../services/adminApi';

interface User {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
  is_active: boolean;
  tenant_count: number;
  created_at: string;
}

export default function AdminUsers() {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState('');
  const navigate = useNavigate();

  useEffect(() => {
    fetchUsers();
  }, []);

  const fetchUsers = async () => {
    try {
      const response = await adminApi.get('/api/admin/users');
      setUsers(response.data.users || []);
    } catch (err) {
      console.error('Failed to fetch users:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleSuspend = async (userId: string) => {
    if (!confirm('Suspend this user?')) return;
    
    try {
      await adminApi.post(`/api/admin/users/${userId}/suspend`);
      fetchUsers();
    } catch (err) {
      alert('Failed to suspend user');
    }
  };

  const handleActivate = async (userId: string) => {
    try {
      await adminApi.post(`/api/admin/users/${userId}/activate`);
      fetchUsers();
    } catch (err) {
      alert('Failed to activate user');
    }
  };

  const filteredUsers = users.filter(u =>
    u.email.toLowerCase().includes(search.toLowerCase()) ||
    `${u.first_name} ${u.last_name}`.toLowerCase().includes(search.toLowerCase())
  );

  if (loading) return <div>Loading...</div>;

  return (
    <div className="min-h-screen bg-gray-50">
      <nav className="bg-white shadow px-6 py-4">
        <button
          onClick={() => navigate('/admin/dashboard')}
          className="text-blue-600 hover:underline"
        >
          ← Back to Dashboard
        </button>
      </nav>

      <div className="p-6">
        <h2 className="text-2xl font-bold mb-6">User Management</h2>

        <input
          type="text"
          placeholder="Search users..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="w-full max-w-md px-4 py-2 border rounded mb-6"
        />

        <div className="bg-white rounded-lg shadow overflow-hidden">
          <table className="w-full">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                  User
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                  Status
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                  Tenants
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                  Created
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {filteredUsers.map((user) => (
                <tr key={user.id}>
                  <td className="px-6 py-4">
                    <div className="font-medium">
                      {user.first_name} {user.last_name}
                    </div>
                    <div className="text-sm text-gray-500">{user.email}</div>
                  </td>
                  <td className="px-6 py-4">
                    <span
                      className={`px-2 py-1 text-xs rounded ${
                        user.is_active
                          ? 'bg-green-100 text-green-800'
                          : 'bg-red-100 text-red-800'
                      }`}
                    >
                      {user.is_active ? 'Active' : 'Suspended'}
                    </span>
                  </td>
                  <td className="px-6 py-4">{user.tenant_count}</td>
                  <td className="px-6 py-4">
                    {new Date(user.created_at).toLocaleDateString()}
                  </td>
                  <td className="px-6 py-4">
                    {user.is_active ? (
                      <button
                        onClick={() => handleSuspend(user.id)}
                        className="text-sm text-red-600 hover:underline"
                      >
                        Suspend
                      </button>
                    ) : (
                      <button
                        onClick={() => handleActivate(user.id)}
                        className="text-sm text-green-600 hover:underline"
                      >
                        Activate
                      </button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
```

---

## Day 4: Impersonation UI

### 1. Impersonation Modal
**File**: `frontend/src/components/Admin/ImpersonationModal.tsx`

```typescript
import { useState } from 'react';
import adminApi from '../../services/adminApi';

interface Props {
  tenantId: string;
  userId: string;
  onSuccess: (token: string) => void;
  onCancel: () => void;
}

export default function ImpersonationModal({ tenantId, userId, onSuccess, onCancel }: Props) {
  const [reason, setReason] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);

    try {
      const response = await adminApi.post('/api/admin/impersonate', {
        user_id: userId,
        tenant_id: tenantId,
        reason
      });
      
      onSuccess(response.data.impersonation_token);
    } catch (err) {
      alert('Impersonation failed');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg p-6 max-w-md w-full">
        <h3 className="text-lg font-bold mb-4">Impersonate User</h3>
        
        <form onSubmit={handleSubmit}>
          <div className="mb-4">
            <label className="block text-sm font-medium mb-2">
              Reason for Impersonation
            </label>
            <textarea
              value={reason}
              onChange={(e) => setReason(e.target.value)}
              className="w-full px-3 py-2 border rounded"
              rows={3}
              placeholder="e.g., Customer support request #1234"
              required
            />
          </div>

          <div className="flex gap-3">
            <button
              type="button"
              onClick={onCancel}
              className="flex-1 px-4 py-2 border rounded hover:bg-gray-50"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={loading}
              className="flex-1 px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50"
            >
              {loading ? 'Starting...' : 'Start Impersonation'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
```

### 2. Impersonation Banner
**File**: `frontend/src/components/Admin/ImpersonationBanner.tsx`

```typescript
import { useState } from 'react';
import adminApi from '../../services/adminApi';

export default function ImpersonationBanner() {
  const [ending, setEnding] = useState(false);

  const handleEndImpersonation = async () => {
    setEnding(true);
    
    try {
      await adminApi.post('/api/admin/end-impersonation');
      localStorage.removeItem('impersonation_token');
      window.location.href = '/admin/dashboard';
    } catch (err) {
      alert('Failed to end impersonation');
      setEnding(false);
    }
  };

  return (
    <div className="bg-yellow-500 text-black px-6 py-3 flex items-center justify-between">
      <div className="flex items-center gap-2">
        <span className="font-bold">⚠️ IMPERSONATION MODE</span>
        <span>You are viewing this account as an administrator</span>
      </div>
      <button
        onClick={handleEndImpersonation}
        disabled={ending}
        className="px-4 py-1 bg-black text-white rounded hover:bg-gray-800 disabled:opacity-50"
      >
        {ending ? 'Ending...' : 'End Impersonation'}
      </button>
    </div>
  );
}
```

---

## Day 5: Analytics & Polish

### 1. Platform Analytics
**File**: `frontend/src/pages/Admin/AdminAnalytics.tsx`

```typescript
import { useEffect, useState } from 'react';
import adminApi from '../../services/adminApi';

export default function AdminAnalytics() {
  const [analytics, setAnalytics] = useState<any>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchAnalytics();
  }, []);

  const fetchAnalytics = async () => {
    try {
      const response = await adminApi.get('/api/admin/analytics');
      setAnalytics(response.data);
    } catch (err) {
      console.error('Failed to fetch analytics:', err);
    } finally {
      setLoading(false);
    }
  };

  if (loading) return <div>Loading...</div>;

  return (
    <div className="p-6">
      <h2 className="text-2xl font-bold mb-6">Platform Analytics</h2>
      
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div className="bg-white p-6 rounded-lg shadow">
          <h3 className="font-bold mb-4">Growth Metrics</h3>
          <div className="space-y-2">
            <div>New Tenants (30d): {analytics?.new_tenants_30d || 0}</div>
            <div>New Users (30d): {analytics?.new_users_30d || 0}</div>
            <div>Active Scans (7d): {analytics?.active_scans_7d || 0}</div>
          </div>
        </div>

        <div className="bg-white p-6 rounded-lg shadow">
          <h3 className="font-bold mb-4">Resource Metrics</h3>
          <div className="space-y-2">
            <div>Total Resources: {analytics?.total_resources || 0}</div>
            <div>Total Findings: {analytics?.total_findings || 0}</div>
            <div>Avg Savings/Tenant: ${analytics?.avg_savings_per_tenant || 0}</div>
          </div>
        </div>
      </div>
    </div>
  );
}
```

---

## Testing Checklist

### Day 1
- [ ] Admin can login with credentials
- [ ] Dashboard shows platform stats
- [ ] Logout works correctly
- [ ] 401 redirects to login

### Day 2
- [ ] Tenant list loads
- [ ] Search filters tenants
- [ ] Suspend/activate works
- [ ] Impersonate button navigates

### Day 3
- [ ] User list loads
- [ ] Search filters users
- [ ] Suspend/activate works

### Day 4
- [ ] Impersonation modal shows
- [ ] Reason field required
- [ ] Impersonation starts
- [ ] Banner shows during impersonation
- [ ] End impersonation works

### Day 5
- [ ] Analytics page loads
- [ ] Metrics display correctly
- [ ] UI polish complete

---

## Next Steps: Week 6

- Backend testing (unit + integration)
- Frontend E2E tests
- Security audit
- Documentation
- Production deployment

---

**Status**: Ready for implementation 🚀
