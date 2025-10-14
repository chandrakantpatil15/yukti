import React from 'react';
import { Resource } from '../types';

interface InstanceCardProps {
  resource: Resource;
  onClick: () => void;
}

export const InstanceCard: React.FC<InstanceCardProps> = ({ resource, onClick }) => {
  const getStatusColor = (status: string) => {
    switch (status) {
      case 'running': return 'bg-green-500';
      case 'stopped': return 'bg-red-500';
      default: return 'bg-gray-500';
    }
  };

  const getUtilizationColor = (utilization: number) => {
    if (utilization > 70) return 'text-green-600';
    if (utilization > 40) return 'text-yellow-600';
    return 'text-red-600';
  };

  return (
    <div 
      className="bg-white rounded-lg shadow-sm border border-gray-200 p-4 hover:shadow-md transition-shadow cursor-pointer"
      onClick={onClick}
    >
      <div className="flex justify-between items-start mb-3">
        <div>
          <h3 className="font-semibold text-gray-900 text-sm">{resource.resourceId}</h3>
          <p className="text-xs text-gray-500">{resource.instanceType}</p>
        </div>
        <div className={`w-3 h-3 rounded-full ${getStatusColor(resource.status)}`}></div>
      </div>
      
      <div className="space-y-2">
        <div className="flex justify-between items-center">
          <span className="text-xs text-gray-600">CPU</span>
          <span className={`text-xs font-medium ${getUtilizationColor(resource.cpuUtilization)}`}>
            {resource.cpuUtilization}%
          </span>
        </div>
        
        <div className="flex justify-between items-center">
          <span className="text-xs text-gray-600">Cost</span>
          <span className="text-xs font-medium text-gray-900">${resource.monthlyCost}</span>
        </div>
        
        <div className="flex justify-between items-center">
          <span className="text-xs text-gray-600">Env</span>
          <span className="text-xs font-medium text-gray-700 capitalize">{resource.environment}</span>
        </div>
      </div>
      
      {resource.cpuUtilization < 30 && (
        <div className="mt-3 px-2 py-1 bg-yellow-100 text-yellow-800 text-xs rounded">
          ⚠️ Low utilization
        </div>
      )}
      
      {resource.status === 'stopped' && (
        <div className="mt-3 px-2 py-1 bg-red-100 text-red-800 text-xs rounded">
          🛑 Stopped instance
        </div>
      )}
    </div>
  );
};