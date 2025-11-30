import React, { useState, useEffect } from 'react';
import { Users, DollarSign, AlertTriangle, TrendingUp, Search, Eye } from 'lucide-react';
import api from '../services/api';

interface Customer {
  tenant_id: string;
  company_name: string;
  email: string;
  onboarding_status: string;
  created_at: string;
  total_savings: number;
  findings_count: number;
}

const AdminDashboard: React.FC = () => {
  const [customers, setCustomers] = useState<Customer[]>([]);
  const [stats, setStats] = useState({
    totalCustomers: 0,
    totalSavings: 0,
    activeTrials: 0,
    mrr: 0,
  });
  const [searchTerm, setSearchTerm] = useState('');

  useEffect(() => {
    fetchCustomers();
    fetchStats();
  }, []);

  const fetchCustomers = async () => {
    try {
      const data = await api.get('/api/admin/customers', {
        'X-Admin-Key': 'yukti-admin-2024',
        'X-Admin-User': 'admin@yukti.com'
      });
      if (data.success && data.customers) {
        setCustomers(data.customers);
      }
    } catch (error) {
      console.error('Error fetching customers:', error);
    }
  };

  const fetchStats = async () => {
    try {
      const data = await api.get('/api/admin/metrics', {
        'X-Admin-Key': 'yukti-admin-2024',
        'X-Admin-User': 'admin@yukti.com'
      });
      if (data.success && data.metrics) {
        setStats({
          totalCustomers: data.metrics.total_customers,
          totalSavings: data.metrics.total_savings,
          activeTrials: data.metrics.active_trials,
          mrr: data.metrics.mrr,
        });
      }
    } catch (error) {
      console.error('Error fetching stats:', error);
    }
  };

  const impersonateCustomer = async (tenantId: string) => {
    try {
      // Backend returns new JWT with impersonated tenant_id
      const data = await api.post('/api/admin/impersonate', { tenant_id: tenantId }, {
        'X-Admin-Key': 'yukti-admin-2024',
        'X-Admin-User': 'admin@yukti.com'
      });
      
      // Store new JWT token (contains impersonated tenant_id)
      if (data.token) {
        localStorage.setItem('yukti_auth_token', data.token);
        const payload = JSON.parse(atob(data.token.split('.')[1]));
        localStorage.setItem('yukti_user', JSON.stringify({
          id: payload.user_id,
          email: payload.email,
          role: payload.role,
          tenant_id: payload.tenant_id
        }));
        window.location.href = '/dashboard';
      } else {
        alert('Impersonation failed: No token returned');
      }
    } catch (error) {
      console.error('Error impersonating customer:', error);
      alert('Failed to impersonate customer');
    }
  };

  const filteredCustomers = customers.filter(c =>
    c.company_name.toLowerCase().includes(searchTerm.toLowerCase()) ||
    c.email.toLowerCase().includes(searchTerm.toLowerCase())
  );

  return (
    <div className="min-h-screen bg-gray-100 p-6">
      <div className="max-w-7xl mx-auto">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900">Admin Dashboard</h1>
          <p className="text-gray-600">Manage customers and platform metrics</p>
        </div>

        {/* Stats Cards */}
        <div className="grid grid-cols-4 gap-6 mb-8">
          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-gray-600 text-sm">Total Customers</p>
                <p className="text-3xl font-bold text-gray-900">{stats.totalCustomers}</p>
              </div>
              <Users className="w-12 h-12 text-blue-600" />
            </div>
          </div>

          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-gray-600 text-sm">Total Savings</p>
                <p className="text-3xl font-bold text-green-600">${stats.totalSavings.toLocaleString()}</p>
              </div>
              <DollarSign className="w-12 h-12 text-green-600" />
            </div>
          </div>

          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-gray-600 text-sm">Active Trials</p>
                <p className="text-3xl font-bold text-orange-600">{stats.activeTrials}</p>
              </div>
              <AlertTriangle className="w-12 h-12 text-orange-600" />
            </div>
          </div>

          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-gray-600 text-sm">MRR</p>
                <p className="text-3xl font-bold text-blue-600">${stats.mrr.toLocaleString()}</p>
              </div>
              <TrendingUp className="w-12 h-12 text-blue-600" />
            </div>
          </div>
        </div>

        {/* Customer List */}
        <div className="bg-white rounded-lg shadow">
          <div className="p-6 border-b">
            <div className="flex justify-between items-center">
              <h2 className="text-xl font-semibold">Customers</h2>
              <div className="relative">
                <Search className="absolute left-3 top-3 w-5 h-5 text-gray-400" />
                <input
                  type="text"
                  placeholder="Search customers..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="pl-10 pr-4 py-2 border rounded-lg w-64"
                />
              </div>
            </div>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Company</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Email</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Savings</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Findings</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Created</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {filteredCustomers.map((customer) => (
                  <tr key={customer.tenant_id} className="hover:bg-gray-50">
                    <td className="px-6 py-4">
                      <div className="font-medium text-gray-900">{customer.company_name}</div>
                      <div className="text-sm text-gray-500">{customer.tenant_id}</div>
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-900">{customer.email}</td>
                    <td className="px-6 py-4">
                      <span className={`px-2 py-1 text-xs rounded-full ${
                        customer.onboarding_status === 'completed'
                          ? 'bg-green-100 text-green-800'
                          : 'bg-yellow-100 text-yellow-800'
                      }`}>
                        {customer.onboarding_status}
                      </span>
                    </td>
                    <td className="px-6 py-4 text-sm font-semibold text-green-600">
                      ${customer.total_savings.toLocaleString()}
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-900">{customer.findings_count}</td>
                    <td className="px-6 py-4 text-sm text-gray-500">{customer.created_at}</td>
                    <td className="px-6 py-4">
                      <button
                        onClick={() => impersonateCustomer(customer.tenant_id)}
                        className="flex items-center gap-2 text-blue-600 hover:text-blue-800"
                      >
                        <Eye className="w-4 h-4" />
                        View
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  );
};

export default AdminDashboard;
