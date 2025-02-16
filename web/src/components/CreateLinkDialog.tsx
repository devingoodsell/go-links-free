import React, { useState } from 'react';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  TextField,
  Alert,
} from '@mui/material';
import { api } from '../utils/api';

interface CreateLinkDialogProps {
  open: boolean;
  onClose: () => void;
  onSuccess: () => void;
}

interface FormData {
  alias: string;
  destinationUrl: string;
  expiresAt?: string;
}

export const CreateLinkDialog: React.FC<CreateLinkDialogProps> = ({
  open,
  onClose,
  onSuccess,
}) => {
  const [formData, setFormData] = useState<FormData>({
    alias: '',
    destinationUrl: '',
  });
  const [error, setError] = useState<string | null>(null);
  const [formErrors, setFormErrors] = useState<Partial<FormData>>({});

  const validateForm = (): boolean => {
    const errors: Partial<FormData> = {};
    
    if (!formData.destinationUrl) {
      errors.destinationUrl = 'Target URL is required';
    } else {
      try {
        new URL(formData.destinationUrl);
      } catch {
        errors.destinationUrl = 'Please enter a valid URL';
      }
    }

    if (!formData.alias) {
      errors.alias = 'Alias is required';
    } else if (!/^[a-zA-Z0-9-]+$/.test(formData.alias)) {
      errors.alias = 'Alias can only contain letters, numbers, and hyphens';
    }

    setFormErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!validateForm()) return;

    try {
      await api.post('/api/links', {
        alias: formData.alias,
        destinationUrl: formData.destinationUrl,
        expiresAt: formData.expiresAt,
        isActive: true,  // Explicitly set to true when creating
      });
      onClose();
      onSuccess();
      // Reset form
      setFormData({ alias: '', destinationUrl: '' });
      setError(null);
      setFormErrors({});
    } catch (error: any) {
      setError(error.response?.data?.error || 'Failed to create link');
    }
  };

  return (
    <Dialog open={open} onClose={onClose}>
      <DialogTitle>Create New Link</DialogTitle>
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
          value={formData.destinationUrl}
          onChange={(e) => setFormData(prev => ({ ...prev, destinationUrl: e.target.value }))}
          error={!!formErrors.destinationUrl}
          helperText={formErrors.destinationUrl}
          placeholder="https://example.com"
        />
        <TextField
          margin="dense"
          label="Alias"
          fullWidth
          required
          value={formData.alias}
          onChange={(e) => setFormData(prev => ({ ...prev, alias: e.target.value }))}
          error={!!formErrors.alias}
          helperText={formErrors.alias || 'This will be used as go/your-alias'}
          placeholder="your-alias"
        />
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Cancel</Button>
        <Button 
          onClick={handleSubmit} 
          variant="contained"
          disabled={!formData.destinationUrl || !formData.alias}
        >
          Create
        </Button>
      </DialogActions>
    </Dialog>
  );
}; 