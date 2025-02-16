import React, { useState } from 'react';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  TextField,
  Switch,
  FormControlLabel,
  CircularProgress,
  Alert
} from '@mui/material';
import { useUpdateLink } from '../../hooks/useUpdateLink';
import type { Link, LinkFormData } from '../../types/link';

interface LinkEditDialogProps {
  open: boolean;
  link?: Link;
  onClose: () => void;
  onSave: (link: Link) => Promise<void>;
}

export const LinkEditDialog: React.FC<LinkEditDialogProps> = ({
  open,
  onClose,
  onSave,
  link
}) => {
  const initialFormData: LinkFormData = link ? {
    alias: link.alias,
    destinationUrl: link.destinationUrl,
    expiresAt: link.expiresAt,
    isActive: link.isActive
  } : {
    alias: '',
    destinationUrl: '',
    isActive: true
  };

  const [formData, setFormData] = useState<LinkFormData>(initialFormData);
  const [error, setError] = useState<string | null>(null);
  const { mutate, isLoading } = useUpdateLink();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    try {
      if (!link) return;

      const updatedLink: Link = {
        ...link,
        ...formData
      };

      await onSave(updatedLink);
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update link');
    }
  };

  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>Edit Link</DialogTitle>
      <form onSubmit={handleSubmit}>
        <DialogContent>
          {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}
          <TextField
            label="Alias"
            fullWidth
            margin="normal"
            value={formData.alias}
            onChange={(e) => setFormData(prev => ({ ...prev, alias: e.target.value }))}
            disabled={isLoading}
          />
          <TextField
            label="Destination URL"
            fullWidth
            margin="normal"
            value={formData.destinationUrl}
            onChange={(e) => setFormData(prev => ({ ...prev, destinationUrl: e.target.value }))}
            disabled={isLoading}
          />
          <TextField
            label="Expires At"
            type="datetime-local"
            fullWidth
            margin="normal"
            value={formData.expiresAt?.split('.')[0] || ''}
            onChange={(e) => setFormData(prev => ({ ...prev, expiresAt: e.target.value }))}
            disabled={isLoading}
            InputLabelProps={{ shrink: true }}
          />
          <FormControlLabel
            control={
              <Switch
                checked={formData.isActive}
                onChange={(e) => setFormData(prev => ({ ...prev, isActive: e.target.checked }))}
                disabled={isLoading}
              />
            }
            label="Active"
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={onClose} disabled={isLoading}>Cancel</Button>
          <Button type="submit" variant="contained" disabled={isLoading}>
            {isLoading ? <CircularProgress size={24} /> : 'Save'}
          </Button>
        </DialogActions>
      </form>
    </Dialog>
  );
}; 