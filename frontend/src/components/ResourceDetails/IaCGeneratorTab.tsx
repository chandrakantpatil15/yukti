import React, { useState } from 'react';
import { Code, Download, Copy, Check } from 'lucide-react';
import api from '../../services/api';

interface IaCGeneratorTabProps {
  resourceId: string;
  resourceType: string;
}

interface IaCResponse {
  code: string;
  rollback_code: string;
  format: string;
  instructions: string[];
}

const IaCGeneratorTab: React.FC<IaCGeneratorTabProps> = ({ resourceId, resourceType }) => {
  const [format, setFormat] = useState<'terraform' | 'cloudformation'>('terraform');
  const [action, setAction] = useState('optimize');
  const [loading, setLoading] = useState(false);
  const [generatedCode, setGeneratedCode] = useState<IaCResponse | null>(null);
  const [copied, setCopied] = useState(false);
  const [error, setError] = useState('');

  const generateCode = async () => {
    setLoading(true);
    setError('');
    try {
      const response = await api.post('/iac/generate', {
        finding_id: resourceId,
        format: format,
        action: action
      });
      setGeneratedCode(response);
    } catch (err) {
      setError('Failed to generate IaC code. Please try again.');
      console.error('IaC generation error:', err);
    } finally {
      setLoading(false);
    }
  };

  const copyToClipboard = async () => {
    if (generatedCode?.code) {
      try {
        await navigator.clipboard.writeText(generatedCode.code);
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
      } catch (err) {
        console.error('Failed to copy:', err);
      }
    }
  };

  const downloadCode = () => {
    if (generatedCode?.code) {
      const extension = format === 'terraform' ? 'tf' : 'yaml';
      const filename = `${resourceId}_optimized.${extension}`;
      const blob = new Blob([generatedCode.code], { type: 'text/plain' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = filename;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
    }
  };

  return (
    <div className="p-4 space-y-6">
      <div>
        <h3 className="text-lg font-semibold text-slate-900 mb-2">Infrastructure as Code Generator</h3>
        <p className="text-slate-600">Generate optimized infrastructure code for this resource.</p>
      </div>

      {/* Configuration */}
      <div className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-slate-700 mb-2">Format</label>
          <div className="flex gap-3">
            <button
              onClick={() => setFormat('terraform')}
              className={`px-4 py-2 rounded-lg border transition-colors ${
                format === 'terraform'
                  ? 'bg-blue-600 text-white border-blue-600'
                  : 'bg-white text-slate-700 border-slate-300 hover:bg-slate-50'
              }`}
            >
              Terraform
            </button>
            <button
              onClick={() => setFormat('cloudformation')}
              className={`px-4 py-2 rounded-lg border transition-colors ${
                format === 'cloudformation'
                  ? 'bg-blue-600 text-white border-blue-600'
                  : 'bg-white text-slate-700 border-slate-300 hover:bg-slate-50'
              }`}
            >
              CloudFormation
            </button>
          </div>
        </div>

        <div>
          <label className="block text-sm font-medium text-slate-700 mb-2">Action</label>
          <select
            value={action}
            onChange={(e) => setAction(e.target.value)}
            className="w-full px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
          >
            <option value="optimize">Optimize Configuration</option>
            <option value="downsize">Downsize Instance</option>
            <option value="rightsize">Right-size Instance</option>
            <option value="schedule">Add Scheduling</option>
          </select>
        </div>

        <button
          onClick={generateCode}
          disabled={loading}
          className="flex items-center gap-2 px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        >
          <Code className="w-4 h-4" />
          {loading ? 'Generating...' : 'Generate Code'}
        </button>
      </div>

      {/* Error */}
      {error && (
        <div className="p-4 bg-red-50 border border-red-200 rounded-lg text-red-700">
          {error}
        </div>
      )}

      {/* Generated Code */}
      {generatedCode && (
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h4 className="text-lg font-semibold text-slate-900">Generated {generatedCode.format} Code</h4>
            <div className="flex items-center gap-2">
              <button
                onClick={copyToClipboard}
                className="flex items-center gap-2 px-3 py-2 text-slate-600 hover:text-slate-800 transition-colors"
              >
                {copied ? <Check className="w-4 h-4 text-green-600" /> : <Copy className="w-4 h-4" />}
                {copied ? 'Copied!' : 'Copy'}
              </button>
              <button
                onClick={downloadCode}
                className="flex items-center gap-2 px-3 py-2 text-slate-600 hover:text-slate-800 transition-colors"
              >
                <Download className="w-4 h-4" />
                Download
              </button>
            </div>
          </div>

          <div className="bg-slate-900 rounded-lg p-4 overflow-x-auto">
            <pre className="text-green-400 text-sm whitespace-pre-wrap">
              <code>{generatedCode.code}</code>
            </pre>
          </div>

          {/* Instructions */}
          {generatedCode.instructions && generatedCode.instructions.length > 0 && (
            <div>
              <h5 className="font-semibold text-slate-900 mb-2">Deployment Instructions</h5>
              <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
                <ol className="list-decimal list-inside space-y-1 text-blue-800">
                  {generatedCode.instructions.map((instruction, index) => (
                    <li key={index} className="text-sm">{instruction}</li>
                  ))}
                </ol>
              </div>
            </div>
          )}

          {/* Rollback Code */}
          {generatedCode.rollback_code && (
            <div>
              <h5 className="font-semibold text-slate-900 mb-2">Rollback Code</h5>
              <div className="bg-slate-900 rounded-lg p-4 overflow-x-auto">
                <pre className="text-red-400 text-sm whitespace-pre-wrap">
                  <code>{generatedCode.rollback_code}</code>
                </pre>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
};

export default IaCGeneratorTab;