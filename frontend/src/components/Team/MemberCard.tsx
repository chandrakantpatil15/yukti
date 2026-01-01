import React, { useState } from 'react';
import { User, MoreVertical, Edit2, Trash2, Shield } from 'lucide-react';
import { getCurrentUser } from '../../lib/auth';

interface TeamMember {
  user_id: string;
  email: string;
  first_name?: string;
  last_name?: string;
  role: string;
  is_active: boolean;
  joined_at: string;
}

interface MemberCardProps {
  member: TeamMember;
  onUpdateRole: (userId: string, newRole: string) => void;
  onRemove: (userId: string) => void;
}

const getRoleColor = (role: string) => {
  switch (role.toLowerCase()) {
    case 'owner':
      return 'bg-purple-100 text-purple-800 border-purple-200';
    case 'admin':
      return 'bg-blue-100 text-blue-800 border-blue-200';
    case 'editor':
      return 'bg-green-100 text-green-800 border-green-200';
    case 'viewer':
      return 'bg-gray-100 text-gray-800 border-gray-200';
    default:
      return 'bg-gray-100 text-gray-800 border-gray-200';
  }
};

const getRoleDisplayName = (role: string) => {
  return role.charAt(0).toUpperCase() + role.slice(1).toLowerCase();
};

const canManageUser = (currentRole: string | undefined, targetRole: string): boolean => {
  if (!currentRole) return false;
  const curr = currentRole.toLowerCase();
  const target = targetRole.toLowerCase();
  
  // Owner can manage everyone except other owners
  if (curr === 'owner' && target !== 'owner') return true;
  // Admin can manage editors and viewers
  if (curr === 'admin' && (target === 'editor' || target === 'viewer')) return true;
  return false;
};

export const MemberCard: React.FC<MemberCardProps> = ({ member, onUpdateRole, onRemove }) => {
  const [showMenu, setShowMenu] = useState(false);
  const [showRoleModal, setShowRoleModal] = useState(false);
  const [selectedRole, setSelectedRole] = useState(member.role);
  const currentUser = getCurrentUser();
  const canManage = canManageUser(currentUser?.role, member.role);
  const isOwner = member.role.toLowerCase() === 'owner';

  const handleRoleUpdate = () => {
    if (selectedRole !== member.role) {
      onUpdateRole(member.user_id, selectedRole);
    }
    setShowRoleModal(false);
    setShowMenu(false);
  };

  const handleRemove = () => {
    if (window.confirm(`Are you sure you want to remove ${member.email} from the team?`)) {
      onRemove(member.user_id);
    }
    setShowMenu(false);
  };

  const getInitials = (email: string, firstName?: string, lastName?: string) => {
    if (firstName && lastName) {
      return `${firstName[0]}${lastName[0]}`.toUpperCase();
    }
    return email.substring(0, 2).toUpperCase();
  };

  const displayName = member.first_name && member.last_name
    ? `${member.first_name} ${member.last_name}`
    : member.email;

  return (
    <>
      <div className="bg-white rounded-lg border border-gray-200 p-4 hover:shadow-md transition-shadow">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4 flex-1">
            {/* Avatar */}
            <div className="w-12 h-12 rounded-full bg-gradient-to-r from-blue-500 to-indigo-500 flex items-center justify-center text-white font-semibold">
              {getInitials(member.email, member.first_name, member.last_name)}
            </div>

            {/* User Info */}
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-2">
                <h3 className="text-sm font-semibold text-gray-900 truncate">{displayName}</h3>
                {!member.is_active && (
                  <span className="px-2 py-0.5 text-xs font-medium bg-red-100 text-red-800 rounded">
                    Inactive
                  </span>
                )}
              </div>
              <p className="text-sm text-gray-500 truncate">{member.email}</p>
              <p className="text-xs text-gray-400 mt-1">
                Joined {new Date(member.joined_at).toLocaleDateString()}
              </p>
            </div>

            {/* Role Badge */}
            <div className={`px-3 py-1 rounded-full text-xs font-semibold border ${getRoleColor(member.role)}`}>
              <div className="flex items-center gap-1">
                <Shield className="w-3 h-3" />
                {getRoleDisplayName(member.role)}
              </div>
            </div>
          </div>

          {/* Actions Menu */}
          {canManage && (
            <div className="relative ml-4">
              <button
                onClick={() => setShowMenu(!showMenu)}
                className="p-2 rounded-lg hover:bg-gray-100 transition-colors"
              >
                <MoreVertical className="w-5 h-5 text-gray-600" />
              </button>

              {showMenu && (
                <>
                  <div
                    className="fixed inset-0 z-10"
                    onClick={() => setShowMenu(false)}
                  />
                  <div className="absolute right-0 mt-2 w-48 bg-white rounded-lg shadow-lg border border-gray-200 z-20">
                    <button
                      onClick={() => {
                        setShowRoleModal(true);
                        setShowMenu(false);
                      }}
                      className="w-full px-4 py-2 text-left text-sm text-gray-700 hover:bg-gray-100 flex items-center gap-2"
                    >
                      <Edit2 className="w-4 h-4" />
                      Change Role
                    </button>
                    {!isOwner && (
                      <button
                        onClick={handleRemove}
                        className="w-full px-4 py-2 text-left text-sm text-red-600 hover:bg-red-50 flex items-center gap-2"
                      >
                        <Trash2 className="w-4 h-4" />
                        Remove from Team
                      </button>
                    )}
                  </div>
                </>
              )}
            </div>
          )}
        </div>
      </div>

      {/* Role Update Modal */}
      {showRoleModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 max-w-md w-full mx-4">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Change Role</h3>
            <p className="text-sm text-gray-600 mb-4">
              Select a new role for {member.email}
            </p>

            <div className="space-y-2 mb-6">
              {['viewer', 'editor', 'admin'].map((role) => (
                <label
                  key={role}
                  className={`flex items-center p-3 rounded-lg border-2 cursor-pointer transition-colors ${
                    selectedRole === role
                      ? 'border-blue-500 bg-blue-50'
                      : 'border-gray-200 hover:border-gray-300'
                  }`}
                >
                  <input
                    type="radio"
                    name="role"
                    value={role}
                    checked={selectedRole === role}
                    onChange={(e) => setSelectedRole(e.target.value)}
                    className="mr-3"
                  />
                  <div>
                    <div className="font-medium text-gray-900">{getRoleDisplayName(role)}</div>
                    <div className="text-xs text-gray-500">
                      {role === 'viewer' && 'Read-only access'}
                      {role === 'editor' && 'Can view and take actions'}
                      {role === 'admin' && 'Full access except billing'}
                    </div>
                  </div>
                </label>
              ))}
            </div>

            <div className="flex justify-end gap-3">
              <button
                onClick={() => {
                  setShowRoleModal(false);
                  setSelectedRole(member.role);
                }}
                className="px-4 py-2 text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200 transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={handleRoleUpdate}
                disabled={selectedRole === member.role}
                className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors disabled:bg-gray-300 disabled:cursor-not-allowed"
              >
                Update Role
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  );
};

