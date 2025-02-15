import React from 'react';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  Skeleton,
  Box,
  Paper,
  SxProps,
  Theme
} from '@mui/material';

interface TableSkeletonProps {
  rowCount?: number;
  columnCount?: number;
  sx?: SxProps<Theme>;
}

export const TableSkeleton: React.FC<TableSkeletonProps> = ({
  rowCount = 5,
  columnCount = 4,
  sx
}) => {
  return (
    <Paper sx={sx}>
      <Box sx={{ overflow: 'auto' }}>
        <Table>
          <TableHead>
            <TableRow>
              {Array.from({ length: columnCount }).map((_, index) => (
                <TableCell key={`header-${index}`}>
                  <Skeleton variant="text" width={100} />
                </TableCell>
              ))}
            </TableRow>
          </TableHead>
          <TableBody>
            {Array.from({ length: rowCount }).map((_, rowIndex) => (
              <TableRow key={`row-${rowIndex}`}>
                {Array.from({ length: columnCount }).map((_, colIndex) => (
                  <TableCell key={`cell-${rowIndex}-${colIndex}`}>
                    <Skeleton 
                      variant="text" 
                      width={colIndex === 0 ? 150 : 100}
                      sx={{ my: 0.5 }}
                    />
                  </TableCell>
                ))}
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </Box>
    </Paper>
  );
}; 