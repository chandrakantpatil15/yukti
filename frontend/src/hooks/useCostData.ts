import { useQuery } from 'react-query';
import { costApi } from '../services/api';

export const useCostSummary = () => {
  return useQuery('costSummary', costApi.getSummary, {
    refetchInterval: 30000, // Refetch every 30 seconds
    staleTime: 10000,
  });
};

export const useRecommendations = () => {
  return useQuery('recommendations', costApi.getRecommendations, {
    refetchInterval: 60000, // Refetch every minute
    staleTime: 30000,
  });
};

export const useResources = () => {
  return useQuery('resources', costApi.getResources, {
    refetchInterval: 30000,
    staleTime: 10000,
  });
};