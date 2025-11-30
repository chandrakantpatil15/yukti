import { useState } from 'react';
import adminApi from '../../services/adminApi';

interface Props {
  userId: string;
  tenantId: string;
  userEmail: string;
  onSuccess: (token: string) => void;
  onCancel: () => void;
}

export default function ImpersonationModal({ userId, tenantId, userEmail, onSuccess, onCancel }: Props) {
  const [reason, setReason] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError('');

    try {
      const response = await adminApi.post('/api/admin/impersonate', {
        user_id: userId,
        tenant_id: tenantId,
        reason
      });
      
      onSuccess(response.data.impersonation_token);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Impersonation failed');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg p-6 max-w-md w-full">
        <h3 className="text-lg font-bold mb-4">Impersonate User</h3>
        
        <div className="mb-4 p-3 bg-gray-50 rounded">
          <p className="text-sm text-gray-600">User: <span className="font-medium">{userEmail}</span></p>
          <p className="text-sm text-gray-600">Tenant ID: <span className="font-medium">{tenantId}</span></p>
        </div>
        
        <form onSubmit={handleSubmit}>
          <div className="mb-4">
            <label className="block text-sm font-medium mb-2">
              Reason for Impersonation <span className="text-red-500">*</span>
            </label>
            <textarea
              value={reason}
              onChange={(e) => setReason(e.target.value)}
              className="w-full px-3 py-2 border rounded"
              rows={3}
              placeholder="e.g., Customer support request #1234"
              required
            />
            <p className="text-xs text-gray-500 mt-1">
              This will be logged for compliance and audit purposes
            </p>
          </div>

          {error && (
            <div className="mb-4 p-3 bg-red-50 text-red-600 rounded text-sm">
              {error}
            </div>
          )}

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
