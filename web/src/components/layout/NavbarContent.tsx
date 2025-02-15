import React from 'react';
import { 
  Box,
  List,
  ListItem,
  ListItemIcon,
  ListItemText,
  ListItemButton,
} from '@mui/material';
import { Link, useLocation } from 'react-router-dom';
import LinkIcon from '@mui/icons-material/Link';
import PeopleIcon from '@mui/icons-material/People';
import DashboardIcon from '@mui/icons-material/Dashboard';

const navItems = [
  { label: 'Dashboard', icon: <DashboardIcon />, to: '/admin' },
  { label: 'Links', icon: <LinkIcon />, to: '/admin/links' },
  { label: 'Users', icon: <PeopleIcon />, to: '/admin/users' },
];

export const NavbarContent = () => {
  const location = useLocation();

  return (
    <Box sx={{ mt: 8 }}>
      <List>
        {navItems.map((item) => (
          <ListItem key={item.to} disablePadding>
            <ListItemButton
              component={Link}
              to={item.to}
              selected={location.pathname === item.to}
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