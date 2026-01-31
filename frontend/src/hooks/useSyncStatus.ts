import { useQuery } from 'react-query';
import api from '../services/api';

export const useSyncStatus = () => {
  return useQuery('syncStatus', async () => {
    const res = await api.get('/internal/sync/status');
    return res;
  }, {
    refetchInterval: 60 * 1000, // refresh every 60s on dashboard
    retry: 1,
  });
};

export default useSyncStatus;
