import {
  faArrowDown,
  faArrowUp,
  faArrowUpShortWide,
  faSearch,
  faX,
} from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Chip, Menu, MenuItem } from '@mui/material';
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
import React, { useEffect, useState } from 'react';

import { theme } from '../../styles/theme/theme';
import Delayed from '../../utils/Delayed';
import { Button } from '../primitives/Button.styles';

export type PaginatedSearchTableElement =
  | string
  | number
  | boolean
  | Date
  | JSX.Element;

export type PaginatedSearchTableRow = {
  [key: string]: PaginatedSearchTableElement;
};

export type SortColumn = {
  // The name of the column by which to sort.
  name: string;

  // The sequence of keys in the row object to access in order to get the
  // value which should be compared for sort purposes.
  sortAccessPath: (string | number)[];
};

export enum SortType {
  None,
  Ascending,
  Descending,
}

type SortConfig = {
  sortColumn: SortColumn;
  sortType: SortType;
};

export interface PaginatedSearchTableProps {
  data: object[];
  columns: string[];
  searchEnabled?: boolean;
  onGetColumnValue?: (row, column) => PaginatedSearchTableElement;
  onShouldInclude?: (rowItem, searchQuery, searchColumn) => boolean;
  onChangeRowsPerPage?: (rowsPerPage) => void;
  savedRowsPerPage?: number;
  sortColumns?: SortColumn[];
  defaultSortConfig?: SortConfig;
}

export const PaginatedSearchTable: React.FC<PaginatedSearchTableProps> = ({
  data,
  columns = [],
  onGetColumnValue,
  searchEnabled = false,
  onShouldInclude,
  onChangeRowsPerPage,
  savedRowsPerPage,
  sortColumns = [],
  defaultSortConfig,
}) => {
  const [page, setPage] = useState(0);
  const [rowsPerPage, setRowsPerPage] = useState(
    savedRowsPerPage ? savedRowsPerPage : 5
  );
  const [searchQuery, setSearchQuery] = useState('');
  // TODO: Add dropdown to select which column to search the table on.
  // TODO: add setSearchColumn to the array below.
  const [searchColumn] = useState('Name');

  const [sortMenuAnchor, setSortMenuAnchor] = useState<HTMLButtonElement>(null);
  const [sortTypeMenuAnchor, setSortTypeMenuAnchor] =
    useState<HTMLLIElement>(null);
  const [sortConfig, setSortConfig] = useState({
    sortColumn: null,
    sortType: SortType.None,
  });

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

  let rowData = [...data].map((row) => {
    const rowData = {};
    columns.forEach((column) => {
      rowData[column] = getColumnValue(row, column);
    });
    return rowData;
  });

  // Default ordering
  if (defaultSortConfig) {
    rowData = rowData.sort((r1, r2) => {
      const col = defaultSortConfig.sortColumn;
      let v1: PaginatedSearchTableRow | PaginatedSearchTableElement = r1;
      let v2: PaginatedSearchTableRow | PaginatedSearchTableElement = r2;
      for (const path of col.sortAccessPath) {
        v1 = v1[path];
        v2 = v2[path];
      }

      if (defaultSortConfig.sortType === SortType.Ascending) {
        if (v1 > v2) {
          return 1;
        } else if (v1 < v2) {
          return -1;
        } else {
          return 0;
        }
      } else {
        // sortType === SortType.Descending
        if (v1 > v2) {
          return -1;
        } else if (v1 < v2) {
          return 1;
        } else {
          return 0;
        }
      }
    });
  }

  const [rows, setRows] = useState(rowData);
  const [orderedRows, setOrderedRows] = useState(rowData);

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

  useEffect(() => {
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
        case 'Name': {
          sortColumns.forEach((column) => {
            if (column.name === searchColumn) {
              let v = rowItem;
              for (const path of column.sortAccessPath) {
                v = v[path];
              }
              shouldInclude = v
                .toLowerCase()
                .includes(searchQuery.toLowerCase());
            }
          });
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
    if (searchQuery.length > 0) {
      const filteredRows = rows.filter((rowItem) => {
        return shouldInclude(rowItem, searchQuery, searchColumn);
      });
      setRows(filteredRows);
    } else {
      if (sortConfig.sortType !== SortType.None) {
        setRows(orderedRows);
      } else {
        setRows(rowData);
      }
    }
  }, [
    searchQuery,
    rowData,
    orderedRows,
    rows,
    searchColumn,
    sortConfig,
    onShouldInclude,
    sortColumns,
  ]);

  useEffect(() => {
    if (
      !sortConfig.sortColumn ||
      !sortConfig.sortColumn.name ||
      sortConfig.sortType === SortType.None
    ) {
      setRows(rowData);
      return;
    }

    const sortedRows = [...rowData].sort((r1, r2) => {
      const col = sortConfig.sortColumn;
      let v1: PaginatedSearchTableRow | PaginatedSearchTableElement = r1;
      let v2: PaginatedSearchTableRow | PaginatedSearchTableElement = r2;
      for (const path of col.sortAccessPath) {
        v1 = v1[path];
        v2 = v2[path];
      }

      if (sortConfig.sortType === SortType.Ascending) {
        if (v1 > v2) {
          return 1;
        } else if (v1 < v2) {
          return -1;
        } else {
          return 0;
        }
      } else {
        // sortType === SortType.Descending
        if (v1 > v2) {
          return -1;
        } else if (v1 < v2) {
          return 1;
        } else {
          return 0;
        }
      }
    });
    setRows(sortedRows);
    setOrderedRows(sortedRows);
  }, [sortConfig, rowData]);

  // Need to slice (.slice(page * rowsPerPage, page * rowsPerPage + rowsPerPage)) after map because we do the API calls in the GetColumnValue function.
  // As a result, if there are less than rowsPerPage number of rows, less hooks are rendered than expected. Thus, we render all hooks.
  // Can investigate optimization as a future step.
  const displayRows = rows
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
                  fontSize: '16px', // This is needed for consistency.
                }}
              >
                <Box padding="8px">{row[column]}</Box>
              </TableCell>
            );
          })}
        </TableRow>
      );
    })
    .slice(page * rowsPerPage, page * rowsPerPage + rowsPerPage);

  return (
    <Delayed>
      {searchEnabled && (
        <Box mb="8px">
          <Box marginBottom="8px" display="flex">
            <TextField
              placeholder="Search"
              value={searchQuery}
              onChange={(event) => setSearchQuery(event.target.value)}
              id="outlined-basic"
              variant="outlined"
              size="small"
              sx={{ minWidth: '300px' }}
              InputProps={{
                startAdornment:
                  searchQuery === '' ? (
                    <Box marginRight="8px">
                      <FontAwesomeIcon
                        icon={faSearch}
                        color={theme.palette.gray[600]}
                      />
                    </Box>
                  ) : (
                    <Box
                      marginRight="8px"
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

            {sortColumns && sortColumns.length > 0 && (
              <Button
                size="small"
                color="secondary"
                endIcon={<FontAwesomeIcon icon={faArrowUpShortWide} />}
                sx={{ ml: 1 }}
                onClick={(e) => setSortMenuAnchor(e.currentTarget)}
              >
                Sort
              </Button>
            )}

            <Menu
              open={!!sortMenuAnchor}
              anchorEl={sortMenuAnchor}
              onClose={() => setSortMenuAnchor(null)}
            >
              {sortColumns.map((column, idx) => (
                // Note that the menu doesn't seem to like non-numerical values, so we use the idx here.
                <MenuItem
                  key={column.name}
                  value={idx}
                  onClick={(e) => setSortTypeMenuAnchor(e.currentTarget)}
                >
                  {column.name}
                </MenuItem>
              ))}
            </Menu>

            <Menu
              open={!!sortTypeMenuAnchor}
              anchorEl={sortTypeMenuAnchor}
              onClose={() => setSortTypeMenuAnchor(null)}
              anchorOrigin={{
                vertical: 'top',
                horizontal: 'right',
              }}
              transformOrigin={{
                vertical: 'top',
                horizontal: 'left',
              }}
            >
              <MenuItem
                onClick={() => {
                  setSortConfig({
                    sortColumn: sortColumns[sortTypeMenuAnchor.value],
                    sortType: SortType.Ascending,
                  });
                  // Close the menus that are open.
                  setSortTypeMenuAnchor(null);
                  setSortMenuAnchor(null);
                }}
              >
                <Box sx={{ mr: 1 }}>
                  <FontAwesomeIcon icon={faArrowUp} />
                </Box>
                Ascending
              </MenuItem>

              <MenuItem
                onClick={() => {
                  setSortConfig({
                    sortColumn: sortColumns[sortTypeMenuAnchor.value],
                    sortType: SortType.Descending,
                  });
                  // Close the menus that are open.
                  setSortTypeMenuAnchor(null);
                  setSortMenuAnchor(null);
                }}
              >
                <Box sx={{ mr: 1 }}>
                  <FontAwesomeIcon icon={faArrowDown} />
                </Box>
                Descending
              </MenuItem>
            </Menu>
          </Box>

          {sortConfig.sortType !== SortType.None && (
            <Chip
              icon={
                <FontAwesomeIcon
                  icon={
                    sortConfig.sortType === SortType.Ascending
                      ? faArrowUp
                      : faArrowDown
                  }
                />
              }
              label={`Sort: ${sortConfig.sortColumn.name}`}
              onDelete={() =>
                setSortConfig({
                  sortType: SortType.None,
                  sortColumn: { name: '', sortAccessPath: [] },
                })
              }
            />
          )}
        </Box>
      )}

      <Paper sx={{ overflow: 'hidden' }}>
        <TableContainer>
          <Table stickyHeader aria-label="sticky table">
            <TableHead>
              <TableRow>
                {columns.map((columnName, columnIndex) => {
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
                      <Box
                        flexDirection="column"
                        padding="8px"
                        sx={{ backgroundColor: theme.palette.gray['50'] }}
                      >
                        <Typography
                          variant="body1"
                          sx={{
                            textTransform: 'capitalize',
                            fontSize: '14px',
                            fontWeight: 800,
                            color: theme.palette.gray['900'],
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
            <TableBody>{displayRows}</TableBody>
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
    </Delayed>
  );
};

export default PaginatedSearchTable;
