import React, { useState, useEffect } from 'react';
import { Shield, User, Clock, Activity } from 'lucide-react';
import api from '../services/api';

interface AuditLog {
  id: string;
  admin_user: string;
  action: string;
  resource_type: string;
  tenant_id: string;
  ip_address: string;
  created_at: string;
}

const AuditLogs: React.FC = () => {
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchAuditLogs();
  }, []);

  const fetchAuditLogs = async () => {
    try {
      const data = await api.get('/api/admin/audit-logs?limit=100', {
        'X-Admin-Key': 'yukti-admin-2024',
        'X-Admin-User': 'admin@yukti.com'
      });
      if (data.success && data.logs) {
        setLogs(data.logs);
      }
    } catch (error) {
      console.error('Error fetching audit logs:', error);
    } finally {
      setLoading(false);
    }
  };

  const getActionColor = (action: string) => {
    if (action.includes('impersonate')) return 'text-orange-600 bg-orange-100';
    if (action.includes('delete')) return 'text-red-600 bg-red-100';
    if (action.includes('create')) return 'text-green-600 bg-green-100';
    return 'text-blue-600 bg-blue-100';
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-100 flex items-center justify-center">
        <div className="text-xl text-gray-600">Loading audit logs...</div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-100 p-6">
      <div className="max-w-7xl mx-auto">
        <div className="mb-8 flex items-center gap-3">
          <Shield className="w-8 h-8 text-red-600" />
          <div>
            <h1 className="text-3xl font-bold text-gray-900">Security Audit Logs</h1>
            <p className="text-gray-600">Monitor all admin activities and tenant access</p>
          </div>
        </div>

        {/* Stats */}
        <div className="grid grid-cols-4 gap-6 mb-8">
          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-gray-600 text-sm">Total Actions</p>
                <p className="text-3xl font-bold text-gray-900">{logs.length}</p>
              </div>
              <Activity className="w-12 h-12 text-blue-600" />
            </div>
          </div>

          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-gray-600 text-sm">Impersonations</p>
                <p className="text-3xl font-bold text-orange-600">
                  {logs.filter(l => l.action.includes('impersonate')).length}
                </p>
              </div>
              <User className="w-12 h-12 text-orange-600" />
            </div>
          </div>

          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-gray-600 text-sm">Admin Users</p>
                <p className="text-3xl font-bold text-purple-600">
                  {new Set(logs.map(l => l.admin_user)).size}
                </p>
              </div>
              <Shield className="w-12 h-12 text-purple-600" />
            </div>
          </div>

          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-gray-600 text-sm">Last 24h</p>
                <p className="text-3xl font-bold text-green-600">
                  {logs.filter(l => {
                    const logDate = new Date(l.created_at);
                    const yesterday = new Date(Date.now() - 24 * 60 * 60 * 1000);
                    return logDate > yesterday;
                  }).length}
                </p>
              </div>
              <Clock className="w-12 h-12 text-green-600" />
            </div>
          </div>
        </div>

        {/* Logs Table */}
        <div className="bg-white rounded-lg shadow">
          <div className="p-6 border-b">
            <h2 className="text-xl font-semibold">Recent Activity</h2>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Timestamp</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Admin User</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Action</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Resource</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Tenant ID</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">IP Address</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {logs.map((log) => (
                  <tr key={log.id} className="hover:bg-gray-50">
                    <td className="px-6 py-4 text-sm text-gray-900">{log.created_at}</td>
                    <td className="px-6 py-4 text-sm font-medium text-gray-900">{log.admin_user}</td>
                    <td className="px-6 py-4">
                      <span className={`px-2 py-1 text-xs rounded-full font-semibold ${getActionColor(log.action)}`}>
                        {log.action}
                      </span>
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-900">{log.resource_type}</td>
                    <td className="px-6 py-4 text-sm text-gray-600">{log.tenant_id || '-'}</td>
                    <td className="px-6 py-4 text-sm text-gray-600">{log.ip_address}</td>
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

export default AuditLogs;
