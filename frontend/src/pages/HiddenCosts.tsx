import React, { useState, useEffect, useMemo } from 'react';
import { 
  AlertCircle, 
  DollarSign, 
  Filter, 
  X, 
  RefreshCw,
  Download,
  ChevronDown,
  Eye,
  Code,
  Shield,
  TrendingUp,
  PieChart,
  BarChart3,
  Activity
} from 'lucide-react';
import api from '../services/api';
import { getCurrentUser } from '../lib/auth';

interface Finding {
  id: string;
  detector_name: string;
  category: string;
  severity: string;
  title: string;
  description: string;
  resource_arn: string;
  estimated_savings: number;
  confidence: number;
  created_at: string;
  status?: string;
  resources_count?: number;
}

interface Filters {
  category: string;
  severity: string;
  status: string;
}

const HiddenCosts: React.FC = () => {
  const [findings, setFindings] = useState<Finding[]>([]);
  const [loading, setLoading] = useState(true);
  const [filters, setFilters] = useState<Filters>({
    category: 'All',
    severity: 'All',
    status: 'All'
  });
  const [selectedFinding, setSelectedFinding] = useState<Finding | null>(null);
  const [expandedRows, setExpandedRows] = useState<Set<string>>(new Set());
  const [showExportMenu, setShowExportMenu] = useState(false);
  const [showIaCModal, setShowIaCModal] = useState(false);
  const [iacFinding, setIaCFinding] = useState<Finding | null>(null);
  const [generatedIaC, setGeneratedIaC] = useState<any>(null);
  const [iacLoading, setIaCLoading] = useState(false);
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(25);
  const [sortField, setSortField] = useState<string>('estimated_savings');
  const [sortDirection, setSortDirection] = useState<'asc' | 'desc'>('desc');
  const [error, setError] = useState('');

  useEffect(() => {
    fetchFindings();
  }, []);

  const fetchFindings = async () => {
    setLoading(true);
    try {
      const user = getCurrentUser();
      if (!user) {
        setError('User not authenticated');
        setLoading(false);
        return;
      }

      const data = await api.get('/api/customers/findings');
      const findingsData = Array.isArray(data.findings) ? data.findings : [];
        // Add mock data for demo purposes
        const enhancedFindings = findingsData.map((f: Finding) => ({
          ...f,
          status: f.status || 'Open',
          resources_count: f.resources_count || Math.floor(Math.random() * 10) + 1
        }));
        setFindings(enhancedFindings);
    } catch (err) {
      console.error('Error fetching findings:', err);
      setError('Failed to load findings');
      setFindings([]);
    } finally {
      setLoading(false);
    }
  };

  // Filter and sort findings
  const filteredAndSortedFindings = useMemo(() => {
    let filtered = findings.filter((finding: Finding) => {
      const matchesCategory = filters.category === 'All' || finding.category === filters.category;
      const matchesSeverity = filters.severity === 'All' || finding.severity === filters.severity;
      const matchesStatus = filters.status === 'All' || finding.status === filters.status;
      
      return matchesCategory && matchesSeverity && matchesStatus;
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
  }, [findings, filters, sortField, sortDirection]);

  // Pagination
  const totalPages = Math.ceil(filteredAndSortedFindings.length / pageSize);
  const paginatedFindings = filteredAndSortedFindings.slice(
    (currentPage - 1) * pageSize,
    currentPage * pageSize
  );

  // Summary calculations
  const summary = useMemo(() => {
    const totalFindings = filteredAndSortedFindings.length;
    const totalSavings = filteredAndSortedFindings.reduce((sum: number, f: Finding) => sum + (f.estimated_savings || 0), 0);
    
    // Severity distribution
    const severityDistribution = filteredAndSortedFindings.reduce((acc: any, f: Finding) => {
      acc[f.severity] = (acc[f.severity] || 0) + 1;
      return acc;
    }, {});

    // Category distribution
    const categoryDistribution = filteredAndSortedFindings.reduce((acc: any, f: Finding) => {
      acc[f.category] = (acc[f.category] || 0) + 1;
      return acc;
    }, {});

    return {
      totalFindings,
      totalSavings,
      severityDistribution,
      categoryDistribution
    };
  }, [filteredAndSortedFindings]);

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'Critical': return 'bg-red-100 text-red-800 border-red-200';
      case 'High': return 'bg-orange-100 text-orange-800 border-orange-200';
      case 'Medium': return 'bg-yellow-100 text-yellow-800 border-yellow-200';
      case 'Low': return 'bg-blue-100 text-blue-800 border-blue-200';
      default: return 'bg-gray-100 text-gray-800 border-gray-200';
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'Open': return 'bg-red-100 text-red-800 border-red-200';
      case 'In Progress': return 'bg-yellow-100 text-yellow-800 border-yellow-200';
      case 'Resolved': return 'bg-green-100 text-green-800 border-green-200';
      default: return 'bg-gray-100 text-gray-800 border-gray-200';
    }
  };

  const handleSort = (field: string) => {
    if (sortField === field) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
    } else {
      setSortField(field);
      setSortDirection('desc');
    }
  };

  const handleRescan = async () => {
    try {
      await api.post('/api/scan', {});
      fetchFindings();
    } catch (error) {
      console.error('Rescan failed:', error);
    }
  };

  const handleExport = (format: string) => {
    console.log(`Exporting findings as ${format}`);
    setShowExportMenu(false);
  };

  const clearFilters = () => {
    setFilters({ category: 'All', severity: 'All', status: 'All' });
    setCurrentPage(1);
  };

  const toggleRowExpansion = (findingId: string) => {
    const newExpanded = new Set(expandedRows);
    if (newExpanded.has(findingId)) {
      newExpanded.delete(findingId);
    } else {
      newExpanded.add(findingId);
    }
    setExpandedRows(newExpanded);
  };

  const handleGenerateIaC = async (finding: Finding) => {
    setIaCFinding(finding);
    setShowIaCModal(true);
    setIaCLoading(true);
    try {
      const response = await api.post('/iac/generate', {
        finding_id: finding.id,
        format: 'terraform',
        action: 'optimize'
      });
      setGeneratedIaC(response);
    } catch (error) {
      console.error('IaC generation failed:', error);
    } finally {
      setIaCLoading(false);
    }
  };

  const activeFilterCount = Object.values(filters).filter(f => f !== 'All').length;

  if (loading) {
    return (
      <div className="min-h-screen bg-slate-50 flex items-center justify-center">
        <div className="text-xl text-slate-600">Loading findings...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen bg-slate-50 p-6 flex items-center justify-center">
        <div className="bg-red-50 border border-red-200 rounded-lg p-6 max-w-md">
          <AlertCircle className="w-6 h-6 text-red-600 mb-2" />
          <p className="text-red-800">{error}</p>
          <button
            onClick={() => window.location.href = '/login'}
            className="mt-4 px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700"
          >
            Go to Login
          </button>
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
              <h1 className="text-3xl font-bold text-slate-900">Hidden Costs</h1>
              <p className="text-slate-600 mt-1">Discover cost optimization opportunities across your infrastructure</p>
            </div>
            
            <div className="flex items-center gap-3">
              <button
                onClick={handleRescan}
                className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
              >
                <RefreshCw className="w-4 h-4" />
                Rescan
              </button>
              
              <div className="relative">
                <button
                  onClick={() => setShowExportMenu(!showExportMenu)}
                  className="flex items-center gap-2 px-4 py-2 border border-slate-300 rounded-lg hover:bg-slate-50 transition-colors"
                >
                  <Download className="w-4 h-4" />
                  Export Findings
                  <ChevronDown className="w-4 h-4" />
                </button>
                
                {showExportMenu && (
                  <div className="absolute right-0 mt-2 w-48 bg-white rounded-lg shadow-lg border border-slate-200 z-10">
                    <div className="py-1">
                      <button onClick={() => handleExport('CSV')} className="w-full text-left px-4 py-2 hover:bg-slate-50">Export as CSV</button>
                      <button onClick={() => handleExport('PDF')} className="w-full text-left px-4 py-2 hover:bg-slate-50">Export as PDF</button>
                      <button onClick={() => handleExport('XLSX')} className="w-full text-left px-4 py-2 hover:bg-slate-50">Export as XLSX</button>
                    </div>
                  </div>
                )}
              </div>
            </div>
          </div>

          {/* Filters */}
          <div className="mt-6 space-y-4">
            <div className="flex flex-wrap items-center gap-4">
              <select
                value={filters.severity}
                onChange={(e) => setFilters({...filters, severity: e.target.value})}
                className="px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500"
              >
                <option value="All">All Severities</option>
                <option value="Critical">Critical</option>
                <option value="High">High</option>
                <option value="Medium">Medium</option>
                <option value="Low">Low</option>
              </select>

              <select
                value={filters.category}
                onChange={(e) => setFilters({...filters, category: e.target.value})}
                className="px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500"
              >
                <option value="All">All Categories</option>
                <option value="Compute">Compute</option>
                <option value="Storage">Storage</option>
                <option value="Network">Network</option>
                <option value="Database">Database</option>
              </select>

              <select
                value={filters.status}
                onChange={(e) => setFilters({...filters, status: e.target.value})}
                className="px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500"
              >
                <option value="All">All Statuses</option>
                <option value="Open">Open</option>
                <option value="In Progress">In Progress</option>
                <option value="Resolved">Resolved</option>
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

        {/* Summary Cards */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <div className="bg-white rounded-lg shadow-sm border border-slate-200 p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-slate-600 text-sm font-medium">Total Findings</p>
                <p className="text-3xl font-bold text-slate-900 mt-1">{summary.totalFindings}</p>
              </div>
              <div className="p-3 bg-red-100 rounded-lg">
                <AlertCircle className="w-6 h-6 text-red-600" />
              </div>
            </div>
          </div>

          <div className="bg-white rounded-lg shadow-sm border border-slate-200 p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-slate-600 text-sm font-medium">Total Savings</p>
                <p className="text-3xl font-bold text-green-600 mt-1">${summary.totalSavings.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}</p>
              </div>
              <div className="p-3 bg-green-100 rounded-lg">
                <DollarSign className="w-6 h-6 text-green-600" />
              </div>
            </div>
          </div>

          <div className="bg-white rounded-lg shadow-sm border border-slate-200 p-6">
            <div className="flex items-center gap-2 mb-4">
              <PieChart className="w-5 h-5 text-slate-600" />
              <h3 className="text-lg font-semibold text-slate-900">By Severity</h3>
            </div>
            <div className="space-y-2">
              {Object.entries(summary.severityDistribution).map(([severity, count]: [string, any]) => {
                const percentage = summary.totalFindings > 0 ? ((count / summary.totalFindings) * 100).toFixed(0) : '0';
                return (
                  <div key={severity} className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <span className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium border ${
                        getSeverityColor(severity)
                      }`}>
                        {severity}
                      </span>
                    </div>
                    <div className="flex items-center gap-2">
                      <span className="text-sm text-slate-600">{percentage}%</span>
                      <span className="text-sm font-medium text-slate-900">{count}</span>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        </div>

        {/* Findings Table */}
        <div className="bg-white rounded-lg shadow-sm border border-slate-200">
          <div className="p-6 border-b border-slate-200">
            <div className="flex items-center justify-between">
              <h3 className="text-lg font-semibold text-slate-900">Cost Optimization Findings</h3>
              <div className="flex items-center gap-4">
                <span className="text-sm text-slate-600">
                  Showing {((currentPage - 1) * pageSize) + 1}-{Math.min(currentPage * pageSize, filteredAndSortedFindings.length)} of {filteredAndSortedFindings.length}
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

          {loading ? (
            <div className="p-8">
              <div className="space-y-4">
                {Array.from({ length: 5 }).map((_, i) => (
                  <div key={i} className="animate-pulse flex items-center space-x-4">
                    <div className="w-8 h-8 bg-slate-200 rounded"></div>
                    <div className="flex-1 space-y-2">
                      <div className="h-4 bg-slate-200 rounded w-3/4"></div>
                      <div className="h-3 bg-slate-200 rounded w-1/2"></div>
                    </div>
                    <div className="w-20 h-4 bg-slate-200 rounded"></div>
                  </div>
                ))}
              </div>
            </div>
          ) : filteredAndSortedFindings.length === 0 ? (
            <div className="p-12 text-center">
              <AlertCircle className="w-12 h-12 text-slate-400 mx-auto mb-4" />
              <h3 className="text-lg font-medium text-slate-900 mb-2">No findings found</h3>
              <p className="text-slate-600">Try adjusting your filters or run a scan to discover optimization opportunities.</p>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead className="bg-slate-50 border-b border-slate-200">
                  <tr>
                    <th className="text-left py-3 px-6 font-medium text-slate-700">
                      <button 
                        onClick={() => handleSort('title')}
                        className="flex items-center gap-1 hover:text-slate-900"
                      >
                        Title
                        {sortField === 'title' && (
                          <span className="text-xs">{sortDirection === 'asc' ? '↑' : '↓'}</span>
                        )}
                      </button>
                    </th>
                    <th className="text-left py-3 px-6 font-medium text-slate-700">
                      <button 
                        onClick={() => handleSort('category')}
                        className="flex items-center gap-1 hover:text-slate-900"
                      >
                        Category
                        {sortField === 'category' && (
                          <span className="text-xs">{sortDirection === 'asc' ? '↑' : '↓'}</span>
                        )}
                      </button>
                    </th>
                    <th className="text-left py-3 px-6 font-medium text-slate-700">
                      <button 
                        onClick={() => handleSort('severity')}
                        className="flex items-center gap-1 hover:text-slate-900"
                      >
                        Severity
                        {sortField === 'severity' && (
                          <span className="text-xs">{sortDirection === 'asc' ? '↑' : '↓'}</span>
                        )}
                      </button>
                    </th>
                    <th className="text-left py-3 px-6 font-medium text-slate-700">Resources</th>
                    <th className="text-left py-3 px-6 font-medium text-slate-700">
                      <button 
                        onClick={() => handleSort('estimated_savings')}
                        className="flex items-center gap-1 hover:text-slate-900"
                      >
                        Potential Savings
                        {sortField === 'estimated_savings' && (
                          <span className="text-xs">{sortDirection === 'asc' ? '↑' : '↓'}</span>
                        )}
                      </button>
                    </th>
                    <th className="text-left py-3 px-6 font-medium text-slate-700">
                      <button 
                        onClick={() => handleSort('status')}
                        className="flex items-center gap-1 hover:text-slate-900"
                      >
                        Status
                        {sortField === 'status' && (
                          <span className="text-xs">{sortDirection === 'asc' ? '↑' : '↓'}</span>
                        )}
                      </button>
                    </th>
                    <th className="text-left py-3 px-6 font-medium text-slate-700">Actions</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-200">
                  {paginatedFindings.map((finding: Finding, index: number) => {
                    const isEven = index % 2 === 0;
                    const isExpanded = expandedRows.has(finding.id);
                    
                    return (
                      <React.Fragment key={finding.id}>
                        <tr 
                          className={`hover:bg-slate-50 transition-colors ${
                            isEven ? 'bg-white' : 'bg-slate-25'
                          }`}
                        >
                          <td className="py-4 px-6">
                            <div className="flex items-center gap-3">
                              <button
                                onClick={() => toggleRowExpansion(finding.id)}
                                className="p-1 hover:bg-slate-100 rounded"
                              >
                                <ChevronDown className={`w-4 h-4 transition-transform ${
                                  isExpanded ? 'rotate-180' : ''
                                }`} />
                              </button>
                              <div>
                                <div className="font-medium text-slate-900">{finding.title}</div>
                                <div className="text-sm text-slate-600">{finding.detector_name}</div>
                              </div>
                            </div>
                          </td>
                          <td className="py-4 px-6">
                            <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                              {finding.category}
                            </span>
                          </td>
                          <td className="py-4 px-6">
                            <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium border ${
                              getSeverityColor(finding.severity)
                            }`}>
                              {finding.severity}
                            </span>
                          </td>
                          <td className="py-4 px-6">
                            <div className="text-sm font-medium text-slate-900">
                              {finding.resources_count || 1} resource{(finding.resources_count || 1) > 1 ? 's' : ''}
                            </div>
                          </td>
                          <td className="py-4 px-6">
                            <div className="font-medium text-green-600">
                              ${(finding.estimated_savings || 0).toFixed(2)}
                            </div>
                            <div className="text-xs text-slate-600">per month</div>
                          </td>
                          <td className="py-4 px-6">
                            <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium border ${
                              getStatusColor(finding.status || 'Open')
                            }`}>
                              {finding.status || 'Open'}
                            </span>
                          </td>
                          <td className="py-4 px-6">
                            <div className="flex items-center gap-2">
                              <button
                                onClick={() => setSelectedFinding(finding)}
                                className="p-1 text-slate-600 hover:text-blue-600 transition-colors"
                                title="View Resources"
                              >
                                <Eye className="w-4 h-4" />
                              </button>
                              <button
                                onClick={() => handleGenerateIaC(finding)}
                                className="p-1 text-slate-600 hover:text-green-600 transition-colors"
                                title="Generate IaC"
                              >
                                <Code className="w-4 h-4" />
                              </button>
                              <button
                                onClick={() => {
                                  console.log('Whitelist functionality coming soon');
                                }}
                                className="p-1 text-slate-600 hover:text-orange-600 transition-colors opacity-50 cursor-not-allowed"
                                title="Whitelist (Coming Soon)"
                              >
                                <Shield className="w-4 h-4" />
                              </button>
                            </div>
                          </td>
                        </tr>
                        
                        {/* Expanded Row */}
                        {isExpanded && (
                          <tr className="bg-slate-50">
                            <td colSpan={7} className="px-6 py-4">
                              <div className="space-y-3">
                                <div>
                                  <h4 className="font-medium text-slate-900 mb-2">Description</h4>
                                  <p className="text-slate-700">{finding.description}</p>
                                </div>
                                <div>
                                  <h4 className="font-medium text-slate-900 mb-2">Recommendation</h4>
                                  <p className="text-slate-700">Optimize resource configuration to reduce costs and improve efficiency.</p>
                                </div>
                                <div>
                                  <h4 className="font-medium text-slate-900 mb-2">Affected Resources</h4>
                                  <div className="bg-white rounded border p-3">
                                    <code className="text-sm text-slate-600">{finding.resource_arn}</code>
                                  </div>
                                </div>
                              </div>
                            </td>
                          </tr>
                        )}
                      </React.Fragment>
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

        {/* IaC Generation Modal */}
        {showIaCModal && iacFinding && (
          <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-6 z-50">
            <div className="bg-white rounded-lg shadow-xl max-w-4xl w-full max-h-[90vh] overflow-y-auto">
              <div className="p-6 border-b border-slate-200 flex justify-between items-start">
                <div>
                  <h2 className="text-2xl font-bold text-slate-900 mb-2">Generated Infrastructure Code</h2>
                  <p className="text-slate-600">Terraform code for: {iacFinding.title}</p>
                </div>
                <button
                  onClick={() => {
                    setShowIaCModal(false);
                    setIaCFinding(null);
                    setGeneratedIaC(null);
                  }}
                  className="text-slate-400 hover:text-slate-600 transition-colors"
                >
                  <X className="w-6 h-6" />
                </button>
              </div>

              <div className="p-6">
                {iacLoading ? (
                  <div className="flex items-center justify-center py-12">
                    <div className="text-slate-600">Generating infrastructure code...</div>
                  </div>
                ) : generatedIaC ? (
                  <div className="space-y-6">
                    <div className="bg-slate-900 rounded-lg p-4 overflow-x-auto">
                      <pre className="text-green-400 text-sm whitespace-pre-wrap">
                        <code>{generatedIaC.code}</code>
                      </pre>
                    </div>
                    
                    {generatedIaC.instructions && generatedIaC.instructions.length > 0 && (
                      <div>
                        <h3 className="font-semibold text-slate-900 mb-2">Deployment Instructions</h3>
                        <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
                          <ol className="list-decimal list-inside space-y-1 text-blue-800">
                            {generatedIaC.instructions.map((instruction: string, index: number) => (
                              <li key={index} className="text-sm">{instruction}</li>
                            ))}
                          </ol>
                        </div>
                      </div>
                    )}
                    
                    <div className="flex gap-3">
                      <button
                        onClick={() => {
                          navigator.clipboard.writeText(generatedIaC.code);
                        }}
                        className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
                      >
                        <Code className="w-4 h-4" />
                        Copy Code
                      </button>
                      <button
                        onClick={() => {
                          const blob = new Blob([generatedIaC.code], { type: 'text/plain' });
                          const url = URL.createObjectURL(blob);
                          const a = document.createElement('a');
                          a.href = url;
                          a.download = `${iacFinding.id}_optimization.tf`;
                          a.click();
                          URL.revokeObjectURL(url);
                        }}
                        className="flex items-center gap-2 px-4 py-2 border border-slate-300 rounded-lg hover:bg-slate-50"
                      >
                        <Download className="w-4 h-4" />
                        Download
                      </button>
                    </div>
                  </div>
                ) : (
                  <div className="text-center py-12 text-slate-600">
                    Failed to generate code. Please try again.
                  </div>
                )}
              </div>
            </div>
          </div>
        )}

        {/* Details Modal */}
        {selectedFinding && (
          <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-6 z-50">
            <div className="bg-white rounded-lg shadow-xl max-w-4xl w-full max-h-[90vh] overflow-y-auto">
              <div className="p-6 border-b border-slate-200 flex justify-between items-start">
                <div>
                  <h2 className="text-2xl font-bold text-slate-900 mb-2">{selectedFinding.title}</h2>
                  <div className="flex items-center gap-3">
                    <span className={`inline-flex items-center px-3 py-1 rounded-full text-sm font-medium border ${
                      getSeverityColor(selectedFinding.severity)
                    }`}>
                      {selectedFinding.severity}
                    </span>
                    <span className="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-blue-100 text-blue-800">
                      {selectedFinding.category}
                    </span>
                    <span className={`inline-flex items-center px-3 py-1 rounded-full text-sm font-medium border ${
                      getStatusColor(selectedFinding.status || 'Open')
                    }`}>
                      {selectedFinding.status || 'Open'}
                    </span>
                  </div>
                </div>
                <button
                  onClick={() => setSelectedFinding(null)}
                  className="text-slate-400 hover:text-slate-600 transition-colors"
                >
                  <X className="w-6 h-6" />
                </button>
              </div>

              <div className="p-6 space-y-6">
                {/* Overview */}
                <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                  <div className="bg-slate-50 rounded-lg p-4">
                    <div className="flex items-center gap-2 mb-2">
                      <DollarSign className="w-5 h-5 text-green-600" />
                      <h3 className="font-semibold text-slate-900">Potential Savings</h3>
                    </div>
                    <div className="text-2xl font-bold text-green-600">
                      ${(selectedFinding.estimated_savings || 0).toLocaleString()}
                    </div>
                    <div className="text-sm text-slate-600">per month</div>
                  </div>

                  <div className="bg-slate-50 rounded-lg p-4">
                    <div className="flex items-center gap-2 mb-2">
                      <Activity className="w-5 h-5 text-blue-600" />
                      <h3 className="font-semibold text-slate-900">Confidence</h3>
                    </div>
                    <div className="flex items-center gap-3">
                      <div className="flex-1 bg-slate-200 rounded-full h-2">
                        <div
                          className="bg-blue-600 h-2 rounded-full"
                          style={{ width: `${(selectedFinding.confidence || 0) * 100}%` }}
                        />
                      </div>
                      <span className="font-semibold text-slate-900">
                        {((selectedFinding.confidence || 0) * 100).toFixed(0)}%
                      </span>
                    </div>
                  </div>

                  <div className="bg-slate-50 rounded-lg p-4">
                    <div className="flex items-center gap-2 mb-2">
                      <AlertCircle className="w-5 h-5 text-orange-600" />
                      <h3 className="font-semibold text-slate-900">Resources Affected</h3>
                    </div>
                    <div className="text-2xl font-bold text-slate-900">
                      {selectedFinding.resources_count || 1}
                    </div>
                    <div className="text-sm text-slate-600">resource{(selectedFinding.resources_count || 1) > 1 ? 's' : ''}</div>
                  </div>
                </div>

                {/* Description */}
                <div>
                  <h3 className="text-lg font-semibold text-slate-900 mb-3">Description</h3>
                  <p className="text-slate-700 leading-relaxed">{selectedFinding.description}</p>
                </div>

                {/* Recommendation */}
                <div>
                  <h3 className="text-lg font-semibold text-slate-900 mb-3">Recommendation</h3>
                  <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
                    <p className="text-blue-800">
                      Optimize resource configuration to reduce costs and improve efficiency. 
                      Consider implementing automated scaling, rightsizing instances, or 
                      migrating to more cost-effective alternatives.
                    </p>
                  </div>
                </div>

                {/* Affected Resources */}
                <div>
                  <h3 className="text-lg font-semibold text-slate-900 mb-3">Affected Resources</h3>
                  <div className="bg-slate-50 rounded-lg p-4">
                    <div className="space-y-2">
                      <div className="flex items-center justify-between p-3 bg-white rounded border">
                        <div>
                          <div className="font-medium text-slate-900">
                            {selectedFinding.resource_arn.split('/').pop() || 'Resource'}
                          </div>
                          <div className="text-sm text-slate-600">
                            {selectedFinding.resource_arn}
                          </div>
                        </div>
                        <button className="text-blue-600 hover:text-blue-800 text-sm font-medium">
                          View Details
                        </button>
                      </div>
                    </div>
                  </div>
                </div>

                {/* IaC Code Preview */}
                <div>
                  <h3 className="text-lg font-semibold text-slate-900 mb-3">Infrastructure as Code Preview</h3>
                  <div className="bg-slate-900 rounded-lg p-4 overflow-x-auto">
                    <pre className="text-green-400 text-sm">
{`# Terraform configuration to optimize this resource
resource "aws_instance" "optimized" {
  instance_type = "t3.medium"  # Rightsized from t3.large
  
  # Enable detailed monitoring
  monitoring = true
  
  # Auto-scaling configuration
  lifecycle {
    create_before_destroy = true
  }
  
  tags = {
    Environment = "production"
    CostOptimized = "true"
  }
}`}
                    </pre>
                  </div>
                </div>

                {/* Action Buttons */}
                <div className="flex flex-col sm:flex-row gap-3 pt-4 border-t border-slate-200">
                  <button className="flex-1 bg-blue-600 text-white px-6 py-3 rounded-lg hover:bg-blue-700 transition-colors flex items-center justify-center gap-2">
                    <Eye className="w-4 h-4" />
                    View Resources
                  </button>
                  <button 
                    onClick={() => {
                      setSelectedFinding(null);
                      handleGenerateIaC(selectedFinding);
                    }}
                    className="flex-1 bg-green-600 text-white px-6 py-3 rounded-lg hover:bg-green-700 transition-colors flex items-center justify-center gap-2"
                  >
                    <Code className="w-4 h-4" />
                    Generate IaC
                  </button>
                  <button 
                    onClick={() => {
                      console.log('Whitelist functionality coming soon');
                    }}
                    className="flex-1 border border-slate-300 px-6 py-3 rounded-lg hover:bg-slate-50 transition-colors flex items-center justify-center gap-2 opacity-50 cursor-not-allowed"
                    title="Whitelist (Coming Soon)"
                  >
                    <Shield className="w-4 h-4" />
                    Whitelist
                  </button>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default HiddenCosts;
