import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import api from '../services/api';
import { setToken } from '../lib/auth';

const Login: React.FC = () => {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const navigate = useNavigate();

  const validateEmail = (email: string) => {
    return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    // Validation
    if (!email || !password) {
      setError('Email and password are required');
      return;
    }

    if (!validateEmail(email)) {
      setError('Please enter a valid email address');
      return;
    }

    if (password.length < 6) {
      setError('Please enter your password');
      return;
    }

    setLoading(true);

    try {
      const data = await api.login(email, password);
      console.log('Login response:', data);

      if (data.success) {
        console.log('[Login] Login successful, token received');
        setToken(data.token);
        
        // Check onboarding status
        try {
          const onboardingStatus = await api.getAWSConnection();
          console.log('[Login] Onboarding status:', onboardingStatus);
          
          if (!onboardingStatus.success || !onboardingStatus.data) {
            // No AWS connection - redirect to onboarding
            console.log('[Login] No AWS connection found, redirecting to onboarding');
            setTimeout(() => {
              window.location.href = '/onboarding';
            }, 500);
          } else {
            // AWS connection exists - redirect to dashboard
            console.log('[Login] AWS connection found, redirecting to dashboard');
            setTimeout(() => {
              window.location.href = '/dashboard';
            }, 500);
          }
        } catch (err) {
          // Error checking onboarding - assume not onboarded, redirect to onboarding
          console.log('[Login] Error checking onboarding status, redirecting to onboarding');
          setTimeout(() => {
            window.location.href = '/onboarding';
          }, 500);
        }
      } else {
        setError(data.error || 'Login failed');
      }
    } catch (err: any) {
      console.error('Login error:', err);
      setError(err.message || 'Network error. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <div className="max-w-md w-full space-y-8">
        <div>
          <h2 className="mt-6 text-center text-3xl font-extrabold text-gray-900">
            Sign in to Yukti
          </h2>
          <p className="mt-2 text-center text-sm text-gray-600">
            Cloud Cost Optimization Platform
          </p>
          <div className="mt-3 p-3 bg-blue-50 border border-blue-200 rounded-md">
            <p className="text-xs text-blue-700 text-center">
              🔒 <strong>Secure Access:</strong> Your financial data is protected with enterprise-grade security
            </p>
          </div>
        </div>
        
        <form className="mt-8 space-y-6" onSubmit={handleSubmit}>
          <div className="rounded-md shadow-sm -space-y-px">
            <div>
              <input
                type="email"
                required
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-t-md focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                placeholder="Email address"
              />
            </div>
            <div>
              <input
                type="password"
                required
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-b-md focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                placeholder="Password"
              />
            </div>
          </div>

          {error && (
            <div className="text-red-600 text-sm text-center">{error}</div>
          )}

          <div>
            <button
              type="submit"
              disabled={loading}
              className="group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:bg-gray-400"
            >
              {loading ? 'Signing in...' : 'Sign in'}
            </button>
          </div>

          <div className="text-center">
            <a href="/signup" className="text-blue-600 hover:text-blue-500">
              Don't have an account? Sign up
            </a>
          </div>
        </form>
      </div>
    </div>
  );
};

export default Login;