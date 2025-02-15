import useSWRMutation from 'swr/mutation';
import { useSWRConfig } from 'swr';
import type { Link } from '../types/link';
import { api } from '../utils/api';
import { useState } from 'react';

interface UseUpdateLinkReturn {
  mutate: (link: Partial<Link> & { id: number }) => Promise<Link>;
  isLoading: boolean;
}

async function updateRequest(url: string, { arg }: { arg: Partial<Link> & { id: number } }) {
  const { data } = await api.put(`/links/${arg.id}`, arg);
  return data;
}

export const useUpdateLink = (): UseUpdateLinkReturn => {
  const { mutate: globalMutate } = useSWRConfig();
  const [isLoading, setIsLoading] = useState(false);
  const { trigger } = useSWRMutation('/links', updateRequest);

  const mutate = async (link: Partial<Link> & { id: number }) => {
    setIsLoading(true);
    try {
      const result = await trigger(link);
      await globalMutate('/links');
      return result;
    } finally {
      setIsLoading(false);
    }
  };

  return {
    mutate,
    isLoading
  };
}; 