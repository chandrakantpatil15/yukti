import React, { useState, useEffect, useCallback } from 'react';
import { DollarSign, TrendingUp, AlertCircle, CheckCircle, RefreshCw, Cloud, User, Settings, Scan } from 'lucide-react';
import Loading from '../components/Loading';
import api from '../services/api';
import { getCurrentUser } from '../lib/auth';

interface DashboardData {
  total_savings: number;
  findings_count: number;
  budget_amount: number;
  current_spend: number;
  ri_savings: number;
}

interface AWSConnection {
  account_id: string;
  role_name: string;
  verified: boolean;
  last_verified_at: string;
}

interface ResourceData {
  ec2_count: number;
  rds_count: number;
  s3_count: number;
  total_count: number;
}

interface DashboardProps {
  onResourceClick?: (resource: any) => void;
}

const Dashboard: React.FC<DashboardProps> = ({ onResourceClick }) => {
  const [dashboardData, setDashboardData] = useState<DashboardData>({
    total_savings: 0,
    findings_count: 0,
    budget_amount: 0,
    current_spend: 0,
    ri_savings: 0,
  });
  const [awsConnection, setAwsConnection] = useState<AWSConnection | null>(null);
  const [resourceData, setResourceData] = useState<ResourceData | null>(null);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState('');
  const [lastUpdated, setLastUpdated] = useState<Date>(new Date());

  const fetchDashboardData = useCallback(async (isRefresh = false) => {
    try {
      if (isRefresh) {
        setRefreshing(true);
      } else {
        setLoading(true);
      }
      
      const user = getCurrentUser();
      if (!user) {
        setError('User not authenticated');
        return;
      }

      // Fetch dashboard data first
      const dashboardResponse = await api.getDashboard();
      
      // Try to fetch AWS connection and resources (optional)
      let awsResponse = null;
      let resourceResponse = null;
      
      try {
        awsResponse = await api.getAWSConnection();
      } catch (e) {
        // AWS connection not configured yet
      }
      
      try {
        resourceResponse = await api.getResourceStats();
      } catch (e) {
        // No resources yet
      }
      
      if (dashboardResponse.success) {
        setDashboardData(dashboardResponse.data);
        setLastUpdated(new Date());
        setError('');
      }
      
      if (awsResponse?.success) {
        setAwsConnection(awsResponse.data);
      }
      
      if (resourceResponse?.success) {
        setResourceData(resourceResponse.data);
      }
    } catch (error) {
      console.error('[Dashboard] Error fetching data:', error);
      if (!isRefresh) {
        setError('Failed to load dashboard data');
      }
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }, []);

  // Initial load
  useEffect(() => {
    fetchDashboardData();
  }, [fetchDashboardData]);

  // Auto-refresh every 1 minute
  useEffect(() => {
    const interval = setInterval(() => {
      fetchDashboardData(true);
    }, 60000); // 60 seconds

    return () => clearInterval(interval);
  }, [fetchDashboardData]);

  const handleManualRefresh = () => {
    fetchDashboardData(true);
  };

  const triggerScan = async () => {
    try {
      setRefreshing(true);
      console.log('[Dashboard] 🚀 Triggering AWS resource scan...');
      
      const response = await api.post('/api/v1/scan', {});
      console.log('[Dashboard] ✅ Scan response:', response);
      
      if (response.success) {
        alert('✅ Scan started! Scanning all AWS regions...\n\nThis will take 30-60 seconds.\nDashboard will auto-refresh when complete.');
        
        // Refresh every 5 seconds for 60 seconds to catch results
        let refreshCount = 0;
        const refreshInterval = setInterval(() => {
          refreshCount++;
          console.log(`[Dashboard] Auto-refresh ${refreshCount}/12`);
          fetchDashboardData(true);
          
          if (refreshCount >= 12) {
            clearInterval(refreshInterval);
            console.log('[Dashboard] ✅ Scan monitoring complete');
          }
        }, 5000);
      } else {
        alert(`❌ Scan failed: ${response.error}\n\nTroubleshooting:\n${response.action || 'Check AWS connection settings'}`);
      }
    } catch (error: any) {
      console.error('[Dashboard] ❌ Scan request failed:', error);
      
      let errorMessage = 'Failed to start scan';
      let troubleshooting = '';
      
      if (error.response?.data) {
        errorMessage = error.response.data.error || errorMessage;
        troubleshooting = error.response.data.action || '';
        
        if (error.response.data.code === 'NO_AWS_CONNECTION') {
          troubleshooting = 'Go to Settings > AWS Connection to configure your account';
        } else if (error.response.data.code === 'AWS_NOT_VERIFIED') {
          troubleshooting = 'Verify your IAM role trust policy and external ID';
        } else if (error.response.data.code === 'SCAN_IN_PROGRESS') {
          troubleshooting = 'A scan is already running. Please wait for it to complete.';
        } else if (error.response.data.code === 'SCAN_COOLDOWN') {
          const retryAfter = error.response.data.retry_after || 300;
          const minutes = Math.ceil(retryAfter / 60);
          troubleshooting = `Please wait ${minutes} minutes before scanning again to avoid AWS API limits.`;
        }
      }
      
      const message = troubleshooting 
        ? `${errorMessage}\n\nNext Steps:\n${troubleshooting}`
        : errorMessage;
        
      alert(message);
    } finally {
      setRefreshing(false);
    }
  };

  const checkScanStatus = async () => {
    try {
      const response = await api.get('/api/v1/scan/status');
      if (response.success) {
        console.log('[Dashboard] Scan status:', response.data);
        return response.data;
      }
    } catch (error) {
      console.error('[Dashboard] Failed to get scan status:', error);
    }
    return null;
  };

  const budgetPercentage = dashboardData.budget_amount > 0 
    ? (dashboardData.current_spend / dashboardData.budget_amount) * 100 
    : 0;

  const user = getCurrentUser();

  if (loading) {
    return <Loading message="Loading dashboard..." />;
  }

  if (error) {
    return (
      <div className="min-h-screen bg-gray-50 p-6 flex items-center justify-center">
        <div className="bg-red-50 border border-red-200 rounded-xl p-8 max-w-md shadow-lg">
          <AlertCircle className="w-12 h-12 text-red-500 mx-auto mb-4" />
          <h2 className="text-xl font-semibold text-red-800 text-center mb-2">Error Loading Dashboard</h2>
          <p className="text-red-700 text-center mb-6">{error}</p>
          <div className="flex gap-3 justify-center">
            <button
              onClick={() => fetchDashboardData()}
              className="px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 transition-colors"
            >
              Retry
            </button>
            <button
              onClick={() => window.location.href = '/login'}
              className="px-4 py-2 bg-gray-600 text-white rounded-lg hover:bg-gray-700 transition-colors"
            >
              Go to Login
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="bg-white min-h-screen">
      <div className="max-w-7xl mx-auto">
        {/* Header Section */}
        <div className="mb-8">
          <div className="flex items-center justify-between mb-4">
            <div>
              <h1 className="text-4xl font-bold text-gray-900 mb-2">FinOps Dashboard</h1>
              <p className="text-gray-600 text-lg">Welcome back, {user?.email || 'User'}!</p>
            </div>
            <div className="flex items-center gap-4">
              <div className="text-right">
                <p className="text-sm text-gray-500">Last updated</p>
                <p className="text-sm font-medium text-gray-700">
                  {lastUpdated.toLocaleTimeString()}
                </p>
              </div>
              <button
                onClick={handleManualRefresh}
                disabled={refreshing}
                className="flex items-center gap-2 px-4 py-2 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors disabled:opacity-50"
              >
                <RefreshCw className={`w-4 h-4 ${refreshing ? 'animate-spin' : ''}`} />
                {refreshing ? 'Refreshing...' : 'Refresh'}
              </button>
            </div>
          </div>
          
          {/* AWS Connection Status */}
          {awsConnection && (
            <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-4 mb-6">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <Cloud className="w-6 h-6 text-orange-500" />
                  <div>
                    <h3 className="font-semibold text-gray-900">AWS Connection</h3>
                    <p className="text-sm text-gray-600">
                      Account: {awsConnection.account_id} • Role: {awsConnection.role_name}
                    </p>
                  </div>
                </div>
                <div className="flex items-center gap-3">
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
                  <div className="flex items-center gap-2">
                    <button
                      onClick={triggerScan}
                      disabled={refreshing || !awsConnection.verified}
                      className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                      <Scan className="w-4 h-4" />
                      {refreshing ? 'Scanning...' : 'Scan Resources'}
                    </button>
                    <button
                      onClick={checkScanStatus}
                      className="flex items-center gap-2 px-3 py-2 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 transition-colors text-sm"
                      title="Check scan status and troubleshooting info"
                    >
                      <Settings className="w-4 h-4" />
                      Debug
                    </button>
                  </div>
                </div>
              </div>
            </div>
          )}
        </div>

        {/* Stats Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
          <div className="bg-white rounded-xl shadow-lg border border-gray-100 p-6 hover:shadow-xl transition-shadow">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-gray-500 text-sm font-medium uppercase tracking-wide">Total Savings</p>
                <p className="text-3xl font-bold text-green-600 mt-2">
                  ${dashboardData.total_savings.toLocaleString()}
                </p>
                <p className="text-sm text-gray-400 mt-1 flex items-center gap-1">
                  <TrendingUp className="w-3 h-3" />
                  per month
                </p>
              </div>
              <div className="bg-green-100 p-3 rounded-full">
                <DollarSign className="w-8 h-8 text-green-600" />
              </div>
            </div>
          </div>

          <div className="bg-white rounded-xl shadow-lg border border-gray-100 p-6 hover:shadow-xl transition-shadow">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-gray-500 text-sm font-medium uppercase tracking-wide">Findings</p>
                <p className="text-3xl font-bold text-blue-600 mt-2">
                  {dashboardData.findings_count}
                </p>
                <p className="text-sm text-gray-400 mt-1 flex items-center gap-1">
                  <AlertCircle className="w-3 h-3" />
                  opportunities
                </p>
              </div>
              <div className="bg-blue-100 p-3 rounded-full">
                <AlertCircle className="w-8 h-8 text-blue-600" />
              </div>
            </div>
          </div>

          <div className="bg-white rounded-xl shadow-lg border border-gray-100 p-6 hover:shadow-xl transition-shadow">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-gray-500 text-sm font-medium uppercase tracking-wide">Budget Usage</p>
                <p className={`text-3xl font-bold mt-2 ${
                  budgetPercentage > 90 ? 'text-red-600' :
                  budgetPercentage > 80 ? 'text-orange-600' :
                  'text-green-600'
                }`}>
                  {budgetPercentage.toFixed(0)}%
                </p>
                <p className="text-sm text-gray-400 mt-1">
                  ${dashboardData.current_spend.toLocaleString()} / ${dashboardData.budget_amount.toLocaleString()}
                </p>
              </div>
              <div className={`p-3 rounded-full ${
                budgetPercentage > 90 ? 'bg-red-100' :
                budgetPercentage > 80 ? 'bg-orange-100' :
                'bg-green-100'
              }`}>
                <TrendingUp className={`w-8 h-8 ${
                  budgetPercentage > 90 ? 'text-red-600' :
                  budgetPercentage > 80 ? 'text-orange-600' :
                  'text-green-600'
                }`} />
              </div>
            </div>
          </div>

          <div className="bg-white rounded-xl shadow-lg border border-gray-100 p-6 hover:shadow-xl transition-shadow">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-gray-500 text-sm font-medium uppercase tracking-wide">RI Savings</p>
                <p className="text-3xl font-bold text-purple-600 mt-2">
                  ${dashboardData.ri_savings.toLocaleString()}
                </p>
                <p className="text-sm text-gray-400 mt-1 flex items-center gap-1">
                  <CheckCircle className="w-3 h-3" />
                  potential
                </p>
              </div>
              <div className="bg-purple-100 p-3 rounded-full">
                <CheckCircle className="w-8 h-8 text-purple-600" />
              </div>
            </div>
          </div>
        </div>

        {/* Budget Progress */}
        {dashboardData.budget_amount > 0 && (
          <div className="bg-white rounded-xl shadow-lg border border-gray-100 p-6 mb-8">
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-xl font-bold text-gray-900">Budget Status</h2>
              {budgetPercentage > 80 && (
                <div className="flex items-center gap-2 px-3 py-1 bg-red-100 text-red-800 rounded-full text-sm font-medium">
                  <AlertCircle className="w-4 h-4" />
                  Alert Threshold Exceeded
                </div>
              )}
            </div>
            <div className="space-y-4">
              <div className="flex justify-between items-center">
                <span className="text-gray-600 font-medium">Current Spend</span>
                <span className="text-2xl font-bold text-gray-900">
                  ${dashboardData.current_spend.toLocaleString()}
                </span>
              </div>
              <div className="relative">
                <div className="w-full bg-gray-200 rounded-full h-6">
                  <div
                    className={`h-6 rounded-full transition-all duration-500 ${
                      budgetPercentage > 90 ? 'bg-gradient-to-r from-red-500 to-red-600' :
                      budgetPercentage > 80 ? 'bg-gradient-to-r from-orange-500 to-orange-600' :
                      'bg-gradient-to-r from-green-500 to-green-600'
                    }`}
                    style={{ width: `${Math.min(budgetPercentage, 100)}%` }}
                  />
                </div>
                <div className="absolute inset-0 flex items-center justify-center">
                  <span className="text-sm font-semibold text-white drop-shadow">
                    {budgetPercentage.toFixed(1)}%
                  </span>
                </div>
              </div>
              <div className="flex justify-between text-sm text-gray-500">
                <span>Budget: ${dashboardData.budget_amount.toLocaleString()}</span>
                <span>Remaining: ${Math.max(0, dashboardData.budget_amount - dashboardData.current_spend).toLocaleString()}</span>
              </div>
            </div>
          </div>
        )}

        {/* Resources Overview */}
        {resourceData && (
          <div className="bg-white rounded-xl shadow-lg border border-gray-100 p-6 mb-8">
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-xl font-bold text-gray-900">AWS Resources</h2>
              <button
                onClick={() => window.location.href = '/resources'}
                className="text-blue-600 hover:text-blue-700 text-sm font-medium"
              >
                View All Resources →
              </button>
            </div>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              <button
                onClick={() => onResourceClick && onResourceClick({ type: 'ec2', count: resourceData.ec2_count })}
                className="text-center p-4 bg-orange-50 rounded-lg hover:bg-orange-100 transition-colors cursor-pointer"
              >
                <div className="text-2xl font-bold text-orange-600">{resourceData.ec2_count}</div>
                <div className="text-sm text-gray-600">EC2 Instances</div>
              </button>
              <button
                onClick={() => onResourceClick && onResourceClick({ type: 'rds', count: resourceData.rds_count })}
                className="text-center p-4 bg-blue-50 rounded-lg hover:bg-blue-100 transition-colors cursor-pointer"
              >
                <div className="text-2xl font-bold text-blue-600">{resourceData.rds_count}</div>
                <div className="text-sm text-gray-600">RDS Databases</div>
              </button>
              <button
                onClick={() => onResourceClick && onResourceClick({ type: 's3', count: resourceData.s3_count })}
                className="text-center p-4 bg-green-50 rounded-lg hover:bg-green-100 transition-colors cursor-pointer"
              >
                <div className="text-2xl font-bold text-green-600">{resourceData.s3_count}</div>
                <div className="text-sm text-gray-600">S3 Buckets</div>
              </button>
              <button
                onClick={() => onResourceClick && onResourceClick({ type: 'total', count: resourceData.total_count })}
                className="text-center p-4 bg-purple-50 rounded-lg hover:bg-purple-100 transition-colors cursor-pointer"
              >
                <div className="text-2xl font-bold text-purple-600">{resourceData.total_count}</div>
                <div className="text-sm text-gray-600">Total Resources</div>
              </button>
            </div>
          </div>
        )}

        {/* Quick Actions */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <button
            onClick={() => window.location.href = '/hidden-costs'}
            className="group bg-white rounded-xl shadow-lg border border-gray-100 p-6 text-left hover:shadow-xl hover:border-blue-200 transition-all duration-200"
          >
            <div className="flex items-center justify-between mb-4">
              <div className="bg-blue-100 p-3 rounded-lg group-hover:bg-blue-200 transition-colors">
                <AlertCircle className="w-6 h-6 text-blue-600" />
              </div>
              <span className="text-2xl font-bold text-blue-600">{dashboardData.findings_count}</span>
            </div>
            <h3 className="text-lg font-bold text-gray-900 mb-2 group-hover:text-blue-600 transition-colors">
              View Hidden Costs
            </h3>
            <p className="text-gray-600 text-sm leading-relaxed">
              Explore {dashboardData.findings_count} cost optimization opportunities and start saving immediately
            </p>
          </button>

          <button
            onClick={() => window.location.href = '/resources'}
            className="group bg-white rounded-xl shadow-lg border border-gray-100 p-6 text-left hover:shadow-xl hover:border-purple-200 transition-all duration-200"
          >
            <div className="flex items-center justify-between mb-4">
              <div className="bg-purple-100 p-3 rounded-lg group-hover:bg-purple-200 transition-colors">
                <CheckCircle className="w-6 h-6 text-purple-600" />
              </div>
              <span className="text-2xl font-bold text-purple-600">${dashboardData.ri_savings.toLocaleString()}</span>
            </div>
            <h3 className="text-lg font-bold text-gray-900 mb-2 group-hover:text-purple-600 transition-colors">
              RI/SP Recommendations
            </h3>
            <p className="text-gray-600 text-sm leading-relaxed">
              Save ${dashboardData.ri_savings.toLocaleString()}/month with Reserved Instances and Savings Plans
            </p>
          </button>

          <button
            onClick={() => window.location.href = '/profile'}
            className="group bg-white rounded-xl shadow-lg border border-gray-100 p-6 text-left hover:shadow-xl hover:border-green-200 transition-all duration-200"
          >
            <div className="flex items-center justify-between mb-4">
              <div className="bg-green-100 p-3 rounded-lg group-hover:bg-green-200 transition-colors">
                <User className="w-6 h-6 text-green-600" />
              </div>
              <span className="text-2xl font-bold text-green-600">{budgetPercentage.toFixed(0)}%</span>
            </div>
            <h3 className="text-lg font-bold text-gray-900 mb-2 group-hover:text-green-600 transition-colors">
              Account Profile
            </h3>
            <p className="text-gray-600 text-sm leading-relaxed">
              View your AWS account details, role configuration, and tenant information
            </p>
          </button>
        </div>
      </div>
    </div>
  );
};

export default Dashboard;
