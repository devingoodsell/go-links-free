import React from 'react';
import {
  Button,
  Menu,
  MenuItem,
  ListItemIcon,
  ListItemText,
  SxProps,
  Theme
} from '@mui/material';
import {
  Delete as DeleteIcon,
  Block as BlockIcon,
  CheckCircle as CheckCircleIcon,
} from '@mui/icons-material';

interface BulkAction {
  type: 'delete' | 'activate' | 'deactivate';
  label: string;
  icon: React.ReactNode;
  color?: 'inherit' | 'primary' | 'secondary' | 'error' | 'info' | 'success' | 'warning';
}

interface BulkActionsMenuProps {
  selectedCount: number;
  onDelete: () => void;
  onActivate: () => void;
  onDeactivate: () => void;
  isLoading?: boolean;
  sx?: SxProps<Theme>;
}

export const BulkActionsMenu: React.FC<BulkActionsMenuProps> = ({
  selectedCount,
  onDelete,
  onActivate,
  onDeactivate,
  isLoading,
  sx
}) => {
  const [anchorEl, setAnchorEl] = React.useState<null | HTMLElement>(null);

  const handleClick = (event: React.MouseEvent<HTMLButtonElement>) => {
    setAnchorEl(event.currentTarget);
  };

  const handleClose = () => {
    setAnchorEl(null);
  };

  const handleAction = (type: 'delete' | 'activate' | 'deactivate') => {
    switch (type) {
      case 'delete':
        onDelete();
        break;
      case 'activate':
        onActivate();
        break;
      case 'deactivate':
        onDeactivate();
        break;
    }
    handleClose();
  };

  const actions: BulkAction[] = [
    {
      type: 'delete',
      label: 'Delete Selected',
      icon: <DeleteIcon />,
      color: 'error'
    },
    {
      type: 'deactivate',
      label: 'Deactivate Selected',
      icon: <BlockIcon />,
      color: 'warning'
    },
    {
      type: 'activate',
      label: 'Activate Selected',
      icon: <CheckCircleIcon />,
      color: 'success'
    }
  ];

  return (
    <>
      <Button
        variant="outlined"
        onClick={handleClick}
        disabled={selectedCount === 0}
        sx={sx}
      >
        Bulk Actions ({selectedCount})
      </Button>
      <Menu
        anchorEl={anchorEl}
        open={Boolean(anchorEl)}
        onClose={handleClose}
      >
        {actions.map((action) => (
          <MenuItem
            key={action.type}
            onClick={() => handleAction(action.type)}
          >
            <ListItemIcon sx={{ color: `${action.color}.main` }}>
              {action.icon}
            </ListItemIcon>
            <ListItemText>{action.label}</ListItemText>
          </MenuItem>
        ))}
      </Menu>
    </>
  );
}; 