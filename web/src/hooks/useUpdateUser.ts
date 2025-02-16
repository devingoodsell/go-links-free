import useSWRMutation from 'swr/mutation';
import { useSWRConfig } from 'swr';
import type { User } from '../types/user';
import { api } from '../utils/api';
import { useState } from 'react';

async function updateUserFetcher(url: string, { arg }: { arg: Partial<User> & { id: number } }) {
  const { data } = await api.put(`/api/users/${arg.id}`, arg);
  return data;
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