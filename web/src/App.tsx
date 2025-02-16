import React, { useEffect } from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import { AppLayout } from './components/layout/AppShell';
import { HomePage } from './pages/HomePage';
import { LinksPage } from './pages/LinksPage';
import { ProfilePage } from './pages/ProfilePage';
import { LoginPage } from './pages/auth/LoginPage';
import { RegisterPage } from './pages/auth/RegisterPage';
import { AdminUsersPage } from './pages/admin/UsersPage';
import { useAuth } from './hooks/useAuth';
import { api } from './utils/api';
import { Box, CircularProgress } from '@mui/material';

const App: React.FC = () => {
  const { user, isLoading } = useAuth();

  useEffect(() => {
    // Test API connection
    api.get('/api/health')
      .then(response => {
        // Health check successful
      })
      .catch(error => {
        console.error('API Connection failed:', error);
      });
  }, []);

  if (isLoading) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh' }}>
        <CircularProgress />
      </Box>
    );
  }

  return (
    <AppLayout>
      <Routes>
        {/* Public routes */}
        <Route path="/" element={<HomePage />} />
        <Route path="/login" element={<LoginPage />} />
        <Route path="/register" element={<RegisterPage />} />
        
        {/* Protected routes */}
        <Route 
          path="/links" 
          element={user ? <LinksPage /> : <Navigate to="/login" replace />} 
        />
        <Route 
          path="/profile" 
          element={user ? <ProfilePage /> : <Navigate to="/login" replace />} 
        />

        {/* Admin routes */}
        <Route 
          path="/admin/users" 
          element={
            user?.isAdmin ? <AdminUsersPage /> : <Navigate to="/" replace />
          } 
        />

        {/* Catch-all route */}
        <Route path="*" element={
          <div style={{ padding: '2rem' }}>
            <h1>404: Page Not Found</h1>
            <p>The page you're looking for doesn't exist.</p>
          </div>
        } />
      </Routes>
    </AppLayout>
  );
};

export default App; 