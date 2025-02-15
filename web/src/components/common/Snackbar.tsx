import React from 'react';
import { Snackbar as MuiSnackbar, Alert, AlertProps } from '@mui/material';

interface SnackbarProps {
  open: boolean;
  message: string;
  severity?: AlertProps['severity'];
  onClose: () => void;
  autoHideDuration?: number;
}

export const Snackbar: React.FC<SnackbarProps> = ({
  open,
  message,
  severity = 'success',
  onClose,
  autoHideDuration = 6000
}) => {
  return (
    <MuiSnackbar
      open={open}
      autoHideDuration={autoHideDuration}
      onClose={onClose}
      anchorOrigin={{ vertical: 'top', horizontal: 'right' }}
    >
      <Alert onClose={onClose} severity={severity}>
        {message}
      </Alert>
    </MuiSnackbar>
  );
}; 