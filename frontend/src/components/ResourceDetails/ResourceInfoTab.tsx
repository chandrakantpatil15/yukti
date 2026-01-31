import React, { useState } from 'react';
import { useQuery } from 'react-query';
import api from '../../services/api';
import { Server, MapPin, Cpu, Network, Shield, Tag as TagIcon, FileText, Clock, Activity } from 'lucide-react';

interface ResourceInfoTabProps {
  resourceId: string;
}

const ResourceInfoTab: React.FC<ResourceInfoTabProps> = ({ resourceId }) => {
  const { data, isLoading } = useQuery(
    ['resource-details', resourceId],
    () => api.getResourceDetails(resourceId)
  );

  const [currentPage, setCurrentPage] = useState(1);
  const ITEMS_PER_PAGE = 8;

  if (isLoading) {
    return (
      <div className="p-4 space-y-4">
        <div className="animate-pulse">
          <div className="h-4 bg-gray-200 rounded w-3/4 mb-2"></div>
          <div className="h-4 bg-gray-200 rounded w-1/2"></div>
        </div>
      </div>
    );
  }

  const resource = data?.data;
  const statusInfo = resource?.status_info || {};
  const networkInfo = resource?.network_info || {};
  const configInfo = resource?.config_info || {};
  const storageInfo = resource?.storage_info || {};
  const tags = resource?.tags || {};
  const metadata = resource?.metadata || {};

  // Get status color based on state
  const getStatusColor = (state: string) => {
    switch (state?.toLowerCase()) {
      case 'running': return 'bg-green-100 text-green-800';
      case 'stopped': return 'bg-red-100 text-red-800';
      case 'stopping': return 'bg-yellow-100 text-yellow-800';
      case 'starting': return 'bg-blue-100 text-blue-800';
      case 'terminated': return 'bg-gray-100 text-gray-800';
      case 'terminating': return 'bg-orange-100 text-orange-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  return (
    <div className="p-4 space-y-6 max-h-[calc(100vh-12rem)] overflow-y-auto">
      {/* Status Overview */}
      <div>
        <h3 className="text-lg font-semibold mb-3 flex items-center gap-2">
          <Activity className="w-5 h-5" />
          Status Overview
        </h3>
        <div className="grid grid-cols-2 gap-3 text-sm">
          <div>
            <span className="text-gray-500">Current State</span>
            <div className="flex items-center gap-2 mt-1">
              <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                getStatusColor(resource?.state)
              }`}>
                {resource?.state || 'Unknown'}
              </span>
            </div>
          </div>
          <div>
            <span className="text-gray-500">Uptime</span>
            <p className="font-medium">
              {statusInfo.uptime_days ? `${statusInfo.uptime_days} days` : 'N/A'}
            </p>
          </div>
          <div>
            <span className="text-gray-500">State Reason</span>
            <p className="font-medium text-xs">{statusInfo.state_reason || 'N/A'}</p>
          </div>
          <div>
            <span className="text-gray-500">Launch Time</span>
            <p className="font-medium text-xs">
              {statusInfo.launch_time ? new Date(statusInfo.launch_time).toLocaleString() : 'N/A'}
            </p>
          </div>
        </div>
      </div>
      {/* Instance Overview */}
      <div>
        <h3 className="text-lg font-semibold mb-3 flex items-center gap-2">
          <Server className="w-5 h-5" />
          Instance Overview
        </h3>
        <div className="grid grid-cols-2 gap-3 text-sm">
          <div>
            <span className="text-gray-500">Instance ID</span>
            <p className="font-medium">{resource?.resource_id}</p>
          </div>
          <div>
            <span className="text-gray-500">Instance Type</span>
            <p className="font-medium">{resource?.instance_type}</p>
          </div>
          <div>
            <span className="text-gray-500">Platform</span>
            <p className="font-medium capitalize">{configInfo.platform || 'Linux'}</p>
          </div>
          <div>
            <span className="text-gray-500">Architecture</span>
            <p className="font-medium">{configInfo.architecture || 'N/A'}</p>
          </div>
        </div>
      </div>

      {/* Location */}
      <div>
        <h3 className="text-lg font-semibold mb-3 flex items-center gap-2">
          <MapPin className="w-5 h-5" />
          Location
        </h3>
        <div className="grid grid-cols-2 gap-3 text-sm">
          <div>
            <span className="text-gray-500">Region</span>
            <p className="font-medium">{resource?.region}</p>
          </div>
          <div>
            <span className="text-gray-500">Availability Zone</span>
            <p className="font-medium">{metadata.availability_zone || 'N/A'}</p>
          </div>
          <div>
            <span className="text-gray-500">VPC ID</span>
            <p className="font-medium">{networkInfo.vpc_id || 'N/A'}</p>
          </div>
          <div>
            <span className="text-gray-500">Subnet ID</span>
            <p className="font-medium">{networkInfo.subnet_id || 'N/A'}</p>
          </div>
        </div>
      </div>

      {/* Network */}
      <div>
        <h3 className="text-lg font-semibold mb-3 flex items-center gap-2">
          <Network className="w-5 h-5" />
          Network
        </h3>
        <div className="grid grid-cols-2 gap-3 text-sm">
          <div>
            <span className="text-gray-500">Private IP</span>
            <p className="font-medium">{networkInfo.private_ip || 'N/A'}</p>
          </div>
          <div>
            <span className="text-gray-500">Public IP</span>
            <p className="font-medium">{networkInfo.public_ip || 'N/A'}</p>
          </div>
          <div>
            <span className="text-gray-500">Private DNS</span>
            <p className="font-medium text-xs">{networkInfo.private_dns || 'N/A'}</p>
          </div>
          <div>
            <span className="text-gray-500">Public DNS</span>
            <p className="font-medium text-xs">{networkInfo.public_dns || 'N/A'}</p>
          </div>
        </div>
        {networkInfo.security_groups && Array.isArray(networkInfo.security_groups) && (
          <div className="mt-3">
            <span className="text-gray-500 text-sm">Security Groups</span>
            <div className="flex flex-wrap gap-2 mt-1">
              {networkInfo.security_groups.map((sg: any, i: number) => (
                <span key={i} className="px-2 py-1 bg-blue-50 text-blue-700 rounded text-xs">
                  {sg.group_id || sg}
                </span>
              ))}
            </div>
          </div>
        )}
      </div>

      {/* Storage */}
      <div>
        <h3 className="text-lg font-semibold mb-3 flex items-center gap-2">
          <FileText className="w-5 h-5" />
          Storage
        </h3>
        <div className="grid grid-cols-2 gap-3 text-sm">
          <div>
            <span className="text-gray-500">Root Device Type</span>
            <p className="font-medium">{storageInfo.root_device_type || 'N/A'}</p>
          </div>
          <div>
            <span className="text-gray-500">Root Device Name</span>
            <p className="font-medium">{storageInfo.root_device_name || 'N/A'}</p>
          </div>
        </div>
        {storageInfo.ebs_volumes && Array.isArray(storageInfo.ebs_volumes) && storageInfo.ebs_volumes.length > 0 && (
          <div className="mt-3">
            <span className="text-gray-500 text-sm">EBS Volumes ({storageInfo.ebs_volumes.length})</span>
            <div className="mt-2 space-y-2">
              {storageInfo.ebs_volumes.slice(0, 3).map((volume: any, i: number) => (
                <div key={i} className="p-2 bg-gray-50 rounded text-sm">
                  <div className="font-medium">{volume.volume_id || 'Unknown'}</div>
                  <div className="text-gray-600">
                    {volume.size && `${volume.size} GiB`} 
                    {volume.device_name && ` • ${volume.device_name}`}
                    {volume.volume_type && ` • ${volume.volume_type}`}
                  </div>
                </div>
              ))}
              {storageInfo.ebs_volumes.length > 3 && (
                <div className="text-xs text-gray-500">+{storageInfo.ebs_volumes.length - 3} more volumes</div>
              )}
            </div>
          </div>
        )}
      </div>

      {/* Configuration */}
      <div>
        <h3 className="text-lg font-semibold mb-3 flex items-center gap-2">
          <Cpu className="w-5 h-5" />
          Configuration
        </h3>
        <div className="grid grid-cols-2 gap-3 text-sm">
          <div>
            <span className="text-gray-500">AMI ID</span>
            <p className="font-medium">{configInfo.ami_id || 'N/A'}</p>
          </div>
          <div>
            <span className="text-gray-500">Key Pair</span>
            <p className="font-medium">{configInfo.key_name || 'N/A'}</p>
          </div>
          <div>
            <span className="text-gray-500">EBS Optimized</span>
            <p className="font-medium">{configInfo.ebs_optimized ? 'Yes' : 'No'}</p>
          </div>
          <div>
            <span className="text-gray-500">Tenancy</span>
            <p className="font-medium capitalize">{configInfo.tenancy || 'default'}</p>
          </div>
        </div>
      </div>

      {/* Monitoring */}
      <div>
        <h3 className="text-lg font-semibold mb-3 flex items-center gap-2">
          <Shield className="w-5 h-5" />
          Monitoring
        </h3>
        <div className="grid grid-cols-2 gap-3 text-sm">
          <div>
            <span className="text-gray-500">Detailed Monitoring</span>
            <p className="font-medium">{configInfo.monitoring ? 'Enabled' : 'Disabled'}</p>
          </div>
          <div>
            <span className="text-gray-500">CloudWatch Metrics</span>
            <p className="font-medium">Available</p>
          </div>
        </div>
      </div>

      {/* Tags with Pagination */}
      {Object.keys(tags).length > 0 && (
        <div>
          <h3 className="text-lg font-semibold mb-3 flex items-center gap-2">
            <TagIcon className="w-5 h-5" />
            Tags ({Object.keys(tags).length})
          </h3>
          {(() => {
            const tagEntries = Object.entries(tags);
            const totalPages = Math.ceil(tagEntries.length / ITEMS_PER_PAGE);
            const startIndex = (currentPage - 1) * ITEMS_PER_PAGE;
            const currentTags = tagEntries.slice(startIndex, startIndex + ITEMS_PER_PAGE);

            return (
              <div>
                <div className="grid grid-cols-1 gap-2">
                  {currentTags.map(([key, value]) => (
                    <div key={key} className="flex justify-between items-center p-2 bg-gray-50 rounded">
                      <span className="font-medium text-sm">{key}</span>
                      <span className="text-sm text-gray-600">{String(value)}</span>
                    </div>
                  ))}
                </div>
                {totalPages > 1 && (
                  <div className="flex items-center justify-between mt-3">
                    <button
                      onClick={() => setCurrentPage(p => Math.max(1, p - 1))}
                      className="px-3 py-1 bg-gray-100 rounded text-sm disabled:opacity-50"
                      disabled={currentPage === 1}
                    >
                      Previous
                    </button>
                    <span className="text-sm text-gray-600">
                      Page {currentPage} of {totalPages}
                    </span>
                    <button
                      onClick={() => setCurrentPage(p => Math.min(totalPages, p + 1))}
                      className="px-3 py-1 bg-gray-100 rounded text-sm disabled:opacity-50"
                      disabled={currentPage === totalPages}
                    >
                      Next
                    </button>
                  </div>
                )}
              </div>
            );
          })()}
        </div>
      )}

      {/* Cost */}
      <div>
        <h3 className="text-lg font-semibold mb-3">Cost Estimate</h3>
        <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
          <div className="text-2xl font-bold text-blue-600">
            ${resource?.monthly_cost?.toFixed(2) || '0.00'}/month
          </div>
          <p className="text-sm text-gray-600 mt-1">Estimated monthly cost</p>
        </div>
      </div>
    </div>
  );
};

export default ResourceInfoTab;