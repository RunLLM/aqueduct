import { faSearch, faX } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
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

import { theme } from '../../styles/theme/theme';
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
  // TODO: Add dropdown to select which column to search the table on.
  const [searchColumn, setSearchColumn] = React.useState('name');

  let rows = data.data;
  const columns = data.schema.fields;

  let filteredRows = [];

  const shouldInclude = (rowItem, searchQuery, searchColumn): boolean => {
    // TODO: Allow users to pass in a function as a prop to support custom search by column.
    // Since table cells can contain complex objects, this implementation is up to the caller.
    // Otherwise, we default to using 'name' as the field to conduct the search on.

    // filter rows by name, which is our default filter column.
    // rowItem.meta.name.toLowercase().includes(searchQuery)
    let shouldInclude = false;
    switch (searchColumn) {
      case 'name': {
        const name = rowItem.name.name as string;
        shouldInclude = name.toLowerCase().includes(searchQuery.toLowerCase());
      }
      // TODO: Create function to make this filtering more generic.
      default: {
        const name = rowItem.name.name as string;
        shouldInclude = name.toLowerCase().includes(searchQuery.toLowerCase());
      }
    }

    return shouldInclude;
  };

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

  if (searchQuery.length > 0) {
    filteredRows = data.data.filter((rowItem, index) => {
      return shouldInclude(rowItem, searchQuery, searchColumn);
    });

    rows = filteredRows;
  }

  return (
    <>
      {searchEnabled && (
        <Box marginBottom="8px">
          <TextField
            placeholder="Search by name ..."
            value={searchQuery}
            onChange={(event) => setSearchQuery(event.target.value)}
            id="outlined-basic"
            variant="outlined"
            fullWidth
            InputProps={{
              startAdornment: (
                <Box marginRight="8px">
                  <FontAwesomeIcon
                    icon={faSearch}
                    color={theme.palette.gray[600]}
                  />
                </Box>
              ),
              endAdornment: (
                <Box
                  marginLeft="8px"
                  color={theme.palette.gray[600]}
                  sx={{
                    '&:hover': {
                      cursor: 'pointer',
                      color: theme.palette.black,
                    },
                  }}
                  onClick={() => {
                    setSearchQuery('');
                  }}
                >
                  <FontAwesomeIcon icon={faX} />
                </Box>
              ),
            }}
          />
        </Box>
      )}

      <Paper sx={{ overflow: 'hidden' }}>
        <TableContainer>
          <Table stickyHeader aria-label="sticky table">
            <TableHead>
              <TableRow>
                {columns.map((column, columnIndex) => {
                  let columnName = column.displayName ? column.displayName : column.name;
                  return (
                    <TableCell
                      padding="none"
                      sx={{
                        borderRight:
                          columnIndex < columns.length - 1
                            ? '1px solid rgba(224, 224, 224, 1);'
                            : 'none',
                      }}
                      key={`table-header-col-${columnIndex}`}
                      align={'left'}
                    >
                      <Box flexDirection="column" padding="8px">
                        <Typography
                          variant="body1"
                          sx={{
                            textTransform: 'capitalize',
                            fontSize: '16px',
                            fontWeight: 400,
                          }}
                        >
                          {columnName}
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
                            padding="none"
                          >
                            <Box padding="8px">
                              {getColumnValue(row, column)}
                            </Box>
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
