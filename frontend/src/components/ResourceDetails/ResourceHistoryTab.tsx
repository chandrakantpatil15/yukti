import React from 'react';

interface ResourceHistoryTabProps {
  resourceId: string;
  resourceType: string;
}

const ResourceHistoryTab: React.FC<ResourceHistoryTabProps> = ({ resourceId, resourceType }) => {
  return (
    <div className="p-4">
      <h3 className="text-lg font-semibold mb-4">Resource History</h3>
      <p className="text-gray-600">No history data available yet.</p>
    </div>
  );
};

export default ResourceHistoryTab;