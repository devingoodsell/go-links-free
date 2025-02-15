import useSWR, { useSWRConfig } from 'swr';
import type { AxiosResponse } from 'axios';
import { api } from '../utils/api';
import { Link } from '../types/link';
import { useMemo, useCallback } from 'react';
import type { MutatorCallback } from 'swr/_internal';
import type { ListResponse } from '../types/api';

interface UseLinksOptions {
  search?: string;
  status?: 'active' | 'expired';
  sortBy?: 'created_desc' | 'created_asc' | 'clicks_desc';
  page?: number;
  pageSize?: number;
}

interface UseLinksReturn {
  data: ListResponse<Link> | undefined;
  error: Error | undefined;
  isLoading: boolean;
  refetch: () => Promise<ListResponse<Link> | undefined>;
  sortBy: string | undefined;
  sortDirection: 'asc' | 'desc' | undefined;
  handleSortChange: (field: string, direction: 'asc' | 'desc') => Promise<void>;
}

export const useLinks = (options: UseLinksOptions = {}): UseLinksReturn => {
  const { data, error, mutate } = useSWR<ListResponse<Link>>(
    ['/links', options],
    async () => {
      const { data } = await api.get('/links', { params: options });
      return data;
    },
    {
      revalidateOnFocus: false,
      revalidateOnReconnect: false,
      dedupingInterval: 60000,
    }
  );

  return {
    data,
    error,
    isLoading: !data && !error,
    refetch: () => mutate(),
    sortBy: options.sortBy,
    sortDirection: options.sortBy?.endsWith('_desc') ? 'desc' : 'asc',
    handleSortChange: async (field: string, direction: 'asc' | 'desc') => {
      const newSortBy = `${field}_${direction}` as UseLinksOptions['sortBy'];
      await mutate(
        async (currentData) => {
          const { data } = await api.get('/links', { 
            params: { ...options, sortBy: newSortBy }
          });
          return data;
        },
        { revalidate: false }
      );
    }
  };
};

export const useUpdateLink = () => {
  const { mutate } = useSWRConfig();
  
  const updateLink = useCallback(async (link: Link) => {
    // Get the current data
    const currentData = await mutate(
      (key: any) => typeof key === 'object' && key.url === '/api/links'
    );
    
    // Optimistically update the UI
    await mutate(
      (key: any) => typeof key === 'object' && key.url === '/api/links',
      ((currentData?: ListResponse<Link>) => currentData ? {
        ...currentData,
        items: currentData.items.map(item => item.id === link.id ? { ...item, ...link } : item)
      } : undefined) as MutatorCallback<ListResponse<Link>>,
      { revalidate: false }
    );
    
    try {
      const result = await api.put(`/api/admin/links/${link.id}`, link)
        .then((res: AxiosResponse<Link>) => res.data);
      return result;
    } catch (error) {
      // Revert to the original data on error
      await mutate(
        (key: any) => typeof key === 'object' && key.url === '/api/links',
        currentData,
        { revalidate: true }
      );
      throw error;
    }
  }, [mutate]);
  
  return { updateLink };
};

export const useDeleteLink = () => {
  const { mutate } = useSWRConfig();
  
  const deleteLink = useCallback(async (id: number) => {
    // Get the current data
    const currentData = await mutate(
      (key: any) => typeof key === 'object' && key.url === '/api/links'
    );
    
    // Optimistically update the UI
    await mutate(
      (key: any) => typeof key === 'object' && key.url === '/api/links',
      ((currentData?: ListResponse<Link>) => currentData ? {
        ...currentData,
        items: currentData.items.filter(item => item.id !== id),
        totalCount: currentData.totalCount - 1
      } : undefined) as MutatorCallback<ListResponse<Link>>,
      { revalidate: false }
    );
    
    try {
      const result = await api.delete(`/api/admin/links/${id}`)
        .then((res: AxiosResponse<void>) => res.data);
      return result;
    } catch (error) {
      // Revert to the original data on error
      await mutate(
        (key: any) => typeof key === 'object' && key.url === '/api/links',
        currentData,
        { revalidate: true }
      );
      throw error;
    }
  }, [mutate]);
  
  return { deleteLink };
};

export const useBulkDeleteLinks = () => {
  const { mutate } = useSWRConfig();
  
  const bulkDeleteLinks = useCallback(async (ids: number[]) => {
    const result = await api.post('/api/admin/links/bulk-delete', { ids })
      .then((res: AxiosResponse<void>) => res.data);
    await mutate(
      (key: any) => typeof key === 'object' && key.url === '/api/links',
      undefined,
      { revalidate: true }
    );
    return result;
  }, [mutate]);
  
  return { bulkDeleteLinks };
};

export const useBulkUpdateLinkStatus = () => {
  const { mutate } = useSWRConfig();
  
  const bulkUpdateStatus = useCallback(async (data: { ids: number[]; isActive: boolean }) => {
    const result = await api.post('/api/admin/links/bulk-status', data)
      .then((res: AxiosResponse<void>) => res.data);
    await mutate(
      (key: any) => typeof key === 'object' && key.url === '/api/links',
      undefined,
      { revalidate: true }
    );
    return result;
  }, [mutate]);
  
  return { bulkUpdateStatus };
}; 