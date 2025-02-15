import useSWRMutation from 'swr/mutation';
import { useSWRConfig } from 'swr';
import type { User } from '../types/user';
import { api } from '../utils/api';
import { useState } from 'react';

async function createUserFetcher(url: string, { arg: user }: { arg: Partial<User> }) {
  const { data } = await api.post('/users', user);
  return data;
}

export const useCreateUser = () => {
  const { mutate: globalMutate } = useSWRConfig();
  const [isLoading, setIsLoading] = useState(false);
  const { trigger } = useSWRMutation('/users', createUserFetcher);

  const createUser = async (user: Partial<User>) => {
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
    createUser,
    isLoading
  };
}; 