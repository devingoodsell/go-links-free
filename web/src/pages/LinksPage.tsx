import { Typography, Box, Button } from '@mui/material';
import AddIcon from '@mui/icons-material/Add';
import { DataGrid } from '../components/common/DataGrid';
import { useLinks } from '../hooks/useLinks';
import type { Link } from '../types/link';

export const LinksPage = () => {
  const { data, isLoading, error, sortBy, sortDirection, handleSortChange } = useLinks();
  
  const columns = [
    { field: 'shortLink' as keyof Link, headerName: 'Short Link' },
    { field: 'targetUrl' as keyof Link, headerName: 'Target URL' },
    { field: 'createdAt' as keyof Link, headerName: 'Created At' },
    { field: 'isActive' as keyof Link, headerName: 'Status', sortable: true },
    { field: 'clicks' as keyof Link, headerName: 'Clicks', sortable: true }
  ];
  
  return (
    <>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Typography variant="h5">Links</Typography>
        <Button 
          variant="contained" 
          startIcon={<AddIcon />}
          onClick={() => {/* TODO: Add link handler */}}
        >
          Add Link
        </Button>
      </Box>
      <DataGrid<Link>
        data={data?.items ?? []}
        columns={columns}
        isLoading={isLoading}
        error={error?.message}
        totalCount={data?.totalCount ?? 0}
        sortBy={sortBy}
        sortDirection={sortDirection}
        onSortChange={handleSortChange}
      />
    </>
  );
}; 