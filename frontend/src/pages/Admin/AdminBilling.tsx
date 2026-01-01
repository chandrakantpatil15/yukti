import React, { useState, useEffect } from 'react';
import adminApi from '../../services/adminApi';

interface Billing {
  id: number;
  tenant_id: number;
  tenant_name: string;
  plan: string;
  amount: number;
  currency: string;
  status: string;
  due_date: string;
  paid_date?: string;
  invoice_url: string;
  notes: string;
}

interface BillingStats {
  total_revenue: number;
  pending_amount: number;
  overdue_amount: number;
  monthly_revenue: number;
  total_billings: number;
  paid_billings: number;
  pending_billings: number;
  overdue_billings: number;
}

const AdminBilling: React.FC = () => {
  const [billings, setBillings] = useState<Billing[]>([]);
  const [stats, setStats] = useState<BillingStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [statusFilter, setStatusFilter] = useState('');
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showEditModal, setShowEditModal] = useState(false);
  const [selectedBilling, setSelectedBilling] = useState<Billing | null>(null);

  useEffect(() => {
    fetchBillings();
    fetchStats();
  }, [statusFilter]);

  const fetchBillings = async () => {
    try {
      const params = statusFilter ? `?status=${statusFilter}` : '';
      const response = await adminApi.get(`/billing${params}`);
      setBillings(response.data.billings || []);
    } catch (error) {
      console.error('Failed to fetch billings:', error);
    } finally {
      setLoading(false);
    }
  };

  const fetchStats = async () => {
    try {
      const response = await adminApi.get('/billing/stats');
      setStats(response.data);
    } catch (error) {
      console.error('Failed to fetch stats:', error);
    }
  };

  const handleCreate = async (data: any) => {
    try {
      await adminApi.post('/billing', data);
      setShowCreateModal(false);
      fetchBillings();
      fetchStats();
    } catch (error) {
      console.error('Failed to create billing:', error);
    }
  };

  const handleUpdate = async (id: number, data: any) => {
    try {
      await adminApi.put(`/billing/${id}`, data);
      setShowEditModal(false);
      fetchBillings();
      fetchStats();
    } catch (error) {
      console.error('Failed to update billing:', error);
    }
  };

  const handleMarkPaid = async (id: number) => {
    try {
      await adminApi.post(`/billing/${id}/mark-paid`, {});
      fetchBillings();
      fetchStats();
    } catch (error) {
      console.error('Failed to mark as paid:', error);
    }
  };

  const handleDelete = async (id: number) => {
    if (!window.confirm('Delete this billing record?')) return;
    try {
      await adminApi.delete(`/billing/${id}`);
      fetchBillings();
      fetchStats();
    } catch (error) {
      console.error('Failed to delete billing:', error);
    }
  };

  const handleExport = async () => {
    try {
      const params = statusFilter ? `?status=${statusFilter}` : '';
      const response = await adminApi.get(`/billing/export${params}`, { responseType: 'blob' });
      const url = window.URL.createObjectURL(new Blob([response.data]));
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', 'billings.csv');
      document.body.appendChild(link);
      link.click();
      link.remove();
    } catch (error) {
      console.error('Failed to export billings:', error);
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'paid': return 'bg-green-100 text-green-800';
      case 'pending': return 'bg-yellow-100 text-yellow-800';
      case 'overdue': return 'bg-red-100 text-red-800';
      case 'cancelled': return 'bg-gray-100 text-gray-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  if (loading) return <div className="p-8">Loading...</div>;

  return (
    <div className="p-8">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold">Billing Management</h1>
        <div className="flex gap-2">
          <button onClick={handleExport} className="px-4 py-2 bg-gray-600 text-white rounded hover:bg-gray-700">
            Export CSV
          </button>
          <button onClick={() => setShowCreateModal(true)} className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700">
            Create Billing
          </button>
        </div>
      </div>

      {stats && (
        <div className="grid grid-cols-4 gap-4 mb-6">
          <div className="bg-white p-4 rounded shadow">
            <div className="text-sm text-gray-600">Total Revenue</div>
            <div className="text-2xl font-bold">${stats.total_revenue.toFixed(2)}</div>
          </div>
          <div className="bg-white p-4 rounded shadow">
            <div className="text-sm text-gray-600">Pending</div>
            <div className="text-2xl font-bold text-yellow-600">${stats.pending_amount.toFixed(2)}</div>
          </div>
          <div className="bg-white p-4 rounded shadow">
            <div className="text-sm text-gray-600">Overdue</div>
            <div className="text-2xl font-bold text-red-600">${stats.overdue_amount.toFixed(2)}</div>
          </div>
          <div className="bg-white p-4 rounded shadow">
            <div className="text-sm text-gray-600">Monthly Revenue</div>
            <div className="text-2xl font-bold text-green-600">${stats.monthly_revenue.toFixed(2)}</div>
          </div>
        </div>
      )}

      <div className="bg-white rounded shadow mb-4 p-4">
        <select value={statusFilter} onChange={(e) => setStatusFilter(e.target.value)} className="px-4 py-2 border rounded">
          <option value="">All Status</option>
          <option value="pending">Pending</option>
          <option value="paid">Paid</option>
          <option value="overdue">Overdue</option>
          <option value="cancelled">Cancelled</option>
        </select>
      </div>

      <div className="bg-white rounded shadow overflow-hidden">
        <table className="min-w-full">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">ID</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Tenant</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Plan</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Amount</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Due Date</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-200">
            {billings.map((billing) => (
              <tr key={billing.id}>
                <td className="px-6 py-4 whitespace-nowrap text-sm">{billing.id}</td>
                <td className="px-6 py-4 whitespace-nowrap text-sm">{billing.tenant_name}</td>
                <td className="px-6 py-4 whitespace-nowrap text-sm">{billing.plan}</td>
                <td className="px-6 py-4 whitespace-nowrap text-sm">${billing.amount.toFixed(2)}</td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <span className={`px-2 py-1 text-xs rounded ${getStatusColor(billing.status)}`}>
                    {billing.status}
                  </span>
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm">{billing.due_date}</td>
                <td className="px-6 py-4 whitespace-nowrap text-sm space-x-2">
                  {billing.status === 'pending' && (
                    <button onClick={() => handleMarkPaid(billing.id)} className="text-green-600 hover:text-green-800">
                      Mark Paid
                    </button>
                  )}
                  <button onClick={() => { setSelectedBilling(billing); setShowEditModal(true); }} className="text-blue-600 hover:text-blue-800">
                    Edit
                  </button>
                  <button onClick={() => handleDelete(billing.id)} className="text-red-600 hover:text-red-800">
                    Delete
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {showCreateModal && <CreateBillingModal onClose={() => setShowCreateModal(false)} onCreate={handleCreate} />}
      {showEditModal && selectedBilling && (
        <EditBillingModal billing={selectedBilling} onClose={() => setShowEditModal(false)} onUpdate={handleUpdate} />
      )}
    </div>
  );
};

const CreateBillingModal: React.FC<{ onClose: () => void; onCreate: (data: any) => void }> = ({ onClose, onCreate }) => {
  const [formData, setFormData] = useState({
    tenant_id: '',
    plan: 'professional',
    amount: '',
    currency: 'USD',
    due_date: '',
    invoice_url: '',
    notes: ''
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onCreate({
      ...formData,
      tenant_id: parseInt(formData.tenant_id),
      amount: parseFloat(formData.amount)
    });
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
      <div className="bg-white rounded-lg p-6 w-96">
        <h2 className="text-xl font-bold mb-4">Create Billing</h2>
        <form onSubmit={handleSubmit} className="space-y-4">
          <input type="number" placeholder="Tenant ID" value={formData.tenant_id} onChange={(e) => setFormData({ ...formData, tenant_id: e.target.value })} className="w-full px-3 py-2 border rounded" required />
          <select value={formData.plan} onChange={(e) => setFormData({ ...formData, plan: e.target.value })} className="w-full px-3 py-2 border rounded">
            <option value="free">Free</option>
            <option value="professional">Professional</option>
            <option value="enterprise">Enterprise</option>
          </select>
          <input type="number" step="0.01" placeholder="Amount" value={formData.amount} onChange={(e) => setFormData({ ...formData, amount: e.target.value })} className="w-full px-3 py-2 border rounded" required />
          <input type="date" placeholder="Due Date" value={formData.due_date} onChange={(e) => setFormData({ ...formData, due_date: e.target.value })} className="w-full px-3 py-2 border rounded" required />
          <input type="url" placeholder="Invoice URL" value={formData.invoice_url} onChange={(e) => setFormData({ ...formData, invoice_url: e.target.value })} className="w-full px-3 py-2 border rounded" />
          <textarea placeholder="Notes" value={formData.notes} onChange={(e) => setFormData({ ...formData, notes: e.target.value })} className="w-full px-3 py-2 border rounded" rows={3} />
          <div className="flex gap-2">
            <button type="submit" className="flex-1 px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700">Create</button>
            <button type="button" onClick={onClose} className="flex-1 px-4 py-2 bg-gray-300 rounded hover:bg-gray-400">Cancel</button>
          </div>
        </form>
      </div>
    </div>
  );
};

const EditBillingModal: React.FC<{ billing: Billing; onClose: () => void; onUpdate: (id: number, data: any) => void }> = ({ billing, onClose, onUpdate }) => {
  const [formData, setFormData] = useState({
    status: billing.status,
    amount: billing.amount.toString(),
    due_date: billing.due_date,
    invoice_url: billing.invoice_url,
    notes: billing.notes
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onUpdate(billing.id, {
      status: formData.status,
      amount: parseFloat(formData.amount),
      due_date: formData.due_date,
      invoice_url: formData.invoice_url,
      notes: formData.notes
    });
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">
      <div className="bg-white rounded-lg p-6 w-96">
        <h2 className="text-xl font-bold mb-4">Edit Billing</h2>
        <form onSubmit={handleSubmit} className="space-y-4">
          <select value={formData.status} onChange={(e) => setFormData({ ...formData, status: e.target.value })} className="w-full px-3 py-2 border rounded">
            <option value="pending">Pending</option>
            <option value="paid">Paid</option>
            <option value="overdue">Overdue</option>
            <option value="cancelled">Cancelled</option>
          </select>
          <input type="number" step="0.01" placeholder="Amount" value={formData.amount} onChange={(e) => setFormData({ ...formData, amount: e.target.value })} className="w-full px-3 py-2 border rounded" required />
          <input type="date" placeholder="Due Date" value={formData.due_date} onChange={(e) => setFormData({ ...formData, due_date: e.target.value })} className="w-full px-3 py-2 border rounded" required />
          <input type="url" placeholder="Invoice URL" value={formData.invoice_url} onChange={(e) => setFormData({ ...formData, invoice_url: e.target.value })} className="w-full px-3 py-2 border rounded" />
          <textarea placeholder="Notes" value={formData.notes} onChange={(e) => setFormData({ ...formData, notes: e.target.value })} className="w-full px-3 py-2 border rounded" rows={3} />
          <div className="flex gap-2">
            <button type="submit" className="flex-1 px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700">Update</button>
            <button type="button" onClick={onClose} className="flex-1 px-4 py-2 bg-gray-300 rounded hover:bg-gray-400">Cancel</button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default AdminBilling;
