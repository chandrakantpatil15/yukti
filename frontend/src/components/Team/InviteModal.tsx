import React, { useState } from 'react';
import { X, Mail, Shield, Loader2 } from 'lucide-react';

interface InviteModalProps {
  isOpen: boolean;
  onClose: () => void;
  onInvite: (email: string, role: string) => Promise<void>;
}

export const InviteModal: React.FC<InviteModalProps> = ({ isOpen, onClose, onInvite }) => {
  const [email, setEmail] = useState('');
  const [role, setRole] = useState('viewer');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  if (!isOpen) return null;

  const validateEmail = (email: string) => {
    return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!email.trim()) {
      setError('Email is required');
      return;
    }

    if (!validateEmail(email)) {
      setError('Please enter a valid email address');
      return;
    }

    setLoading(true);
    try {
      await onInvite(email.trim().toLowerCase(), role);
      // Reset form on success
      setEmail('');
      setRole('viewer');
      setError(null);
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to send invitation');
    } finally {
      setLoading(false);
    }
  };

  const handleClose = () => {
    if (!loading) {
      setEmail('');
      setRole('viewer');
      setError(null);
      onClose();
    }
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-xl shadow-xl max-w-md w-full mx-4 max-h-[90vh] overflow-y-auto">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-gray-200">
          <h2 className="text-xl font-semibold text-gray-900">Invite Team Member</h2>
          <button
            onClick={handleClose}
            disabled={loading}
            className="p-2 rounded-lg hover:bg-gray-100 transition-colors disabled:opacity-50"
          >
            <X className="w-5 h-5 text-gray-600" />
          </button>
        </div>

        {/* Form */}
        <form onSubmit={handleSubmit} className="p-6">
          {error && (
            <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg">
              <p className="text-sm text-red-800">{error}</p>
            </div>
          )}

          {/* Email Input */}
          <div className="mb-4">
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Email Address
            </label>
            <div className="relative">
              <Mail className="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-gray-400" />
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="user@example.com"
                required
                disabled={loading}
                className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent disabled:bg-gray-100 disabled:cursor-not-allowed"
              />
            </div>
          </div>

          {/* Role Selector */}
          <div className="mb-6">
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Role
            </label>
            <div className="space-y-2">
              {[
                { value: 'viewer', label: 'Viewer', description: 'Read-only access' },
                { value: 'editor', label: 'Editor', description: 'Can view and take actions' },
                { value: 'admin', label: 'Admin', description: 'Full access except billing' },
              ].map((roleOption) => (
                <label
                  key={roleOption.value}
                  className={`flex items-start p-3 rounded-lg border-2 cursor-pointer transition-colors ${
                    role === roleOption.value
                      ? 'border-blue-500 bg-blue-50'
                      : 'border-gray-200 hover:border-gray-300'
                  }`}
                >
                  <input
                    type="radio"
                    name="role"
                    value={roleOption.value}
                    checked={role === roleOption.value}
                    onChange={(e) => setRole(e.target.value)}
                    disabled={loading}
                    className="mt-1 mr-3"
                  />
                  <div className="flex-1">
                    <div className="flex items-center gap-2">
                      <Shield className="w-4 h-4 text-gray-600" />
                      <span className="font-medium text-gray-900">{roleOption.label}</span>
                    </div>
                    <p className="text-xs text-gray-500 mt-1">{roleOption.description}</p>
                  </div>
                </label>
              ))}
            </div>
          </div>

          {/* Actions */}
          <div className="flex justify-end gap-3">
            <button
              type="button"
              onClick={handleClose}
              disabled={loading}
              className="px-4 py-2 text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={loading || !email.trim()}
              className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors disabled:bg-gray-300 disabled:cursor-not-allowed flex items-center gap-2"
            >
              {loading ? (
                <>
                  <Loader2 className="w-4 h-4 animate-spin" />
                  Sending...
                </>
              ) : (
                'Send Invitation'
              )}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

