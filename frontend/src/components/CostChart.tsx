import React from 'react';
import { Doughnut, Bar } from 'react-chartjs-2';
import {
  Chart as ChartJS,
  ArcElement,
  Tooltip,
  Legend,
  CategoryScale,
  LinearScale,
  BarElement,
} from 'chart.js';
import { CostSummary } from '../types';

ChartJS.register(ArcElement, Tooltip, Legend, CategoryScale, LinearScale, BarElement);

interface CostChartProps {
  data: CostSummary;
  type: 'doughnut' | 'bar';
  title: string;
}

export const CostChart: React.FC<CostChartProps> = ({ data, type, title }) => {
  const doughnutData = {
    labels: ['Current Cost', 'Potential Savings'],
    datasets: [{
      data: [data.totalMonthlyCost - data.potentialSavings, data.potentialSavings],
      backgroundColor: ['#6b7280', '#10b981'],
      borderWidth: 0,
    }]
  };

  const barData = {
    labels: ['Current', 'After Optimization'],
    datasets: [{
      label: 'Monthly Cost ($)',
      data: [data.totalMonthlyCost, data.totalMonthlyCost - data.potentialSavings],
      backgroundColor: ['#ef4444', '#10b981'],
      borderRadius: 4,
    }]
  };

  const options = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: {
        position: 'bottom' as const,
        labels: {
          padding: 20,
          usePointStyle: true,
        }
      }
    }
  };

  return (
    <div className="bg-white p-6 rounded-lg shadow-sm min-h-[350px]">
      <h3 className="text-lg font-semibold text-gray-900 mb-4">{title}</h3>
      <div className="h-72">
        {type === 'doughnut' ? (
          <Doughnut data={doughnutData} options={options} />
        ) : (
          <Bar data={barData} options={options} />
        )}
      </div>
    </div>
  );
};