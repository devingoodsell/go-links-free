import { useState } from 'react';

interface PaginationState {
  page: number;
  pageSize: number;
}

export const usePaginationErrorHandler = (initialState: PaginationState) => {
  const [page, setPage] = useState(initialState.page);
  const [pageSize, setPageSize] = useState(initialState.pageSize);
  const [error, setError] = useState<string | null>(null);

  const handlePageChange = async (newPage: number, callback?: () => Promise<void>) => {
    try {
      setError(null);
      setPage(newPage);
      await callback?.();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Error changing page');
    }
  };

  const handlePageSizeChange = async (newSize: number, callback?: () => Promise<void>) => {
    try {
      setError(null);
      setPageSize(newSize);
      await callback?.();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Error changing page size');
    }
  };

  const clearError = () => setError(null);

  return {
    page,
    pageSize,
    error,
    handlePageChange,
    handlePageSizeChange,
    clearError,
    setPage,
    setPageSize
  };
}; 