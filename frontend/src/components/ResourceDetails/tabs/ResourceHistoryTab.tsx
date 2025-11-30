import React, { useState, useEffect } from 'react';
import { Clock, ArrowDownCircle, ArrowUpCircle } from 'lucide-react';
import api from '../../../services/api';

interface ResourceHistoryTabProps {
  resourceId: string;
}

interface HistoryEvent {
  id: string;
  timestamp: string;
  event_type: string;
  description: string;
  user: string;
  changes: {
    field: string;
    old_value: string;
    new_value: string;
  }[];
}

const ResourceHistoryTab: React.FC<ResourceHistoryTabProps> = ({ resourceId }) => {
  const [history, setHistory] = useState<HistoryEvent[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(true);

  useEffect(() => {
    const fetchHistory = async () => {
      try {
        setLoading(true);
        const response = await api.get(`/api/v1/resources/${resourceId}/history`, {
          params: {
            page,
            limit: 10,
          },
        });
        
        if (page === 1) {
          setHistory(response.data.events);
        } else {
          setHistory(prev => [...prev, ...response.data.events]);
        }
        
        setHasMore(response.data.has_more);
      } catch (err) {
        setError('Failed to load resource history');
        console.error(err);
      } finally {
        setLoading(false);
      }
    };

    fetchHistory();
  }, [resourceId, page]);

  const loadMore = () => {
    setPage(prev => prev + 1);
  };

  return (
    <div className="space-y-4">
      {history.map((event, index) => (
        <div
          key={event.id}
          className={`relative pb-8 ${
            index === history.length - 1 ? '' : 'border-l border-gray-200'
          }`}
        >
          <div className="relative flex items-start group">
            <span className="h-9 flex items-center">
              <span className="relative z-10 w-8 h-8 flex items-center justify-center bg-white border-2 border-gray-300 rounded-full">
                {event.event_type === 'cost_increase' ? (
                  <ArrowUpCircle className="h-5 w-5 text-red-500" />
                ) : event.event_type === 'cost_decrease' ? (
                  <ArrowDownCircle className="h-5 w-5 text-green-500" />
                ) : (
                  <Clock className="h-5 w-5 text-blue-500" />
                )}
              </span>
            </span>
            <div className="min-w-0 flex-1 ml-4">
              <div className="text-sm font-medium text-gray-900">
                {event.description}
              </div>
              <div className="mt-1 text-sm text-gray-500">
                <span>{new Date(event.timestamp).toLocaleString()}</span>
                <span className="mx-2">•</span>
                <span>{event.user}</span>
              </div>
              {event.changes.length > 0 && (
                <div className="mt-2">
                  <div className="flex flex-col space-y-2">
                    {event.changes.map((change, idx) => (
                      <div
                        key={idx}
                        className="text-sm bg-gray-50 rounded-md p-2"
                      >
                        <span className="font-medium">{change.field}:</span>
                        <div className="grid grid-cols-2 gap-4 mt-1">
                          <div className="text-red-500 line-through">
                            {change.old_value}
                          </div>
                          <div className="text-green-500">{change.new_value}</div>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      ))}

      {hasMore && (
        <div className="flex justify-center">
          <button
            onClick={loadMore}
            disabled={loading}
            className="text-blue-600 hover:text-blue-800 font-medium"
          >
            {loading ? 'Loading...' : 'Load More'}
          </button>
        </div>
      )}
    </div>
  );
};

export default ResourceHistoryTab;