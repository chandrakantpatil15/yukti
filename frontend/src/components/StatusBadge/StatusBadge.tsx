import React from 'react';
import { clsx } from 'clsx';
import { ResourceStatus } from '../../types';

interface StatusBadgeProps {
  status: ResourceStatus;
  label?: string;
  className?: string;
}

const getStatusColor = (status: ResourceStatus): string => {
  switch (status) {
    case 'running':
      return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200 border-green-200 dark:border-green-800';
    case 'stopped':
      return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200 border-yellow-200 dark:border-yellow-800';
    case 'terminated':
      return 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200 border-red-200 dark:border-red-800';
    default:
      return 'bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-200 border-gray-200 dark:border-gray-800';
  }
};

export const StatusBadge = ({ status, label, className }: StatusBadgeProps) => {
  return (
    <span className={clsx(
      'inline-flex px-3 py-1 text-sm font-semibold rounded-full border',
      getStatusColor(status),
      className
    )}>
      {label ?? status.charAt(0).toUpperCase() + status.slice(1)}
    </span>
  );
};