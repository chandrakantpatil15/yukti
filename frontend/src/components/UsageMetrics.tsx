import React from 'react';
import { Activity } from 'lucide-react';

const UsageMetrics: React.FC<{ metrics: any }> = ({ metrics }) => {
  return (
    <div className="flex items-center justify-center py-8">
      <div className="flex flex-col items-center text-neutral-500 dark:text-neutral-400">
        <Activity className="w-12 h-12 mb-4" />
        <p>Usage metrics visualization will be implemented here</p>
      </div>
    </div>
  );
};

export { UsageMetrics };