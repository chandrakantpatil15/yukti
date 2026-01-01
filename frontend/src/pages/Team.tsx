import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from 'react-query';
import { Users, UserPlus, Search, AlertCircle } from 'lucide-react';
import Loading from '../components/Loading';
import api from '../services/api';
import { MemberCard } from '../components/Team/MemberCard';
import { InvitationCard } from '../components/Team/InvitationCard';
import { InviteModal } from '../components/Team/InviteModal';
import { getCurrentUser } from '../lib/auth';

interface TeamMember {
  user_id: string;
  email: string;
  first_name?: string;
  last_name?: string;
  role: string;
  is_active: boolean;
  joined_at: string;
}

interface Invitation {
  id: string;
  email: string;
  role: string;
  status: string;
  expires_at: string;
  created_at: string;
}

const Team: React.FC = () => {
  const queryClient = useQueryClient();
  const currentUser = getCurrentUser();
  const [searchTerm, setSearchTerm] = useState('');
  const [showInviteModal, setShowInviteModal] = useState(false);

  // Fetch team members
  const { data: membersData, isLoading: membersLoading } = useQuery<{ success: boolean; members: TeamMember[] }>(
    'team-members',
    async () => {
      const response = await api.getTeamMembers();
      return response;
    },
    {
      retry: 2,
      refetchOnWindowFocus: false,
    }
  );

  // Fetch pending invitations
  const { data: invitationsData, isLoading: invitationsLoading } = useQuery<{ success: boolean; invitations: Invitation[] }>(
    'team-invitations',
    async () => {
      const response = await api.getTeamInvitations();
      return response;
    },
    {
      retry: 2,
      refetchOnWindowFocus: false,
    }
  );

  // Invite user mutation
  const inviteMutation = useMutation(
    ({ email, role }: { email: string; role: string }) => api.inviteUser(email, role),
    {
      onSuccess: () => {
        queryClient.invalidateQueries('team-invitations');
        setShowInviteModal(false);
      },
      onError: (error: any) => {
        console.error('Failed to invite user:', error);
        throw error;
      },
    }
  );

  // Update role mutation
  const updateRoleMutation = useMutation(
    ({ userId, role }: { userId: string; role: string }) => api.updateMemberRole(userId, role),
    {
      onSuccess: () => {
        queryClient.invalidateQueries('team-members');
      },
    }
  );

  // Remove member mutation
  const removeMemberMutation = useMutation(
    (userId: string) => api.removeMember(userId),
    {
      onSuccess: () => {
        queryClient.invalidateQueries('team-members');
      },
    }
  );

  // Resend invitation mutation
  const resendInvitationMutation = useMutation(
    (invitationId: string) => api.resendInvitation(invitationId),
    {
      onSuccess: () => {
        queryClient.invalidateQueries('team-invitations');
      },
    }
  );

  // Revoke invitation mutation
  const revokeInvitationMutation = useMutation(
    (invitationId: string) => api.revokeInvitation(invitationId),
    {
      onSuccess: () => {
        queryClient.invalidateQueries('team-invitations');
      },
    }
  );

  const handleInvite = async (email: string, role: string) => {
    await inviteMutation.mutateAsync({ email, role });
  };

  const handleUpdateRole = async (userId: string, newRole: string) => {
    await updateRoleMutation.mutateAsync({ userId, role: newRole });
  };

  const handleRemove = async (userId: string) => {
    await removeMemberMutation.mutateAsync(userId);
  };

  const handleResend = async (invitationId: string) => {
    await resendInvitationMutation.mutateAsync(invitationId);
  };

  const handleRevoke = async (invitationId: string) => {
    await revokeInvitationMutation.mutateAsync(invitationId);
  };

  // Filter members by search term
  const filteredMembers = membersData?.members?.filter((member) => {
    const search = searchTerm.toLowerCase();
    return (
      member.email.toLowerCase().includes(search) ||
      member.first_name?.toLowerCase().includes(search) ||
      member.last_name?.toLowerCase().includes(search)
    );
  }) || [];

  const members = membersData?.members || [];
  const invitations = invitationsData?.invitations || [];
  const isLoading = membersLoading || invitationsLoading;

  // Check if current user can manage team
  const canManageTeam = currentUser?.role === 'admin' || currentUser?.role === 'owner';

  if (isLoading) {
    return <Loading message="Loading team members..." />;
  }

  return (
    <div className="min-h-screen bg-gray-50 p-6">
      <div className="max-w-7xl mx-auto">
        {/* Header */}
        <div className="mb-8">
          <div className="flex items-center justify-between mb-4">
            <div>
              <h1 className="text-3xl font-bold text-gray-900 mb-2">Team Management</h1>
              <p className="text-gray-600">Manage your team members and invitations</p>
            </div>
            {canManageTeam && (
              <button
                onClick={() => setShowInviteModal(true)}
                className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors flex items-center gap-2"
              >
                <UserPlus className="w-5 h-5" />
                Invite Member
              </button>
            )}
          </div>
        </div>

        {/* Team Members Section */}
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 mb-6">
          <div className="flex items-center justify-between mb-4">
            <div className="flex items-center gap-2">
              <Users className="w-5 h-5 text-gray-600" />
              <h2 className="text-xl font-semibold text-gray-900">
                Team Members ({members.length})
              </h2>
            </div>

            {/* Search */}
            {members.length > 0 && (
              <div className="relative w-64">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-400" />
                <input
                  type="text"
                  placeholder="Search members..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                />
              </div>
            )}
          </div>

          {members.length === 0 ? (
            <div className="text-center py-12">
              <Users className="w-12 h-12 text-gray-400 mx-auto mb-4" />
              <p className="text-gray-600">No team members yet</p>
            </div>
          ) : (
            <div className="space-y-3">
              {filteredMembers.map((member) => (
                <MemberCard
                  key={member.user_id}
                  member={member}
                  onUpdateRole={handleUpdateRole}
                  onRemove={handleRemove}
                />
              ))}
              {filteredMembers.length === 0 && searchTerm && (
                <div className="text-center py-8">
                  <AlertCircle className="w-8 h-8 text-gray-400 mx-auto mb-2" />
                  <p className="text-gray-600">No members found matching "{searchTerm}"</p>
                </div>
              )}
            </div>
          )}
        </div>

        {/* Pending Invitations Section */}
        {canManageTeam && (
          <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
            <div className="flex items-center gap-2 mb-4">
              <UserPlus className="w-5 h-5 text-gray-600" />
              <h2 className="text-xl font-semibold text-gray-900">
                Pending Invitations ({invitations.length})
              </h2>
            </div>

            {invitations.length === 0 ? (
              <div className="text-center py-12">
                <UserPlus className="w-12 h-12 text-gray-400 mx-auto mb-4" />
                <p className="text-gray-600">No pending invitations</p>
              </div>
            ) : (
              <div className="space-y-3">
                {invitations.map((invitation) => (
                  <InvitationCard
                    key={invitation.id}
                    invitation={invitation}
                    onResend={handleResend}
                    onRevoke={handleRevoke}
                  />
                ))}
              </div>
            )}
          </div>
        )}

        {/* Invite Modal */}
        <InviteModal
          isOpen={showInviteModal}
          onClose={() => setShowInviteModal(false)}
          onInvite={handleInvite}
        />
      </div>
    </div>
  );
};

export default Team;

