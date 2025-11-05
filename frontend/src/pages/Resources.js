import React, { useState, useEffect } from 'react';
import { apiService } from '../services/api';

const Resources = () => {
  const [resources, setResources] = useState([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState('all');

  useEffect(() => {
    fetchResources();
  }, []);

  const fetchResources = async () => {
    try {
      const data = await apiService.getResources();
      setResources(data);
      setLoading(false);
    } catch (error) {
      console.error('Failed to fetch resources:', error);
      setLoading(false);
    }
  };

  const filteredResources = resources.filter(resource => {
    if (filter === 'all') return true;
    return resource.state === filter;
  });

  const handleEmergencyStop = async () => {
    if (window.confirm('Are you sure you want to trigger emergency stop? This will terminate all running instances.')) {
      try {
        await apiService.emergencyStop();
        alert('Emergency stop triggered successfully');
        fetchResources();
      } catch (error) {
        alert('Failed to trigger emergency stop');
      }
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-lg">Loading resources...</div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold text-gray-900">AWS Resources</h1>
        <div className="flex space-x-4">
          <button 
            onClick={fetchResources}
            className="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors"
          >
            🔄 Refresh
          </button>
          <button 
            onClick={handleEmergencyStop}
            className="bg-red-600 text-white px-4 py-2 rounded-lg hover:bg-red-700 transition-colors"
          >
            🚨 Emergency Stop
          </button>
        </div>
      </div>

      {/* Filters */}
      <div className="bg-white p-4 rounded-lg shadow-md">
        <div className="flex space-x-4">
          <button
            onClick={() => setFilter('all')}
            className={`px-4 py-2 rounded-lg transition-colors ${
              filter === 'all' 
                ? 'bg-blue-600 text-white' 
                : 'bg-gray-200 text-gray-700 hover:bg-gray-300'
            }`}
          >
            All ({resources.length})
          </button>
          <button
            onClick={() => setFilter('running')}
            className={`px-4 py-2 rounded-lg transition-colors ${
              filter === 'running' 
                ? 'bg-green-600 text-white' 
                : 'bg-gray-200 text-gray-700 hover:bg-gray-300'
            }`}
          >
            Running ({resources.filter(r => r.state === 'running').length})
          </button>
          <button
            onClick={() => setFilter('stopped')}
            className={`px-4 py-2 rounded-lg transition-colors ${
              filter === 'stopped' 
                ? 'bg-gray-600 text-white' 
                : 'bg-gray-200 text-gray-700 hover:bg-gray-300'
            }`}
          >
            Stopped ({resources.filter(r => r.state === 'stopped').length})
          </button>
        </div>
      </div>

      {/* Resources Table */}
      <div className="bg-white rounded-lg shadow-md overflow-hidden">
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Instance ID</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Type</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Region</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Launch Time</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Monthly Cost</th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Environment</th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {filteredResources.map((resource) => (
                <tr key={resource.instance_id} className="hover:bg-gray-50">
                  <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                    {resource.instance_id}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    <span className="bg-blue-100 text-blue-800 px-2 py-1 rounded-full text-xs">
                      {resource.instance_type}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {resource.region}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <span className={`px-2 py-1 text-xs font-semibold rounded-full ${
                      resource.state === 'running' 
                        ? 'bg-green-100 text-green-800' 
                        : resource.state === 'stopped'
                        ? 'bg-yellow-100 text-yellow-800'
                        : 'bg-gray-100 text-gray-800'
                    }`}>
                      {resource.state}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {resource.launch_time ? new Date(resource.launch_time).toLocaleDateString() : 'N/A'}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    <span className="font-semibold">
                      ${(resource.estimated_monthly_cost || 0).toFixed(2)}
                    </span>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                    {resource.environment || 'N/A'}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {filteredResources.length === 0 && (
        <div className="text-center py-12">
          <div className="text-gray-500 text-lg">No resources found</div>
        </div>
      )}
    </div>
  );
};

export default Resources;