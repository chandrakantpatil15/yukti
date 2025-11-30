import React from 'react';
import { BarChart3 } from 'lucide-react';

interface CostChartProps {
  data: any; // TODO: Add proper type
}

export const CostChart = ({ data }: CostChartProps) => {
  return (
    <div className="h-64 flex items-center justify-center">
      <div className="flex flex-col items-center text-neutral-500 dark:text-neutral-400">
        <BarChart3 className="w-12 h-12 mb-4" />
        <p>Cost data visualization will be implemented here</p>
      </div>
    </div>
  );
};