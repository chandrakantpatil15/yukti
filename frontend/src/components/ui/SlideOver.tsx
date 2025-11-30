import React from 'react';
import { X } from 'lucide-react';
import { clsx } from 'clsx';

interface SlideOverProps {
  isOpen: boolean;
  onClose: () => void;
  title: string;
  subtitle?: string;
  children: React.ReactNode;
  headerBackground?: string;
  headerTextColor?: string;
}

export const SlideOver: React.FC<SlideOverProps> = ({
  isOpen,
  onClose,
  title,
  subtitle,
  children,
  headerBackground = 'bg-gradient-to-r from-primary-600 to-primary-800',
  headerTextColor = 'text-white'
}) => {
  return (
    <>
      {/* Backdrop */}
      {isOpen && (
        <div className="fixed inset-0 bg-black bg-opacity-50 transition-opacity z-40" 
             onClick={onClose}
        />
      )}

      {/* Panel */}
      <div
        className={clsx(
          'fixed inset-y-0 right-0 w-[720px] bg-white dark:bg-neutral-900 shadow-xl z-50 transform transition-transform duration-300',
          isOpen ? 'translate-x-0' : 'translate-x-full'
        )}
      >
        {/* Header */}
        <div className={clsx('px-6 py-4', headerBackground)}>
          <div className="flex justify-between items-start">
            <div>
              <h2 className={clsx('text-2xl font-bold', headerTextColor)}>
                {title}
              </h2>
              {subtitle && (
                <p className={clsx('mt-1 text-sm opacity-90', headerTextColor)}>
                  {subtitle}
                </p>
              )}
            </div>
            <button
              onClick={onClose}
              className={clsx(
                'p-1 rounded-full transition-colors',
                headerTextColor,
                'hover:bg-white hover:bg-opacity-10'
              )}
              aria-label="Close panel"
            >
              <X className="w-6 h-6" />
            </button>
          </div>
        </div>

        {/* Content */}
        <div className="h-[calc(100vh-4rem)] overflow-y-auto">
          {children}
        </div>
      </div>
    </>
  );
};