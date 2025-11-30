import React from 'react';
import { clsx } from 'clsx';

interface Tab {
  id: string;
  label: string;
  icon?: React.ReactNode;
}

interface TabGroupProps {
  tabs: Tab[];
  activeTab: string;
  onChange: (tabId: string) => void;
  variant?: 'line' | 'pill';
}

export const TabGroup: React.FC<TabGroupProps> = ({
  tabs,
  activeTab,
  onChange,
  variant = 'line'
}) => {
  const getTabStyles = (isActive: boolean) => {
    if (variant === 'line') {
      return clsx(
        'px-1 py-4 text-sm font-medium border-b-2 transition-colors',
        'hover:text-neutral-900 dark:hover:text-white',
        isActive
          ? 'border-primary-500 text-primary-600 dark:border-primary-400 dark:text-primary-400'
          : 'border-transparent text-neutral-500 dark:text-neutral-400 hover:border-neutral-300 dark:hover:border-neutral-600'
      );
    }

    return clsx(
      'px-3 py-1.5 text-sm font-medium rounded-full transition-colors',
      'focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500 dark:focus:ring-offset-neutral-900',
      isActive
        ? 'bg-primary-50 text-primary-700 dark:bg-primary-900 dark:text-primary-100'
        : 'text-neutral-500 dark:text-neutral-400 hover:text-neutral-900 dark:hover:text-white hover:bg-neutral-50 dark:hover:bg-neutral-800'
    );
  };

  return (
    <nav className={clsx('flex', variant === 'line' ? 'space-x-8' : 'space-x-2')}>
      {tabs.map((tab) => (
        <button
          key={tab.id}
          onClick={() => onChange(tab.id)}
          className={getTabStyles(activeTab === tab.id)}
        >
          <span className="inline-flex items-center">
            {tab.icon && <span className="mr-2">{tab.icon}</span>}
            {tab.label}
          </span>
        </button>
      ))}
    </nav>
  );
};