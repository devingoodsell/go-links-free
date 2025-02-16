export interface LinkStats {
  dailyCount: number;
  weeklyCount: number;
  totalCount: number;
}

export interface Link {
  id: number;
  alias: string;
  destinationUrl: string;
  createdAt: string;  // Format: "2024-02-15T19:15:26.788045Z"
  updatedAt?: string;
  expiresAt?: string;
  isActive: boolean;
  stats?: LinkStats;
}

export interface LinkFormData {
  alias: string;
  destinationUrl: string;
  expiresAt?: string;
  isActive: boolean;
}

export interface CreateLinkData {
  destinationUrl: string;
  alias: string;
} 