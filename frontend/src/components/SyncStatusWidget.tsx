import React from 'react';
import useSyncStatus from '../hooks/useSyncStatus';

export const SyncStatusWidget: React.FC = () => {
  const { data, isLoading, error } = useSyncStatus();
  const d: any = data || {};

  if (isLoading) return <div className="text-sm text-gray-500">Loading sync status...</div>;
  if (error) return <div className="text-sm text-red-500">Failed to load sync status</div>;

  return (
    <div className="flex items-center space-x-4">
      <div className="text-sm text-gray-600">Last Sync:</div>
      <div className="text-sm font-medium text-gray-900">{d.last_sync_end ?? d.last_sync_start ?? 'Never'}</div>
      <div className="px-2 py-1 rounded-md bg-gray-100 text-xs text-gray-700">{d.status ?? 'unknown'}</div>
      <div className="text-sm text-gray-600">Processed: <span className="font-semibold">{d.processed_records ?? 0}</span></div>
      {d.error_count > 0 && (
        <div className="text-sm text-red-600">Errors: {d.error_count}</div>
      )}
    </div>
  );
};

export default SyncStatusWidget;
