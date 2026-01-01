import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { CreditCard, Calendar, FileText, AlertCircle, CheckCircle, Clock, ArrowRight } from 'lucide-react';
import { useQuery } from 'react-query';
import Loading from '../components/Loading';
import api from '../services/api';

interface BillingInfo {
  subscription_tier: string;
  trial_ends_at: string | null;
  subscription: {
    id: number;
    stripe_subscription_id: string;
    plan_slug: string;
    status: string;
    current_period_start: string;
    current_period_end: string;
    cancel_at_period_end: boolean;
  } | null;
  invoices: Array<{
    id: number;
    invoice_id: string;
    amount_cents: number;
    tax_cents: number;
    currency: string;
    paid: boolean;
    paid_at: string | null;
    pdf_url: string | null;
    invoice_number: string | null;
    created_at: string;
  }>;
}

const Billing: React.FC = () => {
  const navigate = useNavigate();
  const [error, setError] = useState<string | null>(null);

  const { data, isLoading, refetch } = useQuery<{ success: boolean; data: BillingInfo }>(
    'billing-info',
    async () => {
      try {
        const response = await api.getBillingInfo();
        return response;
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load billing information');
        throw err;
      }
    },
    {
      retry: 2,
      refetchOnWindowFocus: false,
    }
  );

  const billingInfo = data?.data;

  const formatCurrency = (cents: number, currency: string = 'USD') => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: currency,
    }).format(cents / 100);
  };

  const formatDate = (dateString: string | null) => {
    if (!dateString) return 'N/A';
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    });
  };

  const getTrialDaysRemaining = () => {
    if (!billingInfo?.trial_ends_at) return null;
    const endDate = new Date(billingInfo.trial_ends_at);
    const now = new Date();
    const diffTime = endDate.getTime() - now.getTime();
    const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));
    return diffDays > 0 ? diffDays : 0;
  };

  const isTrialExpired = () => {
    if (!billingInfo?.trial_ends_at) return false;
    return new Date(billingInfo.trial_ends_at) < new Date();
  };

  const getTierDisplayName = (tier: string) => {
    const tierMap: Record<string, string> = {
      FREE: 'Free',
      PROFESSIONAL: 'Professional',
      ENTERPRISE: 'Enterprise',
      FINANCIAL: 'Financial',
      professional: 'Professional',
      enterprise: 'Enterprise',
      financial: 'Financial',
      free: 'Free',
    };
    return tierMap[tier] || tier;
  };

  const getStatusBadge = (status: string) => {
    const statusMap: Record<string, { color: string; label: string; icon: React.ReactNode }> = {
      active: {
        color: 'bg-green-100 text-green-800',
        label: 'Active',
        icon: <CheckCircle className="w-4 h-4" />,
      },
      trialing: {
        color: 'bg-blue-100 text-blue-800',
        label: 'Trial',
        icon: <Clock className="w-4 h-4" />,
      },
      canceled: {
        color: 'bg-gray-100 text-gray-800',
        label: 'Canceled',
        icon: <AlertCircle className="w-4 h-4" />,
      },
      past_due: {
        color: 'bg-red-100 text-red-800',
        label: 'Past Due',
        icon: <AlertCircle className="w-4 h-4" />,
      },
    };

    const statusInfo = statusMap[status] || {
      color: 'bg-gray-100 text-gray-800',
      label: status,
      icon: <AlertCircle className="w-4 h-4" />,
    };

    return (
      <span className={`inline-flex items-center gap-1 px-3 py-1 rounded-full text-sm font-medium ${statusInfo.color}`}>
        {statusInfo.icon}
        {statusInfo.label}
      </span>
    );
  };

  if (isLoading) {
    return <Loading message="Loading billing information..." />;
  }

  if (error) {
    return (
      <div className="min-h-screen bg-gray-50 p-6 flex items-center justify-center">
        <div className="bg-red-50 border border-red-200 rounded-xl p-8 max-w-md shadow-lg">
          <AlertCircle className="w-12 h-12 text-red-500 mx-auto mb-4" />
          <h2 className="text-xl font-semibold text-red-800 text-center mb-2">Error Loading Billing</h2>
          <p className="text-red-700 text-center mb-6">{error}</p>
          <button
            onClick={() => refetch()}
            className="w-full px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 transition-colors"
          >
            Retry
          </button>
        </div>
      </div>
    );
  }

  const trialDaysRemaining = getTrialDaysRemaining();
  const hasActiveSubscription = billingInfo?.subscription?.status === 'active' || billingInfo?.subscription?.status === 'trialing';

  return (
    <div className="min-h-screen bg-gray-50 p-6">
      <div className="max-w-7xl mx-auto">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900 mb-2">Billing & Subscription</h1>
          <p className="text-gray-600">Manage your subscription and view billing history</p>
        </div>

        {/* Current Plan Card */}
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6 mb-6">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-xl font-semibold text-gray-900">Current Plan</h2>
            {!hasActiveSubscription && billingInfo?.subscription_tier === 'FREE' && (
              <button
                onClick={() => navigate('/billing/upgrade')}
                className="px-4 py-2 bg-indigo-600 text-white rounded-lg hover:bg-indigo-700 transition-colors flex items-center gap-2"
              >
                Upgrade Plan
                <ArrowRight className="w-4 h-4" />
              </button>
            )}
          </div>

          <div className="flex items-center gap-4 mb-4">
            <div className="text-3xl font-bold text-gray-900">
              {getTierDisplayName(billingInfo?.subscription_tier || 'FREE')}
            </div>
            {billingInfo?.subscription && getStatusBadge(billingInfo.subscription.status)}
          </div>

          {/* Trial Status */}
          {billingInfo?.trial_ends_at && (
            <div className="mt-4 p-4 bg-blue-50 border border-blue-200 rounded-lg">
              <div className="flex items-center gap-2 mb-2">
                <Clock className="w-5 h-5 text-blue-600" />
                <span className="font-medium text-blue-900">Trial Period</span>
              </div>
              {isTrialExpired() ? (
                <p className="text-blue-800">
                  Your trial has expired. {hasActiveSubscription ? 'You have an active subscription.' : 'Please subscribe to continue using the service.'}
                </p>
              ) : trialDaysRemaining !== null && trialDaysRemaining > 0 ? (
                <p className="text-blue-800">
                  {trialDaysRemaining} day{trialDaysRemaining !== 1 ? 's' : ''} remaining in your trial period.
                </p>
              ) : (
                <p className="text-blue-800">Your trial period has ended.</p>
              )}
            </div>
          )}

          {/* Subscription Details */}
          {billingInfo?.subscription && (
            <div className="mt-4 grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <p className="text-sm text-gray-600 mb-1">Current Period</p>
                <p className="text-gray-900 font-medium">
                  {formatDate(billingInfo.subscription.current_period_start)} - {formatDate(billingInfo.subscription.current_period_end)}
                </p>
              </div>
              {billingInfo.subscription.cancel_at_period_end && (
                <div>
                  <p className="text-sm text-gray-600 mb-1">Status</p>
                  <p className="text-orange-600 font-medium">Cancels at end of period</p>
                </div>
              )}
            </div>
          )}
        </div>

        {/* Invoices */}
        <div className="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-xl font-semibold text-gray-900">Billing History</h2>
          </div>

          {billingInfo?.invoices && billingInfo.invoices.length > 0 ? (
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Invoice
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Date
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Amount
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Status
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Actions
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {billingInfo.invoices.map((invoice) => (
                    <tr key={invoice.id} className="hover:bg-gray-50">
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="text-sm font-medium text-gray-900">
                          {invoice.invoice_number || invoice.invoice_id.substring(0, 20) + '...'}
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {formatDate(invoice.created_at)}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                        {formatCurrency(invoice.amount_cents + invoice.tax_cents, invoice.currency)}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        {invoice.paid ? (
                          <span className="inline-flex items-center gap-1 px-3 py-1 rounded-full text-sm font-medium bg-green-100 text-green-800">
                            <CheckCircle className="w-4 h-4" />
                            Paid
                          </span>
                        ) : (
                          <span className="inline-flex items-center gap-1 px-3 py-1 rounded-full text-sm font-medium bg-yellow-100 text-yellow-800">
                            <Clock className="w-4 h-4" />
                            Pending
                          </span>
                        )}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm">
                        {invoice.pdf_url && (
                          <a
                            href={invoice.pdf_url}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="text-indigo-600 hover:text-indigo-900 flex items-center gap-1"
                          >
                            <FileText className="w-4 h-4" />
                            Download
                          </a>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <div className="text-center py-12">
              <FileText className="w-12 h-12 text-gray-400 mx-auto mb-4" />
              <p className="text-gray-600">No invoices yet</p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default Billing;

