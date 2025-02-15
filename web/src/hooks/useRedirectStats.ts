import useSWR from 'swr';
import { api } from '../utils/api';

interface RedirectStat {
  timestamp: string;
  value: number;
}

export const useRedirectStats = (period: 'day' | 'week' | 'month' = 'day') => {
  return useSWR<RedirectStat[]>(
    `/stats/redirects?period=${period}`,
    async () => {
      const { data } = await api.get(`/stats/redirects?period=${period}`);
      return data;
    },
    {
      revalidateOnFocus: false,
      revalidateOnReconnect: false,
      dedupingInterval: 60000,
    }
  );
}; 