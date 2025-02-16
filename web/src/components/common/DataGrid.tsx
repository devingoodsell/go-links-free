import React from 'react';
import {
  DataGrid as MuiDataGrid,
  GridColDef,
  GridRenderCellParams,
  GridSortModel,
  GridRowSelectionModel,
  GridPaginationModel,
  GridValidRowModel
} from '@mui/x-data-grid';
import { Box, CircularProgress } from '@mui/material';

export interface Column<T extends GridValidRowModel> {
  field: keyof T | string;
  headerName: string;
  width: number;
  valueGetter?: (params: { row: T; value: any }) => any;
  valueFormatter?: (params: { value: any }) => string;
  renderCell?: (params: GridRenderCellParams<T>) => React.ReactNode;
}

interface DataGridProps<T extends GridValidRowModel> {
  data: T[];
  columns: Column<T>[];
  isLoading?: boolean;
  error?: string;
  totalCount: number;
  page?: number;
  pageSize?: number;
  sortBy?: string;
  sortDirection?: 'asc' | 'desc';
  onPageChange?: (page: number) => void;
  onPageSizeChange?: (pageSize: number) => void;
  onSortChange?: (field: string, direction: 'asc' | 'desc') => void;
  onSelectionChange?: (selection: number[]) => void;
  selectionModel?: number[];
  onRefresh?: () => void;
}

export function DataGrid<T extends GridValidRowModel>({
  data,
  columns,
  isLoading,
  error,
  totalCount,
  page = 0,
  pageSize = 10,
  sortBy,
  sortDirection,
  onPageChange,
  onPageSizeChange,
  onSortChange,
  onSelectionChange,
  selectionModel,
  onRefresh
}: DataGridProps<T>) {
  const handlePaginationModelChange = (model: GridPaginationModel) => {
    onPageChange?.(model.page);
    onPageSizeChange?.(model.pageSize);
  };

  const handleSortModelChange = (model: GridSortModel) => {
    if (model.length > 0 && onSortChange) {
      onSortChange(model[0].field, model[0].sort || 'asc');
    }
  };

  const handleSelectionModelChange = (model: GridRowSelectionModel) => {
    onSelectionChange?.(model as number[]);
  };

  return (
    <Box sx={{ width: '100%', height: '100%' }}>
      <MuiDataGrid
        rows={data}
        columns={columns as GridColDef[]}
        rowCount={totalCount}
        loading={isLoading}
        slots={{
          noRowsOverlay: () => error ? (
            <div style={{ padding: 16 }}>{error}</div>
          ) : null
        }}
        paginationMode="server"
        sortingMode="server"
        filterMode="server"
        initialState={{
          pagination: {
            paginationModel: {
              pageSize,
              page,
            },
          },
        }}
        onPaginationModelChange={handlePaginationModelChange}
        pageSizeOptions={[10, 25, 50]}
        sortModel={sortBy ? [{ field: sortBy, sort: sortDirection }] : []}
        onSortModelChange={handleSortModelChange}
        checkboxSelection
        rowSelectionModel={selectionModel}
        onRowSelectionModelChange={handleSelectionModelChange}
        disableRowSelectionOnClick
        autoHeight
        sx={{
          '& .MuiDataGrid-cell:focus': {
            outline: 'none',
          },
        }}
      />
      {isLoading && (
        <Box
          sx={{
            position: 'absolute',
            top: '50%',
            left: '50%',
            transform: 'translate(-50%, -50%)',
          }}
        >
          <CircularProgress />
        </Box>
      )}
    </Box>
  );
} 