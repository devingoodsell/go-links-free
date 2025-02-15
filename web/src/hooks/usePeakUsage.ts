import useSWR from 'swr';
import { api } from '../utils/api';

interface HourlyStat {
  hour: number;
  redirects: number;
}

interface PeakUsage {
  hourlyStats: HourlyStat[];
  peakHour: number;
  peakRedirects: number;
}

export const usePeakUsage = () => {
  return useSWR<PeakUsage>(
    '/stats/peak-usage',
    async () => {
      const { data } = await api.get('/stats/peak-usage');
      return data;
    },
    {
      revalidateOnFocus: false,
      revalidateOnReconnect: false,
      dedupingInterval: 60000,
    }
  );
}; 