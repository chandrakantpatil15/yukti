import React from 'react';
import { PasswordValidationResult, getPasswordStrengthColor, getPasswordStrengthText } from '../utils/passwordValidation';

interface PasswordStrengthIndicatorProps {
  validation: PasswordValidationResult;
  password: string;
}

const PasswordStrengthIndicator: React.FC<PasswordStrengthIndicatorProps> = ({ validation, password }) => {
  if (!password) return null;

  return (
    <div className="mt-2 space-y-2">
      {/* Strength Bar */}
      <div className="flex items-center space-x-2">
        <div className="flex-1 bg-gray-200 rounded-full h-2">
          <div
            className={`h-2 rounded-full transition-all duration-300 ${getPasswordStrengthColor(validation.score)}`}
            style={{ width: `${validation.score}%` }}
          />
        </div>
        <span className={`text-sm font-medium ${
          validation.score < 40 ? 'text-red-600' :
          validation.score < 60 ? 'text-orange-600' :
          validation.score < 80 ? 'text-yellow-600' :
          'text-green-600'
        }`}>
          {getPasswordStrengthText(validation.score)}
        </span>
      </div>

      {/* Requirements List */}
      <div className="text-xs space-y-1">
        {validation.requirements.map((req, index) => (
          <div key={index} className={`flex items-center space-x-2 ${req.met ? 'text-green-600' : 'text-gray-500'}`}>
            <span className="w-3 h-3 flex items-center justify-center">
              {req.met ? '✓' : '○'}
            </span>
            <span>{req.message}</span>
          </div>
        ))}
      </div>

      {/* Security Notice */}
      {validation.isValid && (
        <div className="mt-3 p-2 bg-green-50 border border-green-200 rounded text-xs text-green-700">
          <strong>🔒 Enterprise Security:</strong> Your password meets our security standards for protecting financial data.
        </div>
      )}
    </div>
  );
};

export default PasswordStrengthIndicator;