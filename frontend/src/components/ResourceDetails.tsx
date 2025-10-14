import React, { useState } from 'react';
import { Resource } from '../types';

interface ResourceDetailsProps {
  resource: Resource;
  onClose: () => void;
}

export const ResourceDetails: React.FC<ResourceDetailsProps> = ({ resource, onClose }) => {
  const [activeTab, setActiveTab] = useState<'overview' | 'resources' | 'billing' | 'compliance' | 'tags'>('overview');
  const getStatusColor = (status: string) => {
    switch (status) {
      case 'running': return 'bg-green-100 text-green-800 border-green-200';
      case 'stopped': return 'bg-red-100 text-red-800 border-red-200';
      default: return 'bg-gray-100 text-gray-800 border-gray-200';
    }
  };

  const getUtilizationLevel = (utilization: number) => {
    if (utilization > 70) return { level: 'High', color: 'text-green-600', bg: 'bg-green-100' };
    if (utilization > 40) return { level: 'Medium', color: 'text-yellow-600', bg: 'bg-yellow-100' };
    return { level: 'Low', color: 'text-red-600', bg: 'bg-red-100' };
  };

  const utilizationInfo = getUtilizationLevel(resource.cpuUtilization);

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg shadow-xl max-w-4xl w-full mx-4 max-h-[90vh] overflow-y-auto">
        {/* Header */}
        <div className="bg-gradient-to-r from-blue-600 to-purple-600 text-white p-6 rounded-t-lg">
          <div className="flex justify-between items-start">
            <div>
              <h2 className="text-2xl font-bold">{resource.resourceId}</h2>
              <p className="text-blue-100 mt-1">{resource.instanceType} • {resource.region}</p>
            </div>
            <button 
              onClick={onClose}
              className="text-white hover:text-gray-200 text-2xl font-bold"
            >
              ×
            </button>
          </div>
        </div>

        {/* Tab Navigation */}
        <div className="border-b border-gray-200">
          <nav className="flex space-x-8 px-6">
            {[
              { id: 'overview', label: '📊 Overview', icon: '📊' },
              { id: 'resources', label: '🔗 Resources', icon: '🔗' },
              { id: 'billing', label: '💰 Billing', icon: '💰' },
              { id: 'compliance', label: '🔒 Security', icon: '🔒' },
              { id: 'tags', label: '🏷️ Tags', icon: '🏷️' }
            ].map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id as any)}
                className={`py-4 px-1 border-b-2 font-medium text-sm ${
                  activeTab === tab.id
                    ? 'border-blue-500 text-blue-600'
                    : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                }`}
              >
                {tab.label}
              </button>
            ))}
          </nav>
        </div>

        <div className="p-6">
          {/* Tab Content */}
          {activeTab === 'overview' && (
            <div>
              {/* Status and Key Metrics */}
              <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
            <div className="text-center">
              <span className={`inline-flex px-3 py-1 text-sm font-semibold rounded-full border ${getStatusColor(resource.status)}`}>
                {resource.status.toUpperCase()}
              </span>
              <p className="text-xs text-gray-500 mt-1">Current Status</p>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-purple-600">${resource.monthlyCost}</div>
              <p className="text-xs text-gray-500">Monthly Cost</p>
            </div>
            <div className="text-center">
              <div className={`text-2xl font-bold ${utilizationInfo.color}`}>{resource.cpuUtilization}%</div>
              <p className="text-xs text-gray-500">CPU Utilization</p>
            </div>
            <div className="text-center">
              <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${utilizationInfo.bg} ${utilizationInfo.color}`}>
                {utilizationInfo.level}
              </span>
              <p className="text-xs text-gray-500 mt-1">Performance</p>
            </div>
          </div>

          {/* Instance Details */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {/* Basic Information */}
            <div className="bg-gray-50 rounded-lg p-4">
              <h3 className="text-lg font-semibold text-gray-900 mb-4">Instance Information</h3>
              <div className="space-y-3">
                <div className="flex justify-between">
                  <span className="text-gray-600">Instance Type:</span>
                  <span className="font-medium">{resource.instanceType}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600">Region:</span>
                  <span className="font-medium">{resource.region}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600">Environment:</span>
                  <span className="font-medium capitalize">{resource.environment}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600">Launch Time:</span>
                  <span className="font-medium">2024-01-15 09:30 UTC</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-gray-600">Uptime:</span>
                  <span className="font-medium">15 days, 6 hours</span>
                </div>
              </div>
            </div>

            {/* Performance Metrics */}
            <div className="bg-gray-50 rounded-lg p-4">
              <h3 className="text-lg font-semibold text-gray-900 mb-4">Performance Metrics</h3>
              <div className="space-y-4">
                <div>
                  <div className="flex justify-between mb-1">
                    <span className="text-gray-600">CPU Utilization</span>
                    <span className="font-medium">{resource.cpuUtilization}%</span>
                  </div>
                  <div className="w-full bg-gray-200 rounded-full h-2">
                    <div 
                      className={`h-2 rounded-full ${resource.cpuUtilization > 70 ? 'bg-green-500' : resource.cpuUtilization > 40 ? 'bg-yellow-500' : 'bg-red-500'}`}
                      style={{ width: `${resource.cpuUtilization}%` }}
                    ></div>
                  </div>
                </div>
                <div>
                  <div className="flex justify-between mb-1">
                    <span className="text-gray-600">Memory Utilization</span>
                    <span className="font-medium">45%</span>
                  </div>
                  <div className="w-full bg-gray-200 rounded-full h-2">
                    <div className="bg-yellow-500 h-2 rounded-full" style={{ width: '45%' }}></div>
                  </div>
                </div>
                <div>
                  <div className="flex justify-between mb-1">
                    <span className="text-gray-600">Network I/O</span>
                    <span className="font-medium">2.3 GB/day</span>
                  </div>
                </div>
                <div>
                  <div className="flex justify-between mb-1">
                    <span className="text-gray-600">Disk I/O</span>
                    <span className="font-medium">156 MB/day</span>
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* Cost Breakdown */}
          <div className="mt-6 bg-blue-50 rounded-lg p-4">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Cost Breakdown</h3>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div className="text-center">
                <div className="text-xl font-bold text-blue-600">${(resource.monthlyCost * 0.7).toFixed(2)}</div>
                <p className="text-sm text-gray-600">Compute Cost</p>
              </div>
              <div className="text-center">
                <div className="text-xl font-bold text-blue-600">${(resource.monthlyCost * 0.2).toFixed(2)}</div>
                <p className="text-sm text-gray-600">Storage Cost</p>
              </div>
              <div className="text-center">
                <div className="text-xl font-bold text-blue-600">${(resource.monthlyCost * 0.1).toFixed(2)}</div>
                <p className="text-sm text-gray-600">Network Cost</p>
              </div>
            </div>
          </div>

          {/* Optimization Recommendations */}
          <div className="mt-6 bg-yellow-50 rounded-lg p-4">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">💡 Optimization Recommendations</h3>
            <div className="space-y-3">
              {resource.cpuUtilization < 30 && (
                <div className="flex items-start space-x-3 p-3 bg-white rounded border-l-4 border-yellow-400">
                  <div className="text-yellow-500 text-xl">⚠️</div>
                  <div>
                    <p className="font-medium text-gray-900">Right-size Instance</p>
                    <p className="text-sm text-gray-600">
                      Consider downsizing to t3.small to save ~${(resource.monthlyCost * 0.4).toFixed(2)}/month
                    </p>
                  </div>
                </div>
              )}
              {resource.status === 'stopped' && (
                <div className="flex items-start space-x-3 p-3 bg-white rounded border-l-4 border-red-400">
                  <div className="text-red-500 text-xl">🛑</div>
                  <div>
                    <p className="font-medium text-gray-900">Terminate Unused Instance</p>
                    <p className="text-sm text-gray-600">
                      This instance has been stopped for 7+ days. Consider terminating to save ${resource.monthlyCost}/month
                    </p>
                  </div>
                </div>
              )}
              <div className="flex items-start space-x-3 p-3 bg-white rounded border-l-4 border-green-400">
                <div className="text-green-500 text-xl">💰</div>
                <div>
                  <p className="font-medium text-gray-900">Reserved Instance Savings</p>
                  <p className="text-sm text-gray-600">
                    Switch to 1-year Reserved Instance to save up to 40% (~${(resource.monthlyCost * 0.4).toFixed(2)}/month)
                  </p>
                </div>
              </div>
            </div>
          </div>

            </div>
          )}

          {/* Associated Resources Tab */}
          {activeTab === 'resources' && (
            <div>
              <h3 className="text-lg font-semibold text-gray-900 mb-4">Associated AWS Resources</h3>
              <div className="space-y-4">
                {resource.associatedResources?.map((assocResource, index) => (
                  <div key={index} className="bg-gray-50 rounded-lg p-4 border">
                    <div className="flex justify-between items-start mb-2">
                      <div>
                        <h4 className="font-medium text-gray-900">{assocResource.resourceType}</h4>
                        <p className="text-sm text-gray-600">{assocResource.resourceId}</p>
                      </div>
                      <div className="text-right">
                        <div className="text-lg font-semibold text-purple-600">${assocResource.monthlyCost}</div>
                        <div className={`text-xs px-2 py-1 rounded-full ${
                          assocResource.status === 'active' ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-800'
                        }`}>
                          {assocResource.status}
                        </div>
                      </div>
                    </div>
                    <div className="flex flex-wrap gap-1 mt-2">
                      {Object.entries(assocResource.tags || {}).map(([key, value]) => (
                        <span key={key} className="text-xs bg-blue-100 text-blue-800 px-2 py-1 rounded">
                          {key}: {value}
                        </span>
                      ))}
                    </div>
                  </div>
                )) || (
                  <div className="text-center py-8 text-gray-500">
                    No associated resources found
                  </div>
                )}
              </div>
            </div>
          )}

          {/* Billing Tab */}
          {activeTab === 'billing' && (
            <div>
              <h3 className="text-lg font-semibold text-gray-900 mb-4">Detailed Billing Breakdown</h3>
              <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                <div className="bg-gray-50 rounded-lg p-4">
                  <h4 className="font-medium text-gray-900 mb-4">Cost Components</h4>
                  <div className="space-y-3">
                    <div className="flex justify-between items-center">
                      <span className="text-gray-600">EC2 Compute</span>
                      <span className="font-medium">${resource.billingBreakdown?.compute || (resource.monthlyCost * 0.6).toFixed(2)}</span>
                    </div>
                    <div className="flex justify-between items-center">
                      <span className="text-gray-600">EBS Storage</span>
                      <span className="font-medium">${resource.billingBreakdown?.storage || (resource.monthlyCost * 0.2).toFixed(2)}</span>
                    </div>
                    <div className="flex justify-between items-center">
                      <span className="text-gray-600">Network/Data Transfer</span>
                      <span className="font-medium">${resource.billingBreakdown?.network || (resource.monthlyCost * 0.1).toFixed(2)}</span>
                    </div>
                    <div className="flex justify-between items-center">
                      <span className="text-gray-600">Associated Services</span>
                      <span className="font-medium">${resource.billingBreakdown?.associatedServices || (resource.monthlyCost * 0.1).toFixed(2)}</span>
                    </div>
                    <hr className="my-2" />
                    <div className="flex justify-between items-center font-semibold">
                      <span>Total Monthly Cost</span>
                      <span className="text-lg text-purple-600">${resource.monthlyCost}</span>
                    </div>
                  </div>
                </div>
                <div className="bg-blue-50 rounded-lg p-4">
                  <h4 className="font-medium text-gray-900 mb-4">Cost Allocation by Tags</h4>
                  <div className="space-y-2">
                    <div className="flex justify-between">
                      <span className="text-sm text-gray-600">Environment: {resource.tags?.Environment || resource.environment}</span>
                      <span className="text-sm font-medium">${(resource.monthlyCost * 1).toFixed(2)}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-sm text-gray-600">Project: {resource.tags?.Project || 'Untagged'}</span>
                      <span className="text-sm font-medium">${(resource.monthlyCost * 1).toFixed(2)}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-sm text-gray-600">Owner: {resource.tags?.Owner || 'Unassigned'}</span>
                      <span className="text-sm font-medium">${(resource.monthlyCost * 1).toFixed(2)}</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* Security Compliance Tab */}
          {activeTab === 'compliance' && (
            <div>
              <h3 className="text-lg font-semibold text-gray-900 mb-4">Security & Compliance</h3>
              <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                <div className="bg-gray-50 rounded-lg p-4">
                  <div className="flex justify-between items-center mb-4">
                    <h4 className="font-medium text-gray-900">Compliance Score</h4>
                    <div className={`text-2xl font-bold ${
                      (resource.securityCompliance?.score || 75) >= 80 ? 'text-green-600' :
                      (resource.securityCompliance?.score || 75) >= 60 ? 'text-yellow-600' : 'text-red-600'
                    }`}>
                      {resource.securityCompliance?.score || 75}%
                    </div>
                  </div>
                  <div className="w-full bg-gray-200 rounded-full h-3 mb-4">
                    <div 
                      className={`h-3 rounded-full ${
                        (resource.securityCompliance?.score || 75) >= 80 ? 'bg-green-500' :
                        (resource.securityCompliance?.score || 75) >= 60 ? 'bg-yellow-500' : 'bg-red-500'
                      }`}
                      style={{ width: `${resource.securityCompliance?.score || 75}%` }}
                    ></div>
                  </div>
                  <p className="text-sm text-gray-600">Last audit: {resource.securityCompliance?.lastAudit || '2024-01-15'}</p>
                </div>
                <div className="bg-red-50 rounded-lg p-4">
                  <h4 className="font-medium text-gray-900 mb-4">Security Issues</h4>
                  <div className="space-y-2">
                    {(resource.securityCompliance?.issues || [
                      { severity: 'medium', category: 'Network', description: 'Security group allows 0.0.0.0/0', recommendation: 'Restrict source IPs' },
                      { severity: 'low', category: 'Tagging', description: 'Missing required tags', recommendation: 'Add Owner and Project tags' }
                    ]).map((issue, index) => (
                      <div key={index} className={`p-2 rounded border-l-4 ${
                        issue.severity === 'critical' ? 'border-red-500 bg-red-50' :
                        issue.severity === 'high' ? 'border-orange-500 bg-orange-50' :
                        issue.severity === 'medium' ? 'border-yellow-500 bg-yellow-50' :
                        'border-blue-500 bg-blue-50'
                      }`}>
                        <div className="flex justify-between items-start">
                          <div>
                            <p className="text-sm font-medium">{issue.category}</p>
                            <p className="text-xs text-gray-600">{issue.description}</p>
                          </div>
                          <span className={`text-xs px-2 py-1 rounded-full ${
                            issue.severity === 'critical' ? 'bg-red-100 text-red-800' :
                            issue.severity === 'high' ? 'bg-orange-100 text-orange-800' :
                            issue.severity === 'medium' ? 'bg-yellow-100 text-yellow-800' :
                            'bg-blue-100 text-blue-800'
                          }`}>
                            {issue.severity}
                          </span>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* Tags Tab */}
          {activeTab === 'tags' && (
            <div>
              <h3 className="text-lg font-semibold text-gray-900 mb-4">Resource Tags & Governance</h3>
              <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                <div className="bg-gray-50 rounded-lg p-4">
                  <h4 className="font-medium text-gray-900 mb-4">Current Tags</h4>
                  <div className="space-y-2">
                    {Object.entries(resource.tags || {
                      Environment: resource.environment,
                      Project: 'web-app',
                      Owner: 'devops-team',
                      CostCenter: '12345'
                    }).map(([key, value]) => (
                      <div key={key} className="flex justify-between items-center p-2 bg-white rounded border">
                        <span className="font-medium text-gray-700">{key}</span>
                        <span className="text-gray-600">{value}</span>
                      </div>
                    ))}
                  </div>
                </div>
                <div className="bg-yellow-50 rounded-lg p-4">
                  <h4 className="font-medium text-gray-900 mb-4">Tag Compliance</h4>
                  <div className="space-y-3">
                    <div className="flex items-center justify-between">
                      <span className="text-sm">Required Tags</span>
                      <span className="text-sm font-medium text-green-600">4/6 Present</span>
                    </div>
                    <div className="space-y-2">
                      {[
                        { tag: 'Environment', present: true, required: true },
                        { tag: 'Project', present: true, required: true },
                        { tag: 'Owner', present: true, required: true },
                        { tag: 'CostCenter', present: true, required: true },
                        { tag: 'Backup', present: false, required: true },
                        { tag: 'Schedule', present: false, required: true }
                      ].map((item, index) => (
                        <div key={index} className="flex items-center justify-between text-sm">
                          <span className={item.required ? 'font-medium' : ''}>{item.tag}</span>
                          <span className={`px-2 py-1 rounded-full text-xs ${
                            item.present ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'
                          }`}>
                            {item.present ? '✓' : '✗'}
                          </span>
                        </div>
                      ))}
                    </div>
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* Action Buttons */}
          <div className="mt-6 flex justify-end space-x-3">
            <button className="px-4 py-2 text-gray-600 border border-gray-300 rounded-md hover:bg-gray-50">
              Export Report
            </button>
            <button className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700">
              Apply Recommendations
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};