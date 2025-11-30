import React, { useEffect, useState } from 'react';
import { useQuery } from 'react-query';
import { getAuthHeader } from '../../lib/auth';
import { Button } from '../ui/button';

const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080';

interface FilterState {
  resourceTypes: string[];
  tags: Record<string, string[]>;
  services: string[];
  accounts: number[];
  regions: string[];
}

interface DynamicFiltersProps {
  onFilterChange: (filters: FilterState) => void;
}

export const DynamicFilters: React.FC<DynamicFiltersProps> = ({ onFilterChange }) => {
  const [filters, setFilters] = useState<FilterState>({
    resourceTypes: [],
    tags: {},
    services: [],
    accounts: [],
    regions: [],
  });

  const authHeader = getAuthHeader();
  const headers = authHeader ? { Authorization: authHeader } : {};

  // Fetch filter options
  const { data: resourceTypes } = useQuery('resourceTypes', async () => {
    const res = await fetch(`${API_BASE_URL}/api/v1/filters/resource-types`, { headers });
    const data = await res.json();
    return data.success ? data.data : [];
  }, { enabled: !!authHeader });

  const { data: tags } = useQuery('tags', async () => {
    const res = await fetch(`${API_BASE_URL}/api/v1/filters/tags?limit=10`, { headers });
    const data = await res.json();
    return data.success ? data.data : { tag_keys: [], tag_values: {} };
  }, { enabled: !!authHeader });

  const { data: services } = useQuery('services', async () => {
    const res = await fetch(`${API_BASE_URL}/api/v1/filters/services`, { headers });
    const data = await res.json();
    return data.success ? data.data : [];
  }, { enabled: !!authHeader });

  const { data: accounts } = useQuery('accounts', async () => {
    const res = await fetch(`${API_BASE_URL}/api/v1/filters/accounts`, { headers });
    const data = await res.json();
    return data.success ? data.data : [];
  }, { enabled: !!authHeader });

  const { data: regions } = useQuery('regions', async () => {
    const res = await fetch(`${API_BASE_URL}/api/v1/filters/regions`, { headers });
    const data = await res.json();
    return data.success ? data.data : [];
  }, { enabled: !!authHeader });

  // Debounced filter update
  useEffect(() => {
    const timer = setTimeout(() => {
      onFilterChange(filters);
    }, 300);

    return () => clearTimeout(timer);
  }, [filters, onFilterChange]);

  const toggleResourceType = (type: string) => {
    setFilters(prev => ({
      ...prev,
      resourceTypes: prev.resourceTypes.includes(type)
        ? prev.resourceTypes.filter(t => t !== type)
        : [...prev.resourceTypes, type],
    }));
  };

  const toggleService = (service: string) => {
    setFilters(prev => ({
      ...prev,
      services: prev.services.includes(service)
        ? prev.services.filter(s => s !== service)
        : [...prev.services, service],
    }));
  };

  const toggleRegion = (region: string) => {
    setFilters(prev => ({
      ...prev,
      regions: prev.regions.includes(region)
        ? prev.regions.filter(r => r !== region)
        : [...prev.regions, region],
    }));
  };

  const clearFilters = () => {
    setFilters({
      resourceTypes: [],
      tags: {},
      services: [],
      accounts: [],
      regions: [],
    });
  };

  return (
    <div className="bg-white p-4 rounded-lg shadow-sm border border-gray-200 space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-lg font-semibold text-gray-900">Filters</h3>
        <Button variant="ghost" size="sm" onClick={clearFilters}>
          Clear All
        </Button>
      </div>

      {/* Resource Types */}
      {resourceTypes && resourceTypes.length > 0 && (
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Resource Types
          </label>
          <div className="flex flex-wrap gap-2">
            {resourceTypes.map((item: any) => (
              <button
                key={item.key}
                onClick={() => toggleResourceType(item.key)}
                className={`px-3 py-1 rounded-full text-sm ${
                  filters.resourceTypes.includes(item.key)
                    ? 'bg-indigo-100 text-indigo-800 border border-indigo-300'
                    : 'bg-gray-100 text-gray-700 border border-gray-300 hover:bg-gray-200'
                }`}
              >
                {item.label} ({item.count})
              </button>
            ))}
          </div>
        </div>
      )}

      {/* Services */}
      {services && services.length > 0 && (
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Services
          </label>
          <div className="flex flex-wrap gap-2">
            {services.map((item: any) => (
              <button
                key={item.key}
                onClick={() => toggleService(item.key)}
                className={`px-3 py-1 rounded-full text-sm ${
                  filters.services.includes(item.key)
                    ? 'bg-indigo-100 text-indigo-800 border border-indigo-300'
                    : 'bg-gray-100 text-gray-700 border border-gray-300 hover:bg-gray-200'
                }`}
              >
                {item.label}
              </button>
            ))}
          </div>
        </div>
      )}

      {/* Regions */}
      {regions && regions.length > 0 && (
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Regions
          </label>
          <div className="flex flex-wrap gap-2">
            {regions.map((item: any) => (
              <button
                key={item.key}
                onClick={() => toggleRegion(item.key)}
                className={`px-3 py-1 rounded-full text-sm ${
                  filters.regions.includes(item.key)
                    ? 'bg-indigo-100 text-indigo-800 border border-indigo-300'
                    : 'bg-gray-100 text-gray-700 border border-gray-300 hover:bg-gray-200'
                }`}
              >
                {item.label} ({item.count})
              </button>
            ))}
          </div>
        </div>
      )}

      {/* Tags (simplified - can be enhanced with typeahead) */}
      {tags && tags.tag_keys && tags.tag_keys.length > 0 && (
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Tags
          </label>
          <div className="space-y-2">
            {tags.tag_keys.slice(0, 5).map((tagKey: any) => (
              <div key={tagKey.key}>
                <span className="text-xs text-gray-600">{tagKey.key}:</span>
                <div className="flex flex-wrap gap-1 mt-1">
                  {tags.tag_values[tagKey.key]?.slice(0, 3).map((value: string) => (
                    <span
                      key={value}
                      className="px-2 py-1 text-xs bg-gray-100 text-gray-700 rounded"
                    >
                      {value}
                    </span>
                  ))}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
};

export default DynamicFilters;

