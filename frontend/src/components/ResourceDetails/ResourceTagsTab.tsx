import React from 'react';
import { useQuery } from 'react-query';
import api from '../../services/api';
import { Tag } from 'lucide-react';

interface ResourceTagsTabProps {
  resourceId: string;
}

const ResourceTagsTab: React.FC<ResourceTagsTabProps> = ({ resourceId }) => {
  const { data, isLoading } = useQuery(
    ['resource-details', resourceId],
    () => api.getResourceDetails(resourceId)
  );

  if (isLoading) {
    return <div className="p-4">Loading tags...</div>;
  }

  const resource = data?.data;
  const tags = resource?.metadata?.tags || resource?.tags || {};

  return (
    <div className="p-4">
      <h3 className="text-lg font-semibold mb-4 flex items-center gap-2">
        <Tag className="w-5 h-5" />
        Resource Tags ({Object.keys(tags).length})
      </h3>
      {Object.keys(tags).length > 0 ? (
        <div className="space-y-3">
          {Object.entries(tags).map(([key, value]) => (
            <div key={key} className="flex items-start gap-3 p-3 bg-gray-50 rounded-lg">
              <span className="font-medium text-gray-700 min-w-[120px]">{key}:</span>
              <span className="text-gray-900 break-all">{String(value)}</span>
            </div>
          ))}
        </div>
      ) : (
        <p className="text-gray-500 italic">No tags found for this resource</p>
      )}
    </div>
  );
};

export default ResourceTagsTab;