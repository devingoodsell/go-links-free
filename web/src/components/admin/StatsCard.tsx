import React from 'react';
import {
  Card,
  CardContent,
  Typography,
  Icon,
  Box,
  useTheme,
  SxProps,
  Theme
} from '@mui/material';

interface StatsCardProps {
  title: string;
  value: number;
  icon: React.ReactNode;
  sx?: SxProps<Theme>;
}

export const StatsCard: React.FC<StatsCardProps> = ({ title, value, icon, sx }) => {
  const theme = useTheme();

  return (
    <Card sx={sx}>
      <CardContent>
        <Box display="flex" alignItems="center" mb={1}>
          <Icon component="span" sx={{ mr: 1 }}>{icon}</Icon>
          <Typography variant="h6" component="div">
            {title}
          </Typography>
        </Box>
        <Typography variant="h4" component="div" sx={{ textAlign: 'center' }}>
          {value.toLocaleString()}
        </Typography>
      </CardContent>
    </Card>
  );
}; 