import React from 'react';
import { X, Tag, DollarSign, Calendar, MapPin, Server, Database, HardDrive } from 'lucide-react';

interface ResourcePanelProps {
  resource: any;
  isOpen: boolean;
  onClose: () => void;
}

export const ResourcePanel: React.FC<ResourcePanelProps> = ({ resource, isOpen, onClose }) => {
  if (!isOpen || !resource) return null;

  const getResourceIcon = (type: string) => {
    switch (type.toLowerCase()) {
      case 'ec2': return <Server className="w-5 h-5" />;
      case 'rds': return <Database className="w-5 h-5" />;
      case 's3': return <HardDrive className="w-5 h-5" />;
      default: return <Server className="w-5 h-5" />;
    }
  };

  const getResourceColor = (type: string) => {
    switch (type.toLowerCase()) {
      case 'ec2': return 'text-orange-600 bg-orange-100';
      case 'rds': return 'text-blue-600 bg-blue-100';
      case 's3': return 'text-green-600 bg-green-100';
      default: return 'text-gray-600 bg-gray-100';
    }
  };

  return (
    <div className="fixed inset-y-0 right-0 w-96 bg-white shadow-2xl border-l border-gray-200 z-50 overflow-y-auto">
      {/* Header */}
      <div className="sticky top-0 bg-white border-b border-gray-200 p-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-3">
            <div className={`p-2 rounded-lg ${getResourceColor(resource.type)}`}>
              {getResourceIcon(resource.type)}
            </div>
            <div>
              <h2 className="text-lg font-semibold text-gray-900">Resource Details</h2>
              <p className="text-sm text-gray-500">{resource.type?.toUpperCase()}</p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
          >
            <X className="w-5 h-5 text-gray-500" />
          </button>
        </div>
      </div>

      {/* Content */}
      <div className="p-4 space-y-6">
        {/* Basic Info */}
        <div>
          <h3 className="text-sm font-medium text-gray-900 mb-3">Basic Information</h3>
          <div className="space-y-3">
            <div>
              <label className="block text-xs font-medium text-gray-500 mb-1">Resource ID</label>
              <p className="text-sm text-gray-900 font-mono bg-gray-50 p-2 rounded">{resource.id}</p>
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-500 mb-1">Name</label>
              <p className="text-sm text-gray-900">{resource.name || 'N/A'}</p>
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-500 mb-1">Type</label>
              <span className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${getResourceColor(resource.type)}`}>
                {getResourceIcon(resource.type)}
                <span className="ml-1">{resource.type?.toUpperCase()}</span>
              </span>
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-500 mb-1">Region</label>
              <div className="flex items-center space-x-1">
                <MapPin className="w-3 h-3 text-gray-400" />
                <span className="text-sm text-gray-900">{resource.region || 'N/A'}</span>
              </div>
            </div>
          </div>
        </div>

        {/* Cost Information */}
        <div>
          <h3 className="text-sm font-medium text-gray-900 mb-3">Cost Information</h3>
          <div className="space-y-3">
            <div className="bg-green-50 border border-green-200 rounded-lg p-3">
              <div className="flex items-center space-x-2">
                <DollarSign className="w-4 h-4 text-green-600" />
                <span className="text-sm font-medium text-green-800">Monthly Cost</span>
              </div>
              <p className="text-lg font-bold text-green-900 mt-1">
                ${resource.monthly_cost?.toFixed(2) || '0.00'}
              </p>
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-500 mb-1">Instance Type</label>
              <p className="text-sm text-gray-900">{resource.instance_type || 'N/A'}</p>
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-500 mb-1">State</label>
              <span className={`inline-flex px-2 py-1 rounded-full text-xs font-medium ${
                resource.state === 'running' ? 'bg-green-100 text-green-800' :
                resource.state === 'stopped' ? 'bg-red-100 text-red-800' :
                'bg-gray-100 text-gray-800'
              }`}>
                {resource.state || 'Unknown'}
              </span>
            </div>
          </div>
        </div>

        {/* Tags */}
        <div>
          <h3 className="text-sm font-medium text-gray-900 mb-3">Tags</h3>
          {resource.tags && Object.keys(resource.tags).length > 0 ? (
            <div className="space-y-2">
              {Object.entries(resource.tags).map(([key, value]) => (
                <div key={key} className="flex items-center justify-between p-2 bg-gray-50 rounded">
                  <div className="flex items-center space-x-2">
                    <Tag className="w-3 h-3 text-gray-400" />
                    <span className="text-xs font-medium text-gray-600">{key}</span>
                  </div>
                  <span className="text-xs text-gray-900">{String(value)}</span>
                </div>
              ))}
            </div>
          ) : (
            <p className="text-sm text-gray-500 italic">No tags available</p>
          )}
        </div>

        {/* Metadata */}
        <div>
          <h3 className="text-sm font-medium text-gray-900 mb-3">Metadata</h3>
          <div className="space-y-3">
            <div>
              <label className="block text-xs font-medium text-gray-500 mb-1">Created</label>
              <div className="flex items-center space-x-1">
                <Calendar className="w-3 h-3 text-gray-400" />
                <span className="text-sm text-gray-900">
                  {resource.created_at ? new Date(resource.created_at).toLocaleDateString() : 'N/A'}
                </span>
              </div>
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-500 mb-1">Last Updated</label>
              <div className="flex items-center space-x-1">
                <Calendar className="w-3 h-3 text-gray-400" />
                <span className="text-sm text-gray-900">
                  {resource.updated_at ? new Date(resource.updated_at).toLocaleDateString() : 'N/A'}
                </span>
              </div>
            </div>
          </div>
        </div>

        {/* Raw Data */}
        <div>
          <h3 className="text-sm font-medium text-gray-900 mb-3">Raw Data</h3>
          <div className="bg-gray-50 rounded-lg p-3">
            <pre className="text-xs text-gray-600 overflow-x-auto">
              {JSON.stringify(resource, null, 2)}
            </pre>
          </div>
        </div>
      </div>
    </div>
  );
};