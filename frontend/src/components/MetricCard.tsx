import React from 'react';

interface MetricCardProps {
  title: string;
  value: string | number;
  change?: string;
  changeType?: 'positive' | 'negative' | 'neutral';
  icon?: string;
}

export const MetricCard: React.FC<MetricCardProps> = ({
  title,
  value,
  change,
  changeType = 'neutral',
  icon
}) => {
  const changeColors = {
    positive: 'bg-green-100 text-green-800',
    negative: 'bg-red-100 text-red-800',
    neutral: 'bg-gray-100 text-gray-800'
  };

  return (
    <div className="bg-white p-6 rounded-lg shadow-sm border-l-4 border-blue-500 min-h-[120px] flex flex-col justify-between">
      <div>
        <div className="text-2xl font-semibold text-gray-900 mb-2">
          {icon && <span className="mr-2">{icon}</span>}
          {value}
        </div>
        <div className="text-sm text-gray-600 uppercase tracking-wide font-medium">
          {title}
        </div>
      </div>
      {change && (
        <div className={`text-xs px-2 py-1 rounded-full inline-block mt-2 ${changeColors[changeType]}`}>
          {change}
        </div>
      )}
    </div>
  );
};