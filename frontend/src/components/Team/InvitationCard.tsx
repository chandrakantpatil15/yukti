import React from 'react';
import { Mail, Clock, X, RefreshCw } from 'lucide-react';

interface Invitation {
  id: string;
  email: string;
  role: string;
  status: string;
  expires_at: string;
  created_at: string;
}

interface InvitationCardProps {
  invitation: Invitation;
  onResend: (invitationId: string) => void;
  onRevoke: (invitationId: string) => void;
}

const getRoleColor = (role: string) => {
  switch (role.toLowerCase()) {
    case 'admin':
      return 'bg-blue-100 text-blue-800';
    case 'editor':
      return 'bg-green-100 text-green-800';
    case 'viewer':
      return 'bg-gray-100 text-gray-800';
    default:
      return 'bg-gray-100 text-gray-800';
  }
};

const getRoleDisplayName = (role: string) => {
  return role.charAt(0).toUpperCase() + role.slice(1).toLowerCase();
};

export const InvitationCard: React.FC<InvitationCardProps> = ({
  invitation,
  onResend,
  onRevoke,
}) => {
  const expiresAt = new Date(invitation.expires_at);
  const isExpired = expiresAt < new Date();
  const daysUntilExpiry = Math.ceil((expiresAt.getTime() - Date.now()) / (1000 * 60 * 60 * 24));

  const handleResend = () => {
    if (window.confirm(`Resend invitation to ${invitation.email}?`)) {
      onResend(invitation.id);
    }
  };

  const handleRevoke = () => {
    if (window.confirm(`Revoke invitation for ${invitation.email}?`)) {
      onRevoke(invitation.id);
    }
  };

  return (
    <div className="bg-white rounded-lg border border-gray-200 p-4 hover:shadow-md transition-shadow">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4 flex-1">
          {/* Icon */}
          <div className="w-12 h-12 rounded-full bg-yellow-100 flex items-center justify-center">
            <Mail className="w-6 h-6 text-yellow-600" />
          </div>

          {/* Invitation Info */}
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2 mb-1">
              <h3 className="text-sm font-semibold text-gray-900">{invitation.email}</h3>
              <span className={`px-2 py-0.5 rounded-full text-xs font-semibold ${getRoleColor(invitation.role)}`}>
                {getRoleDisplayName(invitation.role)}
              </span>
            </div>
            <div className="flex items-center gap-4 text-xs text-gray-500">
              <div className="flex items-center gap-1">
                <Clock className="w-3 h-3" />
                {isExpired ? (
                  <span className="text-red-600">Expired</span>
                ) : daysUntilExpiry > 0 ? (
                  <span>Expires in {daysUntilExpiry} day{daysUntilExpiry !== 1 ? 's' : ''}</span>
                ) : (
                  <span className="text-red-600">Expires today</span>
                )}
              </div>
              <span>Sent {new Date(invitation.created_at).toLocaleDateString()}</span>
            </div>
          </div>
        </div>

        {/* Actions */}
        <div className="flex items-center gap-2 ml-4">
          <button
            onClick={handleResend}
            className="p-2 rounded-lg hover:bg-blue-50 text-blue-600 transition-colors"
            title="Resend invitation"
          >
            <RefreshCw className="w-4 h-4" />
          </button>
          <button
            onClick={handleRevoke}
            className="p-2 rounded-lg hover:bg-red-50 text-red-600 transition-colors"
            title="Revoke invitation"
          >
            <X className="w-4 h-4" />
          </button>
        </div>
      </div>
    </div>
  );
};

