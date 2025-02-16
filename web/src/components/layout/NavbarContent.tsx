import React from 'react';
import { 
  Box,
  List,
  ListItem,
  ListItemIcon,
  ListItemText,
  ListItemButton,
} from '@mui/material';
import { Link as RouterLink, useLocation } from 'react-router-dom';
import HomeIcon from '@mui/icons-material/Home';
import LinkIcon from '@mui/icons-material/Link';
import PersonIcon from '@mui/icons-material/Person';
import AdminPanelSettingsIcon from '@mui/icons-material/AdminPanelSettings';
import { useAuth } from '../../hooks/useAuth';

export const NavbarContent = () => {
  const location = useLocation();
  const { user } = useAuth();

  const navItems = [
    { path: '/', label: 'Home', icon: <HomeIcon /> },
    ...(user ? [
      { path: '/links', label: 'Links', icon: <LinkIcon /> },
      { path: '/profile', label: 'Profile', icon: <PersonIcon /> },
    ] : []),
    ...(user?.isAdmin ? [
      { path: '/admin/users', label: 'User Management', icon: <AdminPanelSettingsIcon /> },
    ] : []),
  ];

  return (
    <Box sx={{ mt: 8 }}>
      <List>
        {navItems.map((item) => (
          <ListItem key={item.path} disablePadding>
            <ListItemButton
              component={RouterLink}
              to={item.path}
              selected={location.pathname === item.path}
            >
              <ListItemIcon>{item.icon}</ListItemIcon>
              <ListItemText primary={item.label} />
            </ListItemButton>
          </ListItem>
        ))}
      </List>
    </Box>
  );
}; 