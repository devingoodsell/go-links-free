import React from 'react';
import {
  Box,
  Container,
  Grid,
  Paper,
  Typography,
  useTheme,
} from '@mui/material';
import {
  LineChart,
  BarChart,
  PieChart,
} from '@mui/x-charts';
import { StatsCard } from '../../components/admin/StatsCard';
import { useSystemStats } from '../../hooks/useSystemStats';
import { useRedirectStats } from '../../hooks/useRedirectStats';
import { usePeakUsage } from '../../hooks/usePeakUsage';

export const AdminDashboard: React.FC = () => {
  const theme = useTheme();
  const { data: systemStats, isLoading: statsLoading } = useSystemStats();
  const { data: redirectStats, isLoading: redirectsLoading } = useRedirectStats();
  const { data: peakUsage, isLoading: peakLoading } = usePeakUsage();

  if (statsLoading || redirectsLoading || peakLoading) {
    return <div>Loading...</div>;
  }

  return (
    <Container maxWidth="lg" sx={{ mt: 4, mb: 4 }}>
      <Grid container spacing={3}>
        {/* Overview Stats */}
        <Grid item xs={12} md={3}>
          <StatsCard
            title="Total Links"
            value={systemStats?.totalLinks || 0}
            icon="link"
          />
        </Grid>
        <Grid item xs={12} md={3}>
          <StatsCard
            title="Active Links"
            value={systemStats?.activeLinks || 0}
            icon="check_circle"
          />
        </Grid>
        <Grid item xs={12} md={3}>
          <StatsCard
            title="Daily Active Users"
            value={systemStats?.dailyActiveUsers || 0}
            icon="people"
          />
        </Grid>
        <Grid item xs={12} md={3}>
          <StatsCard
            title="Total Redirects"
            value={systemStats?.totalRedirects || 0}
            icon="trending_up"
          />
        </Grid>

        {/* Redirects Chart */}
        <Grid item xs={12}>
          <Paper
            sx={{
              p: 2,
              display: 'flex',
              flexDirection: 'column',
              height: 300,
            }}
          >
            <Typography variant="h6" gutterBottom>
              Redirects Over Time
            </Typography>
            <LineChart
              series={[
                {
                  data: redirectStats?.map(stat => stat.value) || [],
                  label: 'Redirects',
                },
              ]}
              xAxis={[
                {
                  data: redirectStats?.map(stat => new Date(stat.timestamp)) || [],
                  scaleType: 'time',
                },
              ]}
            />
          </Paper>
        </Grid>

        {/* Peak Usage */}
        <Grid item xs={12} md={6}>
          <Paper
            sx={{
              p: 2,
              display: 'flex',
              flexDirection: 'column',
              height: 300,
            }}
          >
            <Typography variant="h6" gutterBottom>
              Peak Usage by Hour
            </Typography>
            <BarChart
              series={[
                {
                  data: peakUsage?.hourlyStats.map(stat => stat.redirects) || [],
                  label: 'Redirects',
                },
              ]}
              xAxis={[
                {
                  data: peakUsage?.hourlyStats.map(stat => stat.hour) || [],
                  label: 'Hour',
                },
              ]}
            />
          </Paper>
        </Grid>

        {/* Status Distribution */}
        <Grid item xs={12} md={6}>
          <Paper
            sx={{
              p: 2,
              display: 'flex',
              flexDirection: 'column',
              height: 300,
            }}
          >
            <Typography variant="h6" gutterBottom>
              HTTP Status Distribution
            </Typography>
            <PieChart
              series={[
                {
                  data: [
                    { value: systemStats?.status2xx || 0, label: '2xx' },
                    { value: systemStats?.status3xx || 0, label: '3xx' },
                    { value: systemStats?.status4xx || 0, label: '4xx' },
                    { value: systemStats?.status5xx || 0, label: '5xx' },
                  ],
                },
              ]}
            />
          </Paper>
        </Grid>
      </Grid>
    </Container>
  );
}; 