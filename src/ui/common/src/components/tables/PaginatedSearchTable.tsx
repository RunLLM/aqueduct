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

export type PaginatedSearchTableElement =
  | string
  | number
  | boolean
  | JSX.Element;

export type PaginatedSearchTableRow = {
  [key: string]: PaginatedSearchTableElement;
};

export interface PaginatedSearchTableData {
  schema?: DataSchema;
  data: PaginatedSearchTableRow[];
}

export interface PaginatedSearchTableProps {
  data: PaginatedSearchTableData;
  searchEnabled?: boolean;
  onGetColumnValue?: (row, column) => PaginatedSearchTableElement;
  onShouldInclude?: (rowItem, searchQuery, searchColumn) => boolean;
  onChangeRowsPerPage?: (rowsPerPage) => void;
  savedRowsPerPage?: number;
}

export const PaginatedSearchTable: React.FC<PaginatedSearchTableProps> = ({
  data,
  onGetColumnValue,
  searchEnabled = false,
  onShouldInclude,
  onChangeRowsPerPage,
  savedRowsPerPage
}) => {
  const [page, setPage] = React.useState(0);
  const [rowsPerPage, setRowsPerPage] = React.useState(savedRowsPerPage ? savedRowsPerPage : 5);
  const [searchQuery, setSearchQuery] = React.useState('');
  // TODO: Add dropdown to select which column to search the table on.
  // TODO: add setSearchColumn to the array below.
  const [searchColumn] = React.useState('name');

  let rows = data.data;
  const columns = data.schema.fields;

  let filteredRows = [];

  /**
   * Function used to test whether a row should be included in search results.
   * This function allows for searching over arbitrary columns since it takes in a searchQuery and a column to search through.
   * To allow for more control at the caller's level, this function calls onShouldInclude if there is a function passed in.
   * This allows us to do things like search through a row item that is an array (assuming the callback implements the search for the column) for example.
   * @param rowItem - Row item to test whether or not to include in search results.
   * @param searchQuery - Query to search check e.g. rowItem[searchColumn] === searchQuery
   * @param searchColumn - Column inside row item to use for search.
   * @returns - true or false whether the rowItem[searchColumn] is a match for searchQuery
   */
  const shouldInclude = (rowItem, searchQuery, searchColumn): boolean => {
    // Since table cells can contain complex objects, this implementation is up to the caller.
    // Otherwise, we default to using 'name' (two fields currently in use by the Workflows list table and Data list tables) as the field to conduct the search on.
    if (onShouldInclude) {
      return onShouldInclude(rowItem, searchQuery, searchColumn);
    }

    // filter rows by name, which is our default filter column.
    let shouldInclude = false;
    switch (searchColumn) {
      case 'name': {
        const name = rowItem.name.name as string;
        shouldInclude = name.toLowerCase().includes(searchQuery.toLowerCase());
        break;
      }
      default: {
        // no name column, return true for everything.
        shouldInclude = true;
        break;
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
    if (onChangeRowsPerPage) {
      // Call the callback here and set the appropriate stuff in localstorage.
      onChangeRowsPerPage(+event.target.value);
    }
    setRowsPerPage(+event.target.value);
    setPage(0);
  };

  /*
    Returns the value to be inserted at row, column.
    If a callback is passed in, uses the onGetColumnValue to support things like rendering arbitrary react components.
  */
  const getColumnValue = (row, column) => {
    if (onGetColumnValue) {
      return onGetColumnValue(row, column);
    }

    const value = row[column.name];
    return value;
  };

  if (searchQuery.length > 0) {
    filteredRows = data.data.filter((rowItem) => {
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
                  const columnName = column.displayName
                    ? column.displayName
                    : column.name;
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
                            sx={{
                              borderRight:
                                columnIndex < columns.length - 1
                                  ? '1px solid rgba(224, 224, 224, 1);'
                                  : 'none',
                            }}
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

export default PaginatedSearchTable;
