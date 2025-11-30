import React from 'react';
import { clsx } from 'clsx';

interface CardProps {
  className?: string;
  children: React.ReactNode;
  onClick?: () => void;
  interactive?: boolean;
  selected?: boolean;
  title?: string;
}

export const Card: React.FC<CardProps> = ({
  className,
  children,
  onClick,
  interactive = false,
  selected = false,
  title,
}) => {
  return (
    <div
      className={clsx(
        'bg-white dark:bg-neutral-800 rounded-lg shadow-sm border border-neutral-200 dark:border-neutral-700',
        'transition duration-200 ease-in-out',
        {
          'cursor-pointer hover:border-primary-500 hover:ring-1 hover:ring-primary-500 dark:hover:border-primary-400 dark:hover:ring-primary-400': interactive,
          'border-primary-500 ring-1 ring-primary-500 dark:border-primary-400 dark:ring-primary-400': selected,
        },
        className
      )}
      onClick={onClick}
      role={interactive ? 'button' : undefined}
      tabIndex={interactive ? 0 : undefined}
    >
      {title && (
        <div className="px-6 py-4 border-b border-neutral-200 dark:border-neutral-700">
          <h3 className="text-lg font-medium leading-6 text-neutral-900 dark:text-neutral-100">
            {title}
          </h3>
        </div>
      )}
      <div className="px-6 py-4">{children}</div>
    </div>
  );
};

interface CardHeaderProps {
  className?: string;
  children: React.ReactNode;
}

export const CardHeader: React.FC<CardHeaderProps> = ({ className, children }) => (
  <div className={clsx('px-6 py-4 border-b border-neutral-200', className)}>
    {children}
  </div>
);

interface CardContentProps {
  className?: string;
  children: React.ReactNode;
}

export const CardContent: React.FC<CardContentProps> = ({ className, children }) => (
  <div className={clsx('px-6 py-4', className)}>
    {children}
  </div>
);

interface CardFooterProps {
  className?: string;
  children: React.ReactNode;
}

export const CardFooter: React.FC<CardFooterProps> = ({ className, children }) => (
  <div className={clsx('px-6 py-4 border-t border-neutral-200', className)}>
    {children}
  </div>
);