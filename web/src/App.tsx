import React from 'react';
import { Routes, Route } from 'react-router-dom';
import { AppLayout } from './components/layout/AppShell';
import { LinksPage } from './pages/LinksPage';
import { UsersPage } from './pages/UsersPage';

const App: React.FC = () => {
  return (
    <AppLayout>
      <Routes>
        <Route path="/" element={<LinksPage />} />
        <Route path="/links" element={<LinksPage />} />
        <Route path="/users" element={<UsersPage />} />
      </Routes>
    </AppLayout>
  );
};

export default App; 