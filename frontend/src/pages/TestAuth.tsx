import React, { useEffect, useState } from 'react';
import { getToken, getCurrentUser, isAuthenticated } from '../lib/auth';

const TestAuth: React.FC = () => {
  const [tokenInfo, setTokenInfo] = useState<any>(null);

  useEffect(() => {
    const token = localStorage.getItem('yukti_auth_token');
    const user = localStorage.getItem('yukti_user');
    const authToken = getToken();
    const currentUser = getCurrentUser();
    const authenticated = isAuthenticated();

    setTokenInfo({
      rawToken: token ? token.substring(0, 50) + '...' : 'NULL',
      rawUser: user,
      getToken: authToken ? 'PRESENT' : 'NULL',
      getCurrentUser: currentUser,
      isAuthenticated: authenticated,
    });
  }, []);

  return (
    <div className="min-h-screen bg-gray-100 p-6">
      <div className="max-w-4xl mx-auto bg-white rounded-lg shadow p-6">
        <h1 className="text-2xl font-bold mb-4">Auth Debug Page</h1>
        
        <div className="space-y-4">
          <div>
            <h2 className="font-semibold">Raw Token (localStorage):</h2>
            <pre className="bg-gray-100 p-2 rounded text-xs overflow-auto">
              {tokenInfo?.rawToken || 'Loading...'}
            </pre>
          </div>

          <div>
            <h2 className="font-semibold">Raw User (localStorage):</h2>
            <pre className="bg-gray-100 p-2 rounded text-xs overflow-auto">
              {tokenInfo?.rawUser || 'NULL'}
            </pre>
          </div>

          <div>
            <h2 className="font-semibold">getToken() result:</h2>
            <pre className="bg-gray-100 p-2 rounded text-xs">
              {tokenInfo?.getToken || 'Loading...'}
            </pre>
          </div>

          <div>
            <h2 className="font-semibold">getCurrentUser() result:</h2>
            <pre className="bg-gray-100 p-2 rounded text-xs overflow-auto">
              {JSON.stringify(tokenInfo?.getCurrentUser, null, 2)}
            </pre>
          </div>

          <div>
            <h2 className="font-semibold">isAuthenticated() result:</h2>
            <pre className="bg-gray-100 p-2 rounded text-xs">
              {String(tokenInfo?.isAuthenticated)}
            </pre>
          </div>

          <button
            onClick={() => window.location.href = '/login'}
            className="mt-4 px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
          >
            Go to Login
          </button>

          <button
            onClick={() => window.location.href = '/dashboard'}
            className="mt-4 ml-2 px-4 py-2 bg-green-600 text-white rounded hover:bg-green-700"
          >
            Go to Dashboard
          </button>
        </div>
      </div>
    </div>
  );
};

export default TestAuth;
