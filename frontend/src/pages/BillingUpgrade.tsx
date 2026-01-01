import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Check, ArrowRight, CreditCard, AlertCircle, Loader2 } from 'lucide-react';
import Loading from '../components/Loading';
import api from '../services/api';

interface Plan {
  slug: string;
  name: string;
  price: number;
  currency: string;
  interval: string;
  features: string[];
  recommended?: boolean;
}

const PLANS: Plan[] = [
  {
    slug: 'professional',
    name: 'Professional',
    price: 99,
    currency: 'USD',
    interval: 'month',
    recommended: true,
    features: [
      'Up to 3 AWS accounts',
      '25 hidden cost detectors',
      '10 whitelists',
      'AI-powered forecasting',
      'IaC generation',
      '90-day data retention',
      'Email support',
    ],
  },
  {
    slug: 'enterprise',
    name: 'Enterprise',
    price: 499,
    currency: 'USD',
    interval: 'month',
    features: [
      'Unlimited AWS accounts',
      '50 hidden cost detectors',
      'Unlimited whitelists',
      'Executive reports',
      'SSO integration',
      '1-year data retention',
      'Priority support (4-hour SLA)',
    ],
  },
  {
    slug: 'financial',
    name: 'Financial',
    price: 1999,
    currency: 'USD',
    interval: 'month',
    features: [
      'Everything in Enterprise',
      'Multi-cloud support (Azure, GCP)',
      'Custom detectors',
      'Dedicated success manager',
      '24/7 support (1-hour SLA)',
      'Custom integrations',
      'On-site training',
    ],
  },
];

const BillingUpgrade: React.FC = () => {
  const navigate = useNavigate();
  const [selectedPlan, setSelectedPlan] = useState<string>('professional');
  const [currency, setCurrency] = useState<string>('USD');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleUpgrade = async () => {
    if (!selectedPlan) {
      setError('Please select a plan');
      return;
    }

    setLoading(true);
    setError(null);

    try {
      const response = await api.createCheckoutSession(selectedPlan, currency);
      if (response.success && response.checkout_url) {
        // Redirect to Stripe Checkout
        window.location.href = response.checkout_url;
      } else {
        setError(response.error || 'Failed to create checkout session');
        setLoading(false);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to start checkout process');
      setLoading(false);
    }
  };

  const selectedPlanData = PLANS.find((p) => p.slug === selectedPlan);

  return (
    <div className="min-h-screen bg-gray-50 p-6">
      <div className="max-w-7xl mx-auto">
        <div className="mb-8">
          <button
            onClick={() => navigate('/billing')}
            className="text-indigo-600 hover:text-indigo-700 mb-4 flex items-center gap-2"
          >
            ← Back to Billing
          </button>
          <h1 className="text-3xl font-bold text-gray-900 mb-2">Upgrade Your Plan</h1>
          <p className="text-gray-600">Choose the plan that's right for your organization</p>
        </div>

        {error && (
          <div className="mb-6 bg-red-50 border border-red-200 rounded-lg p-4 flex items-center gap-3">
            <AlertCircle className="w-5 h-5 text-red-600 flex-shrink-0" />
            <p className="text-red-800">{error}</p>
          </div>
        )}

        {/* Currency Selector */}
        <div className="mb-6 flex justify-end">
          <div className="bg-white rounded-lg border border-gray-200 p-2 flex gap-2">
            {['USD', 'EUR', 'GBP', 'INR'].map((curr) => (
              <button
                key={curr}
                onClick={() => setCurrency(curr)}
                className={`px-4 py-2 rounded-md text-sm font-medium transition-colors ${
                  currency === curr
                    ? 'bg-indigo-600 text-white'
                    : 'text-gray-700 hover:bg-gray-100'
                }`}
              >
                {curr}
              </button>
            ))}
          </div>
        </div>

        {/* Plans Grid */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
          {PLANS.map((plan) => (
            <div
              key={plan.slug}
              className={`bg-white rounded-xl border-2 p-6 relative ${
                plan.recommended
                  ? 'border-indigo-500 shadow-lg'
                  : selectedPlan === plan.slug
                  ? 'border-indigo-300 shadow-md'
                  : 'border-gray-200'
              }`}
              onClick={() => setSelectedPlan(plan.slug)}
            >
              {plan.recommended && (
                <div className="absolute -top-3 left-1/2 transform -translate-x-1/2">
                  <span className="bg-indigo-600 text-white px-4 py-1 rounded-full text-sm font-medium">
                    Recommended
                  </span>
                </div>
              )}

              <div className="mb-4">
                <h3 className="text-2xl font-bold text-gray-900 mb-2">{plan.name}</h3>
                <div className="flex items-baseline gap-1">
                  <span className="text-4xl font-bold text-gray-900">
                    ${plan.price}
                  </span>
                  <span className="text-gray-600">/{plan.interval}</span>
                </div>
              </div>

              <button
                className={`w-full py-2 px-4 rounded-lg font-medium mb-6 transition-colors ${
                  selectedPlan === plan.slug
                    ? 'bg-indigo-600 text-white hover:bg-indigo-700'
                    : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                }`}
                onClick={(e) => {
                  e.stopPropagation();
                  setSelectedPlan(plan.slug);
                }}
              >
                {selectedPlan === plan.slug ? 'Selected' : 'Select Plan'}
              </button>

              <ul className="space-y-3">
                {plan.features.map((feature, index) => (
                  <li key={index} className="flex items-start gap-2">
                    <Check className="w-5 h-5 text-green-600 flex-shrink-0 mt-0.5" />
                    <span className="text-gray-700">{feature}</span>
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>

        {/* Selected Plan Summary */}
        {selectedPlanData && (
          <div className="bg-white rounded-xl border border-gray-200 p-6 mb-6">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Selected Plan</h3>
            <div className="flex items-center justify-between">
              <div>
                <p className="text-xl font-bold text-gray-900">{selectedPlanData.name}</p>
                <p className="text-gray-600">
                  ${selectedPlanData.price}/{selectedPlanData.interval} • {currency}
                </p>
              </div>
              <div className="text-right">
                <p className="text-2xl font-bold text-gray-900">
                  ${selectedPlanData.price}
                  <span className="text-lg text-gray-600">/{selectedPlanData.interval}</span>
                </p>
                <p className="text-sm text-gray-500">Billed monthly</p>
              </div>
            </div>
          </div>
        )}

        {/* Checkout Button */}
        <div className="bg-white rounded-xl border border-gray-200 p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-600 mb-1">You'll be redirected to Stripe Checkout</p>
              <p className="text-sm text-gray-500">
                Secure payment processing by Stripe. Cancel anytime.
              </p>
            </div>
            <button
              onClick={handleUpgrade}
              disabled={loading || !selectedPlan}
              className="px-6 py-3 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-colors disabled:bg-gray-400 disabled:cursor-not-allowed flex items-center gap-2 font-medium"
            >
              {loading ? (
                <>
                  <Loader2 className="w-5 h-5 animate-spin" />
                  Processing...
                </>
              ) : (
                <>
                  Continue to Checkout
                  <ArrowRight className="w-5 h-5" />
                </>
              )}
            </button>
          </div>
        </div>

        {/* Security Notice */}
        <div className="mt-6 text-center text-sm text-gray-500">
          <div className="flex items-center justify-center gap-2">
            <CreditCard className="w-4 h-4" />
            <span>Secure payment processing by Stripe. We never store your card details.</span>
          </div>
        </div>
      </div>
    </div>
  );
};

export default BillingUpgrade;

