import { useContext, useCallback } from 'react';
import { AuthContext } from '../contexts/AuthContext';
import { api } from '../utils/api';
import { useNavigate } from 'react-router-dom';
import useSWR from 'swr';
import type { User } from '../types/user';

export const useAuth = () => {
  const context = useContext(AuthContext);
  const navigate = useNavigate();
  const { mutate } = useSWR('/api/auth/me');

  const { data: user, error, isLoading } = useSWR<User>(
    // Only fetch if we have a token
    localStorage.getItem('token') ? '/api/auth/me' : null,
    async () => {
      console.log('Fetching user data...');
      try {
        const { data } = await api.get('/api/auth/me');
        console.log('User data received:', data);
        return data;
      } catch (error) {
        console.error('Error fetching user data:', error);
        localStorage.removeItem('token');
        throw error;
      }
    },
    {
      revalidateOnFocus: false,
      shouldRetryOnError: false,
      onSuccess: (data) => console.log('SWR success:', data),
      onError: (err) => console.error('SWR error:', err),
    }
  );

  const login = useCallback(async (email: string, password: string) => {
    try {
      const { data } = await api.post('/api/auth/login', { email, password });
      localStorage.setItem('token', data.token);
      await mutate(); // Refresh user data
      return data;
    } catch (error: any) {
      throw new Error(error.response?.data?.error || 'Login failed');
    }
  }, [mutate]);

  const register = useCallback(async (email: string, password: string) => {
    try {
      const { data } = await api.post('/api/auth/register', { email, password });
      localStorage.setItem('token', data.token);
      await mutate(); // Refresh user data
      return data;
    } catch (error: any) {
      throw new Error(error.response?.data?.error || 'Registration failed');
    }
  }, [mutate]);

  const logout = useCallback(() => {
    localStorage.removeItem('token');
    mutate(null); // Clear the cached user data
    navigate('/login');
  }, [navigate, mutate]);

  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }

  return {
    user,
    isLoading,
    error,
    login,
    register,
    logout,
  };
}; 