import React from 'react';
import {
  Card,
  CardContent,
  Typography,
  Box,
  Divider,
} from '@mui/material';
import { User } from '../types/user';

interface UserProfileProps {
  user: User;
}

export const UserProfile: React.FC<UserProfileProps> = ({ user }) => {
  const formatDate = (date: string | null) => {
    if (!date) return 'Never';
    return new Date(date).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  return (
    <Card>
      <CardContent>
        <Typography variant="h5" gutterBottom>
          Profile
        </Typography>
        <Divider sx={{ mb: 2 }} />
        
        <Box sx={{ mb: 2 }}>
          <Typography color="textSecondary" gutterBottom>
            Email
          </Typography>
          <Typography variant="body1">
            {user.email}
          </Typography>
        </Box>

        <Box sx={{ mb: 2 }}>
          <Typography color="textSecondary" gutterBottom>
            Member Since
          </Typography>
          <Typography variant="body1">
            {formatDate(user.createdAt)}
          </Typography>
        </Box>

        <Box>
          <Typography color="textSecondary" gutterBottom>
            Last Login
          </Typography>
          <Typography variant="body1">
            {formatDate(user.lastLoginAt)}
          </Typography>
        </Box>
      </CardContent>
    </Card>
  );
}; 