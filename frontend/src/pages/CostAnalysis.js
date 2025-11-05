import React, { useState, useEffect } from 'react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, BarChart, Bar } from 'recharts';
import { apiService } from '../services/api';

const CostAnalysis = () => {
  const [costData, setCostData] = useState([]);
  const [recommendations, setRecommendations] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchCostAnalysis();
  }, []);

  const fetchCostAnalysis = async () => {
    try {
      const resources = await apiService.getResources();
      
      // Generate cost analysis data
      const analysis = generateCostAnalysis(resources);
      setCostData(analysis.costData);
      setRecommendations(analysis.recommendations);
      
      setLoading(false);
    } catch (error) {
      console.error('Failed to fetch cost analysis:', error);
      setLoading(false);
    }
  };

  const generateCostAnalysis = (resources) => {
    // Generate mock historical data for demonstration
    const costData = [
      { month: 'Jan', actual: 45.20, projected: 42.00 },
      { month: 'Feb', actual: 52.10, projected: 48.50 },
      { month: 'Mar', actual: 38.90, projected: 35.20 },
      { month: 'Apr', actual: 61.30, projected: 58.10 },
      { month: 'May', actual: 47.80, projected: 44.60 },
      { month: 'Jun', actual: 55.40, projected: 52.20 }
    ];

    // Generate recommendations based on current resources
    const recommendations = [
      {
        id: 1,
        type: 'rightsizing',
        title: 'Rightsize Underutilized Instances',
        description: 'Switch t3.medium to t3.small for development workloads',
        potentialSavings: 15.60,
        impact: 'Medium',
        effort: 'Low'
      },
      {
        id: 2,
        type: 'scheduling',
        title: 'Implement Auto-Scheduling',
        description: 'Stop non-production instances during off-hours',
        potentialSavings: 28.40,
        impact: 'High',
        effort: 'Medium'
      },
      {
        id: 3,
        type: 'spot',
        title: 'Use Spot Instances',
        description: 'Migrate batch workloads to spot instances',
        potentialSavings: 22.10,
        impact: 'High',
        effort: 'High'
      }
    ];

    return { costData, recommendations };
  };

  const totalPotentialSavings = recommendations.reduce((sum, rec) => sum + rec.potentialSavings, 0);

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-lg">Loading cost analysis...</div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold text-gray-900">Cost Analysis</h1>
        <button 
          onClick={fetchCostAnalysis}
          className="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors"
        >
          🔄 Refresh Analysis
        </button>
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div className="bg-white p-6 rounded-lg shadow-md">
          <div className="flex items-center">
            <div className="text-3xl mr-4">💰</div>
            <div>
              <p className="text-sm font-medium text-gray-600">Current Monthly Cost</p>
              <p className="text-2xl font-bold text-blue-600">$55.40</p>
            </div>
          </div>
        </div>

        <div className="bg-white p-6 rounded-lg shadow-md">
          <div className="flex items-center">
            <div className="text-3xl mr-4">💡</div>
            <div>
              <p className="text-sm font-medium text-gray-600">Potential Savings</p>
              <p className="text-2xl font-bold text-green-600">${totalPotentialSavings.toFixed(2)}</p>
            </div>
          </div>
        </div>

        <div className="bg-white p-6 rounded-lg shadow-md">
          <div className="flex items-center">
            <div className="text-3xl mr-4">📈</div>
            <div>
              <p className="text-sm font-medium text-gray-600">Optimization Score</p>
              <p className="text-2xl font-bold text-orange-600">72%</p>
            </div>
          </div>
        </div>
      </div>

      {/* Cost Trend Chart */}
      <div className="bg-white p-6 rounded-lg shadow-md">
        <h3 className="text-lg font-semibold mb-4">Cost Trend Analysis</h3>
        <ResponsiveContainer width="100%" height={300}>
          <LineChart data={costData}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="month" />
            <YAxis />
            <Tooltip formatter={(value) => [`$${value}`, '']} />
            <Line type="monotone" dataKey="actual" stroke="#3B82F6" strokeWidth={2} name="Actual Cost" />
            <Line type="monotone" dataKey="projected" stroke="#10B981" strokeWidth={2} strokeDasharray="5 5" name="Projected Cost" />
          </LineChart>
        </ResponsiveContainer>
      </div>

      {/* Recommendations */}
      <div className="bg-white p-6 rounded-lg shadow-md">
        <h3 className="text-lg font-semibold mb-4">Cost Optimization Recommendations</h3>
        <div className="space-y-4">
          {recommendations.map((rec) => (
            <div key={rec.id} className="border border-gray-200 rounded-lg p-4 hover:bg-gray-50">
              <div className="flex items-start justify-between">
                <div className="flex-1">
                  <div className="flex items-center space-x-2 mb-2">
                    <span className={`px-2 py-1 text-xs font-semibold rounded-full ${
                      rec.type === 'rightsizing' ? 'bg-blue-100 text-blue-800' :
                      rec.type === 'scheduling' ? 'bg-green-100 text-green-800' :
                      'bg-purple-100 text-purple-800'
                    }`}>
                      {rec.type}
                    </span>
                    <span className={`px-2 py-1 text-xs font-semibold rounded-full ${
                      rec.impact === 'High' ? 'bg-red-100 text-red-800' :
                      rec.impact === 'Medium' ? 'bg-yellow-100 text-yellow-800' :
                      'bg-gray-100 text-gray-800'
                    }`}>
                      {rec.impact} Impact
                    </span>
                    <span className={`px-2 py-1 text-xs font-semibold rounded-full ${
                      rec.effort === 'Low' ? 'bg-green-100 text-green-800' :
                      rec.effort === 'Medium' ? 'bg-yellow-100 text-yellow-800' :
                      'bg-red-100 text-red-800'
                    }`}>
                      {rec.effort} Effort
                    </span>
                  </div>
                  <h4 className="font-semibold text-gray-900 mb-1">{rec.title}</h4>
                  <p className="text-gray-600 text-sm mb-2">{rec.description}</p>
                </div>
                <div className="text-right ml-4">
                  <div className="text-lg font-bold text-green-600">${rec.potentialSavings.toFixed(2)}</div>
                  <div className="text-xs text-gray-500">monthly savings</div>
                </div>
              </div>
              <div className="mt-3 flex space-x-2">
                <button className="bg-blue-600 text-white px-3 py-1 rounded text-sm hover:bg-blue-700 transition-colors">
                  Apply
                </button>
                <button className="bg-gray-200 text-gray-700 px-3 py-1 rounded text-sm hover:bg-gray-300 transition-colors">
                  Learn More
                </button>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Cost Breakdown */}
      <div className="bg-white p-6 rounded-lg shadow-md">
        <h3 className="text-lg font-semibold mb-4">Cost Breakdown by Service</h3>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div>
            <h4 className="font-medium text-gray-700 mb-3">Current Month</h4>
            <div className="space-y-2">
              <div className="flex justify-between items-center">
                <span className="text-sm text-gray-600">EC2 Instances</span>
                <span className="font-semibold">$42.30</span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-sm text-gray-600">EBS Storage</span>
                <span className="font-semibold">$8.90</span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-sm text-gray-600">Data Transfer</span>
                <span className="font-semibold">$4.20</span>
              </div>
              <div className="border-t pt-2 mt-2">
                <div className="flex justify-between items-center font-bold">
                  <span>Total</span>
                  <span>$55.40</span>
                </div>
              </div>
            </div>
          </div>
          <div>
            <h4 className="font-medium text-gray-700 mb-3">After Optimization</h4>
            <div className="space-y-2">
              <div className="flex justify-between items-center">
                <span className="text-sm text-gray-600">EC2 Instances</span>
                <span className="font-semibold text-green-600">$28.20</span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-sm text-gray-600">EBS Storage</span>
                <span className="font-semibold">$8.90</span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-sm text-gray-600">Data Transfer</span>
                <span className="font-semibold">$4.20</span>
              </div>
              <div className="border-t pt-2 mt-2">
                <div className="flex justify-between items-center font-bold">
                  <span>Total</span>
                  <span className="text-green-600">$41.30</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default CostAnalysis;