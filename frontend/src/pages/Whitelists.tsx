import React, { useState, useEffect } from 'react';
import api from '../services/api';

interface Whitelist {
  id: string;
  whitelist_type: string;
  resource_arn?: string;
  reason: string;
  cost_impact_monthly: number;
  created_by: string;
  created_at: string;
  expires_at?: string;
  status: string;
}

export const Whitelists: React.FC = () => {
  const [whitelists, setWhitelists] = useState<Whitelist[]>([]);
  const [totalImpact, setTotalImpact] = useState(0);
  const [summary, setSummary] = useState({ active: 0, expired: 0, pending_approval: 0 });
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [newWhitelist, setNewWhitelist] = useState({
    whitelist_type: 'resource',
    resource_arn: '',
    reason: '',
    expires_in_days: 90,
  });

  useEffect(() => {
    loadWhitelists();
  }, []);

  const loadWhitelists = async () => {
    try {
      const data = await api.get('/api/whitelists');
      if (data.success) {
        setWhitelists(data.whitelists || []);
        setTotalImpact(data.total_cost_impact || 0);
        setSummary(data.summary || { active: 0, expired: 0, pending_approval: 0 });
      }
    } catch (error) {
      console.error('Error loading whitelists:', error);
    }
  };

  const createWhitelist = async () => {
    try {
      await api.post('/api/whitelists', newWhitelist);
      setShowCreateModal(false);
      loadWhitelists();
    } catch (error) {
      console.error('Error creating whitelist:', error);
    }
  };

  const revokeWhitelist = async (id: string) => {
    try {
      await api.delete(`/api/whitelists/${id}`, { reason: 'No longer needed' });
      loadWhitelists();
    } catch (error) {
      console.error('Error revoking whitelist:', error);
    }
  };

  return (
    <div className="p-6 max-w-7xl mx-auto">
      {/* Header */}
      <div className="flex justify-between items-center mb-6">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">Whitelisted Resources</h1>
          <p className="text-gray-600 mt-2">Exclude business-critical resources from recommendations</p>
        </div>
        <button
          onClick={() => setShowCreateModal(true)}
          className="bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700"
        >
          + Add Whitelist
        </button>
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-4 gap-4 mb-6">
        <div className="bg-white p-6 rounded-lg shadow">
          <div className="text-sm text-gray-600">Active</div>
          <div className="text-3xl font-bold text-green-600">{summary.active}</div>
        </div>
        <div className="bg-white p-6 rounded-lg shadow">
          <div className="text-sm text-gray-600">Pending Approval</div>
          <div className="text-3xl font-bold text-yellow-600">{summary.pending_approval}</div>
        </div>
        <div className="bg-white p-6 rounded-lg shadow">
          <div className="text-sm text-gray-600">Expired</div>
          <div className="text-3xl font-bold text-gray-600">{summary.expired}</div>
        </div>
        <div className="bg-white p-6 rounded-lg shadow">
          <div className="text-sm text-gray-600">Total Cost Impact</div>
          <div className="text-3xl font-bold text-red-600">${totalImpact.toFixed(0)}/mo</div>
        </div>
      </div>

      {/* Whitelists Table */}
      <div className="bg-white rounded-lg shadow overflow-hidden">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Resource</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Reason</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Cost Impact</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Expires</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Actions</th>
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {whitelists.map(wl => (
              <tr key={wl.id} className="hover:bg-gray-50">
                <td className="px-6 py-4">
                  <div className="text-sm font-medium text-gray-900">{wl.whitelist_type}</div>
                  <div className="text-sm text-gray-500">{wl.resource_arn}</div>
                </td>
                <td className="px-6 py-4 text-sm text-gray-900 max-w-xs truncate">{wl.reason}</td>
                <td className="px-6 py-4 text-sm font-semibold text-red-600">
                  ${wl.cost_impact_monthly.toFixed(0)}/mo
                </td>
                <td className="px-6 py-4 text-sm text-gray-900">
                  {wl.expires_at ? new Date(wl.expires_at).toLocaleDateString() : 'Never'}
                </td>
                <td className="px-6 py-4">
                  <span className={`px-2 py-1 text-xs font-semibold rounded-full ${
                    wl.status === 'active' ? 'bg-green-100 text-green-800' :
                    wl.status === 'pending_approval' ? 'bg-yellow-100 text-yellow-800' :
                    'bg-gray-100 text-gray-800'
                  }`}>
                    {wl.status}
                  </span>
                </td>
                <td className="px-6 py-4 text-sm">
                  <button
                    onClick={() => revokeWhitelist(wl.id)}
                    className="text-red-600 hover:text-red-800"
                  >
                    Revoke
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Create Modal */}
      {showCreateModal && (
        <>
          <div className="fixed inset-0 bg-black bg-opacity-50 z-40" onClick={() => setShowCreateModal(false)} />
          <div className="fixed inset-0 flex items-center justify-center z-50">
            <div className="bg-white rounded-lg shadow-xl p-6 w-full max-w-md">
              <h2 className="text-2xl font-bold text-gray-900 mb-4">Add Whitelist</h2>
              
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Resource ARN</label>
                  <input
                    type="text"
                    value={newWhitelist.resource_arn}
                    onChange={(e) => setNewWhitelist({ ...newWhitelist, resource_arn: e.target.value })}
                    className="w-full border border-gray-300 rounded-md px-3 py-2"
                    placeholder="arn:aws:ec2:us-east-1:123456789012:instance/i-abc123"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Reason (min 20 chars)</label>
                  <textarea
                    value={newWhitelist.reason}
                    onChange={(e) => setNewWhitelist({ ...newWhitelist, reason: e.target.value })}
                    className="w-full border border-gray-300 rounded-md px-3 py-2"
                    rows={3}
                    placeholder="Explain why this resource should be excluded..."
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Expires In</label>
                  <select
                    value={newWhitelist.expires_in_days}
                    onChange={(e) => setNewWhitelist({ ...newWhitelist, expires_in_days: parseInt(e.target.value) })}
                    className="w-full border border-gray-300 rounded-md px-3 py-2"
                  >
                    <option value={30}>30 days</option>
                    <option value={60}>60 days</option>
                    <option value={90}>90 days</option>
                  </select>
                </div>

                <div className="flex gap-3 mt-6">
                  <button
                    onClick={createWhitelist}
                    className="flex-1 bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700"
                  >
                    Create Whitelist
                  </button>
                  <button
                    onClick={() => setShowCreateModal(false)}
                    className="flex-1 border border-gray-300 px-4 py-2 rounded-md hover:bg-gray-50"
                  >
                    Cancel
                  </button>
                </div>
              </div>
            </div>
          </div>
        </>
      )}
    </div>
  );
};
