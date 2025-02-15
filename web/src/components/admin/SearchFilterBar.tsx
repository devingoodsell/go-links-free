import React from 'react';
import {
  Box,
  TextField,
  IconButton,
  Menu,
  MenuItem,
  Tooltip,
  Button,
  SxProps,
  Theme
} from '@mui/material';
import {
  Search as SearchIcon,
  FilterList as FilterIcon,
  Clear as ClearIcon,
} from '@mui/icons-material';

interface Filter {
  field: string;
  label: string;
  options?: Array<{
    value: string;
    label: string;
  }>;
}

interface SearchFilterBarProps {
  onSearch: (query: string) => void;
  onFilter: (filters: Record<string, string>) => void;
  filters: Filter[];
  placeholder?: string;
  sx?: SxProps<Theme>;
}

export const SearchFilterBar: React.FC<SearchFilterBarProps> = ({
  onSearch,
  onFilter,
  filters,
  placeholder = 'Search...',
  sx
}) => {
  const [searchQuery, setSearchQuery] = React.useState('');
  const [anchorEl, setAnchorEl] = React.useState<null | HTMLElement>(null);
  const [activeFilters, setActiveFilters] = React.useState<Record<string, string>>({});

  const handleSearch = (event: React.ChangeEvent<HTMLInputElement>) => {
    const query = event.target.value;
    setSearchQuery(query);
    onSearch(query);
  };

  const handleFilterClick = (event: React.MouseEvent<HTMLElement>) => {
    setAnchorEl(event.currentTarget);
  };

  const handleFilterClose = () => {
    setAnchorEl(null);
  };

  const handleFilterApply = (field: string, value: string) => {
    const newFilters = { ...activeFilters, [field]: value };
    setActiveFilters(newFilters);
    onFilter(newFilters);
  };

  const handleClearFilters = () => {
    setSearchQuery('');
    setActiveFilters({});
    onSearch('');
    onFilter({});
  };

  return (
    <Box sx={{ display: 'flex', gap: 2, mb: 2, alignItems: 'center', ...sx }}>
      <TextField
        size="small"
        placeholder={placeholder}
        value={searchQuery}
        onChange={handleSearch}
        InputProps={{
          startAdornment: <SearchIcon color="action" sx={{ mr: 1 }} />,
        }}
        sx={{ width: 300 }}
      />
      
      <Tooltip title="Filter list">
        <IconButton onClick={handleFilterClick}>
          <FilterIcon />
        </IconButton>
      </Tooltip>

      {(searchQuery || Object.keys(activeFilters).length > 0) && (
        <Tooltip title="Clear filters">
          <IconButton onClick={handleClearFilters}>
            <ClearIcon />
          </IconButton>
        </Tooltip>
      )}

      <Menu
        anchorEl={anchorEl}
        open={Boolean(anchorEl)}
        onClose={handleFilterClose}
      >
        {filters.map((filter) => (
          <MenuItem key={filter.field}>
            <Box sx={{ display: 'flex', flexDirection: 'column', width: '100%' }}>
              {filter.label}
              {filter.options && (
                <Box sx={{ display: 'flex', gap: 1, mt: 1 }}>
                  {filter.options.map((option) => (
                    <Button
                      key={option.value}
                      size="small"
                      variant={activeFilters[filter.field] === option.value ? 'contained' : 'outlined'}
                      onClick={() => handleFilterApply(filter.field, option.value)}
                    >
                      {option.label}
                    </Button>
                  ))}
                </Box>
              )}
            </Box>
          </MenuItem>
        ))}
      </Menu>
    </Box>
  );
}; 