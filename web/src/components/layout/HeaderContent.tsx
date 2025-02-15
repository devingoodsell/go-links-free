import React from 'react';
import { 
  Box,
  Typography,
  Toolbar,
} from '@mui/material';

export const HeaderContent = () => {
  return (
    <Toolbar>
      <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
        <Typography variant="h6" noWrap component="div">
          Go Links Admin
        </Typography>
      </Box>
    </Toolbar>
  );
}; 