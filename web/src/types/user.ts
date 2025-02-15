export interface UserStats {
  linkCount: number;
  totalClicks: number;
  activeLinks: number;
  expiredLinks: number;
  linksCreated30d: number;
}

export interface User {
  id: number;
  email: string;
  role: 'admin' | 'user';
  isActive: boolean;
  lastLoginAt: string | null;
  createdAt: string;
  updatedAt: string;
  isAdmin: boolean;
  stats?: {
    linkCount: number;
    totalClicks: number;
    activeLinks: number;
    expiredLinks: number;
    linksCreated30d: number;
  };
} 