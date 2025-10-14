import React from 'react';
import { MetricCard } from './MetricCard';
import { ResourceInventory } from './ResourceInventory';
import { ExecutiveSummary } from './ExecutiveSummary';
import { useCostSummary, useRecommendations } from '../hooks/useCostData';

export const Dashboard: React.FC = () => {
  const { data: summary, isLoading: summaryLoading } = useCostSummary();
  const { data: recommendations, isLoading: recsLoading } = useRecommendations();

  if (summaryLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
      </div>
    );
  }

  if (!summary) return <div>No data available</div>;

  const savingsPercent = Math.round((summary.potentialSavings / summary.totalMonthlyCost) * 100);

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white shadow-sm border-b">
        <div className="max-w-7xl mx-auto px-4 py-6">
          <h1 className="text-3xl font-bold text-gray-900">AWS Cost Optimization</h1>
          <p className="text-gray-600 mt-2">Monitor spending and identify savings opportunities</p>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-4 py-8">
        {/* Executive Summary */}
        <ExecutiveSummary />
        
        {/* Metrics Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
          <MetricCard
            title="Monthly Cost"
            value={`$${summary.totalMonthlyCost}`}
            change="Current spending"
            changeType="negative"
            icon="💰"
          />
          <MetricCard
            title="Potential Savings"
            value={`$${summary.potentialSavings}`}
            change={`${savingsPercent}% reduction possible`}
            changeType="positive"
            icon="💡"
          />
          <MetricCard
            title="Total Resources"
            value={summary.totalResources}
            change={`${summary.optimizationOpportunities} need optimization`}
            changeType="neutral"
            icon="🖥️"
          />
          <MetricCard
            title="Average Utilization"
            value={`${summary.averageUtilization}%`}
            change={summary.averageUtilization < 50 ? 'Under-utilized' : 'Well utilized'}
            changeType={summary.averageUtilization < 50 ? 'negative' : 'positive'}
            icon="📈"
          />
        </div>

        {/* Resource Inventory */}
        <ResourceInventory />
        
        {/* Recommendations */}
        <div className="bg-white rounded-lg shadow-sm p-6 mt-8">
          <h2 className="text-xl font-semibold text-gray-900 mb-6">Optimization Recommendations</h2>
          {recsLoading ? (
            <div className="text-center py-8">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500 mx-auto"></div>
              <p className="text-gray-600 mt-2">Loading recommendations...</p>
            </div>
          ) : (
            <div className="space-y-4">
              {recommendations?.map((rec, index) => (
                <div key={index} className="border border-gray-200 rounded-lg p-4 bg-gray-50">
                  <div className="flex justify-between items-start mb-2">
                    <span className="bg-blue-500 text-white px-3 py-1 rounded-full text-sm font-medium">
                      {rec.type.toUpperCase()}
                    </span>
                    <span className="text-lg font-semibold text-green-600">
                      ${rec.savings}/mo
                    </span>
                  </div>
                  <div className="text-gray-700 text-sm">
                    <strong>{rec.resourceId}</strong><br />
                    Change from {rec.from} to {rec.to}<br />
                    Current: ${rec.currentCost}/mo → Optimized: ${rec.optimizedCost}/mo
                  </div>
                  <div className="mt-3">
                    <div className="flex justify-between text-xs text-gray-600 mb-1">
                      <span>Confidence: {rec.confidence}%</span>
                    </div>
                    <div className="w-full bg-gray-200 rounded-full h-2">
                      <div 
                        className="bg-green-500 h-2 rounded-full transition-all duration-300"
                        style={{ width: `${rec.confidence}%` }}
                      ></div>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </main>
    </div>
  );
};