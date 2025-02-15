import { useQuery } from '@tanstack/react-query';
import { SystemStats } from '../types/analytics';
import { api } from '../utils/api';

export const useSystemStats = () => {
  return useQuery({
    queryKey: ['systemStats'],
    queryFn: async (): Promise<SystemStats> => {
      const { data } = await api.get('/analytics/system');
      return data;
    }
  });
};

export const useRedirectStats = () => {
  return useQuery({
    queryKey: ['redirectStats'],
    queryFn: () => api.get('/api/admin/stats/redirects').then(res => res.data),
    refetchInterval: 60000,
  });
};

export const usePeakUsage = () => {
  return useQuery({
    queryKey: ['peakUsage'],
    queryFn: () => api.get('/api/admin/stats/peak-usage').then(res => res.data),
    refetchInterval: 300000, // Refetch every 5 minutes
  });
}; 