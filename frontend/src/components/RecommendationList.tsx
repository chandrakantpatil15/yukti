import React from 'react';
import { Lightbulb } from 'lucide-react';

const RecommendationList: React.FC<{ recommendations: any[] }> = ({ recommendations }) => {
  return (
    <div className="flex items-center justify-center py-8">
      <div className="flex flex-col items-center text-neutral-500 dark:text-neutral-400">
        <Lightbulb className="w-12 h-12 mb-4" />
        <p>Recommendations will be implemented here</p>
      </div>
    </div>
  );
};

export { RecommendationList };