import React from 'react';
import { useQuery } from 'react-query';
import { Building2, CheckCircle, Cloud, DollarSign } from 'lucide-react';
import { Alert } from '../components/ui/alert';
import { Button } from '../components/ui/button';

interface Plan {
  id: string;
  name: string;
  price: number;
  features: string[];
  description: string;
  isPopular?: boolean;
}

const plans: Plan[] = [
  {
    id: 'starter',
    name: 'Starter',
    price: 99,
    description: 'Perfect for small teams just getting started with cloud cost optimization',
    features: [
      'Basic cost analysis',
      'Up to 3 AWS accounts',
      'Basic reporting',
      'Email support',
    ],
  },
  {
    id: 'professional',
    name: 'Professional',
    price: 299,
    description: 'For growing teams with multiple cloud accounts',
    features: [
      'Advanced cost analysis',
      'Up to 10 AWS accounts',
      'Custom reporting',
      'Priority support',
      'Resource tagging',
      'Cost forecasting',
    ],
    isPopular: true,
  },
  {
    id: 'enterprise',
    name: 'Enterprise',
    price: 999,
    description: 'For large organizations with complex cloud infrastructure',
    features: [
      'Full cost analysis',
      'Unlimited AWS accounts',
      'Custom reporting & API access',
      '24/7 dedicated support',
      'Advanced ML forecasting',
      'Custom integrations',
      'SSO & SAML',
    ],
  },
];

interface OnboardingWelcomeProps {
  onNext: (planId: string) => void;
}

const OnboardingWelcome: React.FC<OnboardingWelcomeProps> = ({ onNext }) => {
  const [selectedPlan, setSelectedPlan] = React.useState<string | null>(null);

  return (
    <div className="max-w-6xl mx-auto px-4 py-8">
      <div className="text-center mb-12">
        <h1 className="text-3xl font-bold mb-4">Welcome to Yukti FinOps</h1>
        <p className="text-gray-600 text-lg">
          Choose the plan that best fits your organization's needs
        </p>
      </div>

      <div className="grid md:grid-cols-3 gap-8 mb-8">
        {plans.map((plan) => (
          <div
            key={plan.id}
            className={`relative rounded-lg ${
              plan.isPopular ? 'border-2 border-blue-500' : 'border border-gray-200'
            } bg-white p-6 shadow-sm hover:shadow-md transition-shadow cursor-pointer`}
            onClick={() => setSelectedPlan(plan.id)}
          >
            {plan.isPopular && (
              <div className="absolute -top-3 left-1/2 transform -translate-x-1/2">
                <span className="bg-blue-500 text-white px-3 py-1 rounded-full text-sm">
                  Most Popular
                </span>
              </div>
            )}

            <div className="text-center mb-6">
              <h3 className="text-xl font-semibold mb-2">{plan.name}</h3>
              <p className="text-gray-600 mb-4">{plan.description}</p>
              <div className="text-3xl font-bold mb-1">
                ${plan.price}
                <span className="text-base font-normal text-gray-500">/month</span>
              </div>
            </div>

            <ul className="space-y-3 mb-6">
              {plan.features.map((feature) => (
                <li key={feature} className="flex items-start">
                  <CheckCircle className="w-5 h-5 text-green-500 mr-2 flex-shrink-0" />
                  <span className="text-gray-600">{feature}</span>
                </li>
              ))}
            </ul>

            <Button
              variant={selectedPlan === plan.id ? 'default' : 'outline'}
              className="w-full"
              onClick={() => setSelectedPlan(plan.id)}
            >
              {selectedPlan === plan.id ? 'Selected' : 'Select Plan'}
            </Button>
          </div>
        ))}
      </div>

      <div className="text-center">
        <Button
          size="lg"
          disabled={!selectedPlan}
          onClick={() => selectedPlan && onNext(selectedPlan)}
        >
          Continue
        </Button>
      </div>
    </div>
  );
};

export default OnboardingWelcome;