import React, { useState } from 'react';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Button,
  Box,
  Alert
} from '@mui/material';
import { DateTimePicker } from '@mui/x-date-pickers';
import { useUpdateLink } from '../../hooks/useUpdateLink';
import type { Link } from '../../types/link';

interface LinkEditDialogProps {
  open: boolean;
  onClose: () => void;
  onSave: (link: Link) => Promise<void>;
  link?: Link;
}

interface LinkFormData {
  id?: number;
  shortLink?: string;
  targetUrl?: string;
  expiresAt?: string | null;
  isActive?: boolean;
  clicks?: number;
  createdBy?: number;
  createdAt?: string;
  updatedAt?: string;
  stats?: {
    dailyCount: number;
    weeklyCount: number;
    totalCount: number;
    lastAccessedAt: string;
  };
}

export const LinkEditDialog: React.FC<LinkEditDialogProps> = ({
  open,
  onClose,
  onSave,
  link
}) => {
  const [formData, setFormData] = useState<LinkFormData>(link || {});
  const [error, setError] = useState<string | null>(null);
  const { mutate, isLoading } = useUpdateLink();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    const linkData: Partial<Link> = {
      shortLink: formData.shortLink,
      targetUrl: formData.targetUrl,
      expiresAt: formData.expiresAt || undefined,
      isActive: formData.isActive
    };

    try {
      if (link?.id) {
        await mutate({ ...linkData, id: link.id });
      }
      await onSave(linkData as Link);
      onClose();
    } catch (err) {
      const error = err as Error;
      setError(error.message || 'An error occurred');
    }
  };

  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>{link ? 'Edit Link' : 'Create Link'}</DialogTitle>
      <form onSubmit={handleSubmit}>
        <DialogContent>
          <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
            {error && (
              <Alert severity="error" sx={{ mb: 2 }}>
                {error}
              </Alert>
            )}
            <TextField
              label="Short Link"
              value={formData.shortLink || ''}
              onChange={(e) => setFormData({ ...formData, shortLink: e.target.value })}
              required
            />
            <TextField
              label="Target URL"
              value={formData.targetUrl || ''}
              onChange={(e) => setFormData({ ...formData, targetUrl: e.target.value })}
              required
            />
            <DateTimePicker
              label="Expiration Date"
              value={formData.expiresAt ? new Date(formData.expiresAt) : null}
              onChange={(date) => setFormData({ 
                ...formData, 
                expiresAt: date ? date.toISOString() : null 
              })}
              slotProps={{
                textField: {
                  fullWidth: true,
                  helperText: 'Optional: Set an expiration date for this link'
                }
              }}
            />
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={onClose} disabled={isLoading}>Cancel</Button>
          <Button 
            type="submit" 
            variant="contained" 
            color="primary"
            disabled={isLoading}
          >
            {isLoading ? 'Saving...' : 'Save'}
          </Button>
        </DialogActions>
      </form>
    </Dialog>
  );
}; 