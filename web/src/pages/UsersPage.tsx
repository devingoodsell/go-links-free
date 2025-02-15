import { Typography } from '@mui/material';
import { DataGrid } from '../components/common/DataGrid';
import { useUsers } from '../hooks/useUsers';
import type { User } from '../types/user';

export const UsersPage = () => {
  const { data, isLoading, error } = useUsers();

  const columns = [
    { field: 'email' as keyof User, headerName: 'Email' },
    { field: 'role' as keyof User, headerName: 'Role' },
    { field: 'lastLoginAt' as keyof User, headerName: 'Last Login' },
    { field: 'createdAt' as keyof User, headerName: 'Created At' }
  ];

  return (
    <>
      <Typography variant="h5" gutterBottom>
        Users
      </Typography>

      <DataGrid<User>
        data={data?.items ?? []}
        columns={columns}
        isLoading={isLoading}
        error={error?.message}
        totalCount={data?.totalCount ?? 0}
      />
    </>
  );
}; 