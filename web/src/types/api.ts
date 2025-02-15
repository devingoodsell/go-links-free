export interface ListResponse<T> {
  items: T[];
  totalCount: number;
  hasMore: boolean;
}

export interface ApiError {
  message: string;
  code: string;
} 