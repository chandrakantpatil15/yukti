import React from 'react';
import { useCostSummary, useRecommendations, useResources } from '../hooks/useCostData';

export const ExecutiveSummary: React.FC = () => {
  const { data: summary } = useCostSummary();
  const { data: recommendations } = useRecommendations();
  const { data: resources } = useResources();

  if (!summary || !recommendations || !resources) {
    return <div>Loading executive summary...</div>;
  }

  const highImpactRecs = recommendations.filter(r => r.savings > 50);
  const criticalResources = resources.filter(r => r.cpuUtilization < 20 && r.monthlyCost > 30);
  const wastedSpend = resources.filter(r => r.status === 'stopped').reduce((sum, r) => sum + r.monthlyCost, 0);

  return (
    <div className="bg-gradient-to-r from-blue-600 to-purple-600 text-white rounded-lg p-6 mb-8">
      <h2 className="text-2xl font-bold mb-4">📊 Executive Summary</h2>
      
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <div className="bg-white/10 backdrop-blur rounded-lg p-4">
          <div className="text-3xl font-bold">${summary.totalMonthlyCost}</div>
          <div className="text-sm opacity-90">Current Monthly Spend</div>
          <div className="text-xs mt-1 opacity-75">Across {resources.length} resources</div>
        </div>

        <div className="bg-white/10 backdrop-blur rounded-lg p-4">
          <div className="text-3xl font-bold text-green-300">${summary.potentialSavings}</div>
          <div className="text-sm opacity-90">Potential Savings</div>
          <div className="text-xs mt-1 opacity-75">
            {Math.round((summary.potentialSavings / summary.totalMonthlyCost) * 100)}% cost reduction
          </div>
        </div>

        <div className="bg-white/10 backdrop-blur rounded-lg p-4">
          <div className="text-3xl font-bold text-yellow-300">{highImpactRecs.length}</div>
          <div className="text-sm opacity-90">High-Impact Actions</div>
          <div className="text-xs mt-1 opacity-75">Savings &gt; $50/month each</div>
        </div>

        <div className="bg-white/10 backdrop-blur rounded-lg p-4">
          <div className="text-3xl font-bold text-red-300">{criticalResources.length}</div>
          <div className="text-sm opacity-90">Critical Resources</div>
          <div className="text-xs mt-1 opacity-75">Low utilization, high cost</div>
        </div>
      </div>

      <div className="mt-6 grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="bg-white/10 backdrop-blur rounded-lg p-4">
          <h3 className="font-semibold mb-3">🎯 Top Recommendations</h3>
          <div className="space-y-2">
            {recommendations.slice(0, 3).map((rec, index) => (
              <div key={index} className="flex justify-between items-center text-sm">
                <span className="opacity-90">{rec.type.toUpperCase()} {rec.resourceId.slice(-8)}</span>
                <span className="font-semibold text-green-300">${rec.savings}/mo</span>
              </div>
            ))}
          </div>
        </div>

        <div className="bg-white/10 backdrop-blur rounded-lg p-4">
          <h3 className="font-semibold mb-3">⚠️ Key Issues</h3>
          <div className="space-y-2 text-sm">
            <div className="flex justify-between">
              <span className="opacity-90">Stopped instances wasting:</span>
              <span className="font-semibold text-red-300">${wastedSpend.toFixed(2)}/mo</span>
            </div>
            <div className="flex justify-between">
              <span className="opacity-90">Under-utilized resources:</span>
              <span className="font-semibold text-yellow-300">{criticalResources.length}</span>
            </div>
            <div className="flex justify-between">
              <span className="opacity-90">Average utilization:</span>
              <span className="font-semibold">{summary.averageUtilization}%</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};