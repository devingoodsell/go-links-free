import React, { createContext, useState, useEffect } from 'react';
import { api } from '../utils/api';
import type { User } from '../types/user';

interface AuthContextType {
  user: User | null;
  isLoading: boolean;
  error: Error | null;
  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  register: (email: string, password: string) => Promise<void>;
}

export const AuthContext = createContext<AuthContextType | null>(null);

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true); // Start with loading true
  const [error, setError] = useState<Error | null>(null);

  // Function to restore session
  const restoreSession = async () => {
    const token = localStorage.getItem('token');
    if (!token) {
      setIsLoading(false);
      return;
    }

    try {
      // Set the token in axios headers
      api.defaults.headers.common['Authorization'] = `Bearer ${token}`;
      
      // Try to get user data
      const response = await api.get('/api/auth/me');
      setUser(response.data);
    } catch (err) {
      // If token is invalid, clear it
      localStorage.removeItem('token');
      delete api.defaults.headers.common['Authorization'];
      setError(err as Error);
    } finally {
      setIsLoading(false);
    }
  };

  // Try to restore session on mount
  useEffect(() => {
    restoreSession();
  }, []);

  const login = async (email: string, password: string) => {
    setIsLoading(true);
    try {
      const response = await api.post('/api/auth/login', { email, password });
      const { token, user } = response.data;
      
      // Store token
      localStorage.setItem('token', token);
      
      // Set user in state
      setUser(user);
      
      // Update axios default headers
      api.defaults.headers.common['Authorization'] = `Bearer ${token}`;
    } catch (err) {
      console.error('Login error:', err);
      setError(err as Error);
      throw err;
    } finally {
      setIsLoading(false);
    }
  };

  const logout = async () => {
    setIsLoading(true);
    try {
      await api.post('/api/auth/logout');
    } catch (err) {
      console.error('Logout error:', err);
    } finally {
      setUser(null);
      localStorage.removeItem('token');
      delete api.defaults.headers.common['Authorization'];
      setIsLoading(false);
    }
  };

  const register = async (email: string, password: string) => {
    setIsLoading(true);
    try {
      const response = await api.post('/api/auth/register', { email, password });
      const { token, user } = response.data;
      
      // Store token
      localStorage.setItem('token', token);
      
      // Set user
      setUser(user);
      
      // Update axios default headers
      api.defaults.headers.common['Authorization'] = `Bearer ${token}`;
    } catch (err) {
      setError(err as Error);
      throw err;
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <AuthContext.Provider value={{ 
      user, 
      isLoading, 
      error, 
      login, 
      logout,
      register 
    }}>
      {children}
    </AuthContext.Provider>
  );
}; 