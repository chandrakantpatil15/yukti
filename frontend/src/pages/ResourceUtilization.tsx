import React, { useState, useEffect } from 'react';
import { Cpu, Activity, Network, Download, Power, Minimize2 } from 'lucide-react';
import Loading from '../components/Loading';
import api from '../services/api';

interface IdleResource {
  resource_id: string;
  resource_type: string;
  avg_cpu: number;
  monthly_cost: number;
}

interface Recommendation {
  resource_id: string;
  current_type: string;
  recommended_type: string;
  reason: string;
  monthly_savings: number;
}

const ResourceUtilization: React.FC = () => {
  const [timeRange, setTimeRange] = useState('7d');
  const [idleResources, setIdleResources] = useState<IdleResource[]>([]);
  const [recommendations, setRecommendations] = useState<Recommendation[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchData();
  }, [timeRange]);

  const fetchData = async () => {
    setLoading(true);
    try {
      const [idle, recs] = await Promise.all([
        api.getIdleResources(7),
        api.getRightSizingRecommendations()
      ]);
      setIdleResources(idle);
      setRecommendations(recs);
    } catch (error) {
      console.error('Failed to fetch utilization data:', error);
    } finally {
      setLoading(false);
    }
  };

  const totalSavings = idleResources.reduce((sum, r) => sum + r.monthly_cost, 0);

  if (loading) {
    return <Loading message="Loading resource utilization..." />;
  }

  return (
    <div className="bg-white min-h-screen">
      <div className="max-w-7xl mx-auto">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-4xl font-bold text-gray-900 mb-2">Resource Utilization</h1>
          <p className="text-gray-600 text-lg">Monitor and optimize your AWS resource performance</p>
        </div>

        {/* Filters */}
        <div className="bg-white rounded-xl shadow-lg border border-gray-100 p-6 mb-8">
          <div className="flex gap-4">
            <select className="border border-gray-300 rounded-lg px-4 py-2 text-sm focus:ring-2 focus:ring-blue-500 focus:border-transparent">
              <option>All Resources</option>
              <option>EC2 Only</option>
              <option>RDS Only</option>
            </select>
            <select className="border border-gray-300 rounded-lg px-4 py-2 text-sm focus:ring-2 focus:ring-blue-500 focus:border-transparent">
              <option>All Regions</option>
              <option>us-east-1</option>
              <option>us-west-2</option>
            </select>
            <select
              value={timeRange}
              onChange={(e) => setTimeRange(e.target.value)}
              className="border border-gray-300 rounded-lg px-4 py-2 text-sm focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            >
              <option value="7d">Last 7 Days</option>
              <option value="30d">Last 30 Days</option>
              <option value="90d">Last 90 Days</option>
            </select>
          </div>
        </div>

        {/* Idle Resources */}
        <div className="bg-white rounded-xl shadow-lg border border-gray-100 p-6 mb-8">
          <div className="flex items-center gap-3 mb-6">
            <Activity className="w-6 h-6 text-orange-500" />
            <h2 className="text-2xl font-bold text-gray-900">Idle Resources (CPU &lt; 10% for 7 days)</h2>
          </div>
          <div className="overflow-x-auto mb-6">
            <table className="w-full">
              <thead>
                <tr className="border-b-2 border-gray-200">
                  <th className="text-left py-3 px-4 text-sm font-semibold text-gray-700">Resource ID</th>
                  <th className="text-left py-3 px-4 text-sm font-semibold text-gray-700">Type</th>
                  <th className="text-right py-3 px-4 text-sm font-semibold text-gray-700">Avg CPU</th>
                  <th className="text-right py-3 px-4 text-sm font-semibold text-gray-700">Monthly Cost</th>
                </tr>
              </thead>
              <tbody>
                {idleResources.map((resource, idx) => (
                  <tr key={idx} className="border-b border-gray-100 hover:bg-gray-50 transition-colors">
                    <td className="py-3 px-4 font-mono text-sm text-gray-900">{resource.resource_id}</td>
                    <td className="py-3 px-4 text-gray-600">{resource.resource_type}</td>
                    <td className="text-right py-3 px-4">
                      <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-red-100 text-red-800">
                        {resource.avg_cpu.toFixed(1)}%
                      </span>
                    </td>
                    <td className="text-right py-3 px-4 font-semibold text-gray-900">
                      ${resource.monthly_cost}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          <div className="flex items-center justify-between pt-4 border-t border-gray-200">
            <div className="flex items-center gap-2">
              <span className="text-lg font-semibold text-gray-900">Total Potential Savings:</span>
              <span className="text-2xl font-bold text-green-600">${totalSavings}/month</span>
            </div>
            <div className="flex gap-2">
              <button className="flex items-center gap-2 px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 transition-colors font-medium">
                <Power className="w-4 h-4" />
                Stop All
              </button>
              <button className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors font-medium">
                <Minimize2 className="w-4 h-4" />
                Downsize All
              </button>
              <button className="flex items-center gap-2 px-4 py-2 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 transition-colors">
                <Download className="w-4 h-4" />
                Export Report
              </button>
            </div>
          </div>
        </div>

        {/* Right-Sizing Recommendations */}
        <div className="bg-white rounded-xl shadow-lg border border-gray-100 p-6">
          <div className="flex items-center gap-3 mb-6">
            <Cpu className="w-6 h-6 text-blue-500" />
            <h2 className="text-2xl font-bold text-gray-900">Right-Sizing Recommendations</h2>
          </div>
          <div className="space-y-4">
            {recommendations.length === 0 ? (
              <div className="text-center py-8 text-gray-500">
                <Cpu className="w-12 h-12 mx-auto mb-3 text-gray-300" />
                <p>No right-sizing recommendations available</p>
              </div>
            ) : (
              recommendations.map((rec, idx) => (
                <div key={idx} className="border border-gray-200 rounded-lg p-4 hover:border-blue-300 hover:shadow-md transition-all">
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <p className="font-mono text-sm text-gray-600 mb-2">{rec.resource_id}</p>
                      <div className="flex items-center gap-2 mb-2">
                        <span className="px-3 py-1 bg-gray-100 text-gray-700 rounded-lg text-sm font-medium">
                          {rec.current_type}
                        </span>
                        <span className="text-gray-400">→</span>
                        <span className="px-3 py-1 bg-blue-100 text-blue-700 rounded-lg text-sm font-semibold">
                          {rec.recommended_type}
                        </span>
                      </div>
                      <p className="text-sm text-gray-600 mb-2">
                        <span className="font-medium">Reason:</span> {rec.reason}
                      </p>
                      <p className="text-sm font-semibold text-green-600">
                        💰 Savings: ${rec.monthly_savings}/month
                      </p>
                    </div>
                    <div className="flex gap-2 ml-4">
                      <button className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm font-medium">
                        Apply
                      </button>
                      <button className="px-4 py-2 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 transition-colors text-sm">
                        Dismiss
                      </button>
                    </div>
                  </div>
                </div>
              ))
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default ResourceUtilization;
