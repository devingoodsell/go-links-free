import useSWR, { useSWRConfig } from 'swr';
import useSWRMutation from 'swr/mutation';
import type { AxiosResponse } from 'axios';
import { api } from '../utils/api';
import type { User } from '../types/user';
import type { ListResponse } from '../types/api';
import { useState } from 'react';

interface UseUsersOptions {
  page?: number;
  pageSize?: number;
  search?: string;
  role?: 'admin' | 'user';
  sortBy?: 'created_desc' | 'created_asc' | 'last_login_desc';
}

export const useUsers = (options: UseUsersOptions = {}) => {
  const { data, error, mutate } = useSWR<ListResponse<User>>(
    ['/api/admin/users', options],
    async () => {
      const { data } = await api.get('/api/admin/users', { params: options });
      return data;
    }
  );

  return {
    users: data?.items || [],
    totalCount: data?.totalCount || 0,
    isLoading: !error && !data,
    error,
    refetch: () => mutate()
  };
};

// Update User Hook
async function updateUserFetcher(
  url: string,
  { arg }: { arg: Partial<User> & { id: number } }
) {
  const { data } = await api.put(`/users/${arg.id}`, arg);
  return data;
}

interface UseUpdateUserReturn {
  mutateAsync: (user: Partial<User> & { id: number }) => Promise<User>;
  isLoading: boolean;
}

export const useUpdateUser = () => {
  const { mutate: globalMutate } = useSWRConfig();
  const [isLoading, setIsLoading] = useState(false);
  const { trigger } = useSWRMutation('/users', updateUserFetcher);

  const mutateAsync = async (user: Partial<User> & { id: number }) => {
    setIsLoading(true);
    try {
      const result = await trigger(user);
      await globalMutate('/users');
      return result;
    } finally {
      setIsLoading(false);
    }
  };

  return {
    mutateAsync,
    isLoading
  };
};

// Delete User Hook
async function deleteUserFetcher(url: string, { arg: id }: { arg: number }) {
  await api.delete(`/users/${id}`);
}

interface UseDeleteUserReturn {
  deleteUser: (id: number) => Promise<void>;
  isLoading: boolean;
  mutate: (id: number) => Promise<void>;
}

export const useDeleteUser = () => {
  const { mutate: globalMutate } = useSWRConfig();
  const [isLoading, setIsLoading] = useState(false);
  const { trigger } = useSWRMutation('/users', deleteUserFetcher);

  const deleteUser = async (id: number) => {
    setIsLoading(true);
    try {
      await trigger(id);
      await globalMutate('/users');
    } finally {
      setIsLoading(false);
    }
  };

  return {
    deleteUser,
    isLoading
  };
}; 