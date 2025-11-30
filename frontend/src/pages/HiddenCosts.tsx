import React, { useState, useEffect } from 'react';
import { AlertCircle, DollarSign, Filter, X } from 'lucide-react';
import api from '../services/api';
import { getCurrentUser } from '../lib/auth';

interface Finding {
  id: string;
  detector_name: string;
  category: string;
  severity: string;
  title: string;
  description: string;
  resource_arn: string;
  estimated_savings: number;
  confidence: number;
  created_at: string;
}

const HiddenCosts: React.FC = () => {
  const [findings, setFindings] = useState<Finding[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedCategory, setSelectedCategory] = useState('');
  const [selectedSeverity, setSelectedSeverity] = useState('');
  const [selectedFinding, setSelectedFinding] = useState<Finding | null>(null);
  const [allCategories, setAllCategories] = useState<string[]>([]);
  const [error, setError] = useState('');

  useEffect(() => {
    fetchFindings();
  }, []);

  useEffect(() => {
    if (selectedCategory || selectedSeverity) {
      fetchFindings();
    }
  }, [selectedCategory, selectedSeverity]);

  const fetchFindings = async () => {
    setLoading(true);
    try {
      const user = getCurrentUser();
      if (!user) {
        setError('User not authenticated');
        setLoading(false);
        return;
      }

      // Backend extracts tenant_id from JWT - only send filters
      const filters: any = {};
      if (selectedCategory) {
        filters.category = selectedCategory;
      }
      if (selectedSeverity) {
        filters.severity = selectedSeverity;
      }
      
      const data = await api.getFindings(filters);
      if (data.success) {
        const findingsData = data.findings || [];
        setFindings(findingsData);
        
        // Get all categories for filter dropdown (only on first load)
        if (allCategories.length === 0) {
          const cats = [...new Set(findingsData.map((f: Finding) => f.category))];
          setAllCategories(cats);
        }
      }
    } catch (err) {
      console.error('Error fetching findings:', err);
      setError('Failed to load findings');
      setFindings([]);
    } finally {
      setLoading(false);
    }
  };

  const severities = ['Critical', 'High', 'Medium', 'Low'];
  const totalSavings = findings.reduce((sum, f) => sum + f.estimated_savings, 0);

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'Critical': return 'bg-red-100 text-red-800';
      case 'High': return 'bg-orange-100 text-orange-800';
      case 'Medium': return 'bg-yellow-100 text-yellow-800';
      case 'Low': return 'bg-blue-100 text-blue-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-100 flex items-center justify-center">
        <div className="text-xl text-gray-600">Loading findings...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen bg-gray-100 p-6 flex items-center justify-center">
        <div className="bg-red-50 border border-red-200 rounded-lg p-6 max-w-md">
          <p className="text-red-800">{error}</p>
          <button
            onClick={() => window.location.href = '/login'}
            className="mt-4 px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700"
          >
            Go to Login
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-100 p-6">
      <div className="max-w-7xl mx-auto">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900">Hidden Cost Findings</h1>
          <p className="text-gray-600">
            {findings.length} opportunities • ${totalSavings.toLocaleString()}/month potential savings
          </p>
        </div>

        {/* Filters */}
        <div className="bg-white rounded-lg shadow p-6 mb-6">
          <div className="flex items-center gap-4">
            <Filter className="w-5 h-5 text-gray-600" />
            <select
              value={selectedCategory}
              onChange={(e) => setSelectedCategory(e.target.value)}
              className="px-4 py-2 border rounded-lg"
            >
              <option value="">All Categories</option>
              {allCategories.map(cat => (
                <option key={cat} value={cat}>{cat}</option>
              ))}
            </select>

            <select
              value={selectedSeverity}
              onChange={(e) => setSelectedSeverity(e.target.value)}
              className="px-4 py-2 border rounded-lg"
            >
              <option value="">All Severities</option>
              {severities.map(sev => (
                <option key={sev} value={sev}>{sev}</option>
              ))}
            </select>

            {(selectedCategory || selectedSeverity) && (
              <button
                onClick={() => {
                  setSelectedCategory('');
                  setSelectedSeverity('');
                }}
                className="text-blue-600 hover:text-blue-800 flex items-center gap-1"
              >
                <X className="w-4 h-4" />
                Clear Filters
              </button>
            )}
          </div>
        </div>

        {/* Findings List */}
        <div className="space-y-4">
          {findings.length === 0 ? (
            <div className="bg-white rounded-lg shadow p-12 text-center">
              <AlertCircle className="w-16 h-16 text-gray-400 mx-auto mb-4" />
              <h3 className="text-xl font-semibold text-gray-900 mb-2">No findings found</h3>
              <p className="text-gray-600">Try adjusting your filters or check back later</p>
            </div>
          ) : (
            findings.map((finding) => (
              <div
                key={finding.id}
                className="bg-white rounded-lg shadow p-6 hover:shadow-lg transition-shadow cursor-pointer"
                onClick={() => setSelectedFinding(finding)}
              >
                <div className="flex justify-between items-start">
                  <div className="flex-1">
                    <div className="flex items-center gap-3 mb-2">
                      <span className={`px-3 py-1 rounded-full text-xs font-semibold ${getSeverityColor(finding.severity)}`}>
                        {finding.severity}
                      </span>
                      <span className="text-sm text-gray-600">{finding.category}</span>
                    </div>
                    <h3 className="text-lg font-semibold text-gray-900 mb-2">{finding.title}</h3>
                    <p className="text-gray-600 mb-3">{finding.description}</p>
                    <div className="flex items-center gap-4 text-sm text-gray-500">
                      <span>Resource: {finding.resource_arn}</span>
                      <span>Confidence: {(finding.confidence * 100).toFixed(0)}%</span>
                    </div>
                  </div>
                  <div className="text-right ml-6">
                    <div className="flex items-center gap-2 text-green-600">
                      <DollarSign className="w-5 h-5" />
                      <span className="text-2xl font-bold">${finding.estimated_savings.toLocaleString()}</span>
                    </div>
                    <p className="text-sm text-gray-500">per month</p>
                  </div>
                </div>
              </div>
            ))
          )}
        </div>
      </div>

      {/* Detail Panel */}
      {selectedFinding && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-6 z-50">
          <div className="bg-white rounded-lg shadow-xl max-w-2xl w-full max-h-[90vh] overflow-y-auto">
            <div className="p-6 border-b flex justify-between items-start">
              <div>
                <h2 className="text-2xl font-bold text-gray-900 mb-2">{selectedFinding.title}</h2>
                <div className="flex items-center gap-3">
                  <span className={`px-3 py-1 rounded-full text-xs font-semibold ${getSeverityColor(selectedFinding.severity)}`}>
                    {selectedFinding.severity}
                  </span>
                  <span className="text-sm text-gray-600">{selectedFinding.category}</span>
                </div>
              </div>
              <button
                onClick={() => setSelectedFinding(null)}
                className="text-gray-400 hover:text-gray-600"
              >
                <X className="w-6 h-6" />
              </button>
            </div>

            <div className="p-6 space-y-6">
              <div>
                <h3 className="font-semibold text-gray-900 mb-2">Description</h3>
                <p className="text-gray-600">{selectedFinding.description}</p>
              </div>

              <div>
                <h3 className="font-semibold text-gray-900 mb-2">Estimated Savings</h3>
                <div className="flex items-center gap-2 text-green-600">
                  <DollarSign className="w-6 h-6" />
                  <span className="text-3xl font-bold">${selectedFinding.estimated_savings.toLocaleString()}</span>
                  <span className="text-gray-600">per month</span>
                </div>
              </div>

              <div>
                <h3 className="font-semibold text-gray-900 mb-2">Resource</h3>
                <code className="block bg-gray-100 p-3 rounded text-sm">{selectedFinding.resource_arn}</code>
              </div>

              <div>
                <h3 className="font-semibold text-gray-900 mb-2">Confidence</h3>
                <div className="flex items-center gap-3">
                  <div className="flex-1 bg-gray-200 rounded-full h-3">
                    <div
                      className="bg-blue-600 h-3 rounded-full"
                      style={{ width: `${selectedFinding.confidence * 100}%` }}
                    />
                  </div>
                  <span className="font-semibold">{(selectedFinding.confidence * 100).toFixed(0)}%</span>
                </div>
              </div>

              <div className="flex gap-3">
                <button className="flex-1 bg-blue-600 text-white px-6 py-3 rounded-lg hover:bg-blue-700">
                  Generate IaC
                </button>
                <button className="flex-1 border border-gray-300 px-6 py-3 rounded-lg hover:bg-gray-50">
                  Whitelist
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default HiddenCosts;
