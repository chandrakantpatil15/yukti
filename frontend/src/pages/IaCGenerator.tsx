import React, { useState } from 'react';
import api from '../services/api';

const IaCGenerator: React.FC = () => {
  const [format, setFormat] = useState<'terraform' | 'cloudformation'>('terraform');
  const [generatedCode, setGeneratedCode] = useState<string>('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [findingId, setFindingId] = useState('');

  const handleGenerate = async () => {
    if (!findingId) {
      setError('Please enter a finding ID');
      return;
    }

    setLoading(true);
    setError('');
    
    try {
      const response = await api.post('/api/iac/generate', {
        finding_id: findingId,
        format: format,
      });
      
      if (response.success) {
        setGeneratedCode(response.code || '');
      } else {
        setError(response.error || 'Failed to generate IaC code');
      }
    } catch (err: any) {
      console.error('IaC generation error:', err);
      setError(err.response?.data?.error || 'Failed to generate IaC code');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="p-6">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900">IaC Generator</h1>
        <p className="text-gray-600">Generate Infrastructure as Code for optimization</p>
      </div>

      <div className="grid grid-cols-2 gap-6">
        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="text-lg font-semibold mb-4">Configuration</h2>
          
          <div className="mb-4">
            <label className="block text-sm font-medium text-gray-700 mb-2">Finding ID</label>
            <input
              type="text"
              value={findingId}
              onChange={(e) => setFindingId(e.target.value)}
              placeholder="Enter finding ID from Hidden Costs page"
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>

          <div className="mb-4">
            <label className="block text-sm font-medium text-gray-700 mb-2">Format</label>
            <div className="flex gap-4">
              <button
                onClick={() => setFormat('terraform')}
                className={`px-4 py-2 rounded ${format === 'terraform' ? 'bg-blue-600 text-white' : 'bg-gray-200'}`}
              >
                Terraform
              </button>
              <button
                onClick={() => setFormat('cloudformation')}
                className={`px-4 py-2 rounded ${format === 'cloudformation' ? 'bg-blue-600 text-white' : 'bg-gray-200'}`}
              >
                CloudFormation
              </button>
            </div>
          </div>

          {error && (
            <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-md text-red-700 text-sm">
              {error}
            </div>
          )}

          <button
            onClick={handleGenerate}
            disabled={loading || !findingId}
            className="w-full bg-blue-600 text-white py-2 rounded hover:bg-blue-700 disabled:bg-gray-400"
          >
            {loading ? 'Generating...' : 'Generate Code'}
          </button>
        </div>

        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="text-lg font-semibold mb-4">Generated Code</h2>
          
          {generatedCode ? (
            <div>
              <div className="flex justify-between items-center mb-2">
                <span className="text-sm text-gray-600">Generated {format} code</span>
                <button
                  onClick={() => navigator.clipboard.writeText(generatedCode)}
                  className="text-sm text-blue-600 hover:text-blue-700"
                >
                  Copy to Clipboard
                </button>
              </div>
              <pre className="bg-gray-900 text-gray-100 p-4 rounded overflow-auto max-h-[400px] text-sm font-mono">
                {generatedCode}
              </pre>
            </div>
          ) : (
            <div className="text-center text-gray-500 py-20">
              <p className="mb-2">Enter a finding ID and click Generate</p>
              <p className="text-sm">You can find finding IDs on the Hidden Costs page</p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default IaCGenerator;