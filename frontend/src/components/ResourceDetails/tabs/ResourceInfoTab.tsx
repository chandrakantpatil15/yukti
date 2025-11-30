import React, { useState, useEffect } from 'react';
import { Loader2 } from 'lucide-react';
import api from '../../../services/api';

interface ResourceInfoTabProps {
  resourceId: string;
}

interface ResourceDetails {
  id: string;
  name: string;
  type: string;
  region: string;
  status: string;
  created_at: string;
  updated_at: string;
  cost: {
    current: number;
    projected: number;
    savings_potential: number;
  };
  configuration: Record<string, any>;
}

const ResourceInfoTab: React.FC<ResourceInfoTabProps> = ({ resourceId }) => {
  const [details, setDetails] = useState<ResourceDetails | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchDetails = async () => {
      try {
        setLoading(true);
        const response = await api.get(`/api/v1/resources/${resourceId}`);
        setDetails(response.data);
      } catch (err) {
        setError('Failed to load resource details');
        console.error(err);
      } finally {
        setLoading(false);
      }
    };

    fetchDetails();
  }, [resourceId]);

  if (loading) {
    return (
      <div className="flex items-center justify-center p-8">
        <Loader2 className="h-6 w-6 animate-spin" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-4 text-red-500">
        {error}
      </div>
    );
  }

  if (!details) {
    return null;
  }

  return (
    <div className="space-y-6">
      {/* Basic Info */}
      <div className="grid grid-cols-2 gap-4">
        <div>
          <h3 className="text-sm font-medium text-gray-500">Resource ID</h3>
          <p className="mt-1">{details.id}</p>
        </div>
        <div>
          <h3 className="text-sm font-medium text-gray-500">Name</h3>
          <p className="mt-1">{details.name}</p>
        </div>
        <div>
          <h3 className="text-sm font-medium text-gray-500">Type</h3>
          <p className="mt-1">{details.type}</p>
        </div>
        <div>
          <h3 className="text-sm font-medium text-gray-500">Region</h3>
          <p className="mt-1">{details.region}</p>
        </div>
        <div>
          <h3 className="text-sm font-medium text-gray-500">Status</h3>
          <p className="mt-1">
            <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
              details.status === 'running' ? 'bg-green-100 text-green-800' :
              details.status === 'stopped' ? 'bg-red-100 text-red-800' :
              'bg-gray-100 text-gray-800'
            }`}>
              {details.status}
            </span>
          </p>
        </div>
      </div>

      {/* Cost Information */}
      <div className="mt-6">
        <h3 className="text-lg font-medium">Cost Information</h3>
        <div className="mt-4 grid grid-cols-3 gap-4">
          <div className="bg-blue-50 p-4 rounded-lg">
            <h4 className="text-sm font-medium text-blue-900">Current Cost</h4>
            <p className="mt-2 text-2xl font-semibold text-blue-900">
              ${details.cost.current.toFixed(2)}
            </p>
            <p className="mt-1 text-sm text-blue-700">Per month</p>
          </div>
          <div className="bg-purple-50 p-4 rounded-lg">
            <h4 className="text-sm font-medium text-purple-900">Projected Cost</h4>
            <p className="mt-2 text-2xl font-semibold text-purple-900">
              ${details.cost.projected.toFixed(2)}
            </p>
            <p className="mt-1 text-sm text-purple-700">Next month</p>
          </div>
          <div className="bg-green-50 p-4 rounded-lg">
            <h4 className="text-sm font-medium text-green-900">Potential Savings</h4>
            <p className="mt-2 text-2xl font-semibold text-green-900">
              ${details.cost.savings_potential.toFixed(2)}
            </p>
            <p className="mt-1 text-sm text-green-700">Monthly savings</p>
          </div>
        </div>
      </div>

      {/* Configuration Details */}
      <div className="mt-6">
        <h3 className="text-lg font-medium">Configuration</h3>
        <div className="mt-4 bg-gray-50 rounded-lg p-4">
          <pre className="whitespace-pre-wrap text-sm">
            {JSON.stringify(details.configuration, null, 2)}
          </pre>
        </div>
      </div>
    </div>
  );
};

export default ResourceInfoTab;