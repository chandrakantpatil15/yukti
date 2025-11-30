// Enterprise-grade password validation for financial data security

export interface PasswordRequirement {
  met: boolean;
  message: string;
}

export interface PasswordValidationResult {
  isValid: boolean;
  score: number; // 0-100
  requirements: PasswordRequirement[];
}

export const validatePassword = (password: string): PasswordValidationResult => {
  const requirements: PasswordRequirement[] = [
    {
      met: password.length >= 12,
      message: 'At least 12 characters long'
    },
    {
      met: /[a-z]/.test(password),
      message: 'Contains lowercase letters'
    },
    {
      met: /[A-Z]/.test(password),
      message: 'Contains uppercase letters'
    },
    {
      met: /\d/.test(password),
      message: 'Contains numbers'
    },
    {
      met: /[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]/.test(password),
      message: 'Contains special characters (!@#$%^&*)'
    },
    {
      met: !/(.)\1{2,}/.test(password),
      message: 'No more than 2 consecutive identical characters'
    },
    {
      met: !isCommonPassword(password),
      message: 'Not a commonly used password'
    },
    {
      met: !/^(password|123456|qwerty|admin|user)/i.test(password),
      message: 'Does not start with common weak patterns'
    }
  ];

  const metCount = requirements.filter(req => req.met).length;
  const score = Math.round((metCount / requirements.length) * 100);
  const isValid = metCount === requirements.length;

  return {
    isValid,
    score,
    requirements
  };
};

// Common passwords to reject
const commonPasswords = [
  'password', '123456789', '12345678', '1234567890', 'qwerty123',
  'password123', 'admin123', 'welcome123', 'letmein123', 'monkey123',
  'dragon123', 'sunshine123', 'master123', 'shadow123', 'football123'
];

const isCommonPassword = (password: string): boolean => {
  return commonPasswords.some(common => 
    password.toLowerCase().includes(common.toLowerCase())
  );
};

export const getPasswordStrengthColor = (score: number): string => {
  if (score < 40) return 'bg-red-500';
  if (score < 60) return 'bg-orange-500';
  if (score < 80) return 'bg-yellow-500';
  return 'bg-green-500';
};

export const getPasswordStrengthText = (score: number): string => {
  if (score < 40) return 'Weak';
  if (score < 60) return 'Fair';
  if (score < 80) return 'Good';
  return 'Strong';
};