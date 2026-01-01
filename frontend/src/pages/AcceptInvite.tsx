import React, { useEffect, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { useQuery, useMutation } from 'react-query';
import { CheckCircle, AlertCircle, Loader2, LogIn } from 'lucide-react';
import api from '../services/api';
import { isAuthenticated } from '../lib/auth';

interface InviteDetails {
  success: boolean;
  email?: string;
  role?: string;
  tenant_name?: string;
  expires_at?: string;
  message?: string;
}

const AcceptInvite: React.FC = () => {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const token = searchParams.get('token');
  const [loggedIn, setLoggedIn] = useState(isAuthenticated());

  // Fetch invitation details (public endpoint)
  const { data: inviteDetails, isLoading, error } = useQuery<InviteDetails>(
    ['invite-details', token],
    async () => {
      if (!token) throw new Error('No token provided');
      return await api.getInviteDetails(token);
    },
    {
      enabled: !!token,
      retry: false,
    }
  );

  // Accept invitation mutation
  const acceptMutation = useMutation(
    async () => {
      if (!token) throw new Error('No token provided');
      return await api.acceptInvite(token);
    },
    {
      onSuccess: (data) => {
        // Redirect to dashboard after successful acceptance
        setTimeout(() => {
          navigate('/dashboard');
        }, 2000);
      },
    }
  );

  useEffect(() => {
    // Check if user is logged in
    setLoggedIn(isAuthenticated());
  }, []);

  const handleAccept = () => {
    if (token) {
      acceptMutation.mutate();
    }
  };

  const handleLogin = () => {
    // Store the invite token in localStorage so we can redirect back after login
    if (token) {
      localStorage.setItem('pending_invite_token', token);
    }
    navigate('/login');
  };

  if (!token) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
        <div className="bg-white rounded-xl shadow-lg p-8 max-w-md w-full text-center">
          <AlertCircle className="w-16 h-16 text-red-500 mx-auto mb-4" />
          <h1 className="text-2xl font-bold text-gray-900 mb-2">Invalid Invitation</h1>
          <p className="text-gray-600 mb-6">
            This invitation link is invalid or missing a token.
          </p>
          <button
            onClick={() => navigate('/login')}
            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
          >
            Go to Login
          </button>
        </div>
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <Loader2 className="w-12 h-12 text-blue-600 animate-spin mx-auto mb-4" />
          <p className="text-gray-600">Loading invitation details...</p>
        </div>
      </div>
    );
  }

  if (error || !inviteDetails?.success) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
        <div className="bg-white rounded-xl shadow-lg p-8 max-w-md w-full text-center">
          <AlertCircle className="w-16 h-16 text-red-500 mx-auto mb-4" />
          <h1 className="text-2xl font-bold text-gray-900 mb-2">Invalid Invitation</h1>
          <p className="text-gray-600 mb-6">
            {inviteDetails?.message || 'This invitation is invalid, expired, or has already been used.'}
          </p>
          <button
            onClick={() => navigate('/login')}
            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
          >
            Go to Login
          </button>
        </div>
      </div>
    );
  }

  const getRoleDisplayName = (role?: string) => {
    if (!role) return '';
    return role.charAt(0).toUpperCase() + role.slice(1).toLowerCase();
  };

  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
      <div className="bg-white rounded-xl shadow-lg p-8 max-w-md w-full">
        {acceptMutation.isSuccess ? (
          <div className="text-center">
            <CheckCircle className="w-16 h-16 text-green-500 mx-auto mb-4" />
            <h1 className="text-2xl font-bold text-gray-900 mb-2">Invitation Accepted!</h1>
            <p className="text-gray-600 mb-6">
              You've successfully joined <strong>{inviteDetails.tenant_name}</strong>.
              Redirecting to dashboard...
            </p>
          </div>
        ) : (
          <>
            <div className="text-center mb-6">
              <CheckCircle className="w-16 h-16 text-blue-500 mx-auto mb-4" />
              <h1 className="text-2xl font-bold text-gray-900 mb-2">You've been invited!</h1>
              <p className="text-gray-600">
                You've been invited to join <strong>{inviteDetails.tenant_name}</strong> as a{' '}
                <strong>{getRoleDisplayName(inviteDetails.role)}</strong>.
              </p>
            </div>

            <div className="bg-gray-50 rounded-lg p-4 mb-6">
              <div className="space-y-2 text-sm">
                <div className="flex justify-between">
                  <span className="text-gray-600">Organization:</span>
                  <span className="font-medium text-gray-900">{inviteDetails.tenant_name}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600">Role:</span>
                  <span className="font-medium text-gray-900">{getRoleDisplayName(inviteDetails.role)}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600">Email:</span>
                  <span className="font-medium text-gray-900">{inviteDetails.email}</span>
                </div>
              </div>
            </div>

            {!loggedIn ? (
              <div className="space-y-3">
                <p className="text-sm text-gray-600 text-center mb-4">
                  Please log in to accept this invitation.
                </p>
                <button
                  onClick={handleLogin}
                  className="w-full px-4 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors flex items-center justify-center gap-2"
                >
                  <LogIn className="w-5 h-5" />
                  Log In to Accept
                </button>
              </div>
            ) : (
              <div className="space-y-3">
                <button
                  onClick={handleAccept}
                  disabled={acceptMutation.isLoading}
                  className="w-full px-4 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors disabled:bg-gray-300 disabled:cursor-not-allowed flex items-center justify-center gap-2"
                >
                  {acceptMutation.isLoading ? (
                    <>
                      <Loader2 className="w-5 h-5 animate-spin" />
                      Accepting...
                    </>
                  ) : (
                    <>
                      <CheckCircle className="w-5 h-5" />
                      Accept Invitation
                    </>
                  )}
                </button>
                {acceptMutation.isError && (
                  <div className="p-3 bg-red-50 border border-red-200 rounded-lg">
                    <p className="text-sm text-red-800">
                      {acceptMutation.error instanceof Error
                        ? acceptMutation.error.message
                        : 'Failed to accept invitation. Please try again.'}
                    </p>
                  </div>
                )}
              </div>
            )}
          </>
        )}
      </div>
    </div>
  );
};

export default AcceptInvite;

