import React, { useState } from 'react';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Button,
  Box,
  Alert,
  FormControlLabel,
  Checkbox
} from '@mui/material';
import { useUpdateUser } from '../../hooks/useUpdateUser';
import type { User } from '../../types/user';

interface UserEditDialogProps {
  open: boolean;
  onClose: () => void;
  onSave: (user: User) => Promise<void>;
  user?: User;
}

interface UserFormData extends Partial<User> {
  password?: string;
}

export const UserEditDialog: React.FC<UserEditDialogProps> = ({
  open,
  onClose,
  onSave,
  user
}) => {
  const [formData, setFormData] = useState<UserFormData>(user || {});
  const [error, setError] = useState<string | null>(null);
  const updateUser = useUpdateUser();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    try {
      if (user) {
        await updateUser.mutateAsync({ ...user, ...formData });
      }
      await onSave(formData as User);
      onClose();
    } catch (error) {
      setError(error instanceof Error ? error.message : 'An error occurred');
    }
  };

  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>{user ? 'Edit User' : 'Create User'}</DialogTitle>
      <form onSubmit={handleSubmit}>
        <DialogContent>
          <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
            {error && (
              <Alert severity="error" sx={{ mb: 2 }}>
                {error}
              </Alert>
            )}
            <TextField
              label="Email"
              type="email"
              value={formData.email || ''}
              onChange={(e) => setFormData({ ...formData, email: e.target.value })}
              required
            />
            {!user && (
              <TextField
                label="Password"
                type="password"
                onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                required
              />
            )}
            <FormControlLabel
              control={
                <Checkbox
                  checked={formData.role === 'admin'}
                  onChange={(e) => setFormData({ 
                    ...formData, 
                    role: e.target.checked ? 'admin' : 'user' 
                  })}
                />
              }
              label="Admin User"
            />
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={onClose}>Cancel</Button>
          <Button type="submit" variant="contained" color="primary">
            Save
          </Button>
        </DialogActions>
      </form>
    </Dialog>
  );
}; 