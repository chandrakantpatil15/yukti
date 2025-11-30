import React, { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { Cloud, Copy, Terminal, CheckCircle, AlertCircle } from 'lucide-react';
import { Button } from '../ui/button';
import { Input } from '../ui/input';
import { Alert } from '../ui/alert';
import api from '../../services/api';

const awsAccountSchema = z.object({
  accountId: z.string().min(12, 'AWS Account ID must be 12 digits').max(12),
  roleName: z.string().min(1, 'Role name is required'),
});

type AwsAccountFormData = z.infer<typeof awsAccountSchema>;

interface OnboardingAwsProps {
  onNext: (data: AwsAccountFormData) => void;
  onBack: () => void;
}

const defaultRoleName = 'YuktiFinOpsRole';

const OnboardingAws: React.FC<OnboardingAwsProps> = ({ onNext, onBack }) => {
  const [validationStatus, setValidationStatus] = useState<'idle' | 'validating' | 'success' | 'error'>('idle');
  const [showMockButton] = useState(process.env.NODE_ENV === 'development');

  const {
    register,
    handleSubmit,
    watch,
    formState: { errors, isSubmitting },
  } = useForm<AwsAccountFormData>({
    resolver: zodResolver(awsAccountSchema),
    defaultValues: {
      roleName: defaultRoleName,
    },
  });



  const values = watch();

  const getCFTemplate = () => {
    return `{
  "AWSTemplateFormatVersion": "2010-09-09",
  "Description": "Yukti FinOps Integration Role",
  "Parameters": {
    "ExternalId": {
      "Type": "String",
      "Description": "External ID for cross-account access (provided by Yukti)",
      "Default": "yukti-secure-access"
    }
  },
  "Resources": {
    "YuktiRole": {
      "Type": "AWS::IAM::Role",
      "Properties": {
        "RoleName": "${values.roleName}",
        "AssumeRolePolicyDocument": {
          "Version": "2012-10-17",
          "Statement": [
            {
              "Effect": "Allow",
              "Principal": {
                "AWS": "arn:aws:iam::${process.env.REACT_APP_YUKTI_AWS_ACCOUNT}:root"
              },
              "Action": "sts:AssumeRole",
              "Condition": {
                "StringEquals": {
                  "sts:ExternalId": "yukti-secure-access"
                }
              }
            }
          ]
        },
        "Policies": [
          {
            "PolicyName": "YuktiReadOnlyPolicy",
            "PolicyDocument": {
              "Version": "2012-10-17",
              "Statement": [
                {
                  "Sid": "AllowReadOnlyAccessForAllServices",
                  "Effect": "Allow",
                  "Action": [
                    "*:Describe*",
                    "*:Get*",
                    "*:List*"
                  ],
                  "Resource": "*"
                },
                {
                  "Sid": "AllowTagReadAccess",
                  "Effect": "Allow",
                  "Action": [
                    "tag:GetResources",
                    "tag:GetTagKeys",
                    "tag:GetTagValues"
                  ],
                  "Resource": "*"
                },
                {
                  "Sid": "AllowCloudWatchReadAccess",
                  "Effect": "Allow",
                  "Action": [
                    "cloudwatch:GetMetricData",
                    "cloudwatch:GetMetricStatistics",
                    "cloudwatch:ListMetrics",
                    "logs:Describe*",
                    "logs:Get*",
                    "logs:FilterLogEvents"
                  ],
                  "Resource": "*"
                },
                {
                  "Sid": "AllowS3ReadAccess",
                  "Effect": "Allow",
                  "Action": [
                    "s3:GetObject",
                    "s3:GetObjectAcl",
                    "s3:GetBucketAcl",
                    "s3:GetBucketLocation",
                    "s3:GetBucketPolicy",
                    "s3:ListAllMyBuckets",
                    "s3:ListBucket"
                  ],
                  "Resource": "*"
                },
                {
                  "Sid": "DenySensitiveAPIAccess",
                  "Effect": "Deny",
                  "Action": [
                    "iam:*",
                    "kms:*",
                    "organizations:*",
                    "account:*",
                    "support:*",
                    "ce:*",
                    "cur:*",
                    "savingsplans:*",
                    "aws-portal:*",
                    "budgets:*",
                    "license-manager:*",
                    "secretsmanager:GetSecretValue",
                    "ssm:GetParameter",
                    "ssm:GetParameters",
                    "ssm:GetParametersByPath"
                  ],
                  "Resource": "*"
                }
              ]
            }
          }
        ]
      }
    }
  }
}`;
  };

  const handleCopyTemplate = () => {
    navigator.clipboard.writeText(getCFTemplate());
  };

  const validateAwsCredentials = async (data: AwsAccountFormData) => {
    setValidationStatus('validating');
    try {
      const user = JSON.parse(localStorage.getItem('yukti_user') || '{}');
      
      // TODO: Implement real AWS STS AssumeRole validation
      // For now, mock validation in development
      if (process.env.NODE_ENV === 'development') {
        await new Promise(resolve => setTimeout(resolve, 2000));
        setValidationStatus('success');
        setTimeout(() => onNext(data), 500);
      } else {
        // Production: Call real validation endpoint
        // Backend will generate and use external ID automatically
        const response = await api.connectAWS(
          user.tenant_id,
          data.accountId,
          `arn:aws:iam::${data.accountId}:role/${data.roleName}`,
          'yukti-secure-access' // Backend will replace with tenant-specific ID
        );
        if (response.verified) {
          setValidationStatus('success');
          setTimeout(() => onNext(data), 500);
        } else {
          setValidationStatus('error');
        }
      }
    } catch (error) {
      setValidationStatus('error');
      console.error('Error validating AWS credentials:', error);
    }
  };

  return (
    <div className="max-w-3xl mx-auto px-4 py-8">
      <div className="text-center mb-8">
        <h2 className="text-2xl font-bold mb-2">Connect AWS Account</h2>
        <p className="text-gray-600">Set up access to your AWS account for cost analysis</p>
      </div>

      <form onSubmit={handleSubmit(validateAwsCredentials)} className="space-y-6">
        <div>
          <label className="block text-sm font-medium mb-2">AWS Account ID</label>
          <div className="relative">
            <Cloud className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-5 h-5" />
            <Input
              {...register('accountId')}
              className="pl-10"
              placeholder="Enter your 12-digit AWS account ID"
            />
          </div>
          {errors.accountId && (
            <p className="mt-1 text-sm text-red-600">{errors.accountId.message}</p>
          )}
        </div>



        <div>
          <label className="block text-sm font-medium mb-2">IAM Role Name</label>
          <Input
            {...register('roleName')}
            placeholder="Enter IAM role name"
          />
          {errors.roleName && (
            <p className="mt-1 text-sm text-red-600">{errors.roleName.message}</p>
          )}
        </div>

        <div className="bg-gray-50 rounded-lg p-4">
          <div className="flex items-center justify-between mb-4">
            <div className="flex items-center">
              <Terminal className="w-5 h-5 mr-2" />
              <h3 className="font-medium">CloudFormation Template</h3>
            </div>
            <Button type="button" variant="outline" size="sm" onClick={handleCopyTemplate}>
              <Copy className="w-4 h-4 mr-2" />
              Copy Template
            </Button>
          </div>
          <pre className="bg-gray-900 text-gray-100 p-4 rounded-md text-sm overflow-x-auto">
            {getCFTemplate()}
          </pre>
        </div>

        {showMockButton && (
          <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4 mb-4">
            <p className="text-sm text-yellow-800 mb-2">
              🧪 <strong>Development Mode:</strong> Skip AWS setup for testing
            </p>
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() => {
                onNext({
                  accountId: '999888777666',
                  roleName: values.roleName,
                });
              }}
            >
              Skip AWS Connection (Mock)
            </Button>
          </div>
        )}

        <div className="flex items-center justify-between pt-6">
          <Button type="button" variant="outline" onClick={onBack}>
            Back
          </Button>
          <Button
            type="submit"
            disabled={isSubmitting || validationStatus === 'validating'}
          >
            {validationStatus === 'validating' ? (
              <>Validating...</>
            ) : (
              <>
                {validationStatus === 'success' && (
                  <CheckCircle className="w-4 h-4 mr-2" />
                )}
                Continue
              </>
            )}
          </Button>
        </div>

        {validationStatus === 'error' && (
          <Alert variant="destructive" className="mt-4">
            <AlertCircle className="w-4 h-4 mr-2" />
            Failed to validate AWS credentials. Please check your account ID and role setup.
          </Alert>
        )}
      </form>
    </div>
  );
};

export default OnboardingAws;