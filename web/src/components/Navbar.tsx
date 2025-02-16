import {
  AppBar,
  Toolbar,
  Typography,
  Button,
  IconButton,
  Avatar,
  Menu,
  MenuItem,
  Box,
  CircularProgress,
} from '@mui/material';
import { Link as RouterLink } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import LogoutIcon from '@mui/icons-material/Logout';
import { useState } from 'react';

export const Navbar = () => {
  const { user, logout, isLoading } = useAuth();
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  
  console.log('Navbar render:', { 
    user, 
    isLoading, 
    token: localStorage.getItem('token'),
    hasUser: !!user,
  });

  const handleMenu = (event: React.MouseEvent<HTMLElement>) => {
    setAnchorEl(event.currentTarget);
  };

  const handleClose = () => {
    setAnchorEl(null);
  };

  const handleLogout = () => {
    handleClose();
    logout();
  };

  return (
    <AppBar position="static">
      <Toolbar>
        <Typography variant="h6" component="div" sx={{ flexGrow: 1 }}>
          Go Links
        </Typography>
        <Box>
          {isLoading && <CircularProgress size={24} color="inherit" />}
          {!isLoading && user && (
            <IconButton
              color="inherit"
              onClick={logout}
              title="Logout"
            >
              <LogoutIcon />
            </IconButton>
          )}
          {!isLoading && !user && (
            <Button color="inherit" component={RouterLink} to="/login">
              Login
            </Button>
          )}
        </Box>
      </Toolbar>
    </AppBar>
  );
}; 