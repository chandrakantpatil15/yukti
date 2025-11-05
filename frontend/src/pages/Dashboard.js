import React, { useState, useEffect } from 'react';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell } from 'recharts';
import { apiService } from '../services/api';

const Dashboard = () => {
  const [stats, setStats] = useState({
    totalResources: 0,
    runningResources: 0,
    monthlyCost: 0,
    potentialSavings: 0
  });
  const [resources, setResources] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchDashboardData();
  }, []);

  const fetchDashboardData = async () => {
    try {
      const [resourcesData, healthData] = await Promise.all([
        apiService.getResources(),
        apiService.getHealth()
      ]);
      
      setResources(resourcesData);
      
      // Calculate stats
      const running = resourcesData.filter(r => r.state === 'running').length;
      const totalCost = resourcesData.reduce((sum, r) => sum + (r.estimated_monthly_cost || 0), 0);
      
      setStats({
        totalResources: resourcesData.length,
        runningResources: running,
        monthlyCost: totalCost,
        potentialSavings: totalCost * 0.3 // 30% potential savings
      });
      
      setLoading(false);
    } catch (error) {
      console.error('Failed to fetch dashboard data:', error);
      setLoading(false);
    }
  };

  const costByType = resources.reduce((acc, resource) => {
    const type = resource.instance_type;
    if (!acc[type]) acc[type] = 0;
    acc[type] += resource.estimated_monthly_cost || 0;
    return acc;
  }, {});

  const chartData = Object.entries(costByType).map(([type, cost]) => ({
    name: type,
    cost: parseFloat(cost.toFixed(2))
  }));

  const pieData = [
    { name: 'Running', value: stats.runningResources, color: '#10B981' },
    { name: 'Stopped', value: stats.totalResources - stats.runningResources, color: '#6B7280' }
  ];

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
          <p className="mt-4 text-gray-600">Loading dashboard...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold text-gray-900">FinOps Dashboard</h1>
        <button 
          onClick={fetchDashboardData}
          className="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors"
        >
          🔄 Refresh
        </button>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
        <div className="bg-white p-6 rounded-lg shadow-md">
          <div className="flex items-center">
            <div className="text-3xl mr-4">🖥️</div>
            <div>
              <p className="text-sm font-medium text-gray-600">Total Resources</p>
              <p className="text-2xl font-bold text-gray-900">{stats.totalResources}</p>
            </div>
          </div>
        </div>

        <div className="bg-white p-6 rounded-lg shadow-md">
          <div className="flex items-center">
            <div className="text-3xl mr-4">🟢</div>
            <div>
              <p className="text-sm font-medium text-gray-600">Running</p>
              <p className="text-2xl font-bold text-green-600">{stats.runningResources}</p>
            </div>
          </div>
        </div>

        <div className="bg-white p-6 rounded-lg shadow-md">
          <div className="flex items-center">
            <div className="text-3xl mr-4">💰</div>
            <div>
              <p className="text-sm font-medium text-gray-600">Monthly Cost</p>
              <p className="text-2xl font-bold text-blue-600">${stats.monthlyCost.toFixed(2)}</p>
            </div>
          </div>
        </div>

        <div className="bg-white p-6 rounded-lg shadow-md">
          <div className="flex items-center">
            <div className="text-3xl mr-4">💡</div>
            <div>
              <p className="text-sm font-medium text-gray-600">Potential Savings</p>
              <p className="text-2xl font-bold text-orange-600">${stats.potentialSavings.toFixed(2)}</p>
            </div>
          </div>
        </div>
      </div>

      {/* Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="bg-white p-6 rounded-lg shadow-md">
          <h3 className="text-lg font-semibold mb-4">Cost by Instance Type</h3>
          <ResponsiveContainer width="100%" height={300}>
            <BarChart data={chartData}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="name" />
              <YAxis />
              <Tooltip formatter={(value) => [`$${value}`, 'Monthly Cost']} />
              <Bar dataKey="cost" fill="#3B82F6" />
            </BarChart>
          </ResponsiveContainer>
        </div>

        <div className="bg-white p-6 rounded-lg shadow-md">
          <h3 className="text-lg font-semibold mb-4">Resource Status</h3>
          <ResponsiveContainer width="100%" height={300}>
            <PieChart>
              <Pie
                data={pieData}
                cx="50%"
                cy="50%"
                outerRadius={80}
                dataKey="value"
                label={({ name, value }) => `${name}: ${value}`}
              >
                {pieData.map((entry, index) => (
                  <Cell key={`cell-${index}`} fill={entry.color} />
                ))}
              </Pie>
              <Tooltip />
            </PieChart>
          </ResponsiveContainer>
        </div>
      </div>

      {/* Recent Activity */}
      <div className="bg-white p-6 rounded-lg shadow-md">
        <h3 className="text-lg font-semibold mb-4">Recent Resources</h3>
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Instance</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Type</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Monthly Cost</th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {resources.slice(0, 5).map((resource) => (
                <tr key={resource.instance_id}>
                  <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                    {resource.instance_id}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {resource.instance_type}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className={`px-2 py-1 text-xs font-semibold rounded-full ${
                      resource.state === 'running' 
                        ? 'bg-green-100 text-green-800' 
                        : 'bg-gray-100 text-gray-800'
                    }`}>
                      {resource.state}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    ${(resource.estimated_monthly_cost || 0).toFixed(2)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
};

export default Dashboard;