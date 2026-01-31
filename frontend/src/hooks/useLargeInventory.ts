import { useQuery } from '@tanstack/react-query';
import { Resource } from '../types';

const API_BASE = process.env.REACT_APP_API_BASE || 'http://localhost:8080/api/cost';

export const useLargeInventory = (count: number = 10000) => {
  return useQuery({
    queryKey: ['large-inventory', count],
    queryFn: async (): Promise<Resource[]> => {
      const response = await fetch(`${API_BASE}/resources/bulk?count=${count}`);
      if (!response.ok) throw new Error('Failed to fetch large inventory');
      
      // Handle both streaming and regular JSON responses
      const contentType = response.headers.get('content-type');
      if (contentType?.includes('application/json')) {
        return response.json();
      }
      
      // Handle streaming response
      const reader = response.body?.getReader();
      const decoder = new TextDecoder();
      const resources: Resource[] = [];
      
      if (reader) {
        let buffer = '';
        while (true) {
          const { done, value } = await reader.read();
          if (done) break;
          
          buffer += decoder.decode(value, { stream: true });
          const lines = buffer.split('\n');
          buffer = lines.pop() || '';
          
          for (const line of lines) {
            if (line.trim()) {
              try {
                const resource = JSON.parse(line);
                resources.push(resource);
              } catch (e) {
                // Skip invalid JSON
              }
            }
          }
        }
      }
      
      return resources;
    },
    staleTime: 5 * 60 * 1000, // 5 minutes
    enabled: count > 0
  });
};

export const useStreamingInventory = (count: number = 10000) => {
  return useQuery({
    queryKey: ['streaming-inventory', count],
    queryFn: async (): Promise<Resource[]> => {
      const response = await fetch(`${API_BASE}/resources/live?limit=${count}`);
      if (!response.ok) throw new Error('Failed to fetch streaming inventory');
      return response.json();
    },
    staleTime: 30 * 1000, // 30 seconds for live data
    refetchInterval: 60 * 1000 // Refresh every minute
  });
};