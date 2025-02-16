import React, { useState } from 'react';
import {
  Container,
  Typography,
  Box,
  Button,
  Paper,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  CircularProgress,
  Alert,
  IconButton,
  Dialog as ConfirmDialog,
  DialogContentText,
  Chip
} from '@mui/material';
import AddIcon from '@mui/icons-material/Add';
import { DataGrid } from '../components/common/DataGrid';
import { useLinks } from '../hooks/useLinks';
import type { Link } from '../types/link';
import type { Column } from '../components/common/DataGrid';
import DeleteIcon from '@mui/icons-material/Delete';
import EditIcon from '@mui/icons-material/Edit';
import { CreateLinkDialog } from '../components/CreateLinkDialog';

interface CreateLinkForm {
  destinationUrl: string;
  alias: string;
}

export const LinksPage: React.FC = () => {
  const [page, setPage] = useState(0);
  const [pageSize, setPageSize] = useState(10);
  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [newLink, setNewLink] = useState<CreateLinkForm>({ destinationUrl: '', alias: '' });
  const [error, setError] = useState<string | null>(null);
  const [formErrors, setFormErrors] = useState<Partial<CreateLinkForm>>({});
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false);
  const [linkToDelete, setLinkToDelete] = useState<Link | null>(null);
  const [editDialogOpen, setEditDialogOpen] = useState(false);
  const [linkToEdit, setLinkToEdit] = useState<Link | null>(null);
  const [editForm, setEditForm] = useState({ destinationUrl: '' });

  const { links, totalCount, isLoading, error: fetchError, createLink, deleteLink, updateLink, refetch } = useLinks({
    page,
    pageSize
  });

  const columns: Column<Link>[] = [
    { field: 'alias' as keyof Link, headerName: 'Alias', width: 130 },
    { field: 'destinationUrl' as keyof Link, headerName: 'Target URL', width: 300 },
    { 
      field: 'createdAt', 
      headerName: 'Created', 
      width: 180,
      valueFormatter: (params) => {
        const date = params.value as string;
        return date ? new Date(date).toLocaleString('en-US', {
          year: 'numeric',
          month: 'long',
          day: 'numeric',
          hour: '2-digit',
          minute: '2-digit'
        }) : '';
      }
    },
    { 
      field: 'isActive' as keyof Link, 
      headerName: 'Status', 
      width: 120,
      renderCell: (params) => {
        const isActive = Boolean(params.row.isActive);
        return (
          <Chip 
            label={isActive ? 'Active' : 'Inactive'} 
            color={isActive ? 'success' : 'default'}
          />
        );
      }
    },
    {
      field: 'actions',
      headerName: 'Actions',
      width: 130,
      renderCell: (params) => {
        const link = params.row as Link;
        return (
          <Box>
            <IconButton
              onClick={(e) => {
                e.stopPropagation();
                setLinkToEdit(link);
                setEditForm({ destinationUrl: link.destinationUrl });
                setEditDialogOpen(true);
              }}
              size="small"
              color="primary"
              sx={{ mr: 1 }}
            >
              <EditIcon />
            </IconButton>
            <IconButton
              onClick={(e) => {
                e.stopPropagation();
                setLinkToDelete(link);
                setDeleteConfirmOpen(true);
              }}
              size="small"
              color="error"
            >
              <DeleteIcon />
            </IconButton>
          </Box>
        );
      }
    }
  ];

  const validateForm = (): boolean => {
    const errors: Partial<CreateLinkForm> = {};
    
    if (!newLink.destinationUrl) {
      errors.destinationUrl = 'Target URL is required';
    } else {
      try {
        new URL(newLink.destinationUrl);
      } catch {
        errors.destinationUrl = 'Please enter a valid URL';
      }
    }

    if (!newLink.alias) {
      errors.alias = 'Alias is required';
    } else if (!/^[a-zA-Z0-9-]+$/.test(newLink.alias)) {
      errors.alias = 'Alias can only contain letters, numbers, and hyphens';
    }

    setFormErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const handleCreate = async () => {
    try {
      setError(null);
      if (!validateForm()) {
        return;
      }

      await createLink({
        destinationUrl: newLink.destinationUrl,
        alias: newLink.alias
      });
      setCreateDialogOpen(false);
      setNewLink({ destinationUrl: '', alias: '' });
      setFormErrors({});
      refetch();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create link');
    }
  };

  const handleDelete = async () => {
    if (!linkToDelete) return;
    
    try {
      await deleteLink(linkToDelete.id);
      setDeleteConfirmOpen(false);
      setLinkToDelete(null);
      refetch();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete link');
    }
  };

  const handleEdit = async () => {
    if (!linkToEdit) return;
    
    try {
      await updateLink(linkToEdit.id, {
        destinationUrl: editForm.destinationUrl
      });
      setEditDialogOpen(false);
      setLinkToEdit(null);
      refetch();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update link');
    }
  };

  const handleEditKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && editForm.destinationUrl) {
      e.preventDefault();
      handleEdit();
    }
  };

  if (isLoading) {
    return (
      <Container maxWidth="lg">
        <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '400px' }}>
          <CircularProgress />
        </Box>
      </Container>
    );
  }

  if (fetchError) {
    return (
      <Container maxWidth="lg">
        <Box sx={{ mt: 4, mb: 4 }}>
          <Typography color="error">Error loading links: {fetchError.message}</Typography>
        </Box>
      </Container>
    );
  }

  return (
    <Container maxWidth="lg">
      <Box sx={{ mt: 4, mb: 4 }}>
        <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 3 }}>
          <Typography variant="h4" component="h1">
            My Links
          </Typography>
          <Button
            variant="contained"
            startIcon={<AddIcon />}
            onClick={() => setCreateDialogOpen(true)}
          >
            Create Link
          </Button>
        </Box>

        <Paper sx={{ p: 2, height: 400 }}>
          {links.length === 0 ? (
            <Box sx={{ 
              display: 'flex', 
              flexDirection: 'column',
              alignItems: 'center', 
              justifyContent: 'center',
              height: '100%'
            }}>
              <Typography color="textSecondary" gutterBottom>
                No links yet
              </Typography>
              <Button
                variant="contained"
                startIcon={<AddIcon />}
                onClick={() => setCreateDialogOpen(true)}
                sx={{ mt: 2 }}
              >
                Create your first link
              </Button>
            </Box>
          ) : (
            <DataGrid<Link>
              data={links}
              columns={columns}
              pageSize={pageSize}
              totalCount={totalCount}
              isLoading={isLoading}
              page={page}
              onPageChange={setPage}
              onPageSizeChange={setPageSize}
              error={fetchError?.message}
              onRefresh={refetch}
            />
          )}
        </Paper>

        <CreateLinkDialog 
          open={createDialogOpen}
          onClose={() => setCreateDialogOpen(false)}
          onSuccess={refetch}
        />

        <Dialog open={editDialogOpen} onClose={() => setEditDialogOpen(false)}>
          <DialogTitle>Edit Link</DialogTitle>
          <DialogContent>
            {error && (
              <Alert severity="error" sx={{ mb: 2 }}>
                {error}
              </Alert>
            )}
            <TextField
              autoFocus
              margin="dense"
              label="Target URL"
              fullWidth
              value={editForm.destinationUrl}
              onChange={(e) => setEditForm(prev => ({ ...prev, destinationUrl: e.target.value }))}
              error={!!formErrors.destinationUrl}
              helperText={formErrors.destinationUrl}
              placeholder="https://example.com"
              onKeyPress={handleEditKeyPress}
            />
          </DialogContent>
          <DialogActions>
            <Button onClick={() => setEditDialogOpen(false)}>Cancel</Button>
            <Button 
              onClick={handleEdit} 
              variant="contained"
              disabled={!editForm.destinationUrl}
            >
              Update
            </Button>
          </DialogActions>
        </Dialog>

        <ConfirmDialog open={deleteConfirmOpen} onClose={() => setDeleteConfirmOpen(false)}>
          <DialogTitle>Delete Link</DialogTitle>
          <DialogContent>
            <DialogContentText>
              Are you sure you want to delete the link "{linkToDelete?.alias}"? This action cannot be undone.
            </DialogContentText>
          </DialogContent>
          <DialogActions>
            <Button onClick={() => setDeleteConfirmOpen(false)}>Cancel</Button>
            <Button onClick={handleDelete} color="error" variant="contained">
              Delete
            </Button>
          </DialogActions>
        </ConfirmDialog>
      </Box>
    </Container>
  );
}; 