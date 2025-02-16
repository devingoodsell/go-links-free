import React, { useState } from 'react';
import { 
  Box,
  Typography,
  Paper,
  Container,
} from '@mui/material';
import { DataGrid } from '../../components/common/DataGrid';
import { useUsers } from '../../hooks/useUsers';
import type { User } from '../../types/user';
import type { Column } from '../../components/common/DataGrid';

export const AdminUsersPage: React.FC = () => {
  const [page, setPage] = useState(0);
  const [pageSize, setPageSize] = useState(10);
  const { users, totalCount, isLoading, error, refetch } = useUsers({
    page,
    pageSize
  });

  const columns: Column<User>[] = [
    { field: 'id' as keyof User, headerName: 'ID', width: 90 },
    { field: 'email' as keyof User, headerName: 'Email', width: 200 },
    { field: 'isAdmin' as keyof User, headerName: 'Admin', width: 130 },
    { 
      field: 'createdAt' as keyof User, 
      headerName: 'Created', 
      width: 180,
      valueFormatter: (params) => {
        if (typeof params.value === 'string') {
          return new Date(params.value).toLocaleString();
        }
        return '';
      }
    },
    { 
      field: 'lastLoginAt' as keyof User, 
      headerName: 'Last Login', 
      width: 180,
      valueFormatter: (params) => {
        if (typeof params.value === 'string') {
          return new Date(params.value).toLocaleString();
        }
        return 'Never';
      }
    },
  ];

  return (
    <Container maxWidth="lg">
      <Box sx={{ mt: 4, mb: 4 }}>
        <Typography variant="h4" component="h1" gutterBottom>
          User Management
        </Typography>
        <Paper sx={{ p: 2, height: 400 }}>
          <DataGrid<User>
            data={users}
            columns={columns}
            pageSize={pageSize}
            totalCount={totalCount}
            isLoading={isLoading}
            page={page}
            onPageChange={setPage}
            onPageSizeChange={setPageSize}
            error={error?.message}
            onRefresh={refetch}
          />
        </Paper>
      </Box>
    </Container>
  );
}; 