import { useQuery } from '@tanstack/react-query';
import { Resource } from '../types';

const API_BASE = 'http://localhost:8081/api/live';

export const useLiveInstances = () => {
  return useQuery({
    queryKey: ['live-instances'],
    queryFn: async (): Promise<Resource[]> => {
      const response = await fetch(`${API_BASE}/instances`);
      if (!response.ok) throw new Error('Failed to fetch live instances');
      return response.json();
    },
    staleTime: 30 * 1000, // 30 seconds
    refetchInterval: 60 * 1000, // Refresh every minute
    retry: 3
  });
};

export const useInstanceCost = (instanceId: string) => {
  return useQuery({
    queryKey: ['instance-cost', instanceId],
    queryFn: async () => {
      const response = await fetch(`${API_BASE}/instances/${instanceId}/cost`);
      if (!response.ok) throw new Error('Failed to fetch instance cost');
      return response.json();
    },
    enabled: !!instanceId,
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
};

export const useInstanceMetrics = (instanceId: string) => {
  return useQuery({
    queryKey: ['instance-metrics', instanceId],
    queryFn: async () => {
      const response = await fetch(`${API_BASE}/instances/${instanceId}/metrics`);
      if (!response.ok) throw new Error('Failed to fetch instance metrics');
      return response.json();
    },
    enabled: !!instanceId,
    staleTime: 30 * 1000, // 30 seconds
    refetchInterval: 60 * 1000, // Refresh every minute
  });
};

export const useLiveCostSummary = () => {
  return useQuery({
    queryKey: ['live-cost-summary'],
    queryFn: async () => {
      const response = await fetch(`${API_BASE}/cost/summary`);
      if (!response.ok) throw new Error('Failed to fetch live cost summary');
      return response.json();
    },
    staleTime: 2 * 60 * 1000, // 2 minutes
    refetchInterval: 5 * 60 * 1000, // Refresh every 5 minutes
  });
};