import React from 'react';
import { 
  Container, 
  Typography, 
  Box, 
  Button,
  Paper,
  Grid 
} from '@mui/material';
import { Link as RouterLink } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';

export const HomePage: React.FC = () => {
  const { user } = useAuth();

  return (
    <Container maxWidth="lg">
      <Box sx={{ mt: 8, mb: 4 }}>
        <Typography variant="h3" component="h1" gutterBottom>
          Welcome to GoLinks
        </Typography>
        <Typography variant="h5" color="text.secondary" paragraph>
          Create and manage your short links in one place
        </Typography>
        {!user ? (
          <Box sx={{ mt: 4 }}>
            <Button
              component={RouterLink}
              to="/login"
              variant="contained"
              size="large"
              sx={{ mr: 2 }}
            >
              Login
            </Button>
            <Button
              component={RouterLink}
              to="/register"
              variant="outlined"
              size="large"
            >
              Register
            </Button>
          </Box>
        ) : (
          <Grid container spacing={3} sx={{ mt: 2 }}>
            <Grid item xs={12} md={6}>
              <Paper sx={{ p: 3 }}>
                <Typography variant="h6" gutterBottom>
                  Quick Start
                </Typography>
                <Button
                  component={RouterLink}
                  to="/links"
                  variant="contained"
                  fullWidth
                  sx={{ mt: 2 }}
                >
                  Manage Links
                </Button>
              </Paper>
            </Grid>
            <Grid item xs={12} md={6}>
              <Paper sx={{ p: 3 }}>
                <Typography variant="h6" gutterBottom>
                  Account
                </Typography>
                <Button
                  component={RouterLink}
                  to="/profile"
                  variant="outlined"
                  fullWidth
                  sx={{ mt: 2 }}
                >
                  View Profile
                </Button>
              </Paper>
            </Grid>
          </Grid>
        )}
      </Box>
    </Container>
  );
}; 