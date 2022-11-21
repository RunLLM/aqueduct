import Box from '@mui/material/Box';
import Paper from '@mui/material/Paper';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TablePagination from '@mui/material/TablePagination';
import TableRow from '@mui/material/TableRow';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';
import * as React from 'react';

import { DataSchema } from '../../utils/data';
import { CheckPreview } from '../pages/workflows/components/CheckItem';
import { ExecutionStatusLinkProps } from '../pages/workflows/components/ExecutionStatusLink';
import { MetricPreview } from '../pages/workflows/components/MetricItem';

export type WorkflowTableElement = string | number | boolean | JSX.Element;

export type WorkflowTableRow = {
  [key: string]: WorkflowTableElement;
};

export type WorkflowTableRowData = {
  [key: string]:
  | string
  | number
  | boolean
  | CheckPreview[]
  | MetricPreview[]
  | ExecutionStatusLinkProps;
};

export interface WorkflowTableData {
  schema?: DataSchema;
  data: WorkflowTableRow[];
  meta: WorkflowTableRowData[];
}

export interface WorkflowsTableProps {
  data: WorkflowTableData;
  searchEnabled?: boolean;
  onGetColumnValue?: (row, column) => WorkflowTableElement;
}

export const WorkflowTable: React.FC<WorkflowsTableProps> = ({
  data,
  onGetColumnValue,
  searchEnabled = false,
}) => {
  const [page, setPage] = React.useState(0);
  const [rowsPerPage, setRowsPerPage] = React.useState(5);
  const [searchQuery, setSearchQuery] = React.useState('');
  const [filterColumn, setFilterColumn] = React.useState('name');

  let rows = data.data;
  let columns = data.schema.fields;

  let filteredRows = [];
  //let filteredColumns = columns;

  if (searchQuery.length > 0) {
    filteredRows = data.data.filter((rowItem, index) => {
      // filter rows by name, which is our default filter column.
      // rowItem.meta.name.toLowercase().includes(searchQuery)
      const name = rowItem.name.name as string;
      return name.toLowerCase().includes(searchQuery.toLowerCase());
    });

    rows = filteredRows;
  }

  const handleChangePage = (event: unknown, newPage: number) => {
    setPage(newPage);
  };

  const handleChangeRowsPerPage = (
    event: React.ChangeEvent<HTMLInputElement>
  ) => {
    setRowsPerPage(+event.target.value);
    setPage(0);
  };

  /*
    Returns the value to be inserted at row, column.
    If a callback is passed in, uses the onGetColumnValue to support things like rendering arbitrary react components.
  */
  const getColumnValue = (row, column) => {
    // use callback if passed in
    if (onGetColumnValue) {
      return onGetColumnValue(row, column);
    }

    const value = row[column.name];
    return value;
  };

  return (
    <>
      <TextField
        value={searchQuery}
        onChange={(event) => setSearchQuery(event.target.value)}
        id="outlined-basic"
        label="Search"
        variant="outlined"
      />
      <Paper sx={{ overflow: 'hidden' }}>
        <TableContainer>
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
                      onClick={() => {
                        console.log(
                          'tableColumn clicked colIndex: ',
                          columnIndex
                        );
                        console.log('tableColumn clicked column: ', column);
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
                        return (
                          <TableCell
                            key={`table-col-${columnIndex}`}
                            align={'left'}
                          >
                            {getColumnValue(row, column)}
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
    </>
  );
};

export default WorkflowTable;
