import React, { useState, useEffect } from 'react';
import { Shield, Plus, Trash2, AlertCircle, Server, Database, HardDrive, Cloud } from 'lucide-react';
import api from '../services/api';

interface Whitelist {
  id: string;
  whitelist_type: string;
  resource_id?: string;
  resource_type?: string;
  resource_arn?: string;
  reason: string;
  cost_impact_monthly: number;
  created_by: string;
  created_at: string;
  expires_at?: string;
  status: string;
}

const getResourceIcon = (type: string) => {
  const resourceType = type?.toLowerCase() || '';
  if (resourceType.includes('ec2')) return <Server className="w-4 h-4" />;
  if (resourceType.includes('rds')) return <Database className="w-4 h-4" />;
  if (resourceType.includes('ebs')) return <HardDrive className="w-4 h-4" />;
  if (resourceType.includes('s3')) return <Cloud className="w-4 h-4" />;
  return <Shield className="w-4 h-4" />;
};

export const Whitelists: React.FC = () => {
  const [whitelists, setWhitelists] = useState<Whitelist[]>([]);
  const [loading, setLoading] = useState(true);
  const [totalImpact, setTotalImpact] = useState(0);
  const [summary, setSummary] = useState({ active: 0, expired: 0, pending_approval: 0 });
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [selectedWhitelists, setSelectedWhitelists] = useState<Set<string>>(new Set());
  const [newWhitelist, setNewWhitelist] = useState({
    resource_id: '',
    resource_type: '',
    reason: '',
    expires_in_days: 90,
  });

  useEffect(() => {
    loadWhitelists();
  }, []);

  const loadWhitelists = async () => {
    setLoading(true);
    try {
      const data = await api.get('/whitelists');
      const { whitelists: wl = [], total_cost_impact = 0, summary: sum = { active: 0, expired: 0, pending_approval: 0 } } = data as any;
      setWhitelists(Array.isArray(wl) ? wl : []);
      setTotalImpact(total_cost_impact);
      setSummary(sum);
    } catch (error) {
      console.error('Error loading whitelists:', error);
      setWhitelists([]);
    } finally {
      setLoading(false);
    }
  };

  const createWhitelist = async () => {
    try {
      await api.post('/whitelists', newWhitelist);
      setShowCreateModal(false);
      setNewWhitelist({ resource_id: '', resource_type: '', reason: '', expires_in_days: 90 });
      loadWhitelists();
    } catch (error) {
      console.error('Error creating whitelist:', error);
      alert('Failed to create whitelist');
    }
  };

  const revokeWhitelist = async (id: string) => {
    try {
      await api.delete(`/whitelists/${id}`, { reason: 'Removed from whitelist' });
      console.log('Resource removed from whitelist successfully');
      loadWhitelists();
    } catch (error) {
      console.error('Error revoking whitelist:', error);
    }
  };

  const handleBulkRemove = async () => {
    if (selectedWhitelists.size === 0) {
      console.log('No whitelists selected for removal');
      return;
    }

    try {
      const deletePromises = Array.from(selectedWhitelists).map(id =>
        api.delete(`/whitelists/${id}`, { reason: 'Bulk removal from whitelist' })
      );

      await Promise.all(deletePromises);
      console.log(`${selectedWhitelists.size} whitelist(s) removed successfully`);
      setSelectedWhitelists(new Set());
      loadWhitelists();
    } catch (error) {
      console.error('Bulk remove failed:', error);
    }
  };

  const toggleWhitelistSelection = (id: string) => {
    const newSelection = new Set(selectedWhitelists);
    if (newSelection.has(id)) {
      newSelection.delete(id);
    } else {
      newSelection.add(id);
    }
    setSelectedWhitelists(newSelection);
  };

  const toggleSelectAll = () => {
    if (!whitelists || whitelists.length === 0) return;
    
    if (selectedWhitelists.size === whitelists.length) {
      setSelectedWhitelists(new Set());
    } else {
      setSelectedWhitelists(new Set(whitelists.map(wl => wl.id)));
    }
  };

  return (
    <div className="min-h-screen bg-slate-50">
      <div className="max-w-7xl mx-auto p-6 space-y-6">
        {/* Header */}
        <div className="bg-white rounded-lg shadow-sm border border-slate-200 p-6">
          <div className="flex flex-col lg:flex-row lg:items-center lg:justify-between gap-4">
            <div>
              <h1 className="text-3xl font-bold text-slate-900">Whitelisted Resources</h1>
              <p className="text-slate-600 mt-1">Resources excluded from cost optimization recommendations</p>
            </div>
            <div className="flex gap-2">
              {selectedWhitelists.size > 0 && (
                <button
                  onClick={handleBulkRemove}
                  className="flex items-center gap-2 px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 transition-colors"
                >
                  <Trash2 className="w-4 h-4" />
                  Remove Selected ({selectedWhitelists.size})
                </button>
              )}
              <button
                onClick={() => setShowCreateModal(true)}
                className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
              >
                <Plus className="w-4 h-4" />
                Add to Whitelist
              </button>
            </div>
          </div>
        </div>

        {/* Summary Cards */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
          <div className="bg-white rounded-lg shadow-sm border border-slate-200 p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-slate-600 text-sm font-medium">Active Whitelists</p>
                <p className="text-3xl font-bold text-green-600 mt-1">{summary.active}</p>
              </div>
              <div className="p-3 bg-green-100 rounded-lg">
                <Shield className="w-6 h-6 text-green-600" />
              </div>
            </div>
          </div>

          <div className="bg-white rounded-lg shadow-sm border border-slate-200 p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-slate-600 text-sm font-medium">Pending Approval</p>
                <p className="text-3xl font-bold text-yellow-600 mt-1">{summary.pending_approval}</p>
              </div>
              <div className="p-3 bg-yellow-100 rounded-lg">
                <AlertCircle className="w-6 h-6 text-yellow-600" />
              </div>
            </div>
          </div>

          <div className="bg-white rounded-lg shadow-sm border border-slate-200 p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-slate-600 text-sm font-medium">Expired</p>
                <p className="text-3xl font-bold text-slate-600 mt-1">{summary.expired}</p>
              </div>
              <div className="p-3 bg-slate-100 rounded-lg">
                <AlertCircle className="w-6 h-6 text-slate-600" />
              </div>
            </div>
          </div>

          <div className="bg-white rounded-lg shadow-sm border border-slate-200 p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-slate-600 text-sm font-medium">Cost Impact</p>
                <p className="text-3xl font-bold text-red-600 mt-1">${totalImpact.toFixed(0)}/mo</p>
              </div>
              <div className="p-3 bg-red-100 rounded-lg">
                <AlertCircle className="w-6 h-6 text-red-600" />
              </div>
            </div>
          </div>
        </div>

        {/* Whitelists Table */}
        <div className="bg-white rounded-lg shadow-sm border border-slate-200">
          <div className="p-6 border-b border-slate-200">
            <h3 className="text-lg font-semibold text-slate-900">Whitelisted Resources</h3>
          </div>

          {loading ? (
            <div className="p-8">
              <div className="space-y-4">
                {Array.from({ length: 3 }).map((_, i) => (
                  <div key={i} className="animate-pulse flex items-center space-x-4">
                    <div className="w-8 h-8 bg-slate-200 rounded"></div>
                    <div className="flex-1 space-y-2">
                      <div className="h-4 bg-slate-200 rounded w-3/4"></div>
                      <div className="h-3 bg-slate-200 rounded w-1/2"></div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          ) : !whitelists || whitelists.length === 0 ? (
            <div className="p-12 text-center">
              <Shield className="w-12 h-12 text-slate-400 mx-auto mb-4" />
              <h3 className="text-lg font-medium text-slate-900 mb-2">No whitelisted resources</h3>
              <p className="text-slate-600">Add resources to whitelist to exclude them from cost optimization recommendations.</p>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead className="bg-slate-50 border-b border-slate-200">
                  <tr>
                    <th className="py-3 px-4 text-left">
                      <input
                        type="checkbox"
                        checked={whitelists && selectedWhitelists.size === whitelists.length && whitelists.length > 0}
                        onChange={toggleSelectAll}
                        className="w-4 h-4 rounded border-slate-300"
                      />
                    </th>
                    <th className="text-left py-3 px-6 font-medium text-slate-700">Resource</th>
                    <th className="text-left py-3 px-6 font-medium text-slate-700">Reason</th>
                    <th className="text-left py-3 px-6 font-medium text-slate-700">Cost Impact</th>
                    <th className="text-left py-3 px-6 font-medium text-slate-700">Expires</th>
                    <th className="text-left py-3 px-6 font-medium text-slate-700">Status</th>
                    <th className="text-left py-3 px-6 font-medium text-slate-700">Actions</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-200">
                  {whitelists.map((wl, index) => {
                    const isEven = index % 2 === 0;
                    return (
                      <tr key={wl.id} className={`hover:bg-slate-50 transition-colors ${
                        isEven ? 'bg-white' : 'bg-slate-25'
                      }`}>
                        <td className="py-4 px-4">
                          <input
                            type="checkbox"
                            checked={selectedWhitelists.has(wl.id)}
                            onChange={() => toggleWhitelistSelection(wl.id)}
                            className="w-4 h-4 rounded border-slate-300"
                          />
                        </td>
                        <td className="py-4 px-6">
                          <div className="flex items-center gap-3">
                            <div className="p-2 bg-slate-100 rounded-lg">
                              {getResourceIcon(wl.resource_type || wl.whitelist_type)}
                            </div>
                            <div>
                              <div className="font-medium text-slate-900">
                                {wl.resource_id || wl.resource_arn?.split('/').pop() || 'Unknown Resource'}
                              </div>
                              <div className="text-sm text-slate-600">
                                {wl.resource_type || wl.whitelist_type}
                              </div>
                            </div>
                          </div>
                        </td>
                        <td className="py-4 px-6">
                          <div className="text-sm text-slate-900 max-w-xs">
                            <p className="truncate" title={wl.reason}>{wl.reason}</p>
                          </div>
                        </td>
                        <td className="py-4 px-6">
                          <div className="font-medium text-red-600">
                            ${wl.cost_impact_monthly.toFixed(2)}
                          </div>
                          <div className="text-xs text-slate-600">per month</div>
                        </td>
                        <td className="py-4 px-6 text-sm text-slate-700">
                          {wl.expires_at ? new Date(wl.expires_at).toLocaleDateString() : 'Never'}
                        </td>
                        <td className="py-4 px-6">
                          <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium border ${
                            wl.status === 'active' ? 'bg-green-100 text-green-800 border-green-200' :
                            wl.status === 'pending_approval' ? 'bg-yellow-100 text-yellow-800 border-yellow-200' :
                            'bg-slate-100 text-slate-800 border-slate-200'
                          }`}>
                            {wl.status}
                          </span>
                        </td>
                        <td className="py-4 px-6">
                          <button
                            onClick={() => revokeWhitelist(wl.id)}
                            className="p-1 text-slate-600 hover:text-red-600 transition-colors"
                            title="Revoke Whitelist"
                          >
                            <Trash2 className="w-4 h-4" />
                          </button>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          )}
        </div>

        {/* Create Modal */}
        {showCreateModal && (
          <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-6 z-50">
            <div className="bg-white rounded-lg shadow-xl max-w-md w-full">
              <div className="p-6 border-b border-slate-200">
                <h2 className="text-2xl font-bold text-slate-900">Add Resource to Whitelist</h2>
                <p className="text-slate-600 mt-1">Exclude this resource from cost optimization recommendations</p>
              </div>
              
              <div className="p-6 space-y-4">
                <div>
                  <label className="block text-sm font-medium text-slate-700 mb-2">Resource ID</label>
                  <input
                    type="text"
                    value={newWhitelist.resource_id}
                    onChange={(e) => setNewWhitelist({ ...newWhitelist, resource_id: e.target.value })}
                    className="w-full border border-slate-300 rounded-lg px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    placeholder="i-1234567890abcdef0"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-slate-700 mb-2">Resource Type</label>
                  <select
                    value={newWhitelist.resource_type}
                    onChange={(e) => setNewWhitelist({ ...newWhitelist, resource_type: e.target.value })}
                    className="w-full border border-slate-300 rounded-lg px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                  >
                    <option value="">Select resource type</option>
                    <option value="AWS::EC2::Instance">EC2 Instance</option>
                    <option value="AWS::RDS::DBInstance">RDS Database</option>
                    <option value="AWS::S3::Bucket">S3 Bucket</option>
                    <option value="AWS::EBS::Volume">EBS Volume</option>
                  </select>
                </div>

                <div>
                  <label className="block text-sm font-medium text-slate-700 mb-2">Reason (min 20 characters)</label>
                  <textarea
                    value={newWhitelist.reason}
                    onChange={(e) => setNewWhitelist({ ...newWhitelist, reason: e.target.value })}
                    className="w-full border border-slate-300 rounded-lg px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    rows={3}
                    placeholder="Explain why this resource should be excluded from optimization recommendations..."
                  />
                  <p className="text-xs text-slate-500 mt-1">{newWhitelist.reason.length}/20 characters minimum</p>
                </div>

                <div>
                  <label className="block text-sm font-medium text-slate-700 mb-2">Expires In</label>
                  <select
                    value={newWhitelist.expires_in_days}
                    onChange={(e) => setNewWhitelist({ ...newWhitelist, expires_in_days: parseInt(e.target.value) })}
                    className="w-full border border-slate-300 rounded-lg px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                  >
                    <option value={30}>30 days</option>
                    <option value={60}>60 days</option>
                    <option value={90}>90 days</option>
                    <option value={180}>180 days</option>
                    <option value={365}>1 year</option>
                  </select>
                </div>
              </div>

              <div className="p-6 border-t border-slate-200 flex gap-3">
                <button
                  onClick={createWhitelist}
                  disabled={!newWhitelist.resource_id || !newWhitelist.resource_type || newWhitelist.reason.length < 20}
                  className="flex-1 bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                >
                  Add to Whitelist
                </button>
                <button
                  onClick={() => {
                    setShowCreateModal(false);
                    setNewWhitelist({ resource_id: '', resource_type: '', reason: '', expires_in_days: 90 });
                  }}
                  className="flex-1 border border-slate-300 px-4 py-2 rounded-lg hover:bg-slate-50 transition-colors"
                >
                  Cancel
                </button>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};
