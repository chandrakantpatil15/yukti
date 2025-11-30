import React, { useState, useEffect } from 'react';
import { User, Cloud, Shield, Copy, CheckCircle, AlertCircle, Calendar, Globe, Settings, Edit3 } from 'lucide-react';
import Loading from '../components/Loading';
import api from '../services/api';
import { getCurrentUser } from '../lib/auth';

interface UserProfile {
  email: string;
  tenant_id: number;
  role: string;
  created_at: string;
}

interface AWSConnection {
  account_id: string;
  role_name: string;
  role_arn: string;
  verified: boolean;
  last_verified_at: string;
  regions: string[];
}

const Profile: React.FC = () => {
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [awsConnection, setAwsConnection] = useState<AWSConnection | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [copiedField, setCopiedField] = useState<string | null>(null);

  useEffect(() => {
    fetchProfileData();
  }, []);

  const fetchProfileData = async () => {
    try {
      const user = getCurrentUser();
      if (!user) {
        setError('User not authenticated');
        return;
      }

      setProfile({
        email: user.email,
        tenant_id: user.tenant_id,
        role: user.role || 'user',
        created_at: new Date().toISOString(), // This would come from API in real implementation
      });

      // Fetch AWS connection details
      const awsResponse = await api.getAWSConnection();
      if (awsResponse?.success) {
        setAwsConnection(awsResponse.data);
      }
    } catch (error) {
      console.error('Error fetching profile:', error);
      setError('Failed to load profile data');
    } finally {
      setLoading(false);
    }
  };

  const copyToClipboard = async (text: string, field: string) => {
    try {
      await navigator.clipboard.writeText(text);
      setCopiedField(field);
      setTimeout(() => setCopiedField(null), 2000);
    } catch (error) {
      console.error('Failed to copy:', error);
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  if (loading) {
    return <Loading message="Loading profile..." />;
  }

  if (error) {
    return (
      <div className="min-h-screen bg-gray-50 p-6 flex items-center justify-center">
        <div className="bg-red-50 border border-red-200 rounded-xl p-8 max-w-md shadow-lg">
          <AlertCircle className="w-12 h-12 text-red-500 mx-auto mb-4" />
          <h2 className="text-xl font-semibold text-red-800 text-center mb-2">Error Loading Profile</h2>
          <p className="text-red-700 text-center mb-6">{error}</p>
          <button
            onClick={fetchProfileData}
            className="w-full px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 transition-colors"
          >
            Retry
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 p-6">
      <div className="max-w-4xl mx-auto">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-4xl font-bold text-gray-900 mb-2">Account Profile</h1>
          <p className="text-gray-600 text-lg">Manage your account settings and AWS configuration</p>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
          {/* User Information */}
          <div className="bg-white rounded-xl shadow-lg border border-gray-100 p-6">
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-xl font-bold text-gray-900 flex items-center gap-3">
                <div className="bg-blue-100 p-2 rounded-lg">
                  <User className="w-6 h-6 text-blue-600" />
                </div>
                User Information
              </h2>
              <button className="text-blue-600 hover:text-blue-700 p-2 rounded-lg hover:bg-blue-50 transition-colors">
                <Edit3 className="w-4 h-4" />
              </button>
            </div>

            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-500 mb-1">Email Address</label>
                <div className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                  <span className="text-gray-900 font-medium">{profile?.email}</span>
                  <button
                    onClick={() => copyToClipboard(profile?.email || '', 'email')}
                    className="text-gray-500 hover:text-gray-700 p-1 rounded"
                  >
                    {copiedField === 'email' ? (
                      <CheckCircle className="w-4 h-4 text-green-500" />
                    ) : (
                      <Copy className="w-4 h-4" />
                    )}
                  </button>
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-500 mb-1">Tenant ID</label>
                <div className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                  <span className="text-gray-900 font-medium">{profile?.tenant_id}</span>
                  <button
                    onClick={() => copyToClipboard(String(profile?.tenant_id), 'tenant_id')}
                    className="text-gray-500 hover:text-gray-700 p-1 rounded"
                  >
                    {copiedField === 'tenant_id' ? (
                      <CheckCircle className="w-4 h-4 text-green-500" />
                    ) : (
                      <Copy className="w-4 h-4" />
                    )}
                  </button>
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-500 mb-1">Role</label>
                <div className="p-3 bg-gray-50 rounded-lg">
                  <span className={`inline-flex items-center px-3 py-1 rounded-full text-sm font-medium ${
                    profile?.role === 'admin' 
                      ? 'bg-purple-100 text-purple-800' 
                      : 'bg-blue-100 text-blue-800'
                  }`}>
                    <Shield className="w-3 h-3 mr-1" />
                    {profile?.role?.charAt(0).toUpperCase() + profile?.role?.slice(1)}
                  </span>
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-500 mb-1">Member Since</label>
                <div className="flex items-center p-3 bg-gray-50 rounded-lg">
                  <Calendar className="w-4 h-4 text-gray-500 mr-2" />
                  <span className="text-gray-900">{formatDate(profile?.created_at || '')}</span>
                </div>
              </div>
            </div>
          </div>

          {/* AWS Configuration */}
          <div className="bg-white rounded-xl shadow-lg border border-gray-100 p-6">
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-xl font-bold text-gray-900 flex items-center gap-3">
                <div className="bg-orange-100 p-2 rounded-lg">
                  <Cloud className="w-6 h-6 text-orange-600" />
                </div>
                AWS Configuration
              </h2>
              <button 
                onClick={() => window.location.href = '/onboarding'}
                className="text-orange-600 hover:text-orange-700 p-2 rounded-lg hover:bg-orange-50 transition-colors"
              >
                <Settings className="w-4 h-4" />
              </button>
            </div>

            {awsConnection ? (
              <div className="space-y-4">
                <div className="flex items-center justify-between mb-4">
                  <div className={`flex items-center gap-2 px-3 py-1 rounded-full text-sm font-medium ${
                    awsConnection.verified 
                      ? 'bg-green-100 text-green-800' 
                      : 'bg-red-100 text-red-800'
                  }`}>
                    <div className={`w-2 h-2 rounded-full ${
                      awsConnection.verified ? 'bg-green-500' : 'bg-red-500'
                    }`} />
                    {awsConnection.verified ? 'Connected' : 'Disconnected'}
                  </div>
                  <span className="text-xs text-gray-500">
                    Last verified: {formatDate(awsConnection.last_verified_at)}
                  </span>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-500 mb-1">AWS Account ID</label>
                  <div className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                    <span className="text-gray-900 font-mono">{awsConnection.account_id}</span>
                    <button
                      onClick={() => copyToClipboard(awsConnection.account_id, 'account_id')}
                      className="text-gray-500 hover:text-gray-700 p-1 rounded"
                    >
                      {copiedField === 'account_id' ? (
                        <CheckCircle className="w-4 h-4 text-green-500" />
                      ) : (
                        <Copy className="w-4 h-4" />
                      )}
                    </button>
                  </div>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-500 mb-1">IAM Role Name</label>
                  <div className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                    <span className="text-gray-900 font-mono">{awsConnection.role_name}</span>
                    <button
                      onClick={() => copyToClipboard(awsConnection.role_name, 'role_name')}
                      className="text-gray-500 hover:text-gray-700 p-1 rounded"
                    >
                      {copiedField === 'role_name' ? (
                        <CheckCircle className="w-4 h-4 text-green-500" />
                      ) : (
                        <Copy className="w-4 h-4" />
                      )}
                    </button>
                  </div>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-500 mb-1">Full Role ARN</label>
                  <div className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                    <span className="text-gray-900 font-mono text-sm break-all">{awsConnection.role_arn}</span>
                    <button
                      onClick={() => copyToClipboard(awsConnection.role_arn, 'role_arn')}
                      className="text-gray-500 hover:text-gray-700 p-1 rounded ml-2 flex-shrink-0"
                    >
                      {copiedField === 'role_arn' ? (
                        <CheckCircle className="w-4 h-4 text-green-500" />
                      ) : (
                        <Copy className="w-4 h-4" />
                      )}
                    </button>
                  </div>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-500 mb-1">Regions</label>
                  <div className="p-3 bg-gray-50 rounded-lg">
                    <div className="flex flex-wrap gap-2">
                      {awsConnection.regions.map((region) => (
                        <span
                          key={region}
                          className="inline-flex items-center px-2 py-1 bg-blue-100 text-blue-800 text-xs font-medium rounded-full"
                        >
                          <Globe className="w-3 h-3 mr-1" />
                          {region}
                        </span>
                      ))}
                    </div>
                  </div>
                </div>
              </div>
            ) : (
              <div className="text-center py-8">
                <Cloud className="w-16 h-16 text-gray-300 mx-auto mb-4" />
                <h3 className="text-lg font-medium text-gray-900 mb-2">No AWS Connection</h3>
                <p className="text-gray-600 mb-4">Connect your AWS account to start monitoring costs</p>
                <button
                  onClick={() => window.location.href = '/onboarding'}
                  className="px-6 py-2 bg-orange-600 text-white rounded-lg hover:bg-orange-700 transition-colors"
                >
                  Configure AWS
                </button>
              </div>
            )}
          </div>
        </div>

        {/* Action Buttons */}
        <div className="mt-8 flex justify-center gap-4">
          <button
            onClick={() => window.location.href = '/onboarding'}
            className="px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors flex items-center gap-2"
          >
            <Settings className="w-4 h-4" />
            Account Settings
          </button>
          <button
            onClick={() => window.location.href = '/dashboard'}
            className="px-6 py-3 bg-gray-600 text-white rounded-lg hover:bg-gray-700 transition-colors"
          >
            Back to Dashboard
          </button>
        </div>
      </div>
    </div>
  );
};

export default Profile;