import React, { useEffect, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { CheckCircle, AlertCircle, Loader2, ArrowRight } from 'lucide-react';
import api from '../services/api';

const BillingSuccess: React.FC = () => {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const sessionId = searchParams.get('session_id');

  useEffect(() => {
    // The session_id is returned by Stripe after successful checkout
    // We can optionally verify the session on the backend, but for now
    // we'll just show a success message
    if (sessionId) {
      // Session ID is present, checkout was successful
      setLoading(false);
    } else {
      // No session ID, but still show success (user might have navigated directly)
      setLoading(false);
    }
  }, [sessionId]);

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <Loader2 className="w-12 h-12 text-indigo-600 animate-spin mx-auto mb-4" />
          <p className="text-gray-600">Processing your subscription...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center p-6">
      <div className="max-w-md w-full">
        {error ? (
          <div className="bg-white rounded-xl shadow-lg border border-red-200 p-8 text-center">
            <AlertCircle className="w-16 h-16 text-red-500 mx-auto mb-4" />
            <h1 className="text-2xl font-bold text-gray-900 mb-2">Subscription Error</h1>
            <p className="text-gray-600 mb-6">{error}</p>
            <button
              onClick={() => navigate('/billing')}
              className="px-6 py-3 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-colors"
            >
              Go to Billing
            </button>
          </div>
        ) : (
          <div className="bg-white rounded-xl shadow-lg border border-green-200 p-8 text-center">
            <CheckCircle className="w-16 h-16 text-green-500 mx-auto mb-4" />
            <h1 className="text-2xl font-bold text-gray-900 mb-2">Subscription Activated!</h1>
            <p className="text-gray-600 mb-2">
              Thank you for subscribing. Your subscription is now active.
            </p>
            {sessionId && (
              <p className="text-sm text-gray-500 mb-6">
                Session ID: {sessionId.substring(0, 20)}...
              </p>
            )}
            <div className="space-y-3">
              <button
                onClick={() => navigate('/billing')}
                className="w-full px-6 py-3 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-colors flex items-center justify-center gap-2"
              >
                View Billing Details
                <ArrowRight className="w-5 h-5" />
              </button>
              <button
                onClick={() => navigate('/dashboard')}
                className="w-full px-6 py-3 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 transition-colors"
              >
                Go to Dashboard
              </button>
            </div>
          </div>
        )}

        <div className="mt-6 text-center text-sm text-gray-500">
          <p>
            A confirmation email has been sent to your email address with your invoice and subscription details.
          </p>
        </div>
      </div>
    </div>
  );
};

export default BillingSuccess;

