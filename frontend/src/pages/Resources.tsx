import React, { useState, useMemo } from 'react';
import { useQuery } from 'react-query';
import { 
  Search,
  Filter,
  Server,
  Database,
  HardDrive,
  Cloud,
  AlertCircle,
  Download,
  RefreshCw,
  Eye,
  Play,
  Square,
  X,
  ChevronDown,
  BarChart3,
  PieChart,
  TrendingUp,
  Activity,
  Shield
} from 'lucide-react';
import api from '../services/api';
import ResourcePanel from '../components/ResourceDetails/ResourcePanel';

interface Resource {
  id: string;
  resource_id: string;
  resource_type: string;
  region: string;
  state: string;
  instance_type: string;
  monthly_cost: number;
  tags: Record<string, string>;
  account_id: string;
  metadata?: any;
}

interface Filters {
  type: string;
  region: string;
  status: string;
}

const getResourceIcon = (type: string) => {
  const resourceType = type.toLowerCase();
  if (resourceType.includes('ec2')) return <Server className="w-4 h-4" />;
  if (resourceType.includes('rds')) return <Database className="w-4 h-4" />;
  if (resourceType.includes('ebs')) return <HardDrive className="w-4 h-4" />;
  if (resourceType.includes('s3')) return <Cloud className="w-4 h-4" />;
  return <Cloud className="w-4 h-4" />;
};

const getStatusColor = (status: string) => {
  switch (status?.toLowerCase()) {
    case 'running': case 'available': case 'active': return 'bg-green-100 text-green-800 border-green-200';
    case 'stopped': case 'terminated': return 'bg-red-100 text-red-800 border-red-200';
    case 'stopping': case 'starting': case 'pending': return 'bg-yellow-100 text-yellow-800 border-yellow-200';
    default: return 'bg-gray-100 text-gray-800 border-gray-200';
  }
};

const Resources: React.FC = () => {
  const [selectedResource, setSelectedResource] = useState<Resource | null>(null);
  const [selectedResources, setSelectedResources] = useState<Set<string>>(new Set());
  const [searchTerm, setSearchTerm] = useState('');
  const [filters, setFilters] = useState<Filters>({
    type: 'All',
    region: 'All',
    status: 'All'
  });
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(25);
  const [sortField, setSortField] = useState<string>('monthly_cost');
  const [sortDirection, setSortDirection] = useState<'asc' | 'desc'>('desc');
  const [showExportMenu, setShowExportMenu] = useState(false);

  const { data, isLoading, error, refetch } = useQuery(
    ['resources', filters],
    () => api.getResources(filters),
    {
      refetchInterval: 30000,
    }
  );

  const resources = data?.data || [];

  // Dynamic filter options based on actual data
  const filterOptions = useMemo(() => {
    const types = [...new Set(resources.map((r: Resource) => r.resource_type?.split('::')[1] || 'Unknown'))].sort();
    const regions = [...new Set(resources.map((r: Resource) => r.region))].filter(Boolean).sort();
    const statuses = [...new Set(resources.map((r: Resource) => r.state))].filter(Boolean).sort();
    
    return { types, regions, statuses };
  }, [resources]);

  // Filter and sort resources
  const filteredAndSortedResources = useMemo(() => {
    let filtered = resources.filter((resource: Resource) => {
      const matchesSearch = 
        resource.resource_id?.toLowerCase().includes(searchTerm.toLowerCase()) ||
        resource.resource_type?.toLowerCase().includes(searchTerm.toLowerCase()) ||
        Object.values(resource.tags || {}).some((tag: any) =>
          String(tag).toLowerCase().includes(searchTerm.toLowerCase())
        );
      
      const resourceType = resource.resource_type?.split('::')[1] || 'Unknown';
      const matchesType = filters.type === 'All' || resourceType === filters.type;
      const matchesRegion = filters.region === 'All' || resource.region === filters.region;
      const matchesStatus = filters.status === 'All' || resource.state === filters.status;
      
      return matchesSearch && matchesType && matchesRegion && matchesStatus;
    });

    // Sort
    filtered.sort((a: any, b: any) => {
      let aVal = a[sortField] || 0;
      let bVal = b[sortField] || 0;
      
      if (typeof aVal === 'string') {
        aVal = aVal.toLowerCase();
        bVal = bVal.toLowerCase();
      }
      
      if (sortDirection === 'asc') {
        return aVal > bVal ? 1 : -1;
      } else {
        return aVal < bVal ? 1 : -1;
      }
    });

    return filtered;
  }, [resources, searchTerm, filters, sortField, sortDirection]);

  // Pagination
  const totalPages = Math.ceil(filteredAndSortedResources.length / pageSize);
  const paginatedResources = filteredAndSortedResources.slice(
    (currentPage - 1) * pageSize,
    currentPage * pageSize
  );

  // Summary calculations based on filtered data
  const summary = useMemo(() => {
    const totalResources = filteredAndSortedResources.length;
    const totalCost = filteredAndSortedResources.reduce((sum: number, r: Resource) => sum + (r.monthly_cost || 0), 0);
    const activeResources = filteredAndSortedResources.filter((r: Resource) => 
      ['running', 'available', 'active'].includes(r.state?.toLowerCase())
    ).length;
    
    // Resource type distribution (filtered)
    const typeDistribution = filteredAndSortedResources.reduce((acc: any, r: Resource) => {
      const type = r.resource_type?.split('::')[1] || 'Other';
      acc[type] = (acc[type] || 0) + 1;
      return acc;
    }, {});

    // Region cost breakdown (filtered)
    const regionCosts = filteredAndSortedResources.reduce((acc: any, r: Resource) => {
      acc[r.region] = (acc[r.region] || 0) + (r.monthly_cost || 0);
      return acc;
    }, {});

    return {
      totalResources,
      totalCost,
      activeResources,
      typeDistribution,
      regionCosts
    };
  }, [filteredAndSortedResources]);

  const handleSort = (field: string) => {
    if (sortField === field) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
    } else {
      setSortField(field);
      setSortDirection('desc');
    }
  };

  const handleScan = async () => {
    try {
      await api.post('/api/scan', {});
      refetch();
    } catch (error) {
      console.error('Scan failed:', error);
    }
  };

  const handleWhitelistResources = async () => {
    if (selectedResources.size === 0) {
      console.log('No resources selected for whitelisting');
      return;
    }

    try {
      const resourcesArray = Array.from(selectedResources);
      const whitelistPromises = resourcesArray.map(resourceId => {
        const resource = resources.find((r: Resource) => r.resource_id === resourceId);
        return api.post('/whitelists', {
          whitelist_type: 'resource',
          resource_arn: `arn:aws:ec2:${resource?.region || 'us-east-1'}:${resource?.account_id || '123456789012'}:instance/${resourceId}`,
          reason: 'Excluded from cost optimization recommendations',
          business_justification: 'Business critical resource',
          created_by: 'user'
        });
      });

      await Promise.all(whitelistPromises);
      
      console.log(`${resourcesArray.length} resource(s) added to whitelist`);
      setSelectedResources(new Set());
      refetch();
    } catch (error) {
      console.error('Bulk whitelist failed:', error);
    }
  };

  const handleSelectResource = (resourceId: string, event: React.MouseEvent) => {
    event.stopPropagation();
    const newSelected = new Set(selectedResources);
    if (newSelected.has(resourceId)) {
      newSelected.delete(resourceId);
    } else {
      newSelected.add(resourceId);
    }
    setSelectedResources(newSelected);
    console.log('Selected resources:', Array.from(newSelected));
  };

  const handleSelectAll = (event: React.ChangeEvent<HTMLInputElement>) => {
    event.stopPropagation();
    if (selectedResources.size === paginatedResources.length && paginatedResources.length > 0) {
      setSelectedResources(new Set());
    } else {
      const allIds = paginatedResources.map((r: Resource) => r.resource_id);
      setSelectedResources(new Set(allIds));
    }
  };

  const handleExport = (format: string) => {
    const exportData = filteredAndSortedResources.map((resource: Resource) => ({
      'Resource ID': resource.resource_id,
      'Type': resource.resource_type?.split('::')[1] || 'Unknown',
      'Region': resource.region,
      'Instance Type': resource.instance_type,
      'Status': resource.state,
      'Monthly Cost': `$${(resource.monthly_cost || 0).toFixed(2)}`,
      'Account ID': resource.account_id,
      'Tags': Object.entries(resource.tags || {}).map(([k, v]) => `${k}:${v}`).join('; '),
      'Created': new Date().toISOString().split('T')[0]
    }));

    if (format === 'CSV') {
      const headers = Object.keys(exportData[0] || {});
      const csvContent = [
        headers.join(','),
        ...exportData.map(row => 
          headers.map(header => `"${row[header as keyof typeof row] || ''}"`).join(',')
        )
      ].join('\n');
      
      const blob = new Blob([csvContent], { type: 'text/csv' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `resources_${new Date().toISOString().split('T')[0]}.csv`;
      a.click();
      URL.revokeObjectURL(url);
    } else if (format === 'JSON') {
      const jsonContent = JSON.stringify(exportData, null, 2);
      const blob = new Blob([jsonContent], { type: 'application/json' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `resources_${new Date().toISOString().split('T')[0]}.json`;
      a.click();
      URL.revokeObjectURL(url);
    }
    
    setShowExportMenu(false);
  };

  const clearFilters = () => {
    setFilters({ type: 'All', region: 'All', status: 'All' });
    setSearchTerm('');
    setCurrentPage(1);
  };

  const activeFilterCount = Object.values(filters).filter(f => f !== 'All').length + (searchTerm ? 1 : 0);

  if (error) {
    return (
      <div className="min-h-screen bg-slate-50 p-6">
        <div className="max-w-7xl mx-auto">
          <div className="bg-red-50 border border-red-200 rounded-lg p-4 text-red-700">
            <AlertCircle className="w-5 h-5 inline mr-2" />
            Error loading resources: {(error as Error).message}
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-slate-50">
      <div className="max-w-7xl mx-auto p-6 space-y-6">
        {/* Header */}
        <div className="bg-white rounded-lg shadow-sm border border-slate-200 p-6">
          <div className="flex flex-col lg:flex-row lg:items-center lg:justify-between gap-4">
            <div>
              <h1 className="text-3xl font-bold text-slate-900">Resources</h1>
              <p className="text-slate-600 mt-1">Manage and monitor your AWS infrastructure</p>
            </div>
            
            <div className="flex items-center gap-3">
              {selectedResources.size > 0 && (
                <button
                  onClick={handleWhitelistResources}
                  className="flex items-center gap-2 px-4 py-2 bg-orange-600 text-white rounded-lg hover:bg-orange-700 transition-colors"
                >
                  <Shield className="w-4 h-4" />
                  Whitelist ({selectedResources.size})
                </button>
              )}
              
              <button
                onClick={handleScan}
                className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
              >
                <RefreshCw className="w-4 h-4" />
                Scan Now
              </button>
              
              <div className="relative">
                <button
                  onClick={() => setShowExportMenu(!showExportMenu)}
                  className="flex items-center gap-2 px-4 py-2 border border-slate-300 rounded-lg hover:bg-slate-50 transition-colors"
                >
                  <Download className="w-4 h-4" />
                  Export
                  <ChevronDown className="w-4 h-4" />
                </button>
                
                {showExportMenu && (
                  <div className="absolute right-0 mt-2 w-48 bg-white rounded-lg shadow-lg border border-slate-200 z-10">
                    <div className="py-1">
                      <button onClick={() => handleExport('CSV')} className="w-full text-left px-4 py-2 hover:bg-slate-50">Export as CSV</button>
                      <button onClick={() => handleExport('JSON')} className="w-full text-left px-4 py-2 hover:bg-slate-50">Export as JSON</button>
                    </div>
                  </div>
                )}
              </div>
            </div>
          </div>

          {/* Search and Filters */}
          <div className="mt-6 space-y-4">
            <div className="flex flex-col lg:flex-row gap-4">
              <div className="flex-1 relative">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-slate-400 w-5 h-5" />
                <input
                  type="text"
                  placeholder="Search resources by ID, name, or tags..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="w-full pl-10 pr-4 py-3 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                />
              </div>
            </div>

            <div className="flex flex-wrap items-center gap-4">
              <select
                value={filters.type}
                onChange={(e) => setFilters({...filters, type: e.target.value})}
                className="px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500"
                disabled={isLoading}
              >
                <option value="All">All Types ({resources.length})</option>
                {filterOptions.types.map(type => {
                  const count = resources.filter((r: Resource) => r.resource_type?.includes(type)).length;
                  return (
                    <option key={type} value={type}>{type} ({count})</option>
                  );
                })}
              </select>

              <select
                value={filters.region}
                onChange={(e) => setFilters({...filters, region: e.target.value})}
                className="px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500"
                disabled={isLoading}
              >
                <option value="All">All Regions ({resources.length})</option>
                {filterOptions.regions.map(region => {
                  const count = resources.filter((r: Resource) => r.region === region).length;
                  return (
                    <option key={region} value={region}>{region} ({count})</option>
                  );
                })}
              </select>

              <select
                value={filters.status}
                onChange={(e) => setFilters({...filters, status: e.target.value})}
                className="px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500"
                disabled={isLoading}
              >
                <option value="All">All Statuses ({resources.length})</option>
                {filterOptions.statuses.map(status => {
                  const count = resources.filter((r: Resource) => r.state === status).length;
                  return (
                    <option key={status} value={status}>{status} ({count})</option>
                  );
                })}
              </select>

              {activeFilterCount > 0 && (
                <button
                  onClick={clearFilters}
                  className="flex items-center gap-2 px-3 py-2 text-slate-600 hover:text-slate-800"
                >
                  <X className="w-4 h-4" />
                  Clear ({activeFilterCount})
                </button>
              )}
            </div>
          </div>
        </div>

        {/* Summary Overview */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          <div className="bg-white rounded-lg shadow-sm border border-slate-200 p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-slate-600 text-sm font-medium">Total Resources</p>
                <p className="text-3xl font-bold text-slate-900 mt-1">{summary.totalResources.toLocaleString()}</p>
              </div>
              <div className="p-3 bg-blue-100 rounded-lg">
                <Server className="w-6 h-6 text-blue-600" />
              </div>
            </div>
          </div>

          <div className="bg-white rounded-lg shadow-sm border border-slate-200 p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-slate-600 text-sm font-medium">Monthly Cost</p>
                <p className="text-3xl font-bold text-slate-900 mt-1">${summary.totalCost.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}</p>
              </div>
              <div className="p-3 bg-green-100 rounded-lg">
                <TrendingUp className="w-6 h-6 text-green-600" />
              </div>
            </div>
          </div>

          <div className="bg-white rounded-lg shadow-sm border border-slate-200 p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-slate-600 text-sm font-medium">Active Resources</p>
                <p className="text-3xl font-bold text-slate-900 mt-1">{summary.activeResources.toLocaleString()}</p>
              </div>
              <div className="p-3 bg-green-100 rounded-lg">
                <Activity className="w-6 h-6 text-green-600" />
              </div>
            </div>
          </div>

          <div className="bg-white rounded-lg shadow-sm border border-slate-200 p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-slate-600 text-sm font-medium">Optimization Opportunities</p>
                <p className="text-3xl font-bold text-slate-900 mt-1">{Math.floor(summary.totalResources * 0.15)}</p>
              </div>
              <div className="p-3 bg-orange-100 rounded-lg">
                <AlertCircle className="w-6 h-6 text-orange-600" />
              </div>
            </div>
          </div>
        </div>

        {/* Charts */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Resource Distribution */}
          <div className="bg-white rounded-lg shadow-sm border border-slate-200 p-6">
            <div className="flex items-center gap-2 mb-4">
              <PieChart className="w-5 h-5 text-slate-600" />
              <h3 className="text-lg font-semibold text-slate-900">Resource Distribution</h3>
            </div>
            <div className="space-y-3">
              {Object.entries(summary.typeDistribution).map(([type, count]: [string, any]) => {
                const percentage = ((count / summary.totalResources) * 100).toFixed(1);
                return (
                  <div key={type} className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      {getResourceIcon(type)}
                      <span className="text-slate-700">{type}</span>
                    </div>
                    <div className="flex items-center gap-3">
                      <div className="w-24 bg-slate-200 rounded-full h-2">
                        <div 
                          className="bg-blue-600 h-2 rounded-full" 
                          style={{ width: `${percentage}%` }}
                        ></div>
                      </div>
                      <span className="text-sm text-slate-600 w-12 text-right">{percentage}%</span>
                      <span className="text-sm font-medium text-slate-900 w-8 text-right">{count}</span>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>

          {/* Cost by Region */}
          <div className="bg-white rounded-lg shadow-sm border border-slate-200 p-6">
            <div className="flex items-center gap-2 mb-4">
              <BarChart3 className="w-5 h-5 text-slate-600" />
              <h3 className="text-lg font-semibold text-slate-900">Cost by Region</h3>
            </div>
            <div className="space-y-3">
              {Object.entries(summary.regionCosts)
                .sort(([,a]: [string, any], [,b]: [string, any]) => b - a)
                .slice(0, 5)
                .map(([region, cost]: [string, any]) => {
                  const percentage = summary.totalCost > 0 ? ((cost / summary.totalCost) * 100).toFixed(1) : '0';
                  return (
                    <div key={region} className="flex items-center justify-between">
                      <span className="text-slate-700">{region}</span>
                      <div className="flex items-center gap-3">
                        <div className="w-24 bg-slate-200 rounded-full h-2">
                          <div 
                            className="bg-green-600 h-2 rounded-full" 
                            style={{ width: `${percentage}%` }}
                          ></div>
                        </div>
                        <span className="text-sm text-slate-600 w-12 text-right">{percentage}%</span>
                        <span className="text-sm font-medium text-slate-900 w-16 text-right">${cost.toFixed(2)}</span>
                      </div>
                    </div>
                  );
                })
              }
            </div>
          </div>
        </div>
        {/* Main Resources Table */}
        <div className="bg-white rounded-lg shadow-sm border border-slate-200">
          <div className="p-6 border-b border-slate-200">
            <div className="flex items-center justify-between">
              <h3 className="text-lg font-semibold text-slate-900">Resources</h3>
              <div className="flex items-center gap-4">
                <span className="text-sm text-slate-600">
                  Showing {((currentPage - 1) * pageSize) + 1}-{Math.min(currentPage * pageSize, filteredAndSortedResources.length)} of {filteredAndSortedResources.length}
                </span>
                <select
                  value={pageSize}
                  onChange={(e) => {
                    setPageSize(Number(e.target.value));
                    setCurrentPage(1);
                  }}
                  className="px-3 py-1 border border-slate-300 rounded text-sm"
                >
                  <option value={10}>10 per page</option>
                  <option value={25}>25 per page</option>
                  <option value={50}>50 per page</option>
                </select>
              </div>
            </div>
          </div>

          {isLoading ? (
            <div className="p-8">
              <div className="space-y-4">
                {Array.from({ length: 5 }).map((_, i) => (
                  <div key={i} className="animate-pulse flex items-center space-x-4">
                    <div className="w-8 h-8 bg-slate-200 rounded"></div>
                    <div className="flex-1 space-y-2">
                      <div className="h-4 bg-slate-200 rounded w-1/4"></div>
                      <div className="h-3 bg-slate-200 rounded w-1/6"></div>
                    </div>
                    <div className="w-20 h-4 bg-slate-200 rounded"></div>
                    <div className="w-16 h-4 bg-slate-200 rounded"></div>
                  </div>
                ))}
              </div>
            </div>
          ) : filteredAndSortedResources.length === 0 ? (
            <div className="p-12 text-center">
              <Cloud className="w-12 h-12 text-slate-400 mx-auto mb-4" />
              <h3 className="text-lg font-medium text-slate-900 mb-2">No resources found</h3>
              <p className="text-slate-600">Try adjusting your search or filters, or run a scan to discover resources.</p>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead className="bg-slate-50 border-b border-slate-200">
                  <tr>
                    <th className="text-left py-3 px-6 font-medium text-slate-700 w-12">
                      <input
                        type="checkbox"
                        checked={selectedResources.size === paginatedResources.length && paginatedResources.length > 0}
                        onChange={handleSelectAll}
                        className="rounded border-slate-300 text-blue-600 focus:ring-blue-500 cursor-pointer"
                      />
                    </th>
                    <th className="text-left py-3 px-6 font-medium text-slate-700">
                      <button 
                        onClick={() => handleSort('resource_id')}
                        className="flex items-center gap-1 hover:text-slate-900"
                      >
                        Resource ID
                        {sortField === 'resource_id' && (
                          <span className="text-xs">{sortDirection === 'asc' ? '↑' : '↓'}</span>
                        )}
                      </button>
                    </th>
                    <th className="text-left py-3 px-6 font-medium text-slate-700">
                      <button 
                        onClick={() => handleSort('resource_type')}
                        className="flex items-center gap-1 hover:text-slate-900"
                      >
                        Type
                        {sortField === 'resource_type' && (
                          <span className="text-xs">{sortDirection === 'asc' ? '↑' : '↓'}</span>
                        )}
                      </button>
                    </th>
                    <th className="text-left py-3 px-6 font-medium text-slate-700">
                      <button 
                        onClick={() => handleSort('region')}
                        className="flex items-center gap-1 hover:text-slate-900"
                      >
                        Region
                        {sortField === 'region' && (
                          <span className="text-xs">{sortDirection === 'asc' ? '↑' : '↓'}</span>
                        )}
                      </button>
                    </th>
                    <th className="text-left py-3 px-6 font-medium text-slate-700">
                      <button 
                        onClick={() => handleSort('state')}
                        className="flex items-center gap-1 hover:text-slate-900"
                      >
                        Status
                        {sortField === 'state' && (
                          <span className="text-xs">{sortDirection === 'asc' ? '↑' : '↓'}</span>
                        )}
                      </button>
                    </th>
                    <th className="text-left py-3 px-6 font-medium text-slate-700">
                      <button 
                        onClick={() => handleSort('monthly_cost')}
                        className="flex items-center gap-1 hover:text-slate-900"
                      >
                        Monthly Cost
                        {sortField === 'monthly_cost' && (
                          <span className="text-xs">{sortDirection === 'asc' ? '↑' : '↓'}</span>
                        )}
                      </button>
                    </th>
                    <th className="text-left py-3 px-6 font-medium text-slate-700">Usage</th>
                    <th className="text-left py-3 px-6 font-medium text-slate-700">Tags</th>
                    <th className="text-left py-3 px-6 font-medium text-slate-700">Actions</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-200">
                  {paginatedResources.map((resource: Resource, index: number) => {
                    const isEven = index % 2 === 0;
                    const tagEntries = Object.entries(resource.tags || {}).slice(0, 2);
                    const remainingTags = Object.keys(resource.tags || {}).length - 2;
                    
                    return (
                      <tr 
                        key={resource.id} 
                        className={`hover:bg-slate-50 transition-colors ${
                          isEven ? 'bg-white' : 'bg-slate-25'
                        } ${
                          selectedResources.has(resource.resource_id) ? 'bg-blue-50 border-l-4 border-blue-500' : ''
                        }`}
                      >
                        <td className="py-4 px-6">
                          <input
                            type="checkbox"
                            checked={selectedResources.has(resource.resource_id)}
                            onChange={(e) => handleSelectResource(resource.resource_id, e)}
                            className="rounded border-slate-300 text-blue-600 focus:ring-blue-500 cursor-pointer"
                          />
                        </td>
                        <td 
                          className="py-4 px-6 cursor-pointer"
                          onClick={() => setSelectedResource(resource)}
                        >
                          <div className="flex items-center gap-3">
                            <div className="p-2 bg-slate-100 rounded-lg">
                              {getResourceIcon(resource.resource_type)}
                            </div>
                            <div>
                              <div className="font-medium text-slate-900">{resource.resource_id}</div>
                              <div className="text-sm text-slate-600">{resource.instance_type}</div>
                            </div>
                          </div>
                        </td>
                        <td 
                          className="py-4 px-6 cursor-pointer"
                          onClick={() => setSelectedResource(resource)}
                        >
                          <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                            {resource.resource_type?.split('::')[1] || 'Unknown'}
                          </span>
                        </td>
                        <td 
                          className="py-4 px-6 text-slate-700 cursor-pointer"
                          onClick={() => setSelectedResource(resource)}
                        >{resource.region}</td>
                        <td 
                          className="py-4 px-6 cursor-pointer"
                          onClick={() => setSelectedResource(resource)}
                        >
                          <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium border ${
                            getStatusColor(resource.state)
                          }`}>
                            {resource.state}
                          </span>
                        </td>
                        <td 
                          className="py-4 px-6 cursor-pointer"
                          onClick={() => setSelectedResource(resource)}
                        >
                          <div className="font-medium text-slate-900">
                            ${(resource.monthly_cost || 0).toFixed(2)}
                          </div>
                          <div className="text-xs text-slate-600">per month</div>
                        </td>
                        <td 
                          className="py-4 px-6 cursor-pointer"
                          onClick={() => setSelectedResource(resource)}
                        >
                          <div className="text-sm text-slate-700">
                            {resource.resource_type?.includes('EC2') ? '23% CPU' :
                             resource.resource_type?.includes('RDS') ? '100 GB Storage' :
                             resource.resource_type?.includes('S3') ? '45.2 GB' : 'N/A'}
                          </div>
                        </td>
                        <td 
                          className="py-4 px-6 cursor-pointer"
                          onClick={() => setSelectedResource(resource)}
                        >
                          <div className="flex flex-wrap gap-1">
                            {tagEntries.map(([key, value]) => (
                              <span key={key} className="inline-flex items-center px-2 py-1 rounded text-xs bg-slate-100 text-slate-700">
                                {key}: {String(value)}
                              </span>
                            ))}
                            {remainingTags > 0 && (
                              <span className="inline-flex items-center px-2 py-1 rounded text-xs bg-slate-100 text-slate-600">
                                +{remainingTags} more
                              </span>
                            )}
                          </div>
                        </td>
                        <td className="py-4 px-6">
                          <button
                            onClick={(e) => {
                              e.stopPropagation();
                              setSelectedResource(resource);
                            }}
                            className="p-1 text-slate-600 hover:text-blue-600 transition-colors"
                            title="View Details"
                          >
                            <Eye className="w-4 h-4" />
                          </button>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          )}

          {/* Pagination */}
          {totalPages > 1 && (
            <div className="p-6 border-t border-slate-200">
              <div className="flex items-center justify-between">
                <div className="text-sm text-slate-600">
                  Page {currentPage} of {totalPages}
                </div>
                <div className="flex items-center gap-2">
                  <button
                    onClick={() => setCurrentPage(Math.max(1, currentPage - 1))}
                    disabled={currentPage === 1}
                    className="px-3 py-2 border border-slate-300 rounded-lg text-sm disabled:opacity-50 disabled:cursor-not-allowed hover:bg-slate-50"
                  >
                    Previous
                  </button>
                  
                  {/* Page numbers */}
                  <div className="flex items-center gap-1">
                    {Array.from({ length: Math.min(5, totalPages) }, (_, i) => {
                      const pageNum = Math.max(1, Math.min(totalPages - 4, currentPage - 2)) + i;
                      return (
                        <button
                          key={pageNum}
                          onClick={() => setCurrentPage(pageNum)}
                          className={`px-3 py-2 text-sm rounded-lg ${
                            pageNum === currentPage
                              ? 'bg-blue-600 text-white'
                              : 'border border-slate-300 hover:bg-slate-50'
                          }`}
                        >
                          {pageNum}
                        </button>
                      );
                    })}
                  </div>
                  
                  <button
                    onClick={() => setCurrentPage(Math.min(totalPages, currentPage + 1))}
                    disabled={currentPage === totalPages}
                    className="px-3 py-2 border border-slate-300 rounded-lg text-sm disabled:opacity-50 disabled:cursor-not-allowed hover:bg-slate-50"
                  >
                    Next
                  </button>
                </div>
              </div>
            </div>
          )}
        </div>

        {/* Resource Details Side Panel */}
        {selectedResource && (
          <ResourcePanel
            resourceId={selectedResource.resource_id}
            resourceType={selectedResource.resource_type}
            onClose={() => setSelectedResource(null)}
          />
        )}
      </div>
    </div>
  );
};

export default Resources;