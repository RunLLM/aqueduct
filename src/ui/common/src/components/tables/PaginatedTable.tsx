import Box from '@mui/material/Box';
import Paper from '@mui/material/Paper';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TablePagination from '@mui/material/TablePagination';
import TableRow from '@mui/material/TableRow';
import Typography from '@mui/material/Typography';
import * as React from 'react';

import { Data } from '../../utils/data';

interface PaginatedTableProps {
  data: Data;
}

export const PaginatedTable: React.FC<PaginatedTableProps> = ({ data }) => {
  const [page, setPage] = React.useState(0);
  const [rowsPerPage, setRowsPerPage] = React.useState(5);

  const rows = data.data;
  const columns = data.schema.fields;

  const handleChangePage = (event: unknown, newPage: number) => {
    setPage(newPage);
  };

  const handleChangeRowsPerPage = (
    event: React.ChangeEvent<HTMLInputElement>
  ) => {
    setRowsPerPage(+event.target.value);
    setPage(0);
  };

  return (
    <Paper sx={{ overflow: 'hidden' }}>
      <TableContainer sx={{ maxHeight: '400px' }}>
        <Table stickyHeader aria-label="sticky table">
          <TableHead>
            <TableRow>
              {columns.map((column, columnIndex) => {
                return (
                  <TableCell
                    key={`table-header-col-${columnIndex}`}
                    align={'left'}
                    sx={{
                      backgroundColor: 'blue.900',
                      color: 'white',
                      minWidth: '80px',
                    }}
                  >
                    <Box flexDirection="column">
                      <Typography
                        variant="body1"
                        sx={{
                          textTransform: 'none',
                          fontFamily: 'monospace',
                          fontSize: '16px',
                        }}
                      >
                        {column.name}
                      </Typography>
                      <Typography
                        variant="caption"
                        sx={{
                          textTransform: 'none',
                          fontFamily: 'monospace',
                          fontSize: '12px',
                        }}
                      >
                        {column.type}
                      </Typography>
                    </Box>
                  </TableCell>
                );
              })}
            </TableRow>
          </TableHead>
          <TableBody>
            {rows
              .slice(page * rowsPerPage, page * rowsPerPage + rowsPerPage)
              .map((row, rowIndex) => {
                return (
                  <TableRow
                    hover
                    role="checkbox"
                    tabIndex={-1}
                    key={`table-row-${rowIndex}`}
                  >
                    {columns.map((column, columnIndex) => {
                      const value = row[column.name];
                      let displayedValue = '';
                      if (!!value) {
                        // when the column type is json or object,
                        // the value parsed from backend API will be an arbitrary json object.
                        // Here we need to serialize the object to render it properly.
                        if (
                          column.type === 'json' ||
                          column.type === 'object'
                        ) {
                          displayedValue = JSON.stringify(value);
                        } else {
                          displayedValue = value.toString();
                        }
                      }
                      return (
                        <TableCell
                          key={`table-col-${columnIndex}`}
                          align={'left'}
                        >
                          {displayedValue}
                        </TableCell>
                      );
                    })}
                  </TableRow>
                );
              })}
          </TableBody>
        </Table>
      </TableContainer>
      <TablePagination
        rowsPerPageOptions={[5, 10, 25, 50, 100]}
        component="div"
        count={rows.length}
        rowsPerPage={rowsPerPage}
        page={page}
        onPageChange={handleChangePage}
        onRowsPerPageChange={handleChangeRowsPerPage}
      />
    </Paper>
  );
};

export default PaginatedTable;
