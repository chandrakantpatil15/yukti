import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import adminApi from '../../services/adminApi';

interface Analytics {
  new_tenants_30d: number;
  new_users_30d: number;
  active_scans_7d: number;
  total_resources: number;
  total_findings: number;
  avg_savings_per_tenant: number;
}

export default function AdminAnalytics() {
  const [analytics, setAnalytics] = useState<Analytics | null>(null);
  const [loading, setLoading] = useState(true);
  const navigate = useNavigate();

  useEffect(() => {
    fetchAnalytics();
  }, []);

  const fetchAnalytics = async () => {
    try {
      const response = await adminApi.get('/api/admin/analytics');
      setAnalytics(response.data);
    } catch (err) {
      console.error('Failed to fetch analytics:', err);
    } finally {
      setLoading(false);
    }
  };

  if (loading) return <div className="p-6">Loading...</div>;

  return (
    <div className="min-h-screen bg-gray-50">
      <nav className="bg-white shadow px-6 py-4">
        <button
          onClick={() => navigate('/admin/dashboard')}
          className="text-blue-600 hover:underline"
        >
          ← Back to Dashboard
        </button>
      </nav>

      <div className="p-6">
        <h2 className="text-2xl font-bold mb-6">Platform Analytics</h2>
        
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div className="bg-white p-6 rounded-lg shadow">
            <h3 className="font-bold mb-4 text-lg">Growth Metrics</h3>
            <div className="space-y-3">
              <MetricRow 
                label="New Tenants (30d)" 
                value={analytics?.new_tenants_30d || 0} 
              />
              <MetricRow 
                label="New Users (30d)" 
                value={analytics?.new_users_30d || 0} 
              />
              <MetricRow 
                label="Active Scans (7d)" 
                value={analytics?.active_scans_7d || 0} 
              />
            </div>
          </div>

          <div className="bg-white p-6 rounded-lg shadow">
            <h3 className="font-bold mb-4 text-lg">Resource Metrics</h3>
            <div className="space-y-3">
              <MetricRow 
                label="Total Resources" 
                value={analytics?.total_resources || 0} 
              />
              <MetricRow 
                label="Total Findings" 
                value={analytics?.total_findings || 0} 
              />
              <MetricRow 
                label="Avg Savings/Tenant" 
                value={`$${(analytics?.avg_savings_per_tenant || 0).toFixed(2)}`} 
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

function MetricRow({ label, value }: { label: string; value: string | number }) {
  return (
    <div className="flex justify-between items-center py-2 border-b border-gray-100">
      <span className="text-gray-600">{label}</span>
      <span className="font-semibold text-lg">{value}</span>
    </div>
  );
}
