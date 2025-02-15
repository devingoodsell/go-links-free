import useSWRMutation from 'swr/mutation';
import { useSWRConfig } from 'swr';
import { api } from '../utils/api';
import { useState } from 'react';

async function deleteRequest(url: string, { arg: id }: { arg: number }) {
  await api.delete(`/links/${id}`);
}

export const useDeleteLink = () => {
  const { mutate: globalMutate } = useSWRConfig();
  const [isLoading, setIsLoading] = useState(false);
  const { trigger } = useSWRMutation('/links', deleteRequest);

  const deleteLink = async (id: number) => {
    setIsLoading(true);
    try {
      await trigger(id);
      await globalMutate('/links');
    } finally {
      setIsLoading(false);
    }
  };

  return {
    deleteLink,
    isLoading
  };
}; 