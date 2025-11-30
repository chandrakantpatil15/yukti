import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import api from '../services/api';

const Onboarding: React.FC = () => {
  const [step, setStep] = useState(1);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [awsAccountId, setAwsAccountId] = useState('');
  const [roleArn, setRoleArn] = useState('');
  const navigate = useNavigate();

  // Auto-redirect to dashboard when step 3 is reached
  React.useEffect(() => {
    if (step === 3) {
      const timer = setTimeout(() => {
        navigate('/dashboard');
      }, 3000);
      return () => clearTimeout(timer);
    }
  }, [step, navigate]);

  const handleAWSConnection = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      console.log('Connecting AWS with:', { awsAccountId, roleArn });
      const data = await api.connectAWS(awsAccountId, roleArn, 'yukti-secure-access');
      console.log('AWS connection response:', data);

      if (data.verified) {
        setStep(3);
      } else {
        const errorMsg = data.error || data.message || 'AWS connection failed';
        const errorDetails = data.error_details || '';
        setError(errorDetails ? `${errorMsg}\n\n${errorDetails}` : errorMsg);
      }
    } catch (err: any) {
      console.error('AWS connection error:', err);
      let errorMsg = 'Network error. Please try again.';
      try {
        const errorData = JSON.parse(err.message);
        errorMsg = errorData.error || errorMsg;
        if (errorData.error_details) {
          errorMsg += '\n\n' + errorData.error_details;
        }
      } catch {
        errorMsg = err.message || errorMsg;
      }
      setError(errorMsg);
    } finally {
      setLoading(false);
    }
  };

  const completeOnboarding = () => {
    navigate('/dashboard');
  };

  if (step === 1) {
    const yuktiAccountId = '144403604430';
    const externalIdExample = 'yukti-18-abc123xyz456';
    
    const trustPolicy = `{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {
      "AWS": "arn:aws:iam::${yuktiAccountId}:user/yukti-platform-user"
    },
    "Action": "sts:AssumeRole",
    "Condition": {
      "StringLike": {
        "sts:ExternalId": "yukti-*"
      }
    }
  }]
}`;

    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50 py-8">
        <div className="max-w-4xl w-full space-y-6 px-4">
          <div className="text-center">
            <h2 className="text-3xl font-extrabold text-gray-900">Welcome to Yukti!</h2>
            <p className="mt-2 text-gray-600">Let's connect your AWS account to start optimizing costs</p>
          </div>
          
          <div className="bg-white p-6 rounded-lg shadow">
            <h3 className="text-xl font-semibold mb-4 text-gray-900">Step 1: Create IAM Role in Your AWS Account</h3>
            
            <div className="space-y-4">
              <div className="bg-blue-50 border border-blue-200 p-4 rounded-md">
                <h4 className="font-semibold text-blue-900 mb-2">📋 Quick Setup Instructions</h4>
                <ol className="list-decimal list-inside text-sm text-blue-800 space-y-2">
                  <li>Go to <a href="https://console.aws.amazon.com/iam/home#/roles" target="_blank" rel="noopener noreferrer" className="underline">AWS IAM Console → Roles</a></li>
                  <li>Click "Create role" → Select "AWS account"</li>
                  <li>Choose "Another AWS account"</li>
                  <li>Enter Yukti Account ID: <code className="bg-blue-100 px-2 py-1 rounded font-mono text-xs">{yuktiAccountId}</code></li>
                  <li>Check "Require external ID" → Enter any value starting with: <code className="bg-blue-100 px-2 py-1 rounded font-mono text-xs">yukti-</code></li>
                  <li>Click "Next" → Attach policy: <code className="bg-blue-100 px-2 py-1 rounded font-mono">ReadOnlyAccess</code></li>
                  <li>Click "Next" → Name: <code className="bg-blue-100 px-2 py-1 rounded font-mono">YuktiReadOnlyRole</code></li>
                  <li>Click "Create role"</li>
                </ol>
              </div>

              <div className="bg-gray-50 border border-gray-200 p-4 rounded-md">
                <div className="flex justify-between items-center mb-2">
                  <h4 className="font-semibold text-gray-900">🔐 Trust Policy (Copy & Paste)</h4>
                  <button
                    onClick={() => navigator.clipboard.writeText(trustPolicy)}
                    className="text-xs bg-gray-200 hover:bg-gray-300 px-3 py-1 rounded"
                  >
                    Copy
                  </button>
                </div>
                <pre className="bg-gray-900 text-green-400 p-3 rounded text-xs overflow-x-auto">
                  {trustPolicy}
                </pre>
                <p className="text-xs text-gray-600 mt-2">
                  💡 Use this if you prefer to create the role via AWS CLI or manually edit the trust policy
                </p>
              </div>

              <div className="bg-yellow-50 border border-yellow-200 p-4 rounded-md">
                <h4 className="font-semibold text-yellow-900 mb-2">⚠️ Important Details</h4>
                <ul className="text-sm text-yellow-800 space-y-1">
                  <li>• <strong>Yukti Account ID:</strong> <code className="bg-yellow-100 px-1 rounded">{yuktiAccountId}</code></li>
                  <li>• <strong>External ID Pattern:</strong> <code className="bg-yellow-100 px-1 rounded">yukti-*</code> (auto-generated by platform)</li>
                  <li>• <strong>Required Permission:</strong> ReadOnlyAccess (AWS managed policy)</li>
                  <li>• <strong>Security:</strong> We can only READ your data, never modify anything</li>
                </ul>
              </div>
            </div>
            
            <button
              onClick={() => setStep(2)}
              className="w-full mt-6 bg-blue-600 text-white py-3 px-4 rounded-md hover:bg-blue-700 font-medium"
            >
              ✅ I've Created the IAM Role → Continue
            </button>
          </div>
        </div>
      </div>
    );
  }

  if (step === 2) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="max-w-2xl w-full space-y-8 p-8">
          <div className="text-center">
            <h2 className="text-3xl font-extrabold text-gray-900">Connect AWS Account</h2>
            <p className="mt-2 text-gray-600">Enter your AWS account details</p>
          </div>
          
          <form onSubmit={handleAWSConnection} className="bg-white p-6 rounded-lg shadow space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                AWS Account ID
              </label>
              <input
                type="text"
                required
                value={awsAccountId}
                onChange={(e) => setAwsAccountId(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                placeholder="123456789012"
              />
            </div>
            
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                IAM Role ARN
              </label>
              <input
                type="text"
                required
                value={roleArn}
                onChange={(e) => setRoleArn(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                placeholder="arn:aws:iam::123456789012:role/YuktiReadOnlyRole"
              />
            </div>
            


            {error && (
              <div className="bg-red-50 border border-red-200 rounded-md p-4">
                <div className="flex">
                  <div className="flex-shrink-0">
                    <svg className="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
                      <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
                    </svg>
                  </div>
                  <div className="ml-3">
                    <h3 className="text-sm font-medium text-red-800">Connection Failed</h3>
                    <div className="mt-2 text-sm text-red-700 whitespace-pre-line">{error}</div>
                  </div>
                </div>
              </div>
            )}

            <button
              type="submit"
              disabled={loading}
              className="w-full bg-blue-600 text-white py-2 px-4 rounded-md hover:bg-blue-700 disabled:bg-gray-400"
            >
              {loading ? 'Connecting...' : 'Connect AWS Account'}
            </button>
          </form>
        </div>
      </div>
    );
  }

  if (step === 3) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="max-w-2xl w-full space-y-8 p-8">
          <div className="text-center">
            <div className="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-green-100 mb-4">
              <svg className="h-6 w-6 text-green-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
              </svg>
            </div>
            <h2 className="text-3xl font-extrabold text-gray-900">Setup Complete!</h2>
            <p className="mt-2 text-gray-600">
              Your AWS account is connected. We're analyzing your resources and costs.
            </p>
          </div>
          
          <div className="bg-white p-6 rounded-lg shadow text-center">
            <p className="text-gray-600 mb-4">
              Redirecting to dashboard in 3 seconds...
            </p>
            
            <button
              onClick={completeOnboarding}
              className="bg-blue-600 text-white py-2 px-6 rounded-md hover:bg-blue-700"
            >
              Go to Dashboard Now
            </button>
          </div>
        </div>
      </div>
    );
  }

  return null;
};

export default Onboarding;