export const tokens = {
  colors: {
    primary: {
      50: '#f0f9ff',
      100: '#e0f2fe',
      500: '#0ea5e9',
      600: '#0284c7',
      700: '#0369a1',
    },
    neutral: {
      50: '#f8fafc',
      100: '#f1f5f9',
      200: '#e2e8f0',
      300: '#cbd5e1',
      400: '#94a3b8',
      500: '#64748b',
      600: '#475569',
      700: '#334155',
      800: '#1e293b',
      900: '#0f172a',
    },
    success: {
      100: '#dcfce7',
      500: '#22c55e',
      700: '#15803d',
    },
    warning: {
      100: '#fef3c7',
      500: '#f59e0b',
      700: '#b45309',
    },
    error: {
      100: '#fee2e2',
      500: '#ef4444',
      700: '#b91c1c',
    },
  },
  spacing: {
    '2xs': '4px',
    xs: '8px',
    sm: '12px',
    md: '16px',
    lg: '24px',
    xl: '32px',
    '2xl': '48px',
  },
  typography: {
    display: {
      lg: {
        fontSize: '32px',
        lineHeight: '40px',
        letterSpacing: '-0.02em',
        fontWeight: 700,
      },
      md: {
        fontSize: '24px',
        lineHeight: '32px',
        letterSpacing: '-0.02em',
        fontWeight: 700,
      },
    },
    heading: {
      lg: {
        fontSize: '20px',
        lineHeight: '28px',
        fontWeight: 600,
      },
      md: {
        fontSize: '18px',
        lineHeight: '24px',
        fontWeight: 600,
      },
    },
    body: {
      lg: {
        fontSize: '16px',
        lineHeight: '24px',
      },
      md: {
        fontSize: '14px',
        lineHeight: '20px',
      },
      sm: {
        fontSize: '12px',
        lineHeight: '16px',
      },
    },
  },
  shadows: {
    sm: '0 1px 2px rgba(0, 0, 0, 0.05)',
    md: '0 4px 6px -1px rgba(0, 0, 0, 0.1)',
    lg: '0 10px 15px -3px rgba(0, 0, 0, 0.1)',
  },
} as const;