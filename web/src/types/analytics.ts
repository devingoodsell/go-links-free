export interface SystemStats {
  dailyActiveUsers: number;
  monthlyActiveUsers: number;
  totalLinks: number;
  activeLinks: number;
  expiredLinks: number;
  totalRedirects: number;
  status2xx: number;
  status3xx: number;
  status4xx: number;
  status5xx: number;
  lastUpdated: string;
}

export interface TimeSeriesData {
  timestamp: string;
  value: number;
}

export interface PeakUsageStats {
  hourlyStats: {
    hour: number;
    redirects: number;
    uniqueUsers: number;
  }[];
  peakHour: number;
  peakRedirects: number;
  date: string;
} 