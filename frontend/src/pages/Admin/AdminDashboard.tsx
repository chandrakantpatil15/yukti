import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import adminApi from '../../services/adminApi';

interface PlatformStats {
  total_tenants: number;
  active_tenants: number;
  total_users: number;
  total_resources: number;
  total_findings: number;
  total_savings: number;
}

export default function AdminDashboard() {
  const [stats, setStats] = useState<PlatformStats | null>(null);
  const [loading, setLoading] = useState(true);
  const navigate = useNavigate();

  useEffect(() => {
    fetchStats();
  }, []);

  const fetchStats = async () => {
    try {
      const response = await adminApi.get('/api/admin/stats');
      setStats(response.data);
    } catch (err) {
      console.error('Failed to fetch stats:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleLogout = () => {
    localStorage.removeItem('admin_token');
    localStorage.removeItem('admin_user');
    navigate('/admin/login');
  };

  if (loading) return <div className="p-6">Loading...</div>;

  return (
    <div className="min-h-screen bg-gray-50">
      <nav className="bg-white shadow px-6 py-4 flex justify-between items-center">
        <h1 className="text-xl font-bold">Admin Portal</h1>
        <button
          onClick={handleLogout}
          className="px-4 py-2 text-sm bg-gray-200 rounded hover:bg-gray-300"
        >
          Logout
        </button>
      </nav>

      <div className="p-6">
        <h2 className="text-2xl font-bold mb-6">Platform Overview</h2>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
          <StatCard
            title="Total Tenants"
            value={stats?.total_tenants || 0}
            subtitle={`${stats?.active_tenants || 0} active`}
          />
          <StatCard
            title="Total Users"
            value={stats?.total_users || 0}
          />
          <StatCard
            title="Total Resources"
            value={stats?.total_resources || 0}
          />
          <StatCard
            title="Total Findings"
            value={stats?.total_findings || 0}
          />
          <StatCard
            title="Total Savings"
            value={`$${(stats?.total_savings || 0).toFixed(2)}`}
          />
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <QuickAction
            title="Manage Tenants"
            description="View and manage all customer tenants"
            onClick={() => navigate('/admin/tenants')}
          />
          <QuickAction
            title="Manage Users"
            description="View and manage all platform users"
            onClick={() => navigate('/admin/users')}
          />
          <QuickAction
            title="View Analytics"
            description="Platform growth and resource metrics"
            onClick={() => navigate('/admin/analytics')}
          />
        </div>
      </div>
    </div>
  );
}

function StatCard({ title, value, subtitle }: any) {
  return (
    <div className="bg-white p-6 rounded-lg shadow">
      <h3 className="text-sm text-gray-600 mb-2">{title}</h3>
      <p className="text-3xl font-bold">{value}</p>
      {subtitle && <p className="text-sm text-gray-500 mt-1">{subtitle}</p>}
    </div>
  );
}

function QuickAction({ title, description, onClick }: any) {
  return (
    <div
      onClick={onClick}
      className="bg-white p-6 rounded-lg shadow cursor-pointer hover:shadow-lg transition"
    >
      <h3 className="text-lg font-bold mb-2">{title}</h3>
      <p className="text-gray-600">{description}</p>
    </div>
  );
}
