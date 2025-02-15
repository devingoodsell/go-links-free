import React, { useState } from 'react';
import {
  Box,
  Container,
  Paper,
  Typography,
  IconButton,
  Tooltip,
  Chip,
  Alert,
  Button,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  CircularProgress,
} from '@mui/material';
import {
  DataGrid,
  GridColDef,
  GridRenderCellParams,
} from '@mui/x-data-grid';
import {
  Edit as EditIcon,
  Delete as DeleteIcon,
} from '@mui/icons-material';
import { useUsers, useUpdateUser, useDeleteUser } from '../../hooks/useUsers';
import type { User } from '../../types/user';
import { UserEditDialog } from '../../components/admin/UserEditDialog';
import { SearchFilterBar } from '../../components/admin/SearchFilterBar';
import { TableSkeleton } from '../../components/common/TableSkeleton';
import { Snackbar } from '../../components/common/Snackbar';
import { BulkActionsMenu } from '../../components/admin/BulkActionsMenu';
import { ConfirmationDialog } from '../../components/common/ConfirmationDialog';
import { usePaginationErrorHandler } from '../../hooks/usePaginationErrorHandler';

export const UserManagement: React.FC = () => {
  const [selectedUser, setSelectedUser] = useState<User | undefined>(undefined);
  const [editDialogOpen, setEditDialogOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [filters, setFilters] = useState<Record<string, string>>({});
  const {
    page,
    pageSize,
    error: paginationError,
    handlePageChange,
    handlePageSizeChange,
    clearError,
    setPage,
  } = usePaginationErrorHandler({ page: 0, pageSize: 10 });
  const [notification, setNotification] = useState<{
    open: boolean;
    message: string;
    severity: 'success' | 'error';
  }>({ open: false, message: '', severity: 'success' });
  const [selectedUsers, setSelectedUsers] = useState<number[]>([]);
  const [bulkActionLoading, setBulkActionLoading] = useState(false);
  const [bulkConfirmation, setBulkConfirmation] = useState<{
    open: boolean;
    type: 'delete' | 'activate' | 'deactivate';
    count: number;
  }>({ open: false, type: 'delete', count: 0 });

  const { data, isLoading, error, refetch } = useUsers({
    search: searchQuery,
    role: filters.role as 'admin' | 'user',
    sortBy: filters.sortBy as 'created_desc' | 'created_asc' | 'last_login_desc',
    page,
    pageSize,
  });

  const { mutateAsync: updateUserAsync } = useUpdateUser();
  const { deleteUser, isLoading: isDeleting } = useDeleteUser();

  const columns: GridColDef<User>[] = [
    { field: 'email', headerName: 'Email', flex: 2 },
    {
      field: 'role',
      headerName: 'Role',
      width: 120,
      renderCell: (params: GridRenderCellParams) => (
        <Chip
          label={params.row.isAdmin ? 'Admin' : 'User'}
          color={params.row.isAdmin ? 'primary' : 'default'}
          size="small"
        />
      ),
    },
    {
      field: 'lastLoginAt',
      headerName: 'Last Login',
      width: 180,
      valueFormatter: (params: GridRenderCellParams<User>) => 
        params.value ? new Date(params.value).toLocaleString() : 'Never',
    },
    {
      field: 'createdAt',
      headerName: 'Created',
      width: 180,
      valueFormatter: (params: GridRenderCellParams<User>) => 
        new Date(params.value).toLocaleString(),
    },
    {
      field: 'actions',
      headerName: 'Actions',
      width: 120,
      renderCell: (params: GridRenderCellParams) => (
        <Box>
          <Tooltip title="Edit">
            <IconButton onClick={() => setSelectedUser(params.row)}>
              <EditIcon />
            </IconButton>
          </Tooltip>
          <Tooltip title="Delete">
            <IconButton 
              onClick={() => setSelectedUser(params.row)}
              disabled={params.row.isAdmin} // Prevent deleting admin users
            >
              <DeleteIcon />
            </IconButton>
          </Tooltip>
        </Box>
      ),
    },
  ];

  const handleSearch = (query: string) => {
    setSearchQuery(query);
    setPage(0);
  };

  const handleFilter = (newFilters: Record<string, string>) => {
    setFilters(newFilters);
    setPage(0);
  };

  const filterOptions = [
    {
      field: 'role',
      label: 'Role',
      options: [
        { value: 'admin', label: 'Admin' },
        { value: 'user', label: 'User' },
      ],
    },
    {
      field: 'sortBy',
      label: 'Sort By',
      options: [
        { value: 'created_desc', label: 'Newest' },
        { value: 'created_asc', label: 'Oldest' },
        { value: 'last_login_desc', label: 'Recent Activity' },
      ],
    },
  ];

  const handleBulkDelete = async () => {
    setBulkConfirmation({
      open: true,
      type: 'delete',
      count: selectedUsers.length,
    });
  };

  const handleBulkStatusChange = async (activate: boolean) => {
    setBulkConfirmation({
      open: true,
      type: activate ? 'activate' : 'deactivate',
      count: selectedUsers.length,
    });
  };

  const handleBulkAction = async () => {
    if (selectedUsers.length === 0) return;

    setBulkActionLoading(true);
    try {
      switch (bulkConfirmation.type) {
        case 'delete':
          await Promise.all(selectedUsers.map(id => deleteUser(id)));
          break;
        case 'activate':
        case 'deactivate':
          await Promise.all(selectedUsers.map(id => 
            updateUserAsync({ 
              id, 
              isActive: bulkConfirmation.type === 'activate' 
            })
          ));
          break;
      }
      setSelectedUsers([]);
      await refetch();
    } catch (err) {
      const error = err as Error;
      setNotification({
        open: true,
        message: error.message || `Error performing bulk ${bulkConfirmation.type}`,
        severity: 'error',
      });
    }
  };

  if (isLoading && !data) {
    return <TableSkeleton rowCount={pageSize} columnCount={5} />;
  }

  return (
    <Container maxWidth="xl" sx={{ mt: 4, mb: 4 }}>
      <Paper sx={{ p: 2 }}>
        <Typography variant="h5" gutterBottom>
          User Management
        </Typography>

        {error && (
          <Alert 
            severity="error" 
            sx={{ mb: 2 }}
            action={
              <Button color="inherit" size="small" onClick={() => refetch()}>
                Retry
              </Button>
            }
          >
            {error.message || 'Error loading users. Please try again.'}
          </Alert>
        )}

        {paginationError && (
          <Alert 
            severity="error" 
            sx={{ mb: 2 }}
            onClose={clearError}
            action={
              <Button color="inherit" size="small" onClick={() => refetch()}>
                Retry
              </Button>
            }
          >
            {paginationError}
          </Alert>
        )}

        <Box sx={{ display: 'flex', gap: 2, mb: 2 }}>
          <SearchFilterBar
            onSearch={handleSearch}
            onFilter={handleFilter}
            filters={filterOptions}
            placeholder="Search users..."
          />
          <BulkActionsMenu
            selectedCount={selectedUsers.length}
            onDelete={handleBulkDelete}
            onDeactivate={() => handleBulkStatusChange(false)}
            onActivate={() => handleBulkStatusChange(true)}
            isLoading={bulkActionLoading}
          />
        </Box>

        <Box sx={{ height: 600, width: '100%', position: 'relative' }}>
          {isLoading && (
            <Box 
              sx={{ 
                position: 'absolute',
                top: 0,
                left: 0,
                right: 0,
                bottom: 0,
                bgcolor: 'rgba(255, 255, 255, 0.7)',
                zIndex: 1,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center'
              }}
            >
              <CircularProgress />
            </Box>
          )}
          <DataGrid
            rows={data?.items || []}
            columns={columns}
            initialState={{
              pagination: {
                paginationModel: { page, pageSize },
              },
            }}
            rowCount={data?.totalCount || 0}
            paginationMode="server"
            onPaginationModelChange={({ page: newPage, pageSize: newPageSize }) => {
              handlePageChange(newPage, async () => { await refetch(); });
              handlePageSizeChange(newPageSize, async () => { await refetch(); });
            }}
            pageSizeOptions={[10, 25, 50]}
            checkboxSelection
            onRowSelectionModelChange={(newSelection) => {
              setSelectedUsers(newSelection.map(id => Number(id)));
            }}
            rowSelectionModel={selectedUsers}
          />
        </Box>
      </Paper>

      <UserEditDialog
        open={editDialogOpen}
        user={selectedUser}
        onClose={() => setEditDialogOpen(false)}
        onSave={async (updatedUser) => {
          try {
            await updateUserAsync(updatedUser);
            await refetch();
            setEditDialogOpen(false);
          } catch (err) {
            const error = err as Error;
            setNotification({
              open: true,
              message: error.message || 'Error updating user',
              severity: 'error',
            });
          }
        }}
      />

      <Dialog open={deleteDialogOpen} onClose={() => setDeleteDialogOpen(false)}>
        <DialogTitle>Confirm Delete</DialogTitle>
        <DialogContent>
          Are you sure you want to delete this user?
          {selectedUser?.isAdmin && (
            <Alert severity="warning" sx={{ mt: 2 }}>
              Admin users cannot be deleted.
            </Alert>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDeleteDialogOpen(false)}>Cancel</Button>
          <Button
            onClick={() => {
              if (selectedUser && !selectedUser.isAdmin) {
                deleteUser(selectedUser.id);
              }
            }}
            color="error"
            disabled={selectedUser?.isAdmin || isDeleting}
          >
            {isDeleting ? 'Deleting...' : 'Delete'}
          </Button>
        </DialogActions>
      </Dialog>

      <ConfirmationDialog
        open={bulkConfirmation.open}
        title={`Confirm Bulk ${bulkConfirmation.type}`}
        message={`Are you sure you want to ${bulkConfirmation.type} ${bulkConfirmation.count} users?`}
        confirmLabel={bulkConfirmation.type}
        onConfirm={handleBulkAction}
        onCancel={() => setBulkConfirmation(prev => ({ ...prev, open: false }))}
        isLoading={bulkActionLoading}
        warning={
          bulkConfirmation.type === 'delete' 
            ? 'This action cannot be undone.'
            : undefined
        }
      />

      <Snackbar
        open={notification.open}
        message={notification.message}
        severity={notification.severity}
        onClose={() => setNotification(prev => ({ ...prev, open: false }))}
      />
    </Container>
  );
}; 