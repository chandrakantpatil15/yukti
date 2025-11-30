import React from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { Building2, Mail, Phone, Globe, Users } from 'lucide-react';
import { Button } from '../ui/button';
import { Input } from '../ui/input';
import { Alert } from '../ui/alert';

const companySchema = z.object({
  companyName: z.string().min(2, 'Company name must be at least 2 characters'),
  email: z.string().email('Invalid email address'),
  phone: z.string().min(10, 'Phone number must be at least 10 digits'),
  website: z.string().url('Invalid website URL').optional(),
  employeeCount: z.number().min(1, 'Must have at least 1 employee'),
  industry: z.string().min(2, 'Please select an industry'),
});

type CompanyFormData = z.infer<typeof companySchema>;

interface OnboardingCompanyProps {
  onNext: (data: CompanyFormData) => void;
  onBack: () => void;
}

const industries = [
  'Technology',
  'Healthcare',
  'Finance',
  'Manufacturing',
  'Retail',
  'Education',
  'Other',
];

const OnboardingCompany: React.FC<OnboardingCompanyProps> = ({ onNext, onBack }) => {
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<CompanyFormData>({
    resolver: zodResolver(companySchema),
  });

  const onSubmit = async (data: CompanyFormData) => {
    try {
      await onNext(data);
    } catch (error) {
      console.error('Error submitting company details:', error);
    }
  };

  return (
    <div className="max-w-2xl mx-auto px-4 py-8">
      <div className="text-center mb-8">
        <h2 className="text-2xl font-bold mb-2">Company Details</h2>
        <p className="text-gray-600">Tell us about your organization</p>
      </div>

      <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
        <div>
          <label className="block text-sm font-medium mb-2">Company Name</label>
          <div className="relative">
            <Building2 className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-5 h-5" />
            <Input
              {...register('companyName')}
              className="pl-10"
              placeholder="Enter company name"
            />
          </div>
          {errors.companyName && (
            <p className="mt-1 text-sm text-red-600">{errors.companyName.message}</p>
          )}
        </div>

        <div>
          <label className="block text-sm font-medium mb-2">Business Email</label>
          <div className="relative">
            <Mail className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-5 h-5" />
            <Input
              {...register('email')}
              type="email"
              className="pl-10"
              placeholder="Enter business email"
            />
          </div>
          {errors.email && (
            <p className="mt-1 text-sm text-red-600">{errors.email.message}</p>
          )}
        </div>

        <div>
          <label className="block text-sm font-medium mb-2">Phone Number</label>
          <div className="relative">
            <Phone className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-5 h-5" />
            <Input
              {...register('phone')}
              type="tel"
              className="pl-10"
              placeholder="Enter phone number"
            />
          </div>
          {errors.phone && (
            <p className="mt-1 text-sm text-red-600">{errors.phone.message}</p>
          )}
        </div>

        <div>
          <label className="block text-sm font-medium mb-2">Website (Optional)</label>
          <div className="relative">
            <Globe className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-5 h-5" />
            <Input
              {...register('website')}
              type="url"
              className="pl-10"
              placeholder="Enter website URL"
            />
          </div>
          {errors.website && (
            <p className="mt-1 text-sm text-red-600">{errors.website.message}</p>
          )}
        </div>

        <div>
          <label className="block text-sm font-medium mb-2">Number of Employees</label>
          <div className="relative">
            <Users className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-5 h-5" />
            <Input
              {...register('employeeCount', { valueAsNumber: true })}
              type="number"
              className="pl-10"
              placeholder="Enter number of employees"
              min="1"
            />
          </div>
          {errors.employeeCount && (
            <p className="mt-1 text-sm text-red-600">{errors.employeeCount.message}</p>
          )}
        </div>

        <div>
          <label className="block text-sm font-medium mb-2">Industry</label>
          <select
            {...register('industry')}
            className="w-full rounded-md border border-gray-200 py-2 pl-3 pr-10 text-gray-900 focus:ring-2 focus:ring-blue-500"
          >
            <option value="">Select an industry</option>
            {industries.map((industry) => (
              <option key={industry} value={industry}>
                {industry}
              </option>
            ))}
          </select>
          {errors.industry && (
            <p className="mt-1 text-sm text-red-600">{errors.industry.message}</p>
          )}
        </div>

        <div className="flex justify-between pt-6">
          <Button type="button" variant="outline" onClick={onBack}>
            Back
          </Button>
          <Button type="submit" disabled={isSubmitting}>
            {isSubmitting ? 'Saving...' : 'Continue'}
          </Button>
        </div>
      </form>
    </div>
  );
};

export default OnboardingCompany;