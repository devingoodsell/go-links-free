import React from 'react';
import { 
  Box,
  Typography,
  Paper,
  Container,
  CircularProgress,
  Button,
} from '@mui/material';
import LogoutIcon from '@mui/icons-material/Logout';
import { useAuth } from '../hooks/useAuth';

export const ProfilePage: React.FC = () => {
  const { user, isLoading, logout } = useAuth();

  const formatDate = (date: string | null) => {
    if (!date) return 'Never';
    console.log('Formatting date:', date);
    return new Date(date).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  if (isLoading) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" minHeight="200px">
        <CircularProgress />
      </Box>
    );
  }

  if (!user) {
    return <Typography>Not logged in</Typography>;
  }

  return (
    <Container maxWidth="md">
      <Box sx={{ mt: 4 }}>
        <Paper sx={{ p: 3 }}>
          <Box display="flex" justifyContent="space-between" alignItems="center" mb={2}>
            <Typography variant="h4" gutterBottom>
              Profile
            </Typography>
            <Button
              variant="outlined"
              color="error"
              startIcon={<LogoutIcon />}
              onClick={logout}
            >
              Logout
            </Button>
          </Box>
          <Typography>
            Email: {user.email}
          </Typography>
          <Typography>
            Member Since: {formatDate(user.createdAt)}
          </Typography>
          <Typography>
            Last Login: {formatDate(user.lastLoginAt)}
          </Typography>
        </Paper>
      </Box>
    </Container>
  );
}; 