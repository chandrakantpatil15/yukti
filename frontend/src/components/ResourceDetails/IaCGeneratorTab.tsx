import React from 'react';

interface IaCGeneratorTabProps {
  resourceId: string;
  resourceType: string;
}

const IaCGeneratorTab: React.FC<IaCGeneratorTabProps> = ({ resourceId, resourceType }) => {
  return (
    <div className="p-4">
      <h3 className="text-lg font-semibold mb-4">Infrastructure as Code</h3>
      <p className="text-gray-600 mb-4">Generate Terraform or CloudFormation code for this resource.</p>
      <button className="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700">
        Generate Code
      </button>
    </div>
  );
};

export default IaCGeneratorTab;