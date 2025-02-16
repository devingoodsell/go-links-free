import React, { useState } from 'react';
import {
  Box,
  Container,
  Paper,
  Typography,
  IconButton,
  Tooltip,
  Chip,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
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
import { useLinks } from '../../hooks/useLinks';
import { useUpdateLink } from '../../hooks/useUpdateLink';
import { useDeleteLink } from '../../hooks/useDeleteLink';
import type { Link } from '../../types/link';
import { LinkEditDialog } from '../../components/admin/LinkEditDialog';
import { SearchFilterBar } from '../../components/admin/SearchFilterBar';
import { Snackbar } from '../../components/common/Snackbar';
import { usePaginationErrorHandler } from '../../hooks/usePaginationErrorHandler';
import { BulkActionsMenu } from '../../components/admin/BulkActionsMenu';
import { ConfirmationDialog } from '../../components/common/ConfirmationDialog';

export const LinkManagement: React.FC = () => {
  const [selectedLink, setSelectedLink] = useState<Link | undefined>(undefined);
  const [editDialogOpen, setEditDialogOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [filters, setFilters] = useState<Record<string, string>>({});
  const {
    page,
    pageSize,
    handlePageChange,
    handlePageSizeChange,
    setPage,
  } = usePaginationErrorHandler({ page: 0, pageSize: 10 });
  const [selectedLinks, setSelectedLinks] = useState<number[]>([]);
  const [bulkActionLoading, setBulkActionLoading] = useState(false);
  const [bulkConfirmation, setBulkConfirmation] = useState<{
    open: boolean;
    type: 'delete' | 'activate' | 'deactivate';
    count: number;
  }>({ open: false, type: 'delete', count: 0 });
  const [notification, setNotification] = useState<{
    open: boolean;
    message: string;
    severity: 'success' | 'error';
  }>({ open: false, message: '', severity: 'success' });
  
  const { links, totalCount, isLoading, error, refetch } = useLinks({
    search: searchQuery,
    status: filters.status as 'active' | 'expired',
    sortBy: filters.sortBy as 'created_desc' | 'created_asc' | 'clicks_desc',
    page,
    pageSize,
  });

  const updateLink = useUpdateLink();
  const deleteLink = useDeleteLink();

  const handleEdit = (link: Link) => {
    setSelectedLink(link);
    setEditDialogOpen(true);
  };

  const handleDelete = (link: Link) => {
    setSelectedLink(link);
    setDeleteDialogOpen(true);
  };

  const columns: GridColDef<Link>[] = [
    { field: 'alias', headerName: 'Alias', flex: 1 },
    { field: 'destinationUrl', headerName: 'Destination URL', flex: 2 },
    {
      field: 'status',
      headerName: 'Status',
      width: 120,
      renderCell: (params: GridRenderCellParams<Link>) => {
        const isExpired = params.row.expiresAt && new Date(params.row.expiresAt) < new Date();
        return (
          <Chip
            label={isExpired ? 'Expired' : 'Active'}
            color={isExpired ? 'error' : 'success'}
            size="small"
          />
        );
      },
    },
    {
      field: 'clicks',
      headerName: 'Total Clicks',
      width: 120,
      valueFormatter: (params: { value: any }) => params.value?.toString() || '0',
    },
    {
      field: 'createdAt',
      headerName: 'Created',
      width: 180,
      valueFormatter: (params: { value: any }) => 
        new Date(params.value as string).toLocaleString(),
    },
    {
      field: 'actions',
      headerName: 'Actions',
      width: 120,
      renderCell: (params: GridRenderCellParams<Link>) => (
        <Box>
          <Tooltip title="Edit">
            <IconButton onClick={() => handleEdit(params.row)}>
              <EditIcon />
            </IconButton>
          </Tooltip>
          <Tooltip title="Delete">
            <IconButton onClick={() => handleDelete(params.row)}>
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
      field: 'status',
      label: 'Status',
      options: [
        { value: 'active', label: 'Active' },
        { value: 'expired', label: 'Expired' },
      ],
    },
    {
      field: 'sortBy',
      label: 'Sort By',
      options: [
        { value: 'created_desc', label: 'Newest' },
        { value: 'created_asc', label: 'Oldest' },
        { value: 'clicks_desc', label: 'Most Clicked' },
      ],
    },
  ];

  const handleBulkAction = async () => {
    if (selectedLinks.length === 0) return;

    setBulkActionLoading(true);
    try {
      switch (bulkConfirmation.type) {
        case 'delete':
          await Promise.all(selectedLinks.map(id => deleteLink.deleteLink(id)));
          setNotification({
            open: true,
            message: `Successfully deleted ${selectedLinks.length} links`,
            severity: 'success',
          });
          break;
        case 'activate':
        case 'deactivate':
          await Promise.all(selectedLinks.map(id => 
            updateLink.mutate({ 
              id, 
              isActive: bulkConfirmation.type === 'activate' 
            })
          ));
          break;
      }
      setSelectedLinks([]);
      await refetch();
    } catch (err) {
      const error = err as Error;
      setNotification({
        open: true,
        message: error.message || `Error performing bulk ${bulkConfirmation.type}`,
        severity: 'error',
      });
    } finally {
      setBulkActionLoading(false);
      setBulkConfirmation(prev => ({ ...prev, open: false }));
    }
  };

  return (
    <Container maxWidth="xl" sx={{ mt: 4, mb: 4 }}>
      <Paper sx={{ p: 2 }}>
        <Typography variant="h5" gutterBottom>
          Link Management
        </Typography>
        <Box sx={{ display: 'flex', gap: 2, mb: 2 }}>
          <SearchFilterBar
            onSearch={handleSearch}
            onFilter={handleFilter}
            filters={filterOptions}
            placeholder="Search links..."
          />
          <BulkActionsMenu
            selectedCount={selectedLinks.length}
            onDelete={() => setBulkConfirmation({ type: 'delete', open: true, count: selectedLinks.length })}
            onActivate={() => setBulkConfirmation({ type: 'activate', open: true, count: selectedLinks.length })}
            onDeactivate={() => setBulkConfirmation({ type: 'deactivate', open: true, count: selectedLinks.length })}
            isLoading={bulkActionLoading}
          />
        </Box>
        <Box sx={{ height: 600, width: '100%' }}>
          <DataGrid
            rows={links || []}
            columns={columns}
            initialState={{
              pagination: {
                paginationModel: { page, pageSize },
              },
            }}
            rowCount={totalCount || 0}
            paginationMode="server"
            onPaginationModelChange={({ page, pageSize }) => {
              handlePageChange(page, async () => { await refetch(); });
              handlePageSizeChange(pageSize, async () => { await refetch(); });
            }}
            pageSizeOptions={[10, 25, 50]}
            checkboxSelection
            onRowSelectionModelChange={(newSelection) => {
              setSelectedLinks(newSelection as number[]);
            }}
            rowSelectionModel={selectedLinks}
          />
        </Box>
      </Paper>

      <LinkEditDialog
        open={editDialogOpen}
        link={selectedLink}
        onClose={() => setEditDialogOpen(false)}
        onSave={async (updatedLink) => {
          await updateLink.mutate(updatedLink);
          await refetch();
          setEditDialogOpen(false);
        }}
      />

      <Dialog open={deleteDialogOpen} onClose={() => setDeleteDialogOpen(false)}>
        <DialogTitle>Confirm Delete</DialogTitle>
        <DialogContent>
          Are you sure you want to delete this link?
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDeleteDialogOpen(false)}>Cancel</Button>
          <Button
            onClick={async () => {
              if (selectedLink) {
                // Handle link deletion
                await deleteLink.deleteLink(selectedLink.id);
                setDeleteDialogOpen(false);
              }
            }}
            color="error"
          >
            Delete
          </Button>
        </DialogActions>
      </Dialog>

      <ConfirmationDialog
        open={bulkConfirmation.open}
        title={`Confirm Bulk ${bulkConfirmation.type}`}
        message={`Are you sure you want to ${bulkConfirmation.type} ${bulkConfirmation.count} links?`}
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