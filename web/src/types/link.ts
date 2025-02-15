export interface LinkStats {
  dailyCount: number;
  weeklyCount: number;
  totalCount: number;
  lastAccessedAt: string;
}

export interface Link {
  id: number;
  shortLink: string;
  targetUrl: string;
  createdAt: string;
  updatedAt: string;
  isActive: boolean;
  clicks: number;
  createdBy: number;
  expiresAt?: string;
  stats?: {
    dailyCount: number;
    weeklyCount: number;
    totalCount: number;
    lastAccessedAt: string;
  };
} 