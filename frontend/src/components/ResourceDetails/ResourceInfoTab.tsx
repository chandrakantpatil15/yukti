import React from 'react';
import { useQuery } from 'react-query';
import api from '../../services/api';
import { Server, MapPin, Cpu, Network, Shield, Tag as TagIcon } from 'lucide-react';

interface ResourceInfoTabProps {
  resourceId: string;
}

const ResourceInfoTab: React.FC<ResourceInfoTabProps> = ({ resourceId }) => {
  const { data, isLoading } = useQuery(
    ['resource-details', resourceId],
    () => api.getResourceDetails(resourceId)
  );

  if (isLoading) {
    return <div className="p-4">Loading...</div>;
  }

  const resource = data?.data;
  const metadata = resource?.metadata || {};
  const tags = metadata?.tags || resource?.tags || {};

  return (
    <div className="p-4 space-y-6">
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
            <span className="text-gray-500">State</span>
            <p className="font-medium capitalize">{resource?.state}</p>
          </div>
          <div>
            <span className="text-gray-500">Platform</span>
            <p className="font-medium capitalize">{metadata.platform || 'Linux'}</p>
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
            <p className="font-medium">{metadata.vpc_id || 'N/A'}</p>
          </div>
          <div>
            <span className="text-gray-500">Subnet ID</span>
            <p className="font-medium">{metadata.subnet_id || 'N/A'}</p>
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
            <p className="font-medium">{metadata.private_ip || 'N/A'}</p>
          </div>
          <div>
            <span className="text-gray-500">Public IP</span>
            <p className="font-medium">{metadata.public_ip || 'N/A'}</p>
          </div>
        </div>
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
            <p className="font-medium">{metadata.ami_id || 'N/A'}</p>
          </div>
          <div>
            <span className="text-gray-500">Architecture</span>
            <p className="font-medium">{metadata.architecture || 'N/A'}</p>
          </div>
          <div>
            <span className="text-gray-500">Root Device</span>
            <p className="font-medium">{metadata.root_device_type || 'N/A'}</p>
          </div>
          <div>
            <span className="text-gray-500">Key Pair</span>
            <p className="font-medium">{metadata.key_name || 'N/A'}</p>
          </div>
          <div>
            <span className="text-gray-500">EBS Optimized</span>
            <p className="font-medium">{metadata.ebs_optimized ? 'Yes' : 'No'}</p>
          </div>
          <div>
            <span className="text-gray-500">Tenancy</span>
            <p className="font-medium capitalize">{metadata.tenancy || 'default'}</p>
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
            <p className="font-medium">{metadata.detailed_monitoring ? 'Enabled' : 'Disabled'}</p>
          </div>
          <div>
            <span className="text-gray-500">Launch Time</span>
            <p className="font-medium">{metadata.launch_time ? new Date(metadata.launch_time).toLocaleString() : 'N/A'}</p>
          </div>
        </div>
      </div>

      {/* Tags - Dynamic */}
      {Object.keys(tags).length > 0 && (
        <div>
          <h3 className="text-lg font-semibold mb-3 flex items-center gap-2">
            <TagIcon className="w-5 h-5" />
            Tags ({Object.keys(tags).length})
          </h3>
          <div className="space-y-2">
            {Object.entries(tags).map(([key, value]) => (
              <div key={key} className="flex items-start gap-2 text-sm">
                <span className="px-2 py-1 bg-gray-100 rounded font-medium min-w-[100px]">{key}</span>
                <span className="text-gray-600">=</span>
                <span className="px-2 py-1 bg-blue-50 rounded flex-1 break-all">{String(value)}</span>
              </div>
            ))}
          </div>
        </div>
      )}
      
      {/* All Other Metadata - Dynamic */}
      {Object.keys(metadata).filter(k => !['tags', 'instance_id', 'instance_type', 'state', 'region', 'availability_zone', 'tenancy', 'vpc_id', 'subnet_id', 'private_ip', 'public_ip', 'ami_id', 'architecture', 'root_device_type', 'key_name', 'ebs_optimized', 'detailed_monitoring', 'launch_time', 'platform', 'private_dns', 'public_dns'].includes(k)).length > 0 && (
        <div>
          <h3 className="text-lg font-semibold mb-3">Additional Details</h3>
          <div className="grid grid-cols-2 gap-3 text-sm">
            {Object.entries(metadata)
              .filter(([key]) => !['tags', 'instance_id', 'instance_type', 'state', 'region', 'availability_zone', 'tenancy', 'vpc_id', 'subnet_id', 'private_ip', 'public_ip', 'ami_id', 'architecture', 'root_device_type', 'key_name', 'ebs_optimized', 'detailed_monitoring', 'launch_time', 'platform', 'private_dns', 'public_dns'].includes(key))
              .map(([key, value]) => (
                <div key={key}>
                  <span className="text-gray-500 capitalize">{key.replace(/_/g, ' ')}</span>
                  <p className="font-medium break-all">
                    {Array.isArray(value) ? value.join(', ') : String(value)}
                  </p>
                </div>
              ))}
          </div>
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