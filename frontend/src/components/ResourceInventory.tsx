import React, { useState } from 'react';
import { useResources } from '../hooks/useCostData';
import { useLargeInventory } from '../hooks/useLargeInventory';
import { Resource } from '../types';
import { ResourceDetails } from './ResourceDetails';
import { InstanceCard } from './InstanceCard';

export const ResourceInventory: React.FC = () => {
  const { data: resources, isLoading } = useResources();
  const [filter, setFilter] = useState('all');
  const [sortBy, setSortBy] = useState('cost');
  const [selectedResource, setSelectedResource] = useState<Resource | null>(null);
  const [viewMode, setViewMode] = useState<'table' | 'grid'>('table');

  if (isLoading) {
    return (
      <div className="bg-white rounded-lg shadow-sm p-6">
        <div className="animate-pulse">
          <div className="h-4 bg-gray-200 rounded w-1/4 mb-4"></div>
          <div className="space-y-3">
            {[...Array(5)].map((_, i) => (
              <div key={i} className="h-4 bg-gray-200 rounded"></div>
            ))}
          </div>
        </div>
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className="bg-white rounded-lg shadow-sm p-6">
        <div className="flex items-center justify-center py-12">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
          <span className="ml-3 text-gray-600">Loading resources from database...</span>
        </div>
      </div>
    );
  }

  if (!resources) return <div>No inventory data available</div>;

  const filteredResources = resources.filter(resource => {
    if (filter === 'all') return true;
    if (filter === 'running') return resource.status === 'running';
    if (filter === 'stopped') return resource.status === 'stopped';
    if (filter === 'high-cost') return resource.monthlyCost > 50;
    if (filter === 'low-utilization') return resource.cpuUtilization < 30;
    return true;
  });

  const sortedResources = [...filteredResources].sort((a, b) => {
    switch (sortBy) {
      case 'cost':
        return b.monthlyCost - a.monthlyCost;
      case 'utilization':
        return b.cpuUtilization - a.cpuUtilization;
      case 'name':
        return a.resourceId.localeCompare(b.resourceId);
      default:
        return 0;
    }
  });

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'running': return 'bg-green-100 text-green-800';
      case 'stopped': return 'bg-red-100 text-red-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  const getUtilizationColor = (utilization: number) => {
    if (utilization > 70) return 'text-green-600';
    if (utilization > 40) return 'text-yellow-600';
    return 'text-red-600';
  };

  const totalCost = resources.reduce((sum, r) => sum + r.monthlyCost, 0);
  const runningCount = resources.filter(r => r.status === 'running').length;
  const avgUtilization = resources.reduce((sum, r) => sum + r.cpuUtilization, 0) / resources.length;

  return (
    <div className="bg-white rounded-lg shadow-sm p-6">
      <div className="flex justify-between items-center mb-6">
        <div>
          <h2 className="text-xl font-semibold text-gray-900">Resource Inventory</h2>
          <p className="text-sm text-gray-600">
            {resources?.length || 0} resources from database
          </p>
        </div>
        <div className="flex space-x-4">
          <div className="flex border border-gray-300 rounded-md">
            <button
              onClick={() => setViewMode('table')}
              className={`px-3 py-2 text-sm ${viewMode === 'table' ? 'bg-blue-500 text-white' : 'text-gray-600 hover:bg-gray-50'}`}
            >
              📋 Table
            </button>
            <button
              onClick={() => setViewMode('grid')}
              className={`px-3 py-2 text-sm ${viewMode === 'grid' ? 'bg-blue-500 text-white' : 'text-gray-600 hover:bg-gray-50'}`}
            >
              🔲 Grid
            </button>
          </div>
          <select 
            value={filter} 
            onChange={(e) => setFilter(e.target.value)}
            className="border border-gray-300 rounded-md px-3 py-2 text-sm"
          >
            <option value="all">All Resources</option>
            <option value="running">Running Only</option>
            <option value="stopped">Stopped Only</option>
            <option value="high-cost">High Cost (&gt;$50)</option>
            <option value="low-utilization">Low Utilization (&lt;30%)</option>
          </select>
          <select 
            value={sortBy} 
            onChange={(e) => setSortBy(e.target.value)}
            className="border border-gray-300 rounded-md px-3 py-2 text-sm"
          >
            <option value="cost">Sort by Cost</option>
            <option value="utilization">Sort by Utilization</option>
            <option value="name">Sort by Name</option>
          </select>
        </div>
      </div>

      {/* Summary Stats */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
        <div className="bg-blue-50 p-4 rounded-lg">
          <div className="text-2xl font-bold text-blue-600">{resources.length}</div>
          <div className="text-sm text-blue-600">Total Resources</div>
        </div>
        <div className="bg-green-50 p-4 rounded-lg">
          <div className="text-2xl font-bold text-green-600">{runningCount}</div>
          <div className="text-sm text-green-600">Running</div>
        </div>
        <div className="bg-purple-50 p-4 rounded-lg">
          <div className="text-2xl font-bold text-purple-600">${totalCost.toFixed(2)}</div>
          <div className="text-sm text-purple-600">Total Monthly Cost</div>
        </div>
        <div className="bg-orange-50 p-4 rounded-lg">
          <div className="text-2xl font-bold text-orange-600">{avgUtilization.toFixed(1)}%</div>
          <div className="text-sm text-orange-600">Avg Utilization</div>
        </div>
      </div>

      {/* Resource Views */}
      {viewMode === 'grid' ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
          {sortedResources.map((resource) => (
            <InstanceCard
              key={resource.resourceId}
              resource={resource}
              onClick={() => setSelectedResource(resource)}
            />
          ))}
        </div>
      ) : (
        <div className="overflow-x-auto">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Resource ID
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Instance Type
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Status
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Environment
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                CPU Utilization
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Monthly Cost
              </th>
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {sortedResources.map((resource) => (
              <tr 
                key={resource.resourceId} 
                className="hover:bg-blue-50 cursor-pointer transition-colors"
                onClick={() => setSelectedResource(resource)}
              >
                <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-blue-600 hover:text-blue-800">
                  {resource.resourceId}
                  <div className="text-xs text-gray-500">Click for details</div>
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                  {resource.instanceType}
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${getStatusColor(resource.status)}`}>
                    {resource.status}
                  </span>
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                  {resource.environment}
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm">
                  <span className={`font-medium ${getUtilizationColor(resource.cpuUtilization)}`}>
                    {resource.cpuUtilization}%
                  </span>
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                  ${resource.monthlyCost.toFixed(2)}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        </div>
      )}

      {sortedResources.length === 0 && (
        <div className="text-center py-8 text-gray-500">
          No resources match the current filter criteria.
        </div>
      )}
      
      {/* Resource Details Modal */}
      {selectedResource && (
        <ResourceDetails 
          resource={selectedResource} 
          onClose={() => setSelectedResource(null)} 
        />
      )}
    </div>
  );
};