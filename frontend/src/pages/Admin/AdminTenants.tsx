import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import adminApi from '../../services/adminApi';
import ImpersonationModal from '../../components/Admin/ImpersonationModal';

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
  const [showImpersonateModal, setShowImpersonateModal] = useState(false);
  const [selectedTenant, setSelectedTenant] = useState<Tenant | null>(null);
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

  const handleImpersonate = (tenant: Tenant) => {
    setSelectedTenant(tenant);
    setShowImpersonateModal(true);
  };

  const handleImpersonationSuccess = (token: string) => {
    localStorage.setItem('impersonation_token', token);
    localStorage.setItem('token', token);
    setShowImpersonateModal(false);
    window.location.href = '/dashboard';
  };

  const filteredTenants = tenants.filter(t =>
    t.name.toLowerCase().includes(search.toLowerCase())
  );

  if (loading) return <div className="p-6">Loading...</div>;

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
                        onClick={() => handleImpersonate(tenant)}
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

        {showImpersonateModal && selectedTenant && (
          <ImpersonationModal
            userId="admin-user-id"
            tenantId={selectedTenant.id}
            userEmail={`tenant-${selectedTenant.id}@example.com`}
            onSuccess={handleImpersonationSuccess}
            onCancel={() => setShowImpersonateModal(false)}
          />
        )}
      </div>
    </div>
  );
}
