import React, { useCallback } from 'react';
import useSWR, { useSWRConfig, type KeyedMutator, type SWRResponse } from 'swr';
import type { AxiosResponse } from 'axios';
import { api } from '../utils/api';
import type { Link } from '../types/link';
import type { ListResponse } from '../types/api';

interface UseLinksOptions {
  search?: string;
  status?: 'active' | 'expired';
  sortBy?: 'created_desc' | 'created_asc' | 'clicks_desc';
  page?: number;
  pageSize?: number;
}

interface UseLinksReturn {
  links: Link[];
  totalCount: number;
  isLoading: boolean;
  error: Error | undefined;
  createLink: (linkData: { url: string; alias?: string }) => Promise<Link>;
  updateLink: (alias: string, updates: Partial<Link>) => Promise<Link>;
  deleteLink: (alias: string) => Promise<void>;
  refetch: KeyedMutator<ListResponse<Link>>;
  sortBy?: string;
  sortDirection?: 'asc' | 'desc';
  handleSortChange?: (field: string, direction: 'asc' | 'desc') => Promise<void>;
}

interface CreateLinkData {
  destinationUrl: string;
  alias?: string;
}

export const useLinks = (options: UseLinksOptions = {}) => {
  const { data, error, mutate } = useSWR<ListResponse<Link>>(
    ['/api/links', options],
    async () => {
      const { data } = await api.get('/api/links', { 
        params: {
          page: options.page,
          pageSize: options.pageSize
        }
      });
      // If the response is null, return an empty list response
      return data || { items: [], totalCount: 0 };
    }
  );

  return {
    links: data?.items || [],
    totalCount: data?.totalCount || 0,
    isLoading: typeof data === 'undefined' && !error,
    error,
    createLink: async (linkData: CreateLinkData) => {
      const { data } = await api.post('/api/links', linkData);
      await mutate();
      return data;
    },
    deleteLink: async (id: number) => {
      await api.delete(`/api/links/delete/${id}`);
      await mutate();
    },
    updateLink: async (id: number, updates: { destinationUrl: string }) => {
      await api.put(`/api/links/${id}`, updates);
      await mutate();
    },
    refetch: () => mutate()
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
      } : undefined),
      { revalidate: false }
    );
    
    try {
      const result = await api.put(`/api/links/${link.id}`, link)
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
      } : undefined),
      { revalidate: false }
    );
    
    try {
      const result = await api.delete(`/api/links/${id}`)
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