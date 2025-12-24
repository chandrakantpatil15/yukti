import React, { useState, useEffect } from 'react';
import { Calendar, TrendingUp, Download, Bell, Filter, AlertCircle } from 'lucide-react';
import Loading from '../components/Loading';
import api from '../services/api';

interface CostDriver {
  service: string;
  region: string;
  cost: number;
  change_pct: number;
}

interface Anomaly {
  date: string;
  severity: 'critical' | 'warning';
  description: string;
  cause: string;
}

const CostAnalytics: React.FC = () => {
  const [dateRange, setDateRange] = useState({ start: '2024-01-01', end: '2025-01-31' });
  const [groupBy, setGroupBy] = useState('service');
  const [drivers, setDrivers] = useState<CostDriver[]>([]);
  const [anomalies, setAnomalies] = useState<Anomaly[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchData();
  }, [dateRange, groupBy]);

  const fetchData = async () => {
    setLoading(true);
    try {
      const [driversData, anomaliesData] = await Promise.all([
        api.getCostDrivers(dateRange.start, dateRange.end),
        api.getAnomalies(30)
      ]);
      setDrivers(driversData);
      setAnomalies(anomaliesData);
    } catch (error) {
      console.error('Failed to fetch cost analytics:', error);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return <Loading message="Loading cost analytics..." />;
  }

  return (
    <div className="bg-white min-h-screen">
      <div className="max-w-7xl mx-auto">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-4xl font-bold text-gray-900 mb-2">Cost Analytics</h1>
          <p className="text-gray-600 text-lg">Deep dive into your AWS spending patterns</p>
        </div>

        {/* Filters */}
        <div className="bg-white rounded-xl shadow-lg border border-gray-100 p-6 mb-8">
          <div className="flex items-center gap-4 flex-wrap">
            <div className="flex items-center gap-2">
              <Calendar className="w-5 h-5 text-gray-500" />
              <input
                type="date"
                value={dateRange.start}
                onChange={(e) => setDateRange({ ...dateRange, start: e.target.value })}
                className="border border-gray-300 rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              />
              <span className="text-gray-500">to</span>
              <input
                type="date"
                value={dateRange.end}
                onChange={(e) => setDateRange({ ...dateRange, end: e.target.value })}
                className="border border-gray-300 rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              />
            </div>
            <div className="flex items-center gap-2">
              <Filter className="w-5 h-5 text-gray-500" />
              <select
                value={groupBy}
                onChange={(e) => setGroupBy(e.target.value)}
                className="border border-gray-300 rounded-lg px-3 py-2 text-sm focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              >
                <option value="service">Group by Service</option>
                <option value="region">Group by Region</option>
                <option value="resource_type">Group by Resource Type</option>
              </select>
            </div>
            <button
              onClick={fetchData}
              className="ml-auto px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors font-medium"
            >
              Apply Filters
            </button>
          </div>
        </div>

        {/* Top Cost Drivers */}
        <div className="bg-white rounded-xl shadow-lg border border-gray-100 p-6 mb-8">
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-2xl font-bold text-gray-900">Top 10 Cost Drivers</h2>
            <div className="flex gap-2">
              <button className="flex items-center gap-2 px-4 py-2 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 transition-colors text-sm">
                <Download className="w-4 h-4" />
                Export CSV
              </button>
              <button className="flex items-center gap-2 px-4 py-2 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 transition-colors text-sm">
                <Bell className="w-4 h-4" />
                Set Alert
              </button>
            </div>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b-2 border-gray-200">
                  <th className="text-left py-3 px-4 text-sm font-semibold text-gray-700">Service</th>
                  <th className="text-left py-3 px-4 text-sm font-semibold text-gray-700">Region</th>
                  <th className="text-right py-3 px-4 text-sm font-semibold text-gray-700">Cost</th>
                  <th className="text-right py-3 px-4 text-sm font-semibold text-gray-700">Change</th>
                </tr>
              </thead>
              <tbody>
                {drivers.map((driver, idx) => (
                  <tr key={idx} className="border-b border-gray-100 hover:bg-gray-50 transition-colors">
                    <td className="py-3 px-4 font-medium text-gray-900">{driver.service}</td>
                    <td className="py-3 px-4 text-gray-600">{driver.region}</td>
                    <td className="text-right py-3 px-4 font-semibold text-gray-900">
                      ${driver.cost.toLocaleString()}
                    </td>
                    <td className={`text-right py-3 px-4 font-semibold flex items-center justify-end gap-1 ${
                      driver.change_pct > 0 ? 'text-red-600' : driver.change_pct < 0 ? 'text-green-600' : 'text-gray-600'
                    }`}>
                      {driver.change_pct > 0 ? '↑' : driver.change_pct < 0 ? '↓' : '→'} {Math.abs(driver.change_pct)}%
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>

        {/* Anomaly Detection */}
        <div className="bg-white rounded-xl shadow-lg border border-gray-100 p-6">
          <div className="flex items-center gap-3 mb-6">
            <AlertCircle className="w-6 h-6 text-orange-500" />
            <h2 className="text-2xl font-bold text-gray-900">Anomaly Detection (Last 30 Days)</h2>
          </div>
          <div className="space-y-4">
            {anomalies.length === 0 ? (
              <div className="text-center py-8 text-gray-500">
                <AlertCircle className="w-12 h-12 mx-auto mb-3 text-gray-300" />
                <p>No anomalies detected in the last 30 days</p>
              </div>
            ) : (
              anomalies.map((anomaly, idx) => (
                <div
                  key={idx}
                  className={`p-4 rounded-lg border-l-4 ${
                    anomaly.severity === 'critical'
                      ? 'bg-red-50 border-red-500'
                      : 'bg-yellow-50 border-yellow-500'
                  }`}
                >
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <div className="flex items-center gap-2 mb-2">
                        <span className="text-2xl">{anomaly.severity === 'critical' ? '🔴' : '🟡'}</span>
                        <p className="font-semibold text-gray-900">{anomaly.date}: {anomaly.description}</p>
                      </div>
                      <p className="text-sm text-gray-600 ml-8">Cause: {anomaly.cause}</p>
                    </div>
                    <div className="flex gap-2 ml-4">
                      <button className="text-sm text-blue-600 hover:text-blue-700 font-medium">
                        View Details
                      </button>
                      <button className="text-sm text-blue-600 hover:text-blue-700 font-medium">
                        Create Alert
                      </button>
                    </div>
                  </div>
                </div>
              ))
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default CostAnalytics;
